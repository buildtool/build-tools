# Github Actions
Build-tools can also be used within our official build-tools actions through [GitHub Actions]

You can create a workflow by putting YAML configuration to `.github/workflows/build.yml`.

Below is a simple snippet to use the [setup-buildtools-action] in your workflow:

```yaml
name: Buildtool
on: [push]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v1
      - uses: buildtool/setup-buildtools-action@v1
      - run: build
```

Read more about available [commands](/commands/build):

> For detailed intructions please follow GitHub Actions [syntax].

## Artifact attestations

Build-tools writes the image name and digest to `$GITHUB_OUTPUT` when running in GitHub Actions. This enables integration with [artifact attestations] for supply chain security.

The following step outputs are available:

| Output       | Description                          | Example                              |
|:-------------|:-------------------------------------|:-------------------------------------|
| `image-name` | Full image name without tag          | `ghcr.io/org/my-image`               |
| `digest`     | Image digest                         | `sha256:af534ee896ce2ac80f...`        |

- **With `BUILDKIT_HOST`**: The `build` step outputs both values (images are pushed during build)
- **Without `BUILDKIT_HOST`**: The `build` step outputs `image-name`, the `push` step outputs both

```yaml
name: Build and Attest
on: [push]

jobs:
  build:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
      id-token: write
      attestations: write
    steps:
      - uses: actions/checkout@v4
      - uses: buildtool/setup-buildtools-action@v1

      - name: Build
        id: build
        run: build

      - name: Push
        id: push
        run: push

      - name: Attest
        uses: actions/attest-build-provenance@v2
        with:
          subject-name: ${{ steps.push.outputs.image-name || steps.build.outputs.image-name }}
          subject-digest: ${{ steps.push.outputs.digest || steps.build.outputs.digest }}
```

[Github Actions]: https://github.com/features/actions
[setup-buildtools-action]: https://github.com/buildtool/setup-buildtools-action
[syntax]: https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#About-yaml-syntax-for-workflows
[artifact attestations]: https://docs.github.com/en/actions/security-for-github-actions/using-artifact-attestations/using-artifact-attestations-to-establish-provenance-for-builds
