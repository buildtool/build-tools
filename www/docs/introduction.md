# Introduction

build-tools is a set of highly opinionated tools for creating and building
components/services into [docker](https://www.docker.com/) images and deploying
them to [Kubernetes](https://kubernetes.io/) clusters.

By following the conventions set by the tools, building and deploying applications is made simpler. It streamlines the
process of building and deploying using the same commands on your local development machine and in a CI/CD pipeline.

The basic usage is `build`, `push` and `deploy` (or `promote` if you are using [GitOps](config/gitops.md)). This will
build a docker image of your code using your provided
[Dockerfile](https://docs.docker.com/engine/reference/builder/) making it possible to customize the actual build
process. The built image is then pushed to a [Docker Registry](https://docs.docker.com/registry/)
of your choosing. This is of course optional, but a necessary step to be able to deploy the built image on a (non-local)
Kubernetes cluster. Finally, the code is deployed to the Kubernetes cluster using the
provided [descriptor files](config/k8s.md).


<script id="asciicast-387073" src="https://asciinema.org/a/387073.js" id="asciicast-387073" async data-autoplay="true" data-preload="true" data-theme="solarized-dark"></script>
