# deploy

Deploys the built application to a Kubernetes cluster. Normal usage `deploy <target>`, but additional flags can be used to override:

|      Flag             |                   Description                                                   |
| :-------------------- | :-------------------------------------------------------------------------------|
| `--context`, `-c`           | Use a different context than the one found in configuration                     |
| `--namespace`, `-n`         | Use a different namespace than the one found in configuration                   |
| `--timeout`, `-t`           | Override the default deployment waiting time for completion (default 2 minutes). <br>0 means forever, all other values should contain a corresponding time unit (e.g. 1s, 2m, 3h)|
| `--tag`                    | Override the default tag to use (instead of the current commit tag or the value from CI) |

## Default usage, with `.buildtools.yaml` file
Only the `target` name has to be specified
```sh
$ deploy local
```

### Overriding namespace from config:
```sh
$ deploy --namespace test local
```


## Usage without `.buildtools.yaml` file
In this case we need to at least specify the Kubernetes context to use for deployment:
```sh
$ deploy --context docker-desktop
```

This will set the `namespace` to `default`

### Specifying namespace:
```sh
$ deploy --context docker-desktop --namespace test
```
