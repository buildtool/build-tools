# build

Performs a `docker build`, using a `Dockerfile` to build the application and tags the resulting image. By following the
conventions no additional flags are needed, but the following flags are available:

| Flag                                 | Description                                                                                                                                                                                                                                                 |
|:-------------------------------------|:------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `--file`,`-f` `<path to Dockerfile>` | Used to override the default `Dockerfile` location (which is `$PWD`), or `-` to read from `stdin                                                                                                                                                            |
| `--no-login`                         | Disables login to docker registry (good for local testing)                                                                                                                                                                                                  |
| `--no-pull`                          | Disables pulling of remote images if they already exist (good for local testing)                                                                                                                                                                            |
| `--build-arg key=value`              | Additional Docker [build-arg](https://docs.docker.com/engine/reference/commandline/build/#set-build-time-variables---build-arg)                                                                                                                             |
| `--platform value`                   | Specify target platform(s) for [multi-arch builds](https://docs.docker.com/desktop/multi-arch/). Single platform: `--platform linux/amd64` or multiple platforms: `--platform linux/amd64,linux/arm64`. Multi-platform builds are pushed directly to registry. |

```sh
$ build --file docker/Dockerfile.build --skip-login --build-arg AUTH_TOKEN=abc
```

## Build-args

The
following [build-arg](https://docs.docker.com/engine/reference/commandline/build/#set-build-time-variables---build-arg)
are automatically made available:

| Arg         | Value                                                  |
|:------------|:-------------------------------------------------------|
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
see [Custom build outputs](https://docs.docker.com/engine/reference/commandline/build/#custom-build-outputs). By
specifying a special stage in the `Dockerfile` and name it `export` you can use the `COPY`
directive to copy files from the build context to the local machine. The copied files will be placed in a
folder `exported`

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

## Multi-platform builds

Build-tools supports building Docker images for multiple platforms (architectures) simultaneously using buildkit's native multi-platform support. This is useful for creating images that can run on different architectures like AMD64, ARM64, ARM/v7, etc.

### Basic usage

To build for multiple platforms, provide a comma-separated list of platform identifiers:

```shell
$ build --platform linux/amd64,linux/arm64
```

Common platforms:
- `linux/amd64` - 64-bit x86 (Intel/AMD)
- `linux/arm64` - 64-bit ARM (Apple Silicon, ARM servers)
- `linux/arm/v7` - 32-bit ARM (Raspberry Pi 3+, older ARM devices)
- `linux/arm/v6` - 32-bit ARM (Raspberry Pi 1/2, older ARM devices)

### How it works

Multi-platform builds:
1. Build the image for all specified platforms in parallel using buildkit
2. Create a manifest list that references all platform-specific images
3. Push directly to the configured registry (multi-platform manifests cannot be loaded to local Docker daemon)
4. Tag all platform images with the same tags (commit, branch, latest if applicable)

**Important notes:**
- Multi-platform builds require buildkit (Docker 19.03+)
- Images are automatically pushed to the registry during the build process
- You may need QEMU for cross-platform emulation if building on a single architecture
- Multi-platform builds are typically slower than single-platform builds

### Example

```shell
# Build for AMD64 and ARM64
$ build --platform linux/amd64,linux/arm64

# The built images will be pushed to the registry with manifest list support
# Clients pulling the image will automatically get the correct architecture
```

### Requirements

**Option 1: Use a standalone BuildKit instance (recommended)**

Set the `BUILDKIT_HOST` environment variable to connect directly to a buildkit instance:

```shell
# Example: connect to buildkit running in a container
export BUILDKIT_HOST=docker-container://buildkitd

# Example: connect to buildkit via TCP
export BUILDKIT_HOST=tcp://localhost:1234

# Example: connect to buildkit via Unix socket
export BUILDKIT_HOST=unix:///run/buildkit/buildkitd.sock
```

You can run a standalone buildkit container:

```shell
docker run -d --name buildkitd --privileged moby/buildkit:latest
```

**Option 2: Enable containerd snapshotter in Docker**

If not using `BUILDKIT_HOST`, multi-platform builds require Docker to be configured with the **containerd snapshotter**. This is because Docker's default storage driver doesn't support the image exporter needed for multi-platform manifest lists.

Enable it by adding to `/etc/docker/daemon.json`:

```json
{
  "features": {
    "containerd-snapshotter": true
  }
}
```

Then restart Docker:

```shell
sudo systemctl restart docker
```

**QEMU for Cross-Platform Emulation:**

For cross-platform builds (e.g., building ARM on x86), you may need to set up QEMU:

```shell
docker run --privileged --rm tonistiigi/binfmt --install all
```

This is typically pre-configured in most CI/CD environments (GitHub Actions, GitLab CI, etc.).
