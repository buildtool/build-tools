# `.buildtools.yaml`
Configuration and setup is done in `.buildtools.yaml` file(s).
A typical configuration file consists of a [`registry`](registry.md) config
and a list of [`targets`](targets.md) to use.
The only keys allowed in the file are:

- targets
- registry

*Note:* [Multiple](files.md) files can be used for more advanced usage

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


## Configuration file from environment variables
A `.buildtools.yaml` file can be created by defining an environment variable in the build pipeline named `BUILDTOOLS_CONTENT`.
The value should be a base64-encoded string. This can be useful when setting up CI/CD pipelines where the file system is not
easily accessible.

On MacOS the value can be created and copied to the clipboard using the following snippet:

```sh
$ cat - <<EOF | base64 -w0 | pbcopy
targets:
  local-test:
    context: docker-desktop

EOF
```

**Note:** If `BUILDTOOLS_CONTENT` is set, no other configuration files will be used.

See the following sections for information on how to configure the different parts of the configuration files.
