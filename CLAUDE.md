# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## General Guidelines

- Always update the documentation in `www/docs/` when making user-facing changes (new features, changed behavior, new flags, etc.)

## Project Overview

build-tools is a set of highly opinionated CLI tools for building Docker images and deploying them to Kubernetes clusters. The project provides five main commands: `build`, `push`, `deploy`, `promote`, and `kubecmd`.

## Build and Test Commands

### Building
- **Build all binaries**: The project builds 5 separate binaries from `cmd/` subdirectories (build, push, deploy, promote, kubecmd)
- **No Makefile**: This project does not use a Makefile. Build commands are managed through Go tooling and goreleaser.

### Testing
- **Run all tests**: `CGO_ENABLED=1 go test -p 1 -mod=readonly -race -coverprofile=coverage.txt -covermode=atomic -coverpkg=$(go list ./... | tr '\n' , | sed 's/,$//') ./...`
- **Run single package tests**: `go test ./pkg/config` (or any specific package path)
- **Run single test**: `go test -run TestName ./pkg/config`
- **View coverage**: `go tool cover -func=coverage.txt`

### Linting and Quality
- **Pre-commit hooks**: Run `pre-commit run --all-files` to run all pre-commit hooks
- **Individual checks**:
  - `go mod tidy` - Clean up go.mod and go.sum
  - `gofumpt -l -w .` - Format code
  - `golangci-lint run --enable-all` - Run all linters

## Code Architecture

### Command Structure
Each command in `cmd/` (build, push, deploy, promote, kubecmd) is a standalone CLI tool that:
- Uses the `github.com/apex/log` package for logging
- Imports shared functionality from `pkg/` packages
- Defines command-specific args using structs in the `pkg/args` package
- Leverages the `pkg/cli` package for output formatting

### Package Organization

**Core Packages**:
- `pkg/config` - Configuration loading and parsing. Loads `.buildtools.yaml` files hierarchically from current directory up to root. Merges multiple config files together. Contains `Config` struct that coordinates VCS, CI, and Registry configurations.
- `pkg/vcs` - Version control system abstraction (currently Git-focused)
- `pkg/ci` - CI/CD platform integrations (Azure, Buildkite, GitLab, GitHub Actions, TeamCity). Each CI provider implements the `ci.CI` interface.
- `pkg/registry` - Docker registry integrations (Dockerhub, ACR, ECR, GCR, GitHub, GitLab, Gitea, Quay). Each registry implements the `registry.Registry` interface.

**Implementation Packages**:
- `pkg/build` - Docker image building using buildkit
- `pkg/push` - Push images to registries
- `pkg/deploy` - Deploy to Kubernetes clusters
- `pkg/promote` - GitOps promotion workflow
- `pkg/kubecmd` - Kubernetes command execution helpers
- `pkg/kubectl` - Kubernetes client wrapper
- `pkg/docker` - Docker client operations
- `pkg/file` - File system utilities

**Supporting Packages**:
- `pkg/args` - Command-line argument definitions
- `pkg/cli` - CLI output formatting and logging handlers
- `pkg/version` - Version information

### Key Design Patterns

**Plugin Architecture**: The config system uses a plugin-like approach where:
- Multiple CI providers can be configured but only one is active (auto-detected)
- Multiple registries can be configured but only one is active
- Config system calls `Configured()` on each provider to determine which is active
- `AvailableCI` and `AvailableRegistries` slices hold all configured providers

**Configuration Hierarchy**: The `.buildtools.yaml` files are loaded hierarchically:
1. Start from current directory
2. Walk up to filesystem root
3. Parse each `.buildtools.yaml` found
4. Merge configurations together using `dario.cat/mergo`
5. Environment variables can override config via `BUILDTOOLS_CONTENT` (base64 or plaintext)

**Convention over Configuration**:
- Project directory name becomes the application name (unless overridden by CI env vars or `IMAGE_NAME`)
- Current Git commit ID is used as Docker tag
- Current branch is used as Docker tag
- `master`/`main` branch also gets tagged as `latest`
- Kubernetes descriptors must be in `k8s/` directory
- Dockerfile must be in project root (can be overridden with flags)

### Testing Infrastructure
- Tests use standard `testing` package with `testify/assert`
- Mock logger from `gitlab.com/unboundsoftware/apex-mocks` for testing log output
- `pkg/config/testing.go` and `pkg/testing.go` provide test helpers
- Test data stored in `pkg/build/testdata` and `pkg/file/testdata`

## Configuration

### .buildtools.yaml Structure
The configuration file defines:
- `vcs`: Version control settings
- `ci`: CI platform credentials (azure, buildkite, gitlab, github, teamcity)
- `registry`: Registry credentials (dockerhub, acr, ecr, gcr, github, gitlab, gitea, quay)
- `cache`: Layer cache configuration (ecr)
- `targets`: Kubernetes deployment targets (context, namespace, kubeconfig)
- `git`: Git user configuration for commits (used by promote command)
- `gitops`: GitOps repository configurations (url, path)

Only one CI provider and one registry should be configured at a time.

## Multi-Platform Build Support

The build system supports building Docker images for multiple platforms (architectures) simultaneously:

**Implementation Details**:
- Platform specification: Single platform (`--platform linux/amd64`) or comma-separated list (`--platform linux/amd64,linux/arm64`)
- Helper methods in `pkg/build/build.go`:
  - `isMultiPlatform()` - Detects if building for multiple platforms
  - `platformCount()` - Returns number of target platforms
- Multi-platform build behavior:
  - Buildkit receives comma-separated platform string natively
  - Images are pushed directly to registry (type: "image", push: "true")
  - Multi-platform manifest lists cannot be loaded to local Docker daemon
  - All tags are joined in a single output entry (comma-separated names)
- Single-platform builds load to local daemon (default behavior)
- Platform string is passed directly to buildkit's `ImageBuildOptions.Platform` field

**Key Code Locations**:
- Platform argument definition: `pkg/build/build.go:68`
- Multi-platform detection logic: `pkg/build/build.go:82-90`
- Registry output configuration: `pkg/build/build.go:260-272`
- Tests: `pkg/build/build_test.go:451-532`

**Architecture Note**:
- When `BUILDKIT_HOST` is set:
  - All builds (single and multi-platform) use buildkit client directly via `client.Solve()`
  - Images are pushed to registry during build
  - The `push` command becomes a no-op (images already pushed)
- When `BUILDKIT_HOST` is NOT set:
  - Single-platform builds use Docker's `/build` API with buildkit backend (loads to local daemon)
  - Multi-platform builds connect via Docker's `/grpc` endpoint (requires containerd snapshotter)
- This is necessary because Docker's `/build` API only accepts single platform values
- The `buildMultiPlatform()` function in `pkg/build/build.go` handles buildkit client builds

## ECR Layer Cache Support

The build system supports using AWS ECR as a remote layer cache for buildkit builds:

**Configuration**:
- Set via `.buildtools.yaml`:
  ```yaml
  cache:
    ecr:
      url: 123456789.dkr.ecr.us-east-1.amazonaws.com/cache-repo
      tag: buildcache  # optional, defaults to "buildcache"
  ```
- Or via environment variables: `BUILDTOOLS_CACHE_ECR_URL`, `BUILDTOOLS_CACHE_ECR_TAG`

**Implementation Details**:
- ECR cache is only used when `BUILDKIT_HOST` is set (buildkit client mode)
- Cache import: Pulls cached layers from ECR before build
- Cache export: Pushes all layers with `mode=max` (all stages)
- ECR-specific settings: `image-manifest=true`, `oci-mediatypes=true` (required for ECR compatibility)

**Key Code Locations**:
- Config structs: `pkg/config/config.go` (`CacheConfig`, `ECRCache`)
- Cache import/export: `pkg/build/build.go` (`buildCacheImports()`, `buildCacheExports()`)
- Tests: `pkg/build/build_test.go`, `pkg/config/config_test.go`

## GitHub Actions Outputs (Image Digest & Name)

Both `build` and `push` commands write `image-name` and `digest` to `$GITHUB_OUTPUT` when running in GitHub Actions, enabling artifact attestation workflows.

**How it works**:
- **Buildkit path** (`BUILDKIT_HOST` set): Digest is captured from `SolveResponse.ExporterResponse["containerimage.digest"]` during build. The `build` command writes both `image-name` and `digest` to `$GITHUB_OUTPUT`.
- **Docker API path** (default): Digest is captured from the push response `Aux.Digest` field. The `build` command writes `image-name`, the `push` command writes both `image-name` and `digest`.
- Output is only written when `GITHUB_ACTIONS=true` and `GITHUB_OUTPUT` is set.

**Usage in workflows**:
```yaml
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

**Key Code Locations**:
- GitHub output helper: `pkg/ci/github_output.go` (`WriteGitHubOutput()`)
- Digest capture (buildkit): `pkg/build/build.go` (`buildMultiPlatformWithFactory()`)
- Image name + digest (build): `pkg/build/build.go` (`build()`)
- Image name + digest (push): `pkg/push/push.go` (`doPush()`)
- Digest capture (push): `pkg/registry/registry.go` (`dockerRegistry.PushImage()`)
- Tests: `pkg/ci/github_output_test.go`, `pkg/build/build_test.go`, `pkg/registry/registry_test.go`

## Dependencies
- Go 1.25.3+
- Uses buildkit for Docker builds (via `github.com/moby/buildkit`)
- Uses containerd for container operations
- Kubernetes client libraries (k8s.io/client-go, k8s.io/kubectl)
- Cloud provider SDKs (AWS, Azure, GCP) for registry authentication
- Buildkit 0.11+ recommended for optimal multi-platform support
