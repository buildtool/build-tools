# `.buildtools.yaml`
Configuration and setup is done in `.buildtools.yaml` file(s).
A typical configuration file consists of a [`registry`](registry.md) config
and a list of [`targets`](targets.md) to use.
The only keys allowed in the file are:

- targets
- registry

## Example
The following file specifies a Dockerhub registry and 2 deployment targets: `local-test` and `staging`
```yaml
registry:
  dockerhub:
    namespace: buildtool
targets:
  local-test:
    context: docker-desktop
  staging:
    context: staging-aws-eu-west-1
    namespace: my-test
```

See the following sections for information on how to configure the different parts of the configuration files.
