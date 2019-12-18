---
title: Commands
menu: true
weight: 5
---

## Available commands

### build

Performs a `docker build`, using a `Dockerfile` to build the application and tags the resulting image. By following the conventions no additional flags are needed, but the following flags are available:

|      Flag                        |                   Description                                        |
| :------------------------------- | :-------------------------------------------------------------------- |
| `-f/--file <path to Dockerfile>` | Used to override the default `Dockerfile` location (which is `$PWD`) | 
| `-skiplogin`                     | Disables login to docker registry (good for local testing)           | 
| `-build-arg key=value`           | Additional Docker [build-arg](https://docs.docker.com/engine/reference/commandline/#set-build-time-variables---build-arg) |

```sh
$ build -f docker/Dockerfile.build -skiplogin -build-arg AUTH_TOKEN=abc
```
    
### push

Performs a `docker push` of the image created by `build`.

By following the conventions no additional flags are needed, but the following flags are available:

|      Flag                       |                   Description                                       |
| :------------------------------ | :------------------------------------------------------------------ |
| `-f/--file <path to Dockerfile>`| Used to override the default `Dockerfile` location (which is `$PWD`)|

```sh
$ push -f docker/Dockerfile.build 
```   

### deploy

Deploys the built application to a Kubernetes cluster. Normal usage `deploy <environment>`, but additional flags can be used to override:

|      Flag                          |                   Description                                                   |
| :--------------------------------- | :-------------------------------------------------------------------------------|
| `-c/--context`                     | Use a different context than the one found in configuration                     |
| `-n/--namespace`                   | Use a different namespace than the one found in configuration                   |
| `-t/--timeout`                     | Override the default deployment waiting time for completion (default 2 minutes). 0 means forever, all other values should contain a corresponding time unit (e.g. 1s, 2m, 3h) |

```sh
$ deploy -n testing_namespace local 
```

### service-setup
Setup a new local git repository and scaffolds the project.
Remote Git repository and CI pipeline will be created and configured based on the `.buildtools.yaml` settings. 

Basic usage `service-setup <name>`, it's also possible to scaffold for a certain stack. Supported stacks:

* [go](https://golangci.com/)
* [scala](https://www.scala-lang.org/)

This will generate some basic project files for the different stacks, example
```sh
$ service-setup -s go test
$ tree test
    *TODO*
```
### kubecmd

TODO
