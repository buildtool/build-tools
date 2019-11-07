---
title: Conventions
menu: true
weight: 3
---

* The project folder must be a [Git](https://git-scm.com/) repository, with a least one commit
* `Dockerfile` must be present in the root of the project directory (this can be overriden with [flags](/commands\#build)). The `Dockerfile` will be used to build the project into a runnable docker image.
* Kubernetes descriptor files must be located in the `k8s` folder
* The name of the directory will be used as the name of the docker image (if running in CI `ENV` variables will be used to determine the name of the project being built)
* The current commit id will be used as docker tag

Take a look at the [build-tools-example repository](https://github.com/buildtool/build-tools-examples) to try it out.
