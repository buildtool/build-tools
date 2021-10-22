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

## Export content from build
Buildtools `build` command support exporting content from the actual docker build process,
see [Custom build outputs].
By specifying a special stage in the `Dockerfile` and name it `export` you can use the `COPY`
directive to copy files from the build context to the local machine.
The copied files will be placed in a folder `exported`

### Example
Consider a `Dockerfile` like this:

```dockerfile
FROM debian as build
RUN echo "text to be copied to localhost" >> /testfile

# -- export stage
FROM scratch as export
# Copies the file /testfile from `build` stage to localhost
COPY --from=build  /testfile .

# -- resulting image stage
FROM scratch
# Do other stuff
```

Let's try it:

```shell
$ ls
Dockerfile

$ cat Dockerfile
FROM debian as build
RUN echo "text to be copied to localhost" >> /testfile

# -- export stage
FROM scratch as export
# Copies the file /testfile from `build` stage to localhost
COPY --from=build  /testfile .

# -- resulting image stage
FROM scratch
# Do other stuff
$ build
... <build output>
$ ls
Dockerfile  exported

$ ls exported
testfile

$ cat exported/testfile
text to be copied to localhost
```

[Custom build outputs]: (https://docs.docker.com/engine/reference/commandline/build/#custom-build-outputs)
[build-args]: (https://docs.docker.com/engine/reference/commandline/build/#set-build-time-variables---build-arg)
