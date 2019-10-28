package vcs

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/xanzy/go-gitlab"
	"net/http"
	"testing"
)

func TestGitlab_Name(t *testing.T) {
	vcs := &Gitlab{}
	assert.Equal(t, "Gitlab", vcs.Name())
}

func TestGitlab_Configure(t *testing.T) {
	vcs := &Gitlab{}

	vcs.Configure()
	assert.NotNil(t, vcs.projectsService)
	assert.NotNil(t, vcs.groupsService)
}

func TestGitlab_ValidateConfig_Ok(t *testing.T) {
	vcs := &Gitlab{Group: "group"}

	err := vcs.ValidateConfig()

	assert.Nil(t, err)
}
func TestGitlab_ValidateConfig_Return_Error_If_Group_Not_Set(t *testing.T) {
	vcs := &Gitlab{}

	err := vcs.ValidateConfig()

	assert.EqualError(t, err, "gitlab group must be set")
}

func TestGitlab_Validate_Return_Error_If_Group_Not_Found(t *testing.T) {
	groups := &mockGroups{err: errors.New("group not found")}
	vcs := &Gitlab{Group: "group/sub", groupsService: groups}

	err := vcs.Validate("project")

	assert.Equal(t, "group/sub", groups.gid)
	assert.EqualError(t, err, "group not found")
}

func TestGitlab_Validate_Unexpected_Error_From_Project(t *testing.T) {
	vcs := &Gitlab{
		Group:         "group/sub",
		groupsService: &mockGroups{},
		projectsService: &mockProjects{
			getErr: errors.New("unexpected"),
		},
	}

	err := vcs.Validate("project")

	assert.EqualError(t, err, "unexpected")
}

func TestGitlab_Validate_Project_Exists(t *testing.T) {
	projects := &mockProjects{
		project: &gitlab.Project{},
	}
	vcs := &Gitlab{
		Group:           "group/sub",
		groupsService:   &mockGroups{},
		projectsService: projects,
	}

	err := vcs.Validate("project")

	assert.Equal(t, "group/sub/project", projects.pid)
	assert.EqualError(t, err, "project named 'group/sub/project' already exists at Gitlab")
}

func TestGitlab_Validate_Ok(t *testing.T) {
	vcs := &Gitlab{
		Group:         "group/sub",
		groupsService: &mockGroups{},
		projectsService: &mockProjects{
			response: &gitlab.Response{
				Response: &http.Response{StatusCode: 404},
			},
			getErr: errors.New("404 Project Not Found"),
		},
	}

	err := vcs.Validate("project")

	assert.NoError(t, err)
}

func TestGitlab_Scaffold_Return_Error_If_Group_Not_Found(t *testing.T) {
	groups := &mockGroups{
		err: errors.New("group not found"),
	}
	vcs := &Gitlab{
		Group:         "group/sub",
		groupsService: groups,
	}

	_, err := vcs.Scaffold("project")

	assert.Equal(t, "group/sub", groups.gid)
	assert.EqualError(t, err, "group not found")
}

func TestGitlab_Scaffold_Return_Error_If_Create_Error(t *testing.T) {
	projects := &mockProjects{
		createErr: errors.New("create error"),
	}
	vcs := &Gitlab{
		Group:           "group/sub",
		Visibility:      "private",
		groupsService:   &mockGroups{group: &gitlab.Group{ID: 123}},
		projectsService: projects,
	}

	_, err := vcs.Scaffold("project")

	visibility := gitlab.VisibilityValue("private")
	expectedOpts := &gitlab.CreateProjectOptions{
		Name:                             gitlab.String("project"),
		NamespaceID:                      gitlab.Int(123),
		IssuesEnabled:                    gitlab.Bool(true),
		MergeRequestsEnabled:             gitlab.Bool(true),
		JobsEnabled:                      gitlab.Bool(true),
		WikiEnabled:                      gitlab.Bool(true),
		SnippetsEnabled:                  gitlab.Bool(true),
		ResolveOutdatedDiffDiscussions:   gitlab.Bool(true),
		ContainerRegistryEnabled:         gitlab.Bool(true),
		SharedRunnersEnabled:             gitlab.Bool(true),
		Visibility:                       &visibility,
		PublicBuilds:                     gitlab.Bool(false),
		OnlyAllowMergeIfPipelineSucceeds: gitlab.Bool(true),
		OnlyAllowMergeIfAllDiscussionsAreResolved: gitlab.Bool(true),
		PrintingMergeRequestLinkEnabled:           gitlab.Bool(true),
		InitializeWithReadme:                      gitlab.Bool(true),
	}
	assert.Equal(t, expectedOpts, projects.createOpts)
	assert.EqualError(t, err, "create error")
}

func TestGitlab_Scaffold_Return_URLs_If_Create_Ok(t *testing.T) {
	vcs := &Gitlab{
		Group:         "group/sub",
		groupsService: &mockGroups{group: &gitlab.Group{ID: 123}},
		projectsService: &mockProjects{
			project: &gitlab.Project{
				SSHURLToRepo:  "git@gitlab.com:group/sub/project.git",
				HTTPURLToRepo: "https://gitlab.com/group/sub/project.git",
			},
		},
	}

	info, err := vcs.Scaffold("project")

	assert.NoError(t, err)
	assert.Equal(t, "git@gitlab.com:group/sub/project.git", info.SSHURL)
	assert.Equal(t, "https://gitlab.com/group/sub/project.git", info.HTTPURL)
}

func TestGitlab_Webhook_Add_Error(t *testing.T) {
	projects := &mockProjects{
		hookErr: errors.New("hook error"),
	}
	vcs := &Gitlab{
		Group:           "group/sub",
		groupsService:   &mockGroups{group: &gitlab.Group{ID: 123}},
		projectsService: projects,
	}

	err := vcs.Webhook("project", "https://example.org/hook")

	expectedOpts := &gitlab.AddProjectHookOptions{
		URL:                 gitlab.String("https://example.org/hook"),
		PushEvents:          gitlab.Bool(true),
		MergeRequestsEvents: gitlab.Bool(true),
		TagPushEvents:       gitlab.Bool(true),
	}
	assert.Equal(t, expectedOpts, projects.hookOpts)
	assert.EqualError(t, err, "hook error")
}

type mockProjects struct {
	response   *gitlab.Response
	getErr     error
	createErr  error
	hookErr    error
	pid        interface{}
	createOpts *gitlab.CreateProjectOptions
	hookOpts   *gitlab.AddProjectHookOptions
	project    *gitlab.Project
}

func (m *mockProjects) GetProject(pid interface{}, opt *gitlab.GetProjectOptions, options ...gitlab.OptionFunc) (*gitlab.Project, *gitlab.Response, error) {
	m.pid = pid
	return m.project, m.response, m.getErr
}

func (m *mockProjects) CreateProject(opt *gitlab.CreateProjectOptions, options ...gitlab.OptionFunc) (*gitlab.Project, *gitlab.Response, error) {
	m.createOpts = opt
	return m.project, m.response, m.createErr
}

func (m *mockProjects) AddProjectHook(pid interface{}, opt *gitlab.AddProjectHookOptions, options ...gitlab.OptionFunc) (*gitlab.ProjectHook, *gitlab.Response, error) {
	m.pid = pid
	m.hookOpts = opt
	return nil, m.response, m.hookErr
}

var _ projectsService = &mockProjects{}

type mockGroups struct {
	err   error
	gid   interface{}
	group *gitlab.Group
}

func (m *mockGroups) GetGroup(gid interface{}, options ...gitlab.OptionFunc) (*gitlab.Group, *gitlab.Response, error) {
	m.gid = gid
	return m.group, nil, m.err
}

var _ groupsService = &mockGroups{}
