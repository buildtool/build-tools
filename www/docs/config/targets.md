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

| Parameter     | Default                                | Description                                           |
| :------------ | :------------------------------------- | :---------------------------------------------------  |
| `context`     |                                        | Which context in the Kubernetes configuration to use  |
| `namespace`   | `default`                              | Specific namespace to deploy to                       |
| `kubeconfig`  | value of `KUBECONFIG` `ENV` variable   | Full path to a specific kubeconfig file to use        |


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
