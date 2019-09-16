package ci

import (
	"errors"
	"fmt"
	"github.com/buildkite/go-buildkite/buildkite"
	"github.com/stretchr/testify/assert"
	"gitlab.com/sparetimecoders/build-tools/pkg"
	"gitlab.com/sparetimecoders/build-tools/pkg/templating"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestName_Buildkite(t *testing.T) {
	ci := &BuildkiteCI{}
	assert.Equal(t, "Buildkite", ci.Name())
}

func TestValidate_Buildkite_User_Not_Exist(t *testing.T) {
	ci := &BuildkiteCI{userService: &mockUserService{err: errors.New("unauthorized")}}

	err := ci.Validate("Project")

	assert.EqualError(t, err, "unauthorized")
}

func TestValidate_Buildkite_Organisation_Not_Exist(t *testing.T) {
	ci := &BuildkiteCI{
		userService:         &mockUserService{},
		organizationService: &mockOrganizationService{err: errors.New("not found")},
	}

	err := ci.Validate("Project")

	assert.EqualError(t, err, "not found")
}

func TestValidate_Buildkite_Error_Getting_Pipeline(t *testing.T) {
	ci := &BuildkiteCI{
		userService:         &mockUserService{},
		organizationService: &mockOrganizationService{},
		pipelineService: &mockPipelineService{
			getErr:   errors.New("error"),
			response: &buildkite.Response{Response: &http.Response{StatusCode: 403}},
		},
	}

	err := ci.Validate("Project")

	assert.EqualError(t, err, "error")
}

func TestValidate_Buildkite_Pipeline_Already_Exists(t *testing.T) {
	ci := &BuildkiteCI{
		Organisation:        "org",
		userService:         &mockUserService{},
		organizationService: &mockOrganizationService{},
		pipelineService: &mockPipelineService{
			pipeline: pipeline("", "", ""),
		},
	}

	err := ci.Validate("Project")

	assert.EqualError(t, err, "pipeline named 'org/Project' already exists at Buildkite")
}

func TestValidate_Buildkite_Ok(t *testing.T) {
	ci := &BuildkiteCI{
		userService:         &mockUserService{},
		organizationService: &mockOrganizationService{},
		pipelineService: &mockPipelineService{
			getErr:   errors.New("not found"),
			response: &buildkite.Response{Response: &http.Response{StatusCode: 404}},
		},
	}

	err := ci.Validate("Project")

	assert.NoError(t, err)
}

func TestScaffold_Buildkite_WriteError_PipelineYml(t *testing.T) {
	ci := &BuildkiteCI{}

	dir, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(dir) }()

	name := filepath.Join(dir, ".buildkite")
	_ = ioutil.WriteFile(name, []byte("abc"), 0666)

	_, err := ci.Scaffold(dir, templating.TemplateData{})

	assert.EqualError(t, err, fmt.Sprintf("mkdir %s: not a directory", name))
}

func TestScaffold_Buildkite_WriteError_DockerIgnore(t *testing.T) {
	ci := &BuildkiteCI{}

	dir, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(dir) }()

	name := filepath.Join(dir, ".dockerignore")
	_ = os.MkdirAll(name, 0777)

	_, err := ci.Scaffold(dir, templating.TemplateData{})

	assert.EqualError(t, err, fmt.Sprintf("open %s: is a directory", name))
}

func TestScaffold_Buildkite_CreateError(t *testing.T) {
	ci := &BuildkiteCI{pipelineService: &mockPipelineService{createErr: errors.New("create error")}}

	dir, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(dir) }()

	_, err := ci.Scaffold(dir, templating.TemplateData{})

	assert.EqualError(t, err, "create error")
}

func TestScaffold_Buildkite_Create_Github(t *testing.T) {
	service := &mockPipelineService{pipeline: pipeline("https://hookUrl", "", "")}
	ci := &BuildkiteCI{pipelineService: service}

	dir, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(dir) }()

	data := templating.TemplateData{
		ProjectName:    "Project",
		RepositoryHost: "github.com",
		RepositoryUrl:  "git@repo/",
	}
	hook, err := ci.Scaffold(dir, data)

	assert.NoError(t, err)
	expected := &buildkite.CreatePipeline{
		Name:       "Project",
		Repository: "git@repo/",
		Steps: []buildkite.Step{
			{
				Type:    buildkite.String("script"),
				Name:    buildkite.String("Setup :package:"),
				Command: buildkite.String("buildkite-agent pipeline upload"),
			},
		},
		ProviderSettings: &buildkite.GitHubSettings{
			TriggerMode:                pkg.String("code"),
			BuildPullRequests:          pkg.Bool(true),
			BuildPullRequestForks:      pkg.Bool(false),
			BuildTags:                  pkg.Bool(false),
			PublishCommitStatus:        pkg.Bool(true),
			PublishCommitStatusPerStep: pkg.Bool(true),
		},
		SkipQueuedBranchBuilds:    true,
		CancelRunningBranchBuilds: true,
	}
	assert.Equal(t, expected, service.create)
	assert.Equal(t, "https://hookUrl", *hook)
}

func TestScaffold_Buildkite_Create_Other(t *testing.T) {
	service := &mockPipelineService{pipeline: pipeline("https://hookUrl", "", "")}
	ci := &BuildkiteCI{pipelineService: service}

	dir, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(dir) }()

	data := templating.TemplateData{
		ProjectName:    "Project",
		RepositoryHost: "gitlab.com",
		RepositoryUrl:  "git@repo/",
	}
	hook, err := ci.Scaffold(dir, data)

	assert.NoError(t, err)
	expected := &buildkite.CreatePipeline{
		Name:       "Project",
		Repository: "git@repo/",
		Steps: []buildkite.Step{
			{
				Type:    buildkite.String("script"),
				Name:    buildkite.String("Setup :package:"),
				Command: buildkite.String("buildkite-agent pipeline upload"),
			},
		},
		ProviderSettings:          nil,
		SkipQueuedBranchBuilds:    true,
		CancelRunningBranchBuilds: true,
	}
	assert.Equal(t, expected, service.create)
	assert.Equal(t, "https://hookUrl", *hook)
}

func TestBadges_Buildkite_Error(t *testing.T) {
	ci := &BuildkiteCI{pipelineService: &mockPipelineService{getErr: errors.New("get error")}}

	_, err := ci.Badges("Project")

	assert.EqualError(t, err, "get error")
}

func TestBadges_Buildkite(t *testing.T) {
	ci := &BuildkiteCI{pipelineService: &mockPipelineService{pipeline: pipeline("https://hookUrl", "https://img", "https://link")}}

	badges, err := ci.Badges("Project")

	expected := []templating.Badge{
		{
			Title:    "Build status",
			ImageUrl: "https://img",
			LinkUrl:  "https://link",
		},
	}
	assert.NoError(t, err)
	assert.Equal(t, expected, badges)
}

func pipeline(hookUrl, badgeUrl, webUrl string) *buildkite.Pipeline {
	return &buildkite.Pipeline{
		BadgeURL: pkg.String(badgeUrl),
		WebURL:   pkg.String(webUrl),
		Provider: &buildkite.Provider{WebhookURL: pkg.String(hookUrl)},
	}
}

type mockUserService struct {
	err error
}

func (m mockUserService) Get() (*buildkite.User, *buildkite.Response, error) {
	return nil, nil, m.err
}

var _ userService = &mockUserService{}

type mockOrganizationService struct {
	err error
}

func (m mockOrganizationService) Get(slug string) (*buildkite.Organization, *buildkite.Response, error) {
	return nil, nil, m.err
}

var _ organizationService = &mockOrganizationService{}

type mockPipelineService struct {
	createErr error
	getErr    error
	pipeline  *buildkite.Pipeline
	create    *buildkite.CreatePipeline
	response  *buildkite.Response
}

func (m *mockPipelineService) Create(org string, p *buildkite.CreatePipeline) (*buildkite.Pipeline, *buildkite.Response, error) {
	m.create = p
	return m.pipeline, nil, m.createErr
}

func (m *mockPipelineService) Get(org string, slug string) (*buildkite.Pipeline, *buildkite.Response, error) {
	return m.pipeline, m.response, m.getErr
}

var _ pipelineService = &mockPipelineService{}
