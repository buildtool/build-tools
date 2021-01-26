# push

Performs a [Docker push](https://docs.docker.com/engine/reference/commandline/push/) of the image created by `build`.

By following the conventions no additional flags are needed, but the following flags are available:

|      Flag                       |                   Description                                       |
| :------------------------------ | :------------------------------------------------------------------ |
| `--file <path to Dockerfile>`| Used to override the default `Dockerfile` location (which is `$PWD`)|

```sh
$ push --file docker/Dockerfile.build
```

