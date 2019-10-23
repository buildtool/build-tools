package ci

import (
	"errors"
	"fmt"
	"github.com/buildkite/go-buildkite/buildkite"
	"github.com/sparetimecoders/build-tools/pkg"
	"github.com/sparetimecoders/build-tools/pkg/file"
	"github.com/sparetimecoders/build-tools/pkg/templating"
	"path/filepath"
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

type Buildkite struct {
	Organisation        string `yaml:"organisation" env:"BUILDKITE_ORG"`
	Token               string `yaml:"token" env:"BUILDKITE_TOKEN"`
	pipelineService     pipelineService
	userService         userService
	organizationService organizationService
}

var _ CI = &Buildkite{}

func (c *Buildkite) Name() string {
	return "Buildkite"
}

func (c *Buildkite) ValidateConfig() error {
	if len(c.Token) == 0 {
		return errors.New("token for Buildkite not configured")
	}
	return nil
}

func (c *Buildkite) Validate(name string) error {
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

func (c *Buildkite) Scaffold(dir string, data templating.TemplateData) (*string, error) {
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

func (c *Buildkite) Badges(name string) ([]templating.Badge, error) {
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

func (c *Buildkite) Configure() error {
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
