# push

Performs a [Docker push](https://docs.docker.com/engine/reference/commandline/push/) of the image created by `build`.

By following the conventions no additional flags are needed, but the following flags are available:

|      Flag                       |                   Description                                       |
| :------------------------------ | :------------------------------------------------------------------ |
| `--file`,`-f` `<path to Dockerfile>`| Used to override the default `Dockerfile` location (which is `$PWD`)|

```sh
$ push --file docker/Dockerfile.build
```

## GitHub Actions outputs

When running in GitHub Actions, the `push` command writes the following step outputs to `$GITHUB_OUTPUT`:

| Output       | Description                          |
|:-------------|:-------------------------------------|
| `image-name` | Full image name without tag          |
| `digest`     | Image digest (`sha256:...`)          |

This enables integration with artifact attestations for supply chain security.

!!! note
    When `BUILDKIT_HOST` is set, images are pushed during [`build`](build.md) and `push` becomes a no-op. In that case, the outputs come from the `build` step instead.

See [GitHub Actions](../ci/github.md#artifact-attestations) for a full workflow example.
