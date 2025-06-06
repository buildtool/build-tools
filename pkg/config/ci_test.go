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
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/buildtool/build-tools/pkg"
)

func TestIdentify_Azure(t *testing.T) {
	defer pkg.SetEnv("VSTS_PROCESS_LOOKUP_ID", "1")()
	defer pkg.SetEnv("BUILD_SOURCEVERSION", "abc123")()
	defer pkg.SetEnv("BUILD_REPOSITORY_NAME", "reponame")()
	defer pkg.SetEnv("BUILD_SOURCEBRANCHNAME", "feature/first test")()
	defer pkg.UnsetGithubEnvironment()()

	out := &bytes.Buffer{}
	cfg, err := Load(name)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.NotNil(t, result)
	assert.Equal(t, "abc123", result.Commit())
	assert.Equal(t, "reponame", result.BuildName())
	assert.Equal(t, "feature/first test", result.Branch())
	assert.Equal(t, "feature_first_test", result.BranchReplaceSlash())
	assert.Equal(t, "", out.String())
}

func TestName_Azure(t *testing.T) {
	defer pkg.SetEnv("VSTS_PROCESS_LOOKUP_ID", "1")()
	defer pkg.SetEnv("BUILD_SOURCEVERSION", "abc123")()
	defer pkg.SetEnv("BUILD_REPOSITORY_NAME", "reponame")()
	defer pkg.SetEnv("BUILD_SOURCEBRANCHNAME", "feature/first test")()
	defer pkg.UnsetGithubEnvironment()()

	cfg, err := Load(name)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.Equal(t, "Azure", result.Name())
}

func TestBuildName_Fallback_Azure(t *testing.T) {
	defer pkg.SetEnv("CI", "azure")()
	defer pkg.UnsetGithubEnvironment()()

	dir, _ := os.MkdirTemp("", "build-tools")
	defer func() {
		assert.NoError(t, os.RemoveAll(dir))
	}()
	oldPwd, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(oldPwd) }()

	out := &bytes.Buffer{}
	cfg, err := Load(dir)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.NotNil(t, result)
	assert.Equal(t, filepath.Base(dir), result.BuildName())
	assert.Equal(t, "", out.String())
}

func TestBranch_VCS_Fallback_Azure(t *testing.T) {
	defer pkg.SetEnv("CI", "azure")()
	defer pkg.UnsetGithubEnvironment()()

	dir, _ := os.MkdirTemp("", "build-tools")
	defer func() {
		assert.NoError(t, os.RemoveAll(dir))
	}()

	InitRepoWithCommit(dir)

	out := &bytes.Buffer{}
	cfg, err := Load(dir)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.NotNil(t, result)
	assert.Equal(t, "master", result.Branch())
	assert.Equal(t, "", out.String())
}

func TestCommit_VCS_Fallback_Azure(t *testing.T) {
	defer pkg.SetEnv("CI", "azure")()
	defer pkg.UnsetGithubEnvironment()()

	dir, _ := os.MkdirTemp("", "build-tools")
	defer func() {
		assert.NoError(t, os.RemoveAll(dir))
	}()

	hash, _ := InitRepoWithCommit(dir)

	out := &bytes.Buffer{}
	cfg, err := Load(dir)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.NotNil(t, result)
	assert.Equal(t, hash.String(), result.Commit())
	assert.Equal(t, "", out.String())
}

func TestIdentify_Buildkite(t *testing.T) {
	defer pkg.SetEnv("BUILDKITE_COMMIT", "abc123")()
	defer pkg.SetEnv("BUILDKITE_PIPELINE_SLUG", "reponame")()
	defer pkg.SetEnv("BUILDKITE_BRANCH", "feature/first test")()
	defer pkg.UnsetGithubEnvironment()()

	out := &bytes.Buffer{}
	cfg, err := Load(name)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.NotNil(t, result)
	assert.Equal(t, "abc123", result.Commit())
	assert.Equal(t, "reponame", result.BuildName())
	assert.Equal(t, "feature/first test", result.Branch())
	assert.Equal(t, "feature_first_test", result.BranchReplaceSlash())
	assert.Equal(t, "", out.String())
}

func TestBuildName_Fallback_Buildkite(t *testing.T) {
	defer pkg.SetEnv("CI", "buildkite")()
	defer pkg.SetEnv("BUILDKITE_TOKEN", "abc123")()
	defer pkg.UnsetGithubEnvironment()()

	dir, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(dir) }()
	oldPwd, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(oldPwd) }()

	out := &bytes.Buffer{}
	cfg, err := Load(dir)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.NotNil(t, result)
	assert.Equal(t, filepath.Base(dir), result.BuildName())
	assert.Equal(t, "", out.String())
}

func TestBranch_VCS_Fallback_Buildkite(t *testing.T) {
	defer pkg.SetEnv("CI", "buildkite")()
	defer pkg.SetEnv("BUILDKITE_TOKEN", "abc123")()
	defer pkg.UnsetGithubEnvironment()()

	dir, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(dir) }()

	InitRepoWithCommit(dir)

	out := &bytes.Buffer{}
	cfg, err := Load(dir)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.NotNil(t, result)
	assert.Equal(t, "master", result.Branch())
	assert.Equal(t, "", out.String())
}

func TestCommit_VCS_Fallback_Buildkite(t *testing.T) {
	defer pkg.SetEnv("CI", "buildkite")()
	defer pkg.SetEnv("BUILDKITE_TOKEN", "abc123")()
	defer pkg.UnsetGithubEnvironment()()

	dir, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(dir) }()

	hash, _ := InitRepoWithCommit(dir)

	out := &bytes.Buffer{}
	cfg, err := Load(dir)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.NotNil(t, result)
	assert.Equal(t, hash.String(), result.Commit())
	assert.Equal(t, "", out.String())
}

func TestIdentify_Gitlab(t *testing.T) {
	defer pkg.SetEnv("GITLAB_CI", "1")()
	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "feature/first test")()
	defer pkg.UnsetGithubEnvironment()()

	out := &bytes.Buffer{}
	cfg, err := Load(name)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.NotNil(t, result)
	assert.Equal(t, "abc123", result.Commit())
	assert.Equal(t, "reponame", result.BuildName())
	assert.Equal(t, "feature/first test", result.Branch())
	assert.Equal(t, "feature_first_test", result.BranchReplaceSlash())
	assert.Equal(t, "", out.String())
}

func TestName_Gitlab(t *testing.T) {
	defer pkg.SetEnv("GITLAB_CI", "1")()
	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "feature/first test")()
	defer pkg.UnsetGithubEnvironment()()

	cfg, err := Load(name)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.Equal(t, "Gitlab", result.Name())
}

func TestBuildName_Fallback_Gitlab(t *testing.T) {
	defer pkg.SetEnv("CI", "gitlab")()
	defer pkg.UnsetGithubEnvironment()()

	dir, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(dir) }()
	oldPwd, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(oldPwd) }()

	out := &bytes.Buffer{}
	cfg, err := Load(dir)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.NotNil(t, result)
	assert.Equal(t, filepath.Base(dir), result.BuildName())
	assert.Equal(t, "", out.String())
}

func TestBranch_VCS_Fallback_Gitlab(t *testing.T) {
	defer pkg.SetEnv("CI", "gitlab")()
	defer pkg.UnsetGithubEnvironment()()

	dir, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(dir) }()

	InitRepoWithCommit(dir)

	out := &bytes.Buffer{}
	cfg, err := Load(dir)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.NotNil(t, result)
	assert.Equal(t, "master", result.Branch())
	assert.Equal(t, "", out.String())
}

func TestCommit_VCS_Fallback_Gitlab(t *testing.T) {
	defer pkg.SetEnv("CI", "gitlab")()
	defer pkg.UnsetGithubEnvironment()()

	dir, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(dir) }()

	hash, _ := InitRepoWithCommit(dir)

	out := &bytes.Buffer{}
	cfg, err := Load(dir)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.NotNil(t, result)
	assert.Equal(t, hash.String(), result.Commit())
	assert.Equal(t, "", out.String())
}

func TestNoOp(t *testing.T) {
	defer pkg.SetEnv("CI", "")()
	defer pkg.UnsetGithubEnvironment()()

	// NoOp uses PWD to generate BuildName so have to switch working dir
	oldPwd, _ := os.Getwd()
	_ = os.Chdir(name)
	defer func() { _ = os.Chdir(oldPwd) }()

	InitRepoWithCommit(name)

	out := &bytes.Buffer{}
	cfg, err := Load(name)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.NotNil(t, result)
	assert.Equal(t, filepath.Base(name), result.BuildName())
	assert.Equal(t, "master", result.BranchReplaceSlash())
	assert.False(t, result.Configured())
	assert.Equal(t, "", out.String())
}

func TestName_NoOp(t *testing.T) {
	defer pkg.SetEnv("CI", "")()
	defer pkg.UnsetGithubEnvironment()()

	cfg, err := Load(name)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.Equal(t, "none", result.Name())
}

func TestBranch_VCS_Fallback_NoOp(t *testing.T) {
	defer pkg.SetEnv("CI", "")()
	defer pkg.UnsetGithubEnvironment()()

	dir, _ := os.MkdirTemp("", "build-tools")
	defer func() { _ = os.RemoveAll(dir) }()

	InitRepoWithCommit(dir)

	out := &bytes.Buffer{}
	cfg, err := Load(dir)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.NotNil(t, result)
	assert.Equal(t, "master", result.Branch())
	assert.Equal(t, "", out.String())
}

func TestCommit_VCS_Fallback_NoOp(t *testing.T) {
	defer pkg.SetEnv("CI", "")()
	defer pkg.UnsetGithubEnvironment()()
	dir, _ := os.MkdirTemp("", "build-tools")
	defer func() { _ = os.RemoveAll(dir) }()

	hash, _ := InitRepoWithCommit(dir)

	out := &bytes.Buffer{}
	cfg, err := Load(dir)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.NotNil(t, result)
	assert.Equal(t, hash.String(), result.Commit())
	assert.Equal(t, "", out.String())
}

// Pretty ugly...
func TestCIBuildNameLowerCase(t *testing.T) {
	for _, ci := range InitEmptyConfig().AvailableCI {
		r := reflect.ValueOf(ci).Elem()
		r.FieldByName("CIBuildName").Set(reflect.ValueOf("MixedCase"))
		assert.Equal(t, "mixedcase", ci.BuildName(), "CI %s does not set buildname to lowercase", ci.Name())
	}
}
