
# build-tools
A set of highly opinionated tools for creating and building components/services into [docker](https://www.docker.com/) images and deploying them to [Kubernetes](https://kubernetes.io/) clusters.

By following the conventions set by the tools, building and deploying applications is made simpler.

The only hard requirement is to provide a [Dockerfile](https://docs.docker.com/engine/reference/builder/) which describes how to build and run your application.

The configuration needed is done by environment variables (most likely for CI/CD) and yaml files (for local use).

# Conventions

A `Dockerfile` must be present in the 
# Commands

## build
## push
## deploy
## Using in CI/CD pipelines

## Local usage
After installing (*TODO link*) the tools, clone the example repository (link), cd into it and execute the `build` command.

    $ build
    Using CI none
    
    no Docker registry found

Since we we haven't setup a `.buildtools.yaml` (*TODO LINK in doc*) file nothing has been configured, and to be able to build a docker image we must specify where we (potentially) want to push it later. In other words, setting the [tags](https://docs.docker.com/engine/reference/commandline/tag/) of the created image.
Luckily we can use environment variables as well, let's try:

    $ DOCKERHUB_REPOSITORY=sparetimecoders build
    Using CI none
    
    Using registry Dockerhub

    Login Succeeded
    Using build variables commit 7c76db502b4a70df5480d6ff438ae10e374b420e on branch master

As we can see, the `build` command identified that we are using Dockerhub, and extracted the commit id and branch information from the local git repository.
After the successful build the image is tagged with the commit id and branch.

    Successfully tagged sparetimecoders/buildtools-examples:7c76db502b4a70df5480d6ff438ae10e374b420e
    Successfully tagged sparetimecoders/buildtools-examples:master
    Successfully tagged sparetimecoders/buildtools-examples:latest



    
Now that we have a docker image, let's publish it to Dockerhub (this of course requires write access to the repository).

    $ DOCKERHUB_REPOSITORY=sparetimecoders DOCKERHUB_PASSWORD=<PASSWORD> DOCKERHUB_USERNAME=<USERNAME> push
    ...
    Pushing tag 'sparetimecoders/buildtools-examples:7c76db502b4a70df5480d6ff438ae10e374b420e'
    ...

    
*TODO Link to more environment variables and stuff*

## Basic setup
- Clone this repository
- Add `BUILD_TOOLS_PATH=<PATH TO THIS REPOSITORY>` to your shell environment, typically in `.bash_profile` or something similar


## Setup script
Script to be used for scaffolding a component/service. Handles repository, build-pipeline and basic files-scaffolding.

    ${BUILD_TOOLS_PATH}/service-setup --stack <stack> <name>

See [Build project structure](#Build-project-structure) below

## Common build scripts
Scripts to be used for building and deploying our services (typically when running in a CI/CD environment). They can also be used for local builds/deploy.

    ${BUILD_TOOLS_PATH}/build
    ${BUILD_TOOLS_PATH}/push

## Build project structure
Configuration and setup is done in `.buildtools` files. Those files must be present in the project folder or upwards in the dicectory structure. This lets you create a common `.buildtools` file to be used for a set of projects.
Example:

    $ pwd
    ~/source/
    $ tree
    .
    ├── customer1
    │   ├── project1
    │   └── project2
    └── customer2
        └── project1
        
Here we can choose to put a `.buildtools` file in the different `customer` directories since they (most likely) have different deployment configuration.

    $ cat customer1/.buildtools
    CI=azure
    VCS=azure
    REGISTRY=dockerhub
    ORGANIZATION=com.organization
    DOCKERHUB_REPOSITORY=repository
    DOCKERHUB_USERNAME=user
    DOCKERHUB_PASSWORD=$(lpass show devenv/docker.com --password)
    AZURE_USER=user
    AZURE_TOKEN=token
    AZURE_ORG=organization
    AZURE_PROJECT=project
    PIPELINE_VARIABLES=(
    ["DOCKERHUB_REPOSITORY"]="repository"
    ["DOCKERHUB_USERNAME"]="user"
    ["DOCKERHUB_PASSWORD"]="secret:$(lpass show devenv/docker.com --password)"
    ["BUILDTOOLS_CONTENT"]="$(lpass show devenv/organization/team/buildtools --notes | base64 -w0)"
    ["KUBECONFIG_CONTENT"]="$(lpass show devenv/organization/team/kubeconfig --notes | base64 -w0)"
    )
    valid_environments=(
        ["local"]="--context docker-for-desktop --namespace default"
        ["staging"]="--context docker-for-desktop --namespace staging"
        ["prod"]="--context docker-for-desktop --namespace prod"
    )

This defines three environments (local,staging,prod) all which are to be deployed to a local Kubernetes cluster but in different namespaces. 

    $ cat customer2/.buildtools
    valid_environments=(
        ["local"]="--context docker-for-desktop --namespace local"
        ["prod"]="--context kube-cluster-prod --namespace prod"
    )

Context and namespaces must be provided/created/configured elsewhere.

### Configuration for service-setup
The `service-setup` script need to know where to scaffold things (i.e. which VCS, CI and container registry to use and how to connect to them).
The environment variables `CI`, `VCS` and `REGISTRY` are used to define this.
Depending on the values of those variables, other variables are required as well. See each section below for the required variables.
Values from the associative array `PIPELINE_VARIABLES` will be created as pipeline-variables in the CI-system. If the value is prepended vith `secret:` the variable will be a secret (if the CI-system supports that).

#### Azure Devops
If `CI` and/or `VCS` is set to `azure` the following variables are required:

| Variable | Value |
|----------|-------|
| AZURE_USER | Azure username |
| AZURE_TOKEN | Azure Personal Access Token |
| AZURE_ORG | The organization name in Azure |
| AZURE_PROJECT | The project name in Azure where repository and build-pipeline will be created |

#### Buildkite
If `CI` is set to `buildkite` the following variables are required:

| Variable | Value |
|----------|-------|
| BUILDKITE_TOKEN | Buildkite Access Token |
| BUILDKITE_ORG | The organization name in Buildkite where the build-pipeline will be created |

#### GitlabCI
If `CI` or `VCS` is set to `gitlab` the following variables are required:

| Variable | Value |
|----------|-------|
| GITLAB_TOKEN | Gitlab Access Token |
| GITLAB_GROUP | The group name in Gitlab where project will be created |

### Configuration using environment variables
The `.buildtools` file can be created by defining an environment variable in the build pipeline named `BUILDTOOLS_CONTENT`. The value should be a base64-encoded string.
On MacOS the value can be created and copied to the clipboard using the following snippet:

    cat - <<EOF | base64 -w0 | pbcopy
    valid_environments=(
        ["local"]="--context docker-for-desktop --namespace local"
        ["prod"]="--context kube-cluster-prod --namespace prod"
    )
    EOF

The Kubernetes contexts to deploy to are configured by setting the `KUBECONFIG_CONTENT` variable. The value should be a base64-encoded string containing the Kubernetes config for the clusters you want to be able to deploy to.
On MacOS the value can be created and copied to the clipboard using the following snippet:

    cat ~/.kube/config.d/prod-cluster.yaml | base64 -w0 | pbcopy

The scripts assume that the project follow the directory layout described below.

## Project structure
The project folder must be a [Git](https://git-scm.com/) repository, with a least one commit.

There must be a `k8s` directory in the root of your project file. This directory contains all the `yaml` files used to describe the Kubernetes deployment tasks needed to run this service.
Environment specific files can be handled in two different ways depending on personal preference. They can either be located in sub-directories, for example `k8s/local` for local setup.

    $ cd projecct
    $ tree
    .
    └── k8s
        ├── deploy.yaml
        ├── local
        │   ├── ingress.yaml
        │   └── setup.sh
        └── prod
            └── ingress.yaml

Or they can be defined using a `-<environment>` suffix, i.e. ingress-prod.yaml

    $ cd projecct
    $ tree
    .
    └── k8s
        ├── deploy.yaml
        ├── ingress-local.yaml
        ├── ingress-prod.yaml
        └── setup-local.sh

## Running in a CI/CD environment
The tools recognize which CI/CD environment they are executed in based on which environment variables are present.

| Variable      | CI/CD environment     |
| ------------- | --------------------- |
| VSTS_PROCESS_LOOKUP_ID | Azure Devops |
| BUILDKITE_COMMIT | Buildkite |
| GITLAB_CI | GitlabCI |

## Support for different Docker container registries
The container registry to use when running `docker:push` is also defined by environment variables.

| Variable      | Container registry    | Example value |
| ------------- | --------------------- | ------------- |
| DOCKERHUB_REPOSITORY | Docker Hub | bitnami (resulting in bitnami/\<image> |
| ECR_URL | AWS ECR | 12345678.dkr.ecr.eu-west-1.amazonaws.com |
| CI_REGISTRY_IMAGE | Gitlab Registry | registry.gitlab.com/sparetimecoders/build-tools |
| QUAY_REPOSITORY | Quay.io | quay.io/bitnami |


Other environment variables that need to be defined (either automatically by the CI/CD environment or manually in the build pipeline) for each of the container registries are defined below.

### Docker Hub
| Variable | Description |
| -------- | ----------- |
| DOCKERHUB_USERNAME | Username |
| DOCKERHUB_PASSWORD | Password |

### AWS ECR
| Variable | Description |
| -------- | ----------- |
| ECR_REGION | Optionally specified region, will use eu-west-1 as default value |

### Gitlab Registry
| Variable | Description |
| -------- | ----------- |
| CI_BUILD_TOKEN | The build-token set by GitlabCI |
| CI_REGISTRY | The URL to the registry set by GitlabCI |

### Quay.io
| Variable | Description |
| -------- | ----------- |
| QUAY_USERNAME | Username |
| QUAY_PASSWORD | Password |

## Example Azure Devops pipeline (azure-pipelines.yml)

    resources:
      containers:
      - container: build-tools
        image: registry.gitlab.com/sparetimecoders/build-tools:master
    
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
          DOCKERHUB_PASSWORD: $(DOCKERHUB_PASSWORD)
          QUAY_PASSWORD: $(QUAY_PASSWORD)
      - script: deploy staging
        name: deploy_staging
        condition: succeeded()

## Example Buildkite pipeline (.buildkite/pipeline.yml)

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
    
## Example GitlabCI pipeline (.gitlab-ci.yaml)

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

# Developing

## Generate test mocks
    mockgen -package=mocks -destination=pkg/config/mocks/MockRepositoriesService.go gitlab.com/sparetimecoders/build-tools/pkg/config RepositoriesService