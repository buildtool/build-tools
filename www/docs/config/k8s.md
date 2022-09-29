# Kubernetes

Descriptor files are placed in a `k8s` directory in the root of your project file. This directory contains all the yaml
files used to describe the Kubernetes deployment tasks needed to run this service. Target specific files are handled by
using a `-<target>` "suffix", i.e. `ingress-prod.yaml`.

Files with a `.yaml` suffix will
be [applied](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#apply) to the Kubernetes cluster.

If both a **common** file and a **target-specific** file with the same "basename", i.e. service.yaml and
service-local.yaml, only the **target-specific** file will be applied for a `target` and the **common** file will be
applied to all other targets.

Files with a `.sh` suffix will be run on the machine executing the [`deploy`](/commands/deploy) command. This can be
useful to setup secrets/configurations, mostly for local use. Note that only `.sh` files matching the `target` using the
rules in the above paragraph will be executed, with the difference that no **common** files will be executed.

All other files in `k8s` will be ignored by the `deploy` command.

## Available variables
The following variables are possible to use for substitution in the descriptor files:


| Parameter   | Description                                                 |
|:------------|:------------------------------------------------------------|
| `IMAGE`     | The full image name (`registry/name:tag`)                   |
| `COMMIT`    | The commit SHA (`3b701067e6a6943c773b9dc183fcc39cd31a2ff0`) |
| `TIMESTAMP` | The current time (`2022-02-22T09:16:01+01:00`)              |


## Example

````
$ cd projecct
$ tree
.
└── k8s
    ├── deploy.yaml
    ├── ingress-local.yaml
    ├── ingress-prod.yaml
    ├── service.yaml
    ├── service-local.yaml
    └── setup-local.sh
    └── skipped.sh
````

Given the structure above:

````sh
deploy local
````

Will apply `deploy.yaml`, `ingress-local.yaml` and `service-local.yaml` and execute `setup-local.sh`.

````sh
deploy staging
````

Will apply `deploy.yaml` and `service.yaml`.

````sh
deploy prod
````

Will apply `deploy.yaml`, `ingress-prod.yaml` and `service.yaml`.
