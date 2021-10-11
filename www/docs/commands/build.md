# build

Performs a `docker build`, using a `Dockerfile` to build the application and tags the resulting image.
By following the conventions no additional flags are needed, but the following flags are available:

|      Flag                        |                   Description                                        |
| :------------------------------- | :-------------------------------------------------------------------- |
| `--file`,`-f` `<path to Dockerfile>`    | Used to override the default `Dockerfile` location (which is `$PWD`) |
| `--no-login`                     | Disables login to docker registry (good for local testing)           |
| `--no-pull`                      | Disables pulling of remote images if they already exist (good for local testing)           |
| `--build-arg key=value`          | Additional Docker [build-arg](https://docs.docker.com/engine/reference/commandline/build/#set-build-time-variables---build-arg) |

```sh
$ build --file docker/Dockerfile.build --skip-login --build-arg AUTH_TOKEN=abc
```

## Build-args
The following [build-arg] are automatically made available:

|      Arg    |                   Value                                |
| :---------- | :----------------------------------------------------- |
| `CI_COMMIT` | The commit being built as exposed by [CI](../ci/ci.md) |
| `CI_BRANCH` | The branch being built as exposed by [CI](../ci/ci.md) |

they can be used in a `Dockerfile` like:

```dockerfile
FROM ubuntu
ARG CI_BRANCH

RUN echo "Building $CI_BRANCH"
```

[build-args]: (https://docs.docker.com/engine/reference/commandline/build/#set-build-time-variables---build-arg)
