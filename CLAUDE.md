# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

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
- `pkg/registry` - Docker registry integrations (Dockerhub, ACR, ECR, GCR, GitHub, GitLab, Quay). Each registry implements the `registry.Registry` interface.

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
- `registry`: Registry credentials (dockerhub, acr, ecr, gcr, github, gitlab, quay)
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
- Single-platform builds use Docker's `/build` API with buildkit backend
- Multi-platform builds use buildkit client directly via `client.Solve()`
- Connection priority for multi-platform:
  1. If `BUILDKIT_HOST` env var is set, connect directly to that buildkit instance
  2. Otherwise, connect via Docker's `/grpc` endpoint (requires containerd snapshotter)
- This is necessary because Docker's `/build` API only accepts single platform values
- The `buildMultiPlatform()` function in `pkg/build/build.go` handles this

## Dependencies
- Go 1.25.3+
- Uses buildkit for Docker builds (via `github.com/moby/buildkit`)
- Uses containerd for container operations
- Kubernetes client libraries (k8s.io/client-go, k8s.io/kubectl)
- Cloud provider SDKs (AWS, Azure, GCP) for registry authentication
- Buildkit 0.11+ recommended for optimal multi-platform support
