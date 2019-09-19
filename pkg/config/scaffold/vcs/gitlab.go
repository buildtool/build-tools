package vcs

import (
	"errors"
	"fmt"
	"github.com/xanzy/go-gitlab"
	"path/filepath"
)

type projectsService interface {
	GetProject(pid interface{}, opt *gitlab.GetProjectOptions, options ...gitlab.OptionFunc) (*gitlab.Project, *gitlab.Response, error)
	CreateProject(opt *gitlab.CreateProjectOptions, options ...gitlab.OptionFunc) (*gitlab.Project, *gitlab.Response, error)
	AddProjectHook(pid interface{}, opt *gitlab.AddProjectHookOptions, options ...gitlab.OptionFunc) (*gitlab.ProjectHook, *gitlab.Response, error)
}

type groupsService interface {
	GetGroup(gid interface{}, options ...gitlab.OptionFunc) (*gitlab.Group, *gitlab.Response, error)
}

type Gitlab struct {
	Git
	Group           string `yaml:"group" env:"GITLAB_GROUP"`
	Token           string `yaml:"token" env:"GITLAB_TOKEN"`
	Visibility      string `yaml:"visibility"`
	projectsService projectsService
	groupsService   groupsService
}

func (v *Gitlab) Name() string {
	return "Gitlab"
}

func (v *Gitlab) ValidateConfig() error {
	if len(v.Group) == 0 {
		return errors.New("gitlab group must be set")
	}
	return nil
}

func (v *Gitlab) Scaffold(name string) (*RepositoryInfo, error) {
	group, _, err := v.groupsService.GetGroup(v.Group)
	if err != nil {
		return nil, err
	}

	visibility := gitlab.VisibilityValue(v.Visibility)
	project, _, err := v.projectsService.CreateProject(&gitlab.CreateProjectOptions{
		Name:                             gitlab.String(name),
		NamespaceID:                      gitlab.Int(group.ID),
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
	})
	if err != nil {
		return nil, err
	}
	return &RepositoryInfo{
		SSHURL:  project.SSHURLToRepo,
		HTTPURL: project.HTTPURLToRepo,
	}, nil
}

func (v *Gitlab) Webhook(name, url string) error {
	path := filepath.Join(v.Group, name)
	_, _, err := v.projectsService.AddProjectHook(path, &gitlab.AddProjectHookOptions{
		URL:                 gitlab.String(url),
		PushEvents:          gitlab.Bool(true),
		MergeRequestsEvents: gitlab.Bool(true),
		TagPushEvents:       gitlab.Bool(true),
	})
	return err
}

func (v *Gitlab) Validate(name string) error {
	_, _, err := v.groupsService.GetGroup(v.Group)
	if err != nil {
		return err
	}
	path := filepath.Join(v.Group, name)
	project, response, err := v.projectsService.GetProject(path, nil)
	if err != nil {
		if response == nil || response.StatusCode != 404 {
			return err
		}
	}
	if project != nil {
		return fmt.Errorf("project named '%s/%s' already exists at Gitlab", v.Group, name)
	}
	return nil
}

func (v *Gitlab) Configure() {
	client := gitlab.NewClient(nil, v.Token)
	v.projectsService = client.Projects
	v.groupsService = client.Groups
}

var _ VCS = &Gitlab{}
