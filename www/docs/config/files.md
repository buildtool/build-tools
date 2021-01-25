# Configuration files

The `.buildtools.yaml` files must be present in the project folder or upwards in the directory structure.
This lets you create a common `.buildtools.yaml` file to be used for a set of projects.
The `.buildtools.yaml` files will be merged together, and settings from the file closest to the project being used first.

## Example:

```sh
$ pwd
~/source/
$ tree
.
├── customer1
│ ├── project1
│ └── project2
└── customer2
    └── project1
```

Here we can choose to put a `.buildtools.yaml` file in the different `customer` directories since they (most likely)
have different deployment configuration.
```sh
$ pwd
~/source/
$ cat .buildtools.yaml
registry:
  dockerhub:
    namespace: buildtool
targets:
  local:
    context: docker-desktop
    namespace: default

$ cd customer1
$ cat .buildtools.yaml
targets:
  prod:
    context: production
    namespace: default
    kubeconfig: ~/.kube/config.d/production.yaml

$ cd project2
$ cat .buildtools.yaml
targets:
  staging:
    context: staging
    namespace: project2

```

TODO Describe the different "final" settings
