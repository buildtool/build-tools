
<p align="center">
  <a href="https://github.com/sparetimecoders/build-tools/blob/master/LICENSE"><img alt="LICENSE" src="https://img.shields.io/badge/license-MIT-blue.svg?maxAge=43200"></a>
  <a href="https://github.com/sparetimecoders/build-tools/releases"><img alt="GitHub release (latest by date)" src="https://img.shields.io/github/v/release/sparetimecoders/build-tools"></a>
  <a href="https://github.com/sparetimecoders/build-tools/releases"><img alt="GitHub All Releases" src="https://img.shields.io/github/downloads/sparetimecoders/build-tools/total"></a>
  <a href="pulls"><img alt="GitHub pull requests" src="https://img.shields.io/github/issues-pr/sparetimecoders/build-tools"></a>
</p>

<p align="center">
  <a href="https://github.com/sparetimecoders/build-tools/actions"><img alt="GitHub Actions" src="https://github.com/sparetimecoders/build-tools/workflows/Go/badge.svg"></a>
  <a href="https://codecov.io/github/sparetimecoders/build-tools"><img alt="Coverage Status" src="https://codecov.io/gh/sparetimecoders/build-tools/branch/master/graph/badge.svg"></a>
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
    
# Available commands

## build
Performs a `docker build`, using a `Dockerfile` to build the application and tags the resulting image. By following the conventions no additional flags are needed, but the following flags are available:

|      Flag                         |                   Description                                         |
| :-------------------------------: | :-------------------------------------------------------------------: |
| `-f/--file <path to Dockerfile>`  | Used to override the default `Dockerfile` location (which is `$PWD`)  |
| `-skiplogin`                      | Disables login to docker registry (good for local testing)            |
| `-build-arg key=value`            | Additional Docker [build-arg](https://docs.docker.com/engine/reference/commandline/#set-build-time-variables---build-arg) |

    $ build -f docker/Dockerfile.build -skiplogin -build-arg AUTH_TOKEN=abc
    
## push
Performs a `docker push` of the image created by `build`.

By following the conventions no additional flags are needed, but the following flags are available:

|      Flag                        |                   Description                                        |
| :------------------------------: | :------------------------------------------------------------------: |
| `-f/--file <path to Dockerfile>` | Used to override the default `Dockerfile` location (which is `$PWD`) |

    $ push -f docker/Dockerfile.build 
    
## deploy
Deploys the built application to a Kubernetes cluster. Normal usage `deploy <context>`, but additional flags can be used:

|      Flag                           |                   Description                                  |
| :---------------------------------: | :------------------------------------------------------------: |
| `-c/--context <path to Dockerfile>` | Use a different context than the one found in configuration    |
| `-n/--namespace`                    | Use a different namespace than the one found in configuration  |

    $ deploy -n testing_namespace local 

## service-setup
Setup a new git repository and scaffolds the project.

Basic usage `service-setup <name>`, it's also possible to scaffold for a certain stack. Supported stacks:
* go
* scala

## kubecmd

# Usage
## Conventions

* `Dockerfile` must be present in the root of the project directory (*TODO Override name of file*). The `Dockerfile` will be used to build the project into a runnable docker image.
* The name of the directory will be used as the name of the docker image (*TODO Override by ENV*)
* The current commit id will be used as docker tag
* Kubernetes descriptor files must be located in the `k8s` folder under the root

Take a look at the build-tools-example repository (*TODO link*) to try it out.

### Example
After installing (*TODO link*) the tools, clone the build-tools-example repository (*TODO link*), cd into it and execute the `build` command.

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


### Using in CI/CD pipelines
