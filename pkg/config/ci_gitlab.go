package config

import (
	"gitlab.com/sparetimecoders/build-tools/pkg/file"
	"gitlab.com/sparetimecoders/build-tools/pkg/templating"
	"strings"
)

type GitlabCI struct {
	*ci
	CICommit     string `env:"CI_COMMIT_SHA"`
	CIBuildName  string `env:"CI_PROJECT_NAME"`
	CIBranchName string `env:"CI_COMMIT_REF_NAME"`
}

var _ CI = &GitlabCI{}

func (c GitlabCI) Name() string {
	return "Gitlab"
}

func (c GitlabCI) BranchReplaceSlash() string {
	return strings.ReplaceAll(strings.ReplaceAll(c.Branch(), "/", "_"), " ", "_")
}

func (c GitlabCI) BuildName() string {
	if c.CIBuildName != "" {
		return c.CIBuildName
	}
	return c.ci.BuildName()
}

func (c GitlabCI) Branch() string {
	if len(c.CIBranchName) == 0 && c.VCS != nil {
		return c.VCS.Branch()
	}
	return c.CIBranchName
}

func (c GitlabCI) Commit() string {
	if len(c.CICommit) == 0 && c.VCS != nil {
		return c.VCS.Commit()
	}
	return c.CICommit
}

func (c GitlabCI) Scaffold(dir, name, repository string, data templating.TemplateData) (*string, error) {
	if err := file.WriteTemplated(dir, ".gitlab-ci.yml", gitlabCiYml, data); err != nil {
		return nil, err
	}
	return nil, nil
}

func (c GitlabCI) Badges() string {
	return ""
}

func (c GitlabCI) configure() {}

func (c GitlabCI) configured() bool {
	return c.CIBuildName != ""
}

var gitlabCiYml = `
stages:
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
    - echo Deploy {{ .ProjectName }} to staging.
    - deploy staging
  environment:
    name: staging

deploy-to-prod:
  stage: deploy-prod
  when: on_success
  script:
    - echo Deploy {{ .ProjectName }} to prod.
    - deploy prod
  environment:
    name: prod
  only:
    - master
`
