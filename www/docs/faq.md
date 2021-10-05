# Frequently Asked Questions
Or stuff just good-to-know...

## What happened to X-action?
We deprecated the build, push, deploy Github actions in favour of the new [setup-buildtools-action](/ci/github)
## Dealing with different docker versions
buildtools defaults to using the latest version of the docker client
(the actual version is determined by the docker client library that is used).
This might cause issues if your docker server is running an older version.

Errors like:
```shell
Error response from daemon: client version 1.41 is too new. Maximum supported API version is 1.40
```

The docker client version can be specified with the `env` variable `DOCKER_API_VERSION`
Depending on your setup you might be able to use `export` somewhere "globally"
```shell
export DOCKER_API_VERSION=1.40
```
Or just use it when running the actual command
```shell
DOCKER_API_VERSION=1.40 build
```
