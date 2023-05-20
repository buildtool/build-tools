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

package push

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/apex/log"
	mocks "gitlab.com/unboundsoftware/apex-mocks"

	"github.com/buildtool/build-tools/pkg"
	"github.com/buildtool/build-tools/pkg/config"
	"github.com/buildtool/build-tools/pkg/docker"
	"github.com/buildtool/build-tools/pkg/registry"
	"github.com/buildtool/build-tools/pkg/vcs"
	"github.com/buildtool/build-tools/pkg/version"

	types "github.com/docker/docker/api/types/registry"
	"github.com/stretchr/testify/assert"
)

var name string

func TestMain(m *testing.M) {
	tempDir := setup()
	code := m.Run()
	teardown(tempDir)
	os.Exit(code)
}

func setup() string {
	name, _ = os.MkdirTemp(os.TempDir(), "build-tools")

	return name
}

func teardown(tempDir string) {
	_ = os.RemoveAll(tempDir)
}

func TestPush_BadDockerHost(t *testing.T) {
	defer func() { _ = os.RemoveAll(name) }()

	defer pkg.SetEnv("DOCKER_HOST", "abc-123")()
	code := Push(name, version.Info{})
	assert.Equal(t, -1, code)
}

func TestPush(t *testing.T) {
	defer func() { _ = os.RemoveAll(name) }()
	code := Push(name, version.Info{})
	assert.Equal(t, -5, code)
}

func TestPush_BrokenConfig(t *testing.T) {
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `ci: [] `
	_ = write(name, ".buildtools.yaml", yaml)

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	exitCode := Push(name, version.Info{})

	assert.Equal(t, -2, exitCode)
	logMock.Check(t, []string{
		fmt.Sprintf("debug: Parsing config from file: <green>'%s'</green>\n", filepath.Join(name, ".buildtools.yaml")),
		"error: <red>yaml: unmarshal errors:\n  line 1: cannot unmarshal !!seq into config.CIConfig</red>",
	})
}

func TestPush_NoRegistry(t *testing.T) {
	defer func() { _ = os.RemoveAll(name) }()
	_ = write(name, "Dockerfile", "FROM scratch")

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	client := &docker.MockDocker{}
	cfg := config.InitEmptyConfig()
	cfg.VCS.VCS = &no{}

	exitCode := doPush(client, cfg, name, "Dockerfile")

	assert.Equal(t, -6, exitCode)
	logMock.Check(t, []string{
		"debug: Authentication <yellow>not supported</yellow> for registry <green>No docker registry</green>\n",
		"error: Commit and/or branch information is <red>missing</red>. Perhaps your not in a Git repository or forgot to set environment variables?"})
}
func TestPush_PushError(t *testing.T) {
	defer func() { _ = os.RemoveAll(name) }()
	_ = write(name, "Dockerfile", "FROM scratch")

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	client := &docker.MockDocker{PushError: fmt.Errorf("unable to push layer")}
	cfg := config.InitEmptyConfig()
	cfg.CI.Gitlab.CIBuildName = "project"
	cfg.VCS.VCS = &no{}
	cfg.Registry.Dockerhub.Namespace = "repo"

	exitCode := doPush(client, cfg, name, "Dockerfile")

	assert.NotNil(t, exitCode)
	assert.Equal(t, -6, exitCode)
	logMock.Check(t, []string{
		"debug: Logged in\n",
		"error: Commit and/or branch information is <red>missing</red>. Perhaps your not in a Git repository or forgot to set environment variables?",
	})
}

func TestPush_PushFeatureBranch(t *testing.T) {
	defer func() { _ = os.RemoveAll(name) }()
	_ = write(name, "Dockerfile", "FROM scratch")

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	pushOut := `{"status":"Push successful"}`
	client := &docker.MockDocker{PushOutput: &pushOut}
	cfg := config.InitEmptyConfig()
	cfg.CI.Gitlab.CIBuildName = "reponame"
	cfg.CI.Gitlab.CICommit = "abc123"
	cfg.CI.Gitlab.CIBranchName = "feature1"
	cfg.Registry.Dockerhub.Namespace = "repo"

	exitCode := doPush(client, cfg, name, "Dockerfile")

	assert.Equal(t, 0, exitCode)
	assert.Equal(t, []string{"repo/reponame:abc123", "repo/reponame:feature1"}, client.Images)
	logMock.Check(t, []string{
		"debug: Logged in\n",
		"info: Pushing tag '<green>repo/reponame:abc123</green>'\n",
		"info: Pushing tag '<green>repo/reponame:feature1</green>'\n"})
}

func TestPush_PushMasterBranch(t *testing.T) {
	defer func() { _ = os.RemoveAll(name) }()
	_ = write(name, "Dockerfile", "FROM scratch")

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	pushOut := `{"status":"Push successful"}`
	client := &docker.MockDocker{PushOutput: &pushOut}
	cfg := config.InitEmptyConfig()
	cfg.CI.Gitlab.CIBuildName = "reponame"
	cfg.CI.Gitlab.CICommit = "abc123"
	cfg.CI.Gitlab.CIBranchName = "master"
	cfg.Registry.Dockerhub.Namespace = "repo"
	exitCode := doPush(client, cfg, name, "Dockerfile")

	assert.Equal(t, 0, exitCode)
	assert.Equal(t, []string{
		"repo/reponame:abc123", "repo/reponame:master", "repo/reponame:latest"}, client.Images)
	logMock.Check(t, []string{"debug: Logged in\n",
		"info: Pushing tag '<green>repo/reponame:abc123</green>'\n",
		"info: Pushing tag '<green>repo/reponame:master</green>'\n",
		"info: Pushing tag '<green>repo/reponame:latest</green>'\n"})
}
func TestPush_PushMainBranch(t *testing.T) {
	defer func() { _ = os.RemoveAll(name) }()
	_ = write(name, "Dockerfile", "FROM scratch")

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	pushOut := `{"status":"Push successful"}`
	client := &docker.MockDocker{PushOutput: &pushOut}
	cfg := config.InitEmptyConfig()
	cfg.CI.Gitlab.CIBuildName = "reponame"
	cfg.CI.Gitlab.CICommit = "abc123"
	cfg.CI.Gitlab.CIBranchName = "main"
	cfg.Registry.Dockerhub.Namespace = "repo"
	exitCode := doPush(client, cfg, name, "Dockerfile")

	assert.Equal(t, 0, exitCode)
	assert.Equal(t, []string{"repo/reponame:abc123", "repo/reponame:main", "repo/reponame:latest"}, client.Images)
	logMock.Check(t, []string{"debug: Logged in\n",
		"info: Pushing tag '<green>repo/reponame:abc123</green>'\n",
		"info: Pushing tag '<green>repo/reponame:main</green>'\n",
		"info: Pushing tag '<green>repo/reponame:latest</green>'\n"})
}

func TestPush_Multistage(t *testing.T) {
	defer func() { _ = os.RemoveAll(name) }()
	dockerfile := `
FROM scratch as build
RUN echo apa > file
FROM scratch as test
RUN echo cepa > file2
FROM scratch
COPY --from=build file .
COPY --from=test file2 .
`
	_ = write(name, "Dockerfile", dockerfile)

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	pushOut := `{"status":"Push successful"}`
	client := &docker.MockDocker{PushOutput: &pushOut}
	cfg := config.InitEmptyConfig()
	cfg.CI.Gitlab.CIBuildName = "reponame"
	cfg.CI.Gitlab.CICommit = "abc123"
	cfg.CI.Gitlab.CIBranchName = "master"
	cfg.Registry.Dockerhub.Namespace = "repo"

	exitCode := doPush(client, cfg, name, "Dockerfile")

	assert.Equal(t, 0, exitCode)
	assert.Equal(t, []string{"repo/reponame:build", "repo/reponame:test", "repo/reponame:abc123", "repo/reponame:master", "repo/reponame:latest"}, client.Images)
	logMock.Check(t, []string{"debug: Logged in\n",
		"info: Pushing tag '<green>repo/reponame:build</green>'\n",
		"info: Pushing tag '<green>repo/reponame:test</green>'\n",
		"info: Pushing tag '<green>repo/reponame:abc123</green>'\n",
		"info: Pushing tag '<green>repo/reponame:master</green>'\n",
		"info: Pushing tag '<green>repo/reponame:latest</green>'\n"})
}

func TestPush_Output(t *testing.T) {
	defer func() { _ = os.RemoveAll(name) }()
	_ = write(name, "Dockerfile", "FROM scratch")

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	pushOut := `{"status":"The push refers to repository [registry.gitlab.com/project/image]"}
{"status":"Preparing","progressDetail":{},"id":"c49bda176134"}
{"status":"Preparing","progressDetail":{},"id":"cb13bd9b95b6"}
{"status":"Preparing","progressDetail":{},"id":"5905e8d02856"}
{"status":"Preparing","progressDetail":{},"id":"e3ef84c7b541"}
{"status":"Preparing","progressDetail":{},"id":"6096558c3d50"}
{"status":"Preparing","progressDetail":{},"id":"3b12aae5d4ca"}
{"status":"Preparing","progressDetail":{},"id":"ac7b6b272904"}
{"status":"Preparing","progressDetail":{},"id":"5b1304247ae3"}
{"status":"Preparing","progressDetail":{},"id":"75e70aa52609"}
{"status":"Preparing","progressDetail":{},"id":"dda151859818"}
{"status":"Preparing","progressDetail":{},"id":"fbd2732ad777"}
{"status":"Preparing","progressDetail":{},"id":"ba9de9d8475e"}
{"status":"Waiting","progressDetail":{},"id":"dda151859818"}
{"status":"Waiting","progressDetail":{},"id":"3b12aae5d4ca"}
{"status":"Waiting","progressDetail":{},"id":"ac7b6b272904"}
{"status":"Waiting","progressDetail":{},"id":"ba9de9d8475e"}
{"status":"Waiting","progressDetail":{},"id":"5b1304247ae3"}
{"status":"Waiting","progressDetail":{},"id":"75e70aa52609"}
{"status":"Waiting","progressDetail":{},"id":"fbd2732ad777"}
{"status":"Layer already exists","progressDetail":{},"id":"6096558c3d50"}
{"status":"Layer already exists","progressDetail":{},"id":"c49bda176134"}
{"status":"Layer already exists","progressDetail":{},"id":"e3ef84c7b541"}
{"status":"Pushing","progressDetail":{"current":512,"total":13088},"progress":"[=\u003e                                                 ]     512B/13.09kB","id":"cb13bd9b95b6"}
{"status":"Pushing","progressDetail":{"current":16896,"total":13088},"progress":"[==================================================\u003e]   16.9kB","id":"cb13bd9b95b6"}
{"status":"Pushing","progressDetail":{"current":512,"total":3511},"progress":"[=======\u003e                                           ]     512B/3.511kB","id":"5905e8d02856"}
{"status":"Pushing","progressDetail":{"current":6144,"total":3511},"progress":"[==================================================\u003e]  6.144kB","id":"5905e8d02856"}
{"status":"Layer already exists","progressDetail":{},"id":"ac7b6b272904"}
{"status":"Layer already exists","progressDetail":{},"id":"3b12aae5d4ca"}
{"status":"Layer already exists","progressDetail":{},"id":"5b1304247ae3"}
{"status":"Layer already exists","progressDetail":{},"id":"75e70aa52609"}
{"status":"Layer already exists","progressDetail":{},"id":"dda151859818"}
{"status":"Layer already exists","progressDetail":{},"id":"fbd2732ad777"}
{"status":"Layer already exists","progressDetail":{},"id":"ba9de9d8475e"}
{"status":"Pushed","progressDetail":{},"id":"5905e8d02856"}
{"status":"Pushed","progressDetail":{},"id":"cb13bd9b95b6"}
{"status":"cd38b8b25e3e62d05589ad6b4639e2e222086604: digest: sha256:af534ee896ce2ac80f3413318329e45e3b3e74b89eb337b9364b8ac1e83498b7 size: 2828"}
{"progressDetail":{},"aux":{"Tag":"cd38b8b25e3e62d05589ad6b4639e2e222086604","Digest":"sha256:af534ee896ce2ac80f3413318329e45e3b3e74b89eb337b9364b8ac1e83498b7","Size":2828}}
`
	client := &docker.MockDocker{PushOutput: &pushOut}
	cfg := config.InitEmptyConfig()
	cfg.CI.Gitlab.CIBuildName = "reponame"
	cfg.CI.Gitlab.CICommit = "abc123"
	cfg.CI.Gitlab.CIBranchName = "master"
	cfg.Registry.Dockerhub.Namespace = "repo"

	exitCode := doPush(client, cfg, name, "Dockerfile")

	assert.Equal(t, 0, exitCode)
	assert.Equal(t, []string{"repo/reponame:abc123", "repo/reponame:master", "repo/reponame:latest"}, client.Images)
	logMock.Check(t, []string{"debug: Logged in\n",
		"info: Pushing tag '<green>repo/reponame:abc123</green>'\n",
		"info: Pushing tag '<green>repo/reponame:master</green>'\n",
		"info: Pushing tag '<green>repo/reponame:latest</green>'\n"})
}

func TestPush_BrokenOutput(t *testing.T) {
	defer func() { _ = os.RemoveAll(name) }()
	_ = write(name, "Dockerfile", "FROM scratch")

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	pushOut := `Broken output`
	client := &docker.MockDocker{PushOutput: &pushOut}
	cfg := config.InitEmptyConfig()
	cfg.CI.Gitlab.CIBuildName = "reponame"
	cfg.CI.Gitlab.CICommit = "abc123"
	cfg.CI.Gitlab.CIBranchName = "master"
	cfg.Registry.Dockerhub.Namespace = "repo"
	exitCode := doPush(client, cfg, name, "Dockerfile")

	assert.Equal(t, -7, exitCode)
	logMock.Check(t, []string{
		"debug: Logged in\n",
		"info: Pushing tag '<green>repo/reponame:abc123</green>'\n",
		"error: Unable to parse response: Broken output, Error: invalid character 'B' looking for beginning of value\n",
		"error: <red>invalid character 'B' looking for beginning of value</red>"})
}

func TestPush_ErrorDetail(t *testing.T) {
	defer func() { _ = os.RemoveAll(name) }()
	_ = write(name, "Dockerfile", "FROM scratch")

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	pushOut := `{"status":"", "errorDetail":{"message":"error details"}}`
	client := &docker.MockDocker{PushOutput: &pushOut}
	cfg := config.InitEmptyConfig()
	cfg.CI.Gitlab.CIBuildName = "reponame"
	cfg.CI.Gitlab.CICommit = "abc123"
	cfg.CI.Gitlab.CIBranchName = "master"
	cfg.Registry.Dockerhub.Namespace = "repo"
	exitCode := doPush(client, cfg, name, "Dockerfile")

	assert.Equal(t, -7, exitCode)
	logMock.Check(t, []string{
		"debug: Logged in\n",
		"info: Pushing tag '<green>repo/reponame:abc123</green>'\n",
		"error: <red>error details</red>"})
}

func TestPush_Create_Error(t *testing.T) {
	defer func() { _ = os.RemoveAll(name) }()
	_ = write(name, "Dockerfile", "FROM scratch")

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	pushOut := `Broken output`
	client := &docker.MockDocker{PushOutput: &pushOut}
	cfg := config.InitEmptyConfig()
	cfg.AvailableRegistries = []registry.Registry{&mockRegistry{}}
	cfg.CI.Gitlab.CIBuildName = "reponame"
	cfg.CI.Gitlab.CICommit = "abc123"
	cfg.CI.Gitlab.CIBranchName = "master"
	cfg.Registry.Dockerhub.Namespace = "repo"
	exitCode := doPush(client, cfg, name, "Dockerfile")

	assert.Equal(t, -4, exitCode)
	logMock.Check(t, []string{
		"error: <red>create error</red>"})
}

func TestPush_UnreadableDockerfile(t *testing.T) {
	defer func() { _ = os.RemoveAll(name) }()
	dockerfile := filepath.Join(name, "Dockerfile")
	_ = os.MkdirAll(dockerfile, 0777)

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	pushOut := `Broken output`
	client := &docker.MockDocker{PushOutput: &pushOut}
	cfg := config.InitEmptyConfig()
	cfg.CI.Gitlab.CIBuildName = "reponame"
	cfg.CI.Gitlab.CICommit = "abc123"
	cfg.CI.Gitlab.CIBranchName = "master"
	cfg.Registry.Dockerhub.Namespace = "repo"
	exitCode := doPush(client, cfg, name, "Dockerfile")

	assert.Equal(t, -5, exitCode)
	logMock.Check(t, []string{
		"debug: Logged in\n",
		fmt.Sprintf("error: <red>read %s: is a directory</red>", dockerfile)})
}

type mockRegistry struct {
}

func (m mockRegistry) Configured() bool {
	return true
}

func (m mockRegistry) Name() string {
	panic("implement me")
}

func (m mockRegistry) Login(client docker.Client) error {
	return nil
}

func (m mockRegistry) GetAuthConfig() types.AuthConfig {
	return types.AuthConfig{}
}

func (m mockRegistry) GetAuthInfo() string {
	return ""
}

func (m mockRegistry) RegistryUrl() string {
	panic("implement me")
}

func (m mockRegistry) Create(repository string) error {
	return errors.New("create error")
}

func (m mockRegistry) PushImage(client docker.Client, auth, image string) error {
	panic("implement me")
}

var _ registry.Registry = &mockRegistry{}

type no struct {
	vcs.CommonVCS
}

func (v no) Identify(dir string) bool {
	v.CurrentCommit = ""
	v.CurrentBranch = ""

	return true
}

func (v no) Name() string {
	return "none"
}

var _ vcs.VCS = &no{}

func write(dir, file, content string) error {
	if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, file)), 0777); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, file), []byte(fmt.Sprintln(strings.TrimSpace(content))), 0666)
}
