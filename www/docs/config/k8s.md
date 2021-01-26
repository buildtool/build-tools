# Kubernetes

Descriptor files are placed in a `k8s` directory in the root of your project file.
This directory contains all the yaml files used to describe the Kubernetes deployment tasks needed to run this service.
Target specific files are handled by using a `-<target>` "suffix", i.e. `ingress-prod.yaml`.

Files with a `.yaml` suffix will be [applied](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#apply) to the Kubernetes cluster.

Files with a `.sh` suffix will be run on the machine executing the [`deploy`](/commands/deploy) command.
This can be useful to setup secrets/configurations, mostly for local use.
Note that only `.sh` files matching the `target` will be executed.

All other files in `k8s` will be ignored by the `deploy` command.

## Example
````yaml
$ cd projecct
$ tree
.
└── k8s
    ├── deploy.yaml
    ├── ingress-local.yaml
    ├── ingress-prod.yaml
    └── setup-local.sh
    └── skipped.sh

````

Given the structure above:

````sh
deploy local
````
Will apply `deploy.yaml` and `ingress-local.yaml` and execute `setup-local.sh`.

````sh
deploy staging
````
Will apply `deploy.yaml` and nothing else.

````sh
deploy prod
````
Will apply `deploy.yaml` and `ingress-prod.yaml`.

