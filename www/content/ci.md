---
title: Continuous Integration
menu: true
weight: 9
---

Commands recognize which CI/CD environment they are executed in based on which environment variables are present.

| Environment variable      | CI/CD                             |
| :-------------------------| :-------------------------------- |
| `BUILDKITE_PIPELINE_SLUG` | [Buildkite](#Buildkite)           |
| `CI_PROJECT_NAME`         | [Gitlab CI](#Gitlab CI)           |
| `RUNNER_WORKSPACE`        | [Github Actions](#Github Actions) |
| `TEAMCITY_PROJECT_NAME`   | [TeamCity](#TeamCity)             |
| `BUILD_REPOSITORY_NAME`   | [Azure Devops](#Azure Devops)     |


## Examples

### Buildkite

[Buildkite] is configured with `.buildkite/pipeline.yml` file in your project.

```yaml
steps:
  - command: |-
      build
      push
    label: build

  - wait

  - command: |-
      ${BUILD_TOOLS_PATH}/deploy staging
    label: Deploy to staging
    branches: "master"

  - block: ":rocket: Release PROD"
    branches: "master"

  - command: |-
      ${BUILD_TOOLS_PATH}/deploy prod
    label: Deploy PROD
    branches: "master"
```

### Gitlab CI

[Gitlab CI] is configured with a `.gitlab-ci.yaml` file in your project.

````yaml
stages:
  - build
  - deploy-staging
  - deploy-prod

variables:
  DOCKER_HOST: tcp://docker:2375/

image: buildtool/build-tools:latest

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
    - echo Deploy to staging.
    - deploy staging
  environment:
    name: staging

deploy-to-prod:
  stage: deploy-prod
  when: on_success
  script:
    - echo Deploy to PROD.
    - deploy prod
  environment:
    name: prod
  only:
    - master
````

### Github Actions

TODO

### TeamCity
[TeamCity] can be configured with a `.teamcity/settings.kts` file in your project. 
    
```kotlin
import jetbrains.buildServer.configs.kotlin.v2018_2.*
import jetbrains.buildServer.configs.kotlin.v2018_2.buildSteps.ScriptBuildStep
import jetbrains.buildServer.configs.kotlin.v2018_2.buildSteps.script
import jetbrains.buildServer.configs.kotlin.v2018_2.triggers.finishBuildTrigger
import jetbrains.buildServer.configs.kotlin.v2018_2.triggers.vcs

version = "2019.1"

project {
    buildType(BuildAndPush)
}

object BuildAndPush : BuildType({
    name = "BuildAndPush"

    steps {
        script {
            name = "build and push"
            scriptContent = """
                build && push
            """.trimIndent()
            dockerImage = "buildtool/buildtools"
            dockerImagePlatform = ScriptBuildStep.ImagePlatform.Linux
            dockerPull = true
            dockerRunParameters = """
                -v /var/run/docker.sock:/var/run/docker.sock
                --rm
            """.trimIndent()
        }
    }

    triggers {
        vcs {}
    }
})

```

### Azure Devops 

[Azure Devops] is configured with a `azure-pipelines.yml` file in your project.

````yaml
resources:
  containers:
  - container: build-tools
    image: buildtool/build-tools:latest

jobs:
- job: build_and_deploy
  pool:
    vmImage: 'Ubuntu 16.04'
  container: build-tools
  steps:
  - script: |
      build
      push
    name: build
    env:
      QUAY_PASSWORD: $(QUAY_PASSWORD)
  - script: deploy staging
    name: deploy_staging
    condition: succeeded()
````

[Buildkite]: https://buildkite.com
[Gitlab CI]: https://docs.gitlab.com/ce/ci
[github actions]: https://github.com/features/actions
[teamcity]: https://www.jetbrains.com/teamcity
[azure devops]: https://azure.microsoft.com/en-us/services/devops/pipelines/