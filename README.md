
<p align="center">
  <a href="https://github.com/sparetimecoders/build-tools/actions"><img alt="GitHub Actions" src="https://github.com/sparetimecoders/build-tools/workflows/Go/badge.svg"></a>
  <a href="https://github.com/sparetimecoders/build-tools/releases"><img alt="GitHub release (latest by date)" src="https://img.shields.io/github/v/release/sparetimecoders/build-tools"></a>
  <a href="pulls"><img alt="GitHub pull requests" src="https://img.shields.io/github/issues-pr/sparetimecoders/build-tools"></a>
  <a href="https://github.com/sparetimecoders/build-tools/releases"><img alt="GitHub All Releases" src="https://img.shields.io/github/downloads/sparetimecoders/build-tools/total"></a>
</p>

<p align="center">
  <a href="https://github.com/sparetimecoders/build-tools/blob/master/LICENSE"><img alt="LICENSE" src="https://img.shields.io/badge/license-MIT-blue.svg?maxAge=43200"></a>
  <a href="https://codecov.io/github/sparetimecoders/build-tools"><img alt="Coverage Status" src="https://codecov.io/gh/sparetimecoders/build-tools/branch/master/graph/badge.svg"></a>
  <a href="https://codebeat.co/projects/github-com-sparetimecoders-build-tools-master"><img alt="codebeat badge" src="https://codebeat.co/badges/434836f7-e0ab-4af9-8ef8-60cde2738764" /></a>
  <a href="https://goreportcard.com/report/github.com/sparetimecoders/build-tools"><img alt="goreportcard badge" src="https://goreportcard.com/badge/github.com/sparetimecoders/build-tools" /></a>
  <a href="https://libraries.io/github/sparetimecoders/build-tools"><img alt="" src="https://img.shields.io/librariesio/github/sparetimecoders/build-tools"></a>
</p>


# build-tools
*WIP*

A set of highly opinionated tools for creating and building components/services into [docker](https://www.docker.com/) images and deploying them to [Kubernetes](https://kubernetes.io/) clusters.

By following the conventions set by the tools, building and deploying applications is made simpler.

The only hard requirement is to provide a [Dockerfile](https://docs.docker.com/engine/reference/builder/) which describes how to build and run your application.

The configuration needed is done by environment variables (most likely for CI/CD) and yaml files (for local use).

# Installation
You can install the pre-compiled binary (in several different ways), use Docker or compile from source.

Here are the steps for each of them:

## Homebrew tap

    $ brew install sparetimecoders/taps/build-tools
## Shell script

    $ curl -sfL https://raw.githubusercontent.com/sparetimecoders/build-tools/master/install.sh | sh
## Manually

Download the pre-compiled binaries from the [releases](https://github.com/sparetimecoders/build-tools/releases) page and copy to the desired location.
## Docker
You can also use it within a Docker container. To do that, you’ll need to execute something more-or-less like the following:

    $ docker run --rm --privileged \
      -v $PWD:/repo \
      -v /var/run/docker.sock:/var/run/docker.sock \
      -w /repo \
      -e DOCKER_USERNAME \
      -e DOCKER_PASSWORD \
      sparetimecoders/build-tools build
      
## Compiling from source

    # clone it outside GOPATH
    $ git clone https://github.com/sparetimecoders/build-tools
    $ cd build-tools
    
    # get dependencies using go modules (needs go 1.11+)
    $ go get ./...
    
    # build
    $ go build ./cmd/build
    
    # check it works
    ./build -version
    
# Usage
If the project follow the conventions and some configuration for the CI/CD tool are present (probably using `ENV` variables), a pipeline could be as simple as:

    $ build
    $ push
    $ deploy <environment>
    
to build, push and deploy the project to a Kubernetes cluster.    

## Conventions

* `Dockerfile` must be present in the root of the project directory (this can be overriden with [flags](#build)). The `Dockerfile` will be used to build the project into a runnable docker image.
* Kubernetes descriptor files must be located in the `k8s` folder
* The name of the directory will be used as the name of the docker image (if running in CI `ENV` variables will be used to determine the name of the project being built)
* The current commit id will be used as docker tag

Take a look at the [build-tools-example repository](https://github.com/sparetimecoders/build-tools-examples) to try it out.

### Project structure
The project folder must be a [Git](https://git-scm.com/) repository, with a least one commit.

There must be a `k8s` directory in the root of your project. This directory contains all the `yaml` files used to describe the Kubernetes tasks needed to run this service.
Environment specific files can be handled in two different ways depending on personal preference. 
They can either be located in sub-directories, for example `k8s/local` for local setup.

    $ cd projecct
    $ tree
    .
    └── k8s
        ├── deploy.yaml
        ├── local
        │   ├── ingress.yaml
        └── prod
            └── ingress.yaml

Or they can be defined using a `-<environment>` suffix, i.e. `ingress-prod.yaml`

    $ cd projecct
    $ tree
    .
    └── k8s
        ├── deploy.yaml
        ├── ingress-local.yaml
        ├── ingress-prod.yaml

### Configuration
Configuration and setup is done in `.buildtools.yaml` files. 
Those files must be present in the project folder or upwards in the directory structure. 
This lets you create a common `.buildtools.yaml` file to be used for a set of projects.
The `.buildtools.yaml` files will be merged together, and settings from file closest to the project being used first.

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
        
Here we can choose to put a `.buildtools.yaml` file in the different `customer` directories since they (most likely) have different deployment configuration.

    $ pwd
    ~/source/
    $ cat .buildtools.yaml
    registry:
      dockerhub:
        repository: sparetimecoders
    environments:
      - name: local
        context: docker-desktop
        namespace: default

    $ cd customer1
    $ cat .buildtools.yaml
    environments:
      - name: prod
        context: production
        namespace: default
        kubeconfig: ~/.kube/config.d/production.yaml

    $ cd project2
    $ cat .buildtools.yaml
    environments:
      - name: staging
        context: staging
        namespace: project2

    $ build -printconfig
    
    $ cd ..
    
    $ build -printconfig
        
Context and namespaces must be provided/created/configured elsewhere.

# .build-tools.yaml

`environments` specifies the different deployment 'targets' to use for the project.
The environments matches Kubernetes cluster [configurations](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/#define-clusters-users-and-contexts) to deploy projects.
The only required configuration is `context`.

    environments:
      <name>:
        context:
        namespace:
        kubeconfig:


| Parameter     | Default                                | Description                                           |
| :------------ | :------------------------------------- | :---------------------------------------------------  |
| `context`     |                                        | Which context in the Kubernetes configuration to use  |
| `namespace`   | `default`                              | Specific namespace to deploy to                       |
| `kubeconfig`  | value of `KUBECONFIG` `ENV` variable   | Full path to a specific kubeconfig file to use        |


The `registry` defines the docker registry used for the project. This will primarily be used for CI pipelines to push built docker images.
Locally it can be used to build images with correct tags, making it possible to deploy locally built images.

*TODO* Resulting image names for different providers?

The following registries are supported:

`dockerhub`

[Docker hub](https://hub.docker.com/) registry

| Parameter         | Description                           |
| :---------------- | :-----------------------------------  |
| `namespace`       |  The namespace to publish to          |
| `username`        |  User to authenticate                 |
| `password`        |  Password for `user` authentication   |

`ecr`

AWS [ECR](https://docs.aws.amazon.com/ecr/index.html) docker registry.
AWS Credentials must be supplied as `ENV` variables, read more [here](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html).

| Parameter | Description                                                                                |
| :-------- | :----------------------------------------------------------------------------------------- |
| `url`     | The ECR registry URL                                                                       |
| `region`  | Specify a region (if it's possible to derive from the `url` parameter it can be omitted)   |


`github`

Github [package registry](https://help.github.com/en/github/managing-packages-with-github-package-registry/about-github-package-registry).

To authenticate `token` or a combination of `username` and `password` must be provided.

| Parameter       | Description                                           |
| :-------------- | :---------------------------------------------------  |
| `repository`    | The repository part of the docker image name          |
| `username`      | User to authenticate                                  |
| `password`      | Password for `user` authentication                    |
| `token`         | A personal access token to use for authentication     |


`gitlab`

Gitlab [container registry](https://docs.gitlab.com/ee/user/packages/container_registry/).

| Parameter        | Description                                           |
| :--------------- | :---------------------------------------------------  |
| `registry`       | The repository part of the docker image name          |
| `repository`     | The repository part of the docker image name          |
| `token`          | A personal access token to use for authentication     |

`quay`

Quay [docker registry](https://docs.quay.io/)

| Parameter       | Description                                           |
| :-------------- | :---------------------------------------------------  |
| `repository`    | The repository part of the docker image name          |
| `username`      | User to authenticate                                  |
| `password`      | Password for `user` authentication                    |

### Example

After [installing](#installation) the tools, clone the [build-tools-example repository](https://github.com/sparetimecoders/build-tools-examples), cd into it and execute the `build` command.

    $ build
    Using CI none
    
    no Docker registry found

Since we we haven't setup a`.buildtools.yaml` (*TODO LINK in doc*) file, nothing has been configured, and to be able to build a docker image we must specify where we (potentially) want to push it later. In other words, setting the [tags](https://docs.docker.com/engine/reference/commandline/tag/) of the created image.
Luckily we can use environment variables as well, let's try:

    $ DOCKERHUB_REPOSITORY=sparetimecoders build
    Using CI none
    
    Using registry Dockerhub

    Login Succeeded
    Using build variables commit 7c76db502b4a70df5480d6ff438ae10e374b420e on branch master

As we can see, the `build` command identified that we are using Dockerhub, and extracted the commit id and branch information from the local git repository.
Notice that the name of the current directory is used as the image name.
After the successful build the image is tagged with the commit id and branch.

    Successfully tagged sparetimecoders/buildtools-examples:7c76db502b4a70df5480d6ff438ae10e374b420e
    Successfully tagged sparetimecoders/buildtools-examples:master
    Successfully tagged sparetimecoders/buildtools-examples:latest



    
Now that we have a docker image, let's publish it to the docker repository (this of course requires write access to the repository).

    $ DOCKERHUB_REPOSITORY=sparetimecoders DOCKERHUB_PASSWORD=<PASSWORD> DOCKERHUB_USERNAME=<USERNAME> push
    ...
    Pushing tag 'sparetimecoders/buildtools-examples:7c76db502b4a70df5480d6ff438ae10e374b420e'
    ...

    
    
    
*TODO Link to more environment variables and stuff*

# Available commands

## build
Performs a `docker build`, using a `Dockerfile` to build the application and tags the resulting image. By following the conventions no additional flags are needed, but the following flags are available:

|      Flag                        |                   Description                                        |
| :------------------------------- | :-------------------------------------------------------------------- |
| `-f/--file <path to Dockerfile>` | Used to override the default `Dockerfile` location (which is `$PWD`) | 
| `-skiplogin`                     | Disables login to docker registry (good for local testing)           | 
| `-build-arg key=value`           | Additional Docker [build-arg](https://docs.docker.com/engine/reference/commandline/#set-build-time-variables---build-arg) |

    $ build -f docker/Dockerfile.build -skiplogin -build-arg AUTH_TOKEN=abc
    
## push
Performs a `docker push` of the image created by `build`.

By following the conventions no additional flags are needed, but the following flags are available:

|      Flag                       |                   Description                                       |
| :------------------------------ | :------------------------------------------------------------------ |
| `-f/--file <path to Dockerfile>`| Used to override the default `Dockerfile` location (which is `$PWD`)|

    $ push -f docker/Dockerfile.build 
    
## deploy
Deploys the built application to a Kubernetes cluster. Normal usage `deploy <environment>`, but additional flags can be used to override:

|      Flag                          |                   Description                                 |
| :--------------------------------- | :------------------------------------------------------------ |
| `-c/--context`                     | Use a different context than the one found in configuration   |
| `-n/--namespace`                   | Use a different namespace than the one found in configuration |

    $ deploy -n testing_namespace local 

## service-setup
Setup a new local git repository and scaffolds the project.
Remote Git repository and CI pipeline will be created and configured based on the `.buildtools.yaml` settings. 

Basic usage `service-setup <name>`, it's also possible to scaffold for a certain stack. Supported stacks:
* go
* scala

This will generate some basic project files for the different stacks, example

    $ service-setup -s go test
    $ tree test
    *TODO*

## kubecmd


### Using in CI/CD pipelines
