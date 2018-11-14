# build-tools
A repository with tools for building components/services into [docker](https://www.docker.com/) images and deploying the to [Kubernetes](https://kubernetes.io/) clusters.


## Common build scripts
Scripts to be used for building and deploying our services (typically when running in a CI/CD environment). They can also be used for local builds/deploy.

    source ${BUILD_TOOLS_PATH}/docker.sh

To be able to use the build scripts locally:

- Clone this repository
- Add `BUILD_TOOLS_PATH=<PATH TO THIS REPOSITORY>` to your shell environment, typically in `.bash_profile` or something similar

## Build project structure
The scripts assume that the project follow the directory layout described below,it also depends on configuration in `environments.sh` to determine valid target environments and deployment commands.


Example:

    $ cat environments.sh
    ...
    valid_environments=(
    ["local"]="--context docker-for-desktop --namespace default"
    ["staging"]="--context docker-for-desktop --namespace staging"
    ["prod"]="--context docker-for-desktop --namespace prod"
    )
    ...

This defines three environments (local,staging,prod) all which are to be deployed to a local Kubernetes cluster but in different namespaces. 
Context and namespaces must be provided/created/configured elsewhere.

### Project structure
The project folder must be a [Git](https://git-scm.com/) repository, with a least one commit.

There must be a `deployment_files` directory in the root of your project file. This directory contains all the `yaml` files used to describe the Kubernetes deployment tasks needed to run this service.
Environment specific files are to be located in sub-directories, for example `deployment_files/local` for local setup.

    $ cd projecct
    $ tree
    .
    └── deployment_files
        ├── deploy.yaml
        ├── local
        │   ├── local-ingress.yaml
        │   └── setup-local.sh
        └── prod
            └── prod-ingress.yaml


