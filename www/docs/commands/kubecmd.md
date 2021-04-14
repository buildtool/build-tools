# kubecmd

Generates a kubectl command, using the configuration from .buildtools.yaml if found.
Normal usage `kubecmd <target>`, but additional flags can be used to override:

|      Flag             |                   Description                                                   |
| :-------------------- | :-------------------------------------------------------------------------------|
| `--context`, `-c`           | Use a different context than the one found in configuration                     |
| `--namespace`, `-n`         | Use a different namespace than the one found in configuration                   |

## Default usage, with `.buildtools.yaml` file
Only the `target` name has to be specified
```sh
$ kubecmd local
kubectl --context docker-desktop --namespace default
```

### Overriding namespace from config:
```sh
$ kubecmd --namespace test local
kubectl --context docker-desktop --namespace test
```
