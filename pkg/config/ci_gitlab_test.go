package config

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/xanzy/go-gitlab"
	"gitlab.com/sparetimecoders/build-tools/pkg/templating"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestIdentify_Gitlab(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("GITLAB_CI", "1")
	_ = os.Setenv("CI_COMMIT_SHA", "abc123")
	_ = os.Setenv("CI_PROJECT_NAME", "reponame")
	_ = os.Setenv("CI_COMMIT_REF_NAME", "feature/first test")

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
	assert.NoError(t, err)
	result, err := cfg.CurrentCI()
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "abc123", result.Commit())
	assert.Equal(t, "reponame", result.BuildName())
	assert.Equal(t, "feature/first test", result.Branch())
	assert.Equal(t, "feature_first_test", result.BranchReplaceSlash())
	assert.Equal(t, "", out.String())
}

func TestName_Gitlab(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("GITLAB_CI", "1")
	_ = os.Setenv("CI_COMMIT_SHA", "abc123")
	_ = os.Setenv("CI_PROJECT_NAME", "reponame")
	_ = os.Setenv("CI_COMMIT_REF_NAME", "feature/first test")

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
	assert.NoError(t, err)
	result, err := cfg.CurrentCI()
	assert.NoError(t, err)
	assert.Equal(t, "Gitlab", result.Name())
}

func TestBuildName_Fallback_Gitlab(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI", "gitlab")

	dir, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(dir) }()
	oldPwd, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(oldPwd) }()

	out := &bytes.Buffer{}
	cfg, err := Load(dir, out)
	assert.NoError(t, err)
	result, err := cfg.CurrentCI()
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, filepath.Base(dir), result.BuildName())
	assert.Equal(t, "", out.String())
}

func TestBranch_VCS_Fallback_Gitlab(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI", "gitlab")

	dir, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(dir) }()

	InitRepoWithCommit(dir)

	out := &bytes.Buffer{}
	cfg, err := Load(dir, out)
	assert.NoError(t, err)
	result, err := cfg.CurrentCI()
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "master", result.Branch())
	assert.Equal(t, "", out.String())
}

func TestCommit_VCS_Fallback_Gitlab(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI", "gitlab")

	dir, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(dir) }()

	hash, _ := InitRepoWithCommit(dir)

	out := &bytes.Buffer{}
	cfg, err := Load(dir, out)
	assert.NoError(t, err)
	result, err := cfg.CurrentCI()
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, hash.String(), result.Commit())
	assert.Equal(t, "", out.String())
}

func TestValidate_Gitlab_User_Not_Exist(t *testing.T) {
	ci := &GitlabCI{usersService: &mockUsersService{err: errors.New("unauthorized")}}

	err := ci.Validate("Project")

	assert.EqualError(t, err, "unauthorized")
}

func TestValidate_Gitlab_Organisation_Not_Exist(t *testing.T) {
	ci := &GitlabCI{
		usersService:  &mockUsersService{},
		groupsService: &mockGroups{err: errors.New("not found")},
	}

	err := ci.Validate("Project")

	assert.EqualError(t, err, "not found")
}

func TestValidate_Gitlab_Error_Getting_Pipeline(t *testing.T) {
	ci := &GitlabCI{
		usersService:  &mockUsersService{},
		groupsService: &mockGroups{},
		projectsService: &mockProjects{
			getErr:   errors.New("error"),
			response: &gitlab.Response{Response: &http.Response{StatusCode: 403}},
		},
	}

	err := ci.Validate("Project")

	assert.EqualError(t, err, "error")
}

func TestValidate_Gitlab_Pipeline_Already_Exists(t *testing.T) {
	ci := &GitlabCI{
		Group:         "org",
		usersService:  &mockUsersService{},
		groupsService: &mockGroups{},
		projectsService: &mockProjects{
			project: &gitlab.Project{},
		},
	}

	err := ci.Validate("Project")

	assert.EqualError(t, err, "project named 'org/Project' already exists at Gitlab")
}

func TestValidate_Gitlab_Ok(t *testing.T) {
	ci := &GitlabCI{
		usersService:  &mockUsersService{},
		groupsService: &mockGroups{},
		projectsService: &mockProjects{
			getErr:   errors.New("not found"),
			response: &gitlab.Response{Response: &http.Response{StatusCode: 404}},
		},
	}

	err := ci.Validate("Project")

	assert.NoError(t, err)
}

func TestScaffold_Error(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(dir) }()

	name := filepath.Join(dir, "dummy")
	_ = ioutil.WriteFile(name, []byte("abc"), 0666)

	ci := &GitlabCI{}

	_, err := ci.Scaffold(name, templating.TemplateData{})
	assert.EqualError(t, err, fmt.Sprintf("mkdir %s: not a directory", name))
}

func TestScaffold(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(dir) }()

	ci := &GitlabCI{}

	_, err := ci.Scaffold(dir, templating.TemplateData{ProjectName: "Project"})
	assert.NoError(t, err)

	buff, err := ioutil.ReadFile(filepath.Join(dir, ".gitlab-ci.yml"))
	assert.NoError(t, err)
	assert.Equal(t, expectedGitlabCiYml, string(buff))
}

func TestBadges_Error(t *testing.T) {
	ci := &GitlabCI{badgesService: &mockBadges{err: errors.New("badge error")}}

	_, err := ci.Badges("project")
	assert.EqualError(t, err, "badge error")
}

func TestBadges(t *testing.T) {
	ci := &GitlabCI{
		badgesService: &mockBadges{
			badges: []*gitlab.ProjectBadge{
				{ImageURL: "build.svg", RenderedLinkURL: "https://buildlink", RenderedImageURL: "https://buildimg"},
				{ImageURL: "coverage.svg", RenderedLinkURL: "https://coverlink", RenderedImageURL: "https://coverimg"},
				{ImageURL: "other.svg", RenderedLinkURL: "https://otherlink", RenderedImageURL: "https://otherimg"},
			},
		},
	}

	badges, err := ci.Badges("project")
	assert.NoError(t, err)
	expected := []templating.Badge{
		{Title: "Build status", ImageUrl: "https://buildimg", LinkUrl: "https://buildlink"},
		{Title: "Coverage report", ImageUrl: "https://coverimg", LinkUrl: "https://coverlink"},
		{ImageUrl: "https://otherimg", LinkUrl: "https://otherlink"},
	}
	assert.Equal(t, expected, badges)
}

type mockUsersService struct {
	err error
}

func (m mockUsersService) CurrentUser(options ...gitlab.OptionFunc) (*gitlab.User, *gitlab.Response, error) {
	return nil, nil, m.err
}

var _ usersService = &mockUsersService{}

type mockBadges struct {
	err    error
	badges []*gitlab.ProjectBadge
}

func (m mockBadges) ListProjectBadges(pid interface{}, opt *gitlab.ListProjectBadgesOptions, options ...gitlab.OptionFunc) ([]*gitlab.ProjectBadge, *gitlab.Response, error) {
	return m.badges, nil, m.err
}

var _ badgesService = &mockBadges{}

var expectedGitlabCiYml = `stages:
  - build
  - deploy-staging
  - deploy-prod

variables:
  DOCKER_HOST: tcp://docker:2375/

image: registry.gitlab.com/sparetimecoders/build-tools:master

build:
  stage: build
  services:
    - docker:dind
  script:
  - build
  - push

deploy-to-staging:
  stage: deploy-staging
  when: on_success
  script:
    - echo Deploy Project to staging.
    - deploy staging
  environment:
    name: staging

deploy-to-prod:
  stage: deploy-prod
  when: on_success
  script:
    - echo Deploy Project to prod.
    - deploy prod
  environment:
    name: prod
  only:
    - master
`
