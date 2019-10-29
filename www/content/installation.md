---
title: Installation
weight: 2
menu: true
---

You can install the pre-compiled binary (in several different ways), use Docker or compile from source.

## Installation pre-built binaries
**Homebrew tap**

```sh 
$ brew install buildtool/taps/build-tools
```

**Shell script**
```sh
$ curl -sfL https://raw.githubusercontent.com/buildtool/build-tools/master/install.sh | sh
```
**Manually**

Download the pre-compiled binaries from the [releases](https://github.com/buildtool/build-tools/releases) page and copy to the desired location.

## Docker
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
```sh

    # clone it outside GOPATH
    $ git clone https://github.com/buildtool/build-tools
    $ cd build-tools
    
    # get dependencies using go modules (needs go 1.11+)
    $ go get ./...
    
    # build
    $ go build ./cmd/build
    
    # check it works
    ./build -version
```
