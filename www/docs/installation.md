# Installation

You can install the pre-compiled binary (in several different ways), use Docker or compile from source.

## Installation pre-built binaries
**Homebrew tap**

```sh
$ brew install buildtool/taps/build-tools
```

**Shell script**
```sh
$ curl -sfL https://raw.githubusercontent.com/buildtool/build-tools/main/install.sh | sh
```
**Manually**

Download the pre-compiled binaries from the [releases](https://github.com/buildtool/build-tools/releases) page and copy to the desired location.

## Running with Docker
You can also use it within a Docker container. To do that, you’ll need to execute something more-or-less like the following:
```sh
$ docker run --rm --privileged \
  -v $PWD:/repo \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -w /repo \
  -e DOCKER_USERNAME \
  -e DOCKER_PASSWORD \
  buildtool/build-tools build
```

## Compiling from source

Here you have two options:

If you want to contribute to the project, please follow the
steps on our [contributing guide](/contributing).

If you just want to build from source for whatever reason, follow these steps:

**Clone:**

```sh
git clone https://github.com/buildtool/build-tools
cd build-tools
```

**Get the dependencies:**

```sh
go get ./...
```

**Build:**

```sh
go build  ./cmd/build/build.go
```

**Verify it works:**

```sh
./build --version
```
