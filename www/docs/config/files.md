# Configuration files

The `.buildtools.yaml` file(s) must be present in the project folder or upwards
in the directory structure.
This lets you create a common `.buildtools.yaml` file to be used for a set of projects.
The `.buildtools.yaml` files will then be merged together,
and settings from the file closest to the project being used first.

## Example:

```sh
$ tree
.
├── customer1
│ ├── project1
│ └── project2
└── customer2
    └── project1
```

Here we can choose to put a `.buildtools.yaml` file in the different `customer` directories
since they (most likely) have different deployment configuration.

But both `project1` and `project2` for `cutomer1` use the same repository, so we can share that.

```sh
$ cat .buildtools.yaml
targets:
  local:
    context: docker-desktop

$ cat customer1/.buildtools.yaml
registry:
  dockerhub:
    namespace: buildtool
targets:
  prod:
    context: production
    kubeconfig: ~/.kube/config.d/production.yaml

$ cat customer1/project1/.buildtools.yaml
targets:
  staging:
    context: test

$ cat customer1/project2/.buildtools.yaml
targets:
  staging:
    context: staging
    namespace: project2

$ cat customer2/project1/.buildtools.yaml
targets:
  staging:
    context: local
    namespace: customer2
```



TODO Describe the different "final" settings
