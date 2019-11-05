---
title: Kubernetes
menu: true
weight: 14
---

here must be a `k8s` directory in the root of your project file. 
This directory contains all the yaml files used to describe the Kubernetes deployment tasks needed to run this service. 
Environment specific files are handled by using a `-<environment>` suffix, i.e. `ingress-prod.yaml`.

````yaml
$ cd projecct
$ tree
.
└── k8s
    ├── deploy.yaml
    ├── ingress-local.yaml
    ├── ingress-prod.yaml
    └── setup-local.sh
````

Files with a `.yaml` suffix will be [applied](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#apply) to the Kubernetes cluster.
Files with a `.sh` suffix will be run on the machine executing the [`deploy`](/commands#deploy) command 