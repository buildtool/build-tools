# `.buildtools.yaml`
Configuration and setup is done in `.buildtools.yaml` file(s).
A typical configuration file consists of a [`registry`](registry.md) config
and a list of [`targets`](targets.md) to use.


|      Key             |                   Description       |
| :------------------- | :---------------------------------- |
| registry  | [registry](registry.md) registry to push to    |
| targets   | [targets](targets.md) to deploy to             |
| git       |  [git](git.md) configuration block             |
| gitops    |  [git repos](gitops.md) to push descriptors to |


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
This can be useful when setting up CI/CD pipelines where the file system is not
easily accessible.

On MacOS the value can be created and copied to the clipboard using the following snippet:

```sh
$ cat - <<EOF | base64 -w0 | pbcopy
targets:
  local-test:
    context: docker-desktop

EOF
```

`BUILDTOOLS_CONTENT` can be either a base64 encoded string or plain text.

**Note:** If `BUILDTOOLS_CONTENT` is set, no other configuration files will be used.

See the following sections for information on how to configure the different parts of the configuration files.
