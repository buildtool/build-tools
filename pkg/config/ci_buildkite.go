package config

import (
	"fmt"
	"github.com/buildkite/go-buildkite/buildkite"
	"gitlab.com/sparetimecoders/build-tools/pkg"
	"gitlab.com/sparetimecoders/build-tools/pkg/file"
	"gitlab.com/sparetimecoders/build-tools/pkg/templating"
	"path/filepath"
	"strings"
)

type pipelineService interface {
	Create(org string, p *buildkite.CreatePipeline) (*buildkite.Pipeline, *buildkite.Response, error)
	Get(org string, slug string) (*buildkite.Pipeline, *buildkite.Response, error)
}

type userService interface {
	Get() (*buildkite.User, *buildkite.Response, error)
}

type organizationService interface {
	Get(slug string) (*buildkite.Organization, *buildkite.Response, error)
}
type BuildkiteCI struct {
	*ci
	CICommit            string `env:"BUILDKITE_COMMIT"`
	CIBuildName         string `env:"BUILDKITE_PIPELINE_SLUG"`
	CIBranchName        string `env:"BUILDKITE_BRANCH_NAME"`
	Organisation        string `yaml:"organisation" env:"BUILDKITE_ORG"`
	Token               string `yaml:"token" env:"BUILDKITE_TOKEN"`
	pipelineService     pipelineService
	userService         userService
	organizationService organizationService
}

var _ CI = &BuildkiteCI{}

func (c *BuildkiteCI) Name() string {
	return "Buildkite"
}

func (c *BuildkiteCI) BranchReplaceSlash() string {
	return strings.ReplaceAll(strings.ReplaceAll(c.Branch(), "/", "_"), " ", "_")
}

func (c *BuildkiteCI) BuildName() string {
	if c.CIBuildName != "" {
		return c.CIBuildName
	}
	return c.ci.BuildName()
}

func (c *BuildkiteCI) Branch() string {
	if len(c.CIBranchName) == 0 {
		return c.VCS.Branch()
	}
	return c.CIBranchName
}

func (c *BuildkiteCI) Commit() string {
	if len(c.CICommit) == 0 {
		return c.VCS.Commit()
	}
	return c.CICommit
}

func (c *BuildkiteCI) Validate(name string) error {
	if _, _, err := c.userService.Get(); err != nil {
		return err
	}
	if _, _, err := c.organizationService.Get(c.Organisation); err != nil {
		return err
	}
	pipeline, response, err := c.pipelineService.Get(c.Organisation, name)
	if err != nil {
		if response == nil || response.StatusCode != 404 {
			return err
		}
	}
	if pipeline != nil {
		return fmt.Errorf("pipeline named '%s/%s' already exists at Buildkite", c.Organisation, name)
	}

	return nil
}

func (c *BuildkiteCI) Scaffold(dir string, data templating.TemplateData) (*string, error) {
	if err := file.Write(filepath.Join(dir, ".buildkite"), "pipeline.yml", pipelineYml); err != nil {
		return nil, err
	}
	if err := file.Append(filepath.Join(dir, ".dockerignore"), ".buildkite"); err != nil {
		return nil, err
	}
	provider := getProviderFromRepositoryHost(data.RepositoryHost)
	pipeline, _, err := c.pipelineService.Create(c.Organisation, &buildkite.CreatePipeline{
		Name:       data.ProjectName,
		Repository: data.RepositoryUrl,
		Steps: []buildkite.Step{
			{
				Type:    buildkite.String("script"),
				Name:    buildkite.String("Setup :package:"),
				Command: buildkite.String("buildkite-agent pipeline upload"),
			},
		},
		ProviderSettings:          provider,
		SkipQueuedBranchBuilds:    true,
		CancelRunningBranchBuilds: true,
	})
	if err != nil {
		return nil, err
	}

	var hookUrl *string
	if pipeline != nil && pipeline.Provider != nil {
		hookUrl = pipeline.Provider.WebhookURL
	}
	return hookUrl, err
}

func (c *BuildkiteCI) Badges(name string) ([]templating.Badge, error) {
	pipeline, _, err := c.pipelineService.Get(c.Organisation, name)
	if err != nil {
		return nil, err
	}
	badges := []templating.Badge{
		{
			Title:    "Build status",
			ImageUrl: *pipeline.BadgeURL,
			LinkUrl:  *pipeline.WebURL,
		},
	}
	return badges, nil
}

func (c *BuildkiteCI) configure() error {
	config, err := buildkite.NewTokenConfig(c.Token, false)
	if err != nil {
		return err
	}
	client := buildkite.NewClient(config.Client())

	c.pipelineService = client.Pipelines
	c.userService = client.User
	c.organizationService = client.Organizations

	return nil
}

func (c *BuildkiteCI) configured() bool {
	return c.CIBuildName != ""
}

func getProviderFromRepositoryHost(host string) buildkite.ProviderSettings {
	if host == "github.com" {
		return &buildkite.GitHubSettings{
			TriggerMode:                pkg.String("code"),
			BuildPullRequests:          pkg.Bool(true),
			BuildPullRequestForks:      pkg.Bool(false),
			BuildTags:                  pkg.Bool(false),
			PublishCommitStatus:        pkg.Bool(true),
			PublishCommitStatusPerStep: pkg.Bool(true),
		}
	}
	return nil
}

var pipelineYml = `
steps:
  - command: |-
      build
      push
    label: build

  - wait

  - command: |-
      deploy staging
    label: ":rocket: Deploy to STAGING"
    branches: "master"

  - block: ":rocket: Release PROD"
    branches: "master"

  - command: |-
      deploy prod
    label: ":rocket: Deploy to PROD"
    branches: "master"
`
