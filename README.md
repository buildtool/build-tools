
<p align="center">
  <a href="https://github.com/sparetimecoders/build-tools/blob/master/LICENSE"><img alt="LICENSE" src="https://img.shields.io/badge/license-MIT-blue.svg?maxAge=43200"></a>
  <a href="https://github.com/sparetimecoders/build-tools/releases"><img alt="releases" src="https://img.shields.io/github/release/sparetimecoders/build-tools.svg?maxAge=43200"></a>
  <a href="https://github.com/sparetimecoders/build-tools/releases"><img alt="Github All Releases" src="https://img.shields.io/github/downloads/sparetimecoders/build-tools/total.svg?maxAge=43200"></a>
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

# Available commands

## build
## push
## deploy

# Conventions

* `Dockerfile` must be present in the root of the project directory (*TODO Override name of file*). The `Dockerfile` will be used to build the project into a runnable docker image.
* The name of the directory will be used as the name of the docker image (*TODO Override by ENV*)
* The current commit id will be used as docker tag
* Kubernetes descriptor files must be located in the `k8s` folder under the root

Take a look at the build-tools-example repository (*TODO link*) to try it out.

## Using in CI/CD pipelines

## Example usage
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

