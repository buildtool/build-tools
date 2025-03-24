// MIT License
//
// Copyright (c) 2018 buildtool
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package config

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/buildtool/build-tools/pkg"
	"github.com/buildtool/build-tools/pkg/registry"
)

func TestDockerhub_Identify(t *testing.T) {
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	defer pkg.SetEnv("DOCKERHUB_USERNAME", "user")()
	defer pkg.SetEnv("DOCKERHUB_PASSWORD", "pass")()

	out := &bytes.Buffer{}
	cfg, err := Load(name)
	assert.NoError(t, err)
	registry := cfg.CurrentRegistry()
	assert.NotNil(t, registry)
	assert.Equal(t, "repo", registry.RegistryUrl())
	assert.Equal(t, "", out.String())
}

func TestDockerhub_Name(t *testing.T) {
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	defer pkg.SetEnv("DOCKERHUB_USERNAME", "user")()
	defer pkg.SetEnv("DOCKERHUB_PASSWORD", "pass")()

	cfg, err := Load(name)
	assert.NoError(t, err)
	registry := cfg.CurrentRegistry()
	assert.Equal(t, "Dockerhub", registry.Name())
}

func TestEcr_Identify(t *testing.T) {
	defer pkg.SetEnv("ECR_URL", "1234.dkr.ecr.eu-west-1.amazonaws.com")()
	defer pkg.SetEnv("ECR_REGION", "region")()

	out := &bytes.Buffer{}
	cfg, err := Load(name)
	assert.NoError(t, err)
	registry := cfg.CurrentRegistry()
	assert.NotNil(t, registry)
	assert.Equal(t, "1234.dkr.ecr.eu-west-1.amazonaws.com", registry.RegistryUrl())
	assert.Equal(t, "", out.String())
}

func TestEcr_Name(t *testing.T) {
	defer pkg.SetEnv("ECR_URL", "1234.dkr.ecr.eu-west-1.amazonaws.com")()
	defer pkg.SetEnv("ECR_REGION", "region")()

	cfg, err := Load(name)
	assert.NoError(t, err)
	registry := cfg.CurrentRegistry()
	assert.Equal(t, "ECR", registry.Name())
}

func TestEcr_Identify_MissingDockerRegistry(t *testing.T) {
	defer pkg.SetEnv("ECR_URL", "url")()
	defer pkg.SetEnv("ECR_REGION", "region")()
	defer pkg.SetEnv("AWS_CA_BUNDLE", "/missing/bundle")()

	out := &bytes.Buffer{}
	cfg, err := Load(name)
	assert.NoError(t, err)
	reg := cfg.CurrentRegistry()
	assert.Equal(t, registry.NoDockerRegistry{}, reg)
	assert.Equal(t, "", out.String())
}

func TestGitlab_Identify(t *testing.T) {
	defer pkg.SetEnv("CI_REGISTRY", "registry.gitlab.com")()
	defer pkg.SetEnv("CI_REGISTRY_USER", "gitlab-ci-token")()
	defer pkg.SetEnv("CI_REGISTRY_IMAGE", "registry.gitlab.com/group/image")()
	defer pkg.SetEnv("CI_JOB_TOKEN", "token")()

	out := &bytes.Buffer{}
	cfg, err := Load(name)
	assert.NoError(t, err)
	registry := cfg.CurrentRegistry()
	assert.Equal(t, "registry.gitlab.com/group", registry.RegistryUrl())
	assert.Equal(t, "eyJ1c2VybmFtZSI6ImdpdGxhYi1jaS10b2tlbiIsInBhc3N3b3JkIjoidG9rZW4iLCJzZXJ2ZXJhZGRyZXNzIjoicmVnaXN0cnkuZ2l0bGFiLmNvbSJ9", registry.GetAuthInfo())
	assert.Equal(t, "", out.String())
}

func TestGitlab_Name(t *testing.T) {
	defer pkg.SetEnv("CI_REGISTRY", "registry.gitlab.com")()
	defer pkg.SetEnv("CI_REGISTRY_IMAGE", "registry.gitlab.com/group/image")()
	defer pkg.SetEnv("CI_JOB_TOKEN", "token")()

	cfg, err := Load(name)
	assert.NoError(t, err)
	registry := cfg.CurrentRegistry()
	assert.Equal(t, "Gitlab", registry.Name())
}

func TestGitlab_RepositoryWithoutSlash(t *testing.T) {
	defer pkg.SetEnv("CI_REGISTRY", "registry.gitlab.com")()
	defer pkg.SetEnv("CI_REGISTRY_IMAGE", "registry.gitlab.com")()
	defer pkg.SetEnv("CI_JOB_TOKEN", "token")()

	out := &bytes.Buffer{}
	cfg, err := Load(name)
	assert.NoError(t, err)
	registry := cfg.CurrentRegistry()
	assert.NotNil(t, registry)
	assert.Equal(t, "registry.gitlab.com", registry.RegistryUrl())
	assert.Equal(t, "", out.String())
}

func TestGitlab_RegistryFallback(t *testing.T) {
	defer pkg.SetEnv("CI_REGISTRY", "registry.gitlab.com")()
	defer pkg.SetEnv("CI_REGISTRY_IMAGE", "")()
	defer pkg.SetEnv("CI_JOB_TOKEN", "token")()

	dir, _ := os.MkdirTemp("", "build-tools")
	defer func() {
		assert.NoError(t, os.RemoveAll(dir))
	}()
	oldPwd, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(oldPwd) }()

	out := &bytes.Buffer{}
	cfg, err := Load(name)
	assert.NoError(t, err)
	registry := cfg.CurrentRegistry()
	assert.NotNil(t, registry)
	assert.Equal(t, "registry.gitlab.com", registry.RegistryUrl())
	assert.Equal(t, "", out.String())
}

func TestQuay_Identify(t *testing.T) {
	defer pkg.SetEnv("QUAY_REPOSITORY", "repo")()
	defer pkg.SetEnv("QUAY_USERNAME", "user")()
	defer pkg.SetEnv("QUAY_PASSWORD", "pass")()

	out := &bytes.Buffer{}
	cfg, err := Load(name)
	assert.NoError(t, err)
	registry := cfg.CurrentRegistry()
	assert.NotNil(t, registry)
	assert.Equal(t, "quay.io/repo", registry.RegistryUrl())
	assert.Equal(t, "", out.String())
}

func TestQuay_Name(t *testing.T) {
	defer pkg.SetEnv("QUAY_REPOSITORY", "repo")()
	defer pkg.SetEnv("QUAY_USERNAME", "user")()
	defer pkg.SetEnv("QUAY_PASSWORD", "pass")()

	cfg, err := Load(name)
	assert.NoError(t, err)
	registry := cfg.CurrentRegistry()
	assert.Equal(t, "Quay.io", registry.Name())
}
