# Targets

`targets` specifies the different 'deployment targets' to use for the project.
The target match Kubernetes cluster [configurations](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/#define-clusters-users-and-contexts)
to deploy projects.
Setting up Kubernetes contexts and namespaces is not handled by these tools.

The only required configuration is `context` and `<name>` must be unique.

```yaml
targets:
  <name>:
    context:
    namespace:
    kubeconfig:
```

| Parameter     | Default                                       | Description                                           |
| :------------ | :-------------------------------------------- | :---------------------------------------------------  |
| `context`     |                                               | Which context in the Kubernetes configuration to use  |
| `namespace`   | `default`                                     | Specific namespace to deploy to                       |
| `kubeconfig`  | value of `KUBECONFIG` environment variable    | Full path to a specific kubeconfig file to use        |

The `KUBECONFIG_CONTENT` environment variable (probably most useful in CI/CD pipelines) can be used to provide the
content of a "kubeconfig" file. If set, buildtools will create a temporary file with that content to use as the `kubeconfig` value.
`KUBECONFIG_CONTENT` can be either a base64 encoded string or plain text.

**Note:** the `kubeconfig` parameter in config file overrides both the `KUBECONFIG` and `KUBECONFIG_CONTENT` environment
variables if set.

## Examples

````yaml
targets:
  local:
    context: docker-desktop
    namespace: default
  local-test:
    context: docker-desktop
    namespace: test
````
