# Configuration files

Configuration and setup is done in `.buildtools.yaml` files. 
Those files must be present in the project folder or upwards in the directory structure. 
This lets you create a common `.buildtools.yaml` file to be used for a set of projects.
The `.buildtools.yaml` files will be merged together, and settings from file closest to the project being used first.
 
Example:
```sh
$ pwd
~/source/
$ tree
.
├── customer1
│   ├── project1
│   └── project2
└── customer2
    └── project1
```
        
Here we can choose to put a `.buildtools.yaml` file in the different `customer` directories since they (most likely) have different deployment configuration.
```sh
$ pwd
~/source/
$ cat .buildtools.yaml
registry:
  dockerhub:
    namespace: buildtool
environments:
  - name: local
    context: docker-desktop
    namespace: default

$ cd customer1
$ cat .buildtools.yaml
environments:
  - name: prod
    context: production
    namespace: default
    kubeconfig: ~/.kube/config.d/production.yaml

$ cd project2
$ cat .buildtools.yaml
environments:
  - name: staging
    context: staging
    namespace: project2

$ build -printconfig

$ cd ..

$ build -printconfig
```

### `.buildtools.yaml`
A typical configuration file consists of a `registry` config and a list of `environments` to use.

```yaml
registry:
  dockerhub:
    namespace: buildtool
environments:
  local-test:
    context: docker-desktop
  staging:
    context: staging-aws-eu-west-1
    namespace: my-test 
```    

### Configuration from environment
As an option file can be created by defining an environment variable in the build pipeline named `BUILDTOOLS_CONTENT`. 
The value should be a base64-encoded string. On MacOS the value can be created and copied to the clipboard using the following snippet:

```sh
$ cat - <<EOF | base64 -w0 | pbcopy
environments:
  local-test:
    context: docker-desktop
)
EOF
```

See the following sections for information on how to configure the different parts of the configuration files.