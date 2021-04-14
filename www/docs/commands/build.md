# build

Performs a `docker build`, using a `Dockerfile` to build the application and tags the resulting image. By following the conventions no additional flags are needed, but the following flags are available:

|      Flag                        |                   Description                                        |
| :------------------------------- | :-------------------------------------------------------------------- |
| `--file`,`-f` `<path to Dockerfile>`    | Used to override the default `Dockerfile` location (which is `$PWD`) |
| `--no-login`                     | Disables login to docker registry (good for local testing)           |
| `--no-pull`                      | Disables pulling of remote images if they already exist (good for local testing)           |
| `--build-arg key=value`          | Additional Docker [build-arg](https://docs.docker.com/engine/reference/commandline/build/#set-build-time-variables---build-arg) |

```sh
$ build --file docker/Dockerfile.build --skip-login --build-arg AUTH_TOKEN=abc
```
