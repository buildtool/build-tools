package build

import (
	"encoding/json"
	"errors"
	"fmt"

	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/apex/log"
	mocks "gitlab.com/unboundsoftware/apex-mocks"

	"github.com/buildtool/build-tools/pkg"
	"github.com/buildtool/build-tools/pkg/args"
	"github.com/buildtool/build-tools/pkg/docker"
	"github.com/buildtool/build-tools/pkg/version"

	"github.com/docker/docker/pkg/archive"
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
	name, _ = ioutil.TempDir(os.TempDir(), "build-tools")
	os.Clearenv()

	return name
}

func teardown(tempDir string) {
	_ = os.RemoveAll(tempDir)
}

func TestBuild_BrokenConfig(t *testing.T) {
	yaml := `ci: [] `
	file := filepath.Join(name, ".buildtools.yaml")
	_ = ioutil.WriteFile(file, []byte(yaml), 0777)
	defer func() { _ = os.Remove(file) }()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	client := &docker.MockDocker{}
	buildContext, _ := archive.Generate("Dockerfile", "FROM scratch")
	code := build(client, name, ioutil.NopCloser(buildContext), Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		BuildArgs:  nil,
		NoLogin:    false,
		NoPull:     false,
	})

	absPath, _ := filepath.Abs(filepath.Join(name, ".buildtools.yaml"))
	assert.Equal(t, -3, code)
	logMock.Check(t, []string{fmt.Sprintf("debug: Parsing config from file: <green>'%s'</green>\n", absPath),
		"error: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!seq into config.CIConfig"})
}

func TestBuild_NoRegistry(t *testing.T) {
	filename := filepath.Join(name, "Dockerfile")
	_ = ioutil.WriteFile(filename, []byte("FROM scratch"), 0777)
	defer func() { _ = os.RemoveAll(filename) }()

	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "feature1")()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	tmpDockerClient := dockerClient
	dockerClient = func() (client docker.Client, e error) {
		return &docker.MockDocker{}, nil
	}
	defer func() { dockerClient = tmpDockerClient }()

	code := DoBuild(name, version.Info{})
	assert.Equal(t, 0, code)
	logMock.Check(t, []string{"debug: Using CI <green>Gitlab</green>\n",
		"debug: Using registry <green>No docker registry</green>\n",
		"debug: Authenticating against registry <green>No docker registry</green>\n",
		"debug: Authentication <yellow>not supported</yellow> for registry <green>No docker registry</green>\n",
		"debug: Using build variables commit <green>abc123</green> on branch <green>feature1</green>\n",
		"debug: performing docker build with options (auths removed):\ntags:\n    - noregistry/reponame:abc123\n    - noregistry/reponame:feature1\nsuppressoutput: false\nremotecontext: \"\"\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    CI_BRANCH: feature1\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - noregistry/reponame:feature1\n    - noregistry/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: \"\"\nversion: \"\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful"})
}

func TestBuild_LoginError(t *testing.T) {
	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "feature1")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	defer pkg.SetEnv("DOCKERHUB_USERNAME", "user")()
	defer pkg.SetEnv("DOCKERHUB_PASSWORD", "pass")()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	client := &docker.MockDocker{LoginError: fmt.Errorf("invalid username/password")}
	buildContext, _ := archive.Generate("Dockerfile", "FROM scratch")
	code := build(client, name, ioutil.NopCloser(buildContext), Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		BuildArgs:  nil,
		NoLogin:    false,
		NoPull:     false,
	})

	assert.Equal(t, -4, code)
	logMock.Check(t, []string{
		"debug: Using CI <green>Gitlab</green>\n",
		"debug: Using registry <green>Dockerhub</green>\n",
		"debug: Authenticating against registry <green>Dockerhub</green>\n",
		"error: Unable to login\n",
		"error: invalid username/password"})
}

func TestBuild_BuildError(t *testing.T) {
	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "feature1")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	defer pkg.SetEnv("DOCKERHUB_USERNAME", "user")()
	defer pkg.SetEnv("DOCKERHUB_PASSWORD", "pass")()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	client := &docker.MockDocker{BuildError: []error{fmt.Errorf("build error")}}
	buildContext, _ := archive.Generate("Dockerfile", "FROM scratch")
	code := build(client, name, ioutil.NopCloser(buildContext), Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		BuildArgs:  nil,
		NoLogin:    false,
		NoPull:     false,
	})

	assert.Equal(t, -7, code)
	logMock.Check(t, []string{
		"debug: Using CI <green>Gitlab</green>\n",
		"debug: Using registry <green>Dockerhub</green>\n",
		"debug: Authenticating against registry <green>Dockerhub</green>\n",
		"debug: Logged in\n",
		"debug: Using build variables commit <green>abc123</green> on branch <green>feature1</green>\n",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:abc123\n    - repo/reponame:feature1\nsuppressoutput: false\nremotecontext: \"\"\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    CI_BRANCH: feature1\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:feature1\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: \"\"\nversion: \"\"\nbuildid: \"\"\noutputs: []\n\n",
		"error: build error",
	})
}

func TestBuild_BuildResponseError(t *testing.T) {
	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "feature1")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	defer pkg.SetEnv("DOCKERHUB_USERNAME", "user")()
	defer pkg.SetEnv("DOCKERHUB_PASSWORD", "pass")()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	client := &docker.MockDocker{ResponseError: fmt.Errorf("build error")}
	buildContext, _ := archive.Generate("Dockerfile", "FROM scratch")
	code := build(client, name, ioutil.NopCloser(buildContext), Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		BuildArgs:  nil,
		NoLogin:    false,
		NoPull:     false,
	})

	assert.Equal(t, -7, code)
	logMock.Check(t, []string{
		"debug: Using CI <green>Gitlab</green>\n",
		"debug: Using registry <green>Dockerhub</green>\n",
		"debug: Authenticating against registry <green>Dockerhub</green>\n",
		"debug: Logged in\n",
		"debug: Using build variables commit <green>abc123</green> on branch <green>feature1</green>\n",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:abc123\n    - repo/reponame:feature1\nsuppressoutput: false\nremotecontext: \"\"\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    CI_BRANCH: feature1\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:feature1\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: \"\"\nversion: \"\"\nbuildid: \"\"\noutputs: []\n\n",
		"error: error Code: 123 Message: build error",
	})
}

func TestBuild_BrokenOutput(t *testing.T) {
	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "feature1")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	defer pkg.SetEnv("DOCKERHUB_USERNAME", "user")()
	defer pkg.SetEnv("DOCKERHUB_PASSWORD", "pass")()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	client := &docker.MockDocker{BrokenOutput: true}
	buildContext, _ := archive.Generate("Dockerfile", "FROM scratch")
	code := build(client, name, ioutil.NopCloser(buildContext), Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		BuildArgs:  nil,
		NoLogin:    false,
		NoPull:     false,
	})

	assert.Equal(t, -7, code)
	logMock.Check(t, []string{
		"debug: Using CI <green>Gitlab</green>\n",
		"debug: Using registry <green>Dockerhub</green>\n",
		"debug: Authenticating against registry <green>Dockerhub</green>\n",
		"debug: Logged in\n",
		"debug: Using build variables commit <green>abc123</green> on branch <green>feature1</green>\n",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:abc123\n    - repo/reponame:feature1\nsuppressoutput: false\nremotecontext: \"\"\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    CI_BRANCH: feature1\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:feature1\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: \"\"\nversion: \"\"\nbuildid: \"\"\noutputs: []\n\n",
		"error: unable to parse response: {\"code\":123,, Error: unexpected end of JSON input",
	})
}

func TestBuild_WithBuildArgs(t *testing.T) {
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "master")()
	defer pkg.SetEnv("CI_COMMIT_SHA", "sha")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	client := &docker.MockDocker{}
	buildContext, _ := archive.Generate("Dockerfile", "FROM scratch")
	code := build(client, name, ioutil.NopCloser(buildContext), Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		BuildArgs:  []string{"buildargs1=1", "buildargs2=2"},
		NoLogin:    false,
		NoPull:     false,
	})
	assert.Equal(t, 0, code)

	assert.Equal(t, 4, len(client.BuildOptions[0].BuildArgs))
	assert.Equal(t, "1", *client.BuildOptions[0].BuildArgs["buildargs1"])
	assert.Equal(t, "2", *client.BuildOptions[0].BuildArgs["buildargs2"])
}

func TestBuild_WithStrangeBuildArg(t *testing.T) {
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "master")()
	defer pkg.SetEnv("CI_COMMIT_SHA", "sha")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	defer pkg.SetEnv("buildargs4", "env-value")()
	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	client := &docker.MockDocker{}
	buildContext, _ := archive.Generate("Dockerfile", "FROM scratch")
	code := build(client, name, ioutil.NopCloser(buildContext), Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		BuildArgs:  []string{"buildargs1=1=1", "buildargs2", "buildargs3=", "buildargs4"},
		NoLogin:    false,
		NoPull:     false,
	})
	assert.Equal(t, 0, code)

	assert.Equal(t, 4, len(client.BuildOptions[0].BuildArgs))
	assert.Equal(t, "1=1", *client.BuildOptions[0].BuildArgs["buildargs1"])
	assert.Equal(t, "env-value", *client.BuildOptions[0].BuildArgs["buildargs4"])
	logMock.Check(t, []string{
		"debug: Using CI <green>Gitlab</green>\n",
		"debug: Using registry <green>Dockerhub</green>\n",
		"debug: Authenticating against registry <green>Dockerhub</green>\n",
		"debug: Logged in\n",
		"debug: Using build variables commit <green>sha</green> on branch <green>master</green>\n",
		"debug: ignoring build-arg buildargs2\n",
		"debug: ignoring build-arg buildargs3\n",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:sha\n    - repo/reponame:master\n    - repo/reponame:latest\nsuppressoutput: false\nremotecontext: \"\"\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    CI_BRANCH: master\n    CI_COMMIT: sha\n    buildargs1: 1=1\n    buildargs4: env-value\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:master\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: \"\"\nversion: \"\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful"})
}

func TestBuild_WithSkipLogin(t *testing.T) {
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "master")()
	defer pkg.SetEnv("CI_COMMIT_SHA", "sha")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	client := &docker.MockDocker{}
	buildContext, _ := archive.Generate("Dockerfile", "FROM scratch")
	code := build(client, name, ioutil.NopCloser(buildContext), Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		BuildArgs:  nil,
		NoLogin:    true,
		NoPull:     false,
	})
	assert.Equal(t, 0, code)
	logMock.Check(t, []string{
		"debug: Using CI <green>Gitlab</green>\n",
		"debug: Using registry <green>Dockerhub</green>\n",
		"debug: Login <yellow>disabled</yellow>\n",
		"debug: Using build variables commit <green>sha</green> on branch <green>master</green>\n",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:sha\n    - repo/reponame:master\n    - repo/reponame:latest\nsuppressoutput: false\nremotecontext: \"\"\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    CI_BRANCH: master\n    CI_COMMIT: sha\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:master\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: \"\"\nversion: \"\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful"})
}

func TestBuild_FeatureBranch(t *testing.T) {
	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "feature1")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	defer pkg.SetEnv("DOCKERHUB_USERNAME", "user")()
	defer pkg.SetEnv("DOCKERHUB_PASSWORD", "pass")()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	client := &docker.MockDocker{}
	buildContext, _ := archive.Generate("Dockerfile", "FROM scratch")
	code := build(client, name, ioutil.NopCloser(buildContext), Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		BuildArgs:  nil,
		NoLogin:    false,
		NoPull:     false,
	})

	assert.Equal(t, 0, code)
	assert.Equal(t, "Dockerfile", client.BuildOptions[0].Dockerfile)
	assert.Equal(t, 2, len(client.BuildOptions[0].BuildArgs))
	assert.Equal(t, "abc123", *client.BuildOptions[0].BuildArgs["CI_COMMIT"])
	assert.Equal(t, "feature1", *client.BuildOptions[0].BuildArgs["CI_BRANCH"])
	assert.Equal(t, int64(-1), client.BuildOptions[0].MemorySwap)
	assert.Equal(t, true, client.BuildOptions[0].Remove)
	assert.Equal(t, int64(256*1024*1024), client.BuildOptions[0].ShmSize)
	assert.Equal(t, []string{"repo/reponame:abc123", "repo/reponame:feature1"}, client.BuildOptions[0].Tags)
	assert.Equal(t, "Dockerfile", client.BuildOptions[0].Dockerfile)
	assert.Equal(t, int64(-1), client.BuildOptions[0].MemorySwap)
	assert.Equal(t, true, client.BuildOptions[0].Remove)
	assert.Equal(t, int64(256*1024*1024), client.BuildOptions[0].ShmSize)
	assert.Equal(t, []string{"repo/reponame:abc123", "repo/reponame:feature1"}, client.BuildOptions[0].Tags)
	logMock.Check(t, []string{"debug: Using CI <green>Gitlab</green>\n",
		"debug: Using registry <green>Dockerhub</green>\n",
		"debug: Authenticating against registry <green>Dockerhub</green>\n",
		"debug: Logged in\n",
		"debug: Using build variables commit <green>abc123</green> on branch <green>feature1</green>\n",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:abc123\n    - repo/reponame:feature1\nsuppressoutput: false\nremotecontext: \"\"\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    CI_BRANCH: feature1\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:feature1\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: \"\"\nversion: \"\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful"})
}

func TestBuild_MasterBranch(t *testing.T) {
	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "master")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	defer pkg.SetEnv("DOCKERHUB_USERNAME", "user")()
	defer pkg.SetEnv("DOCKERHUB_PASSWORD", "pass")()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	client := &docker.MockDocker{}
	buildContext, _ := archive.Generate("Dockerfile", "FROM scratch")
	code := build(client, name, ioutil.NopCloser(buildContext), Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		BuildArgs:  nil,
		NoLogin:    false,
		NoPull:     false,
	})

	assert.Equal(t, 0, code)
	assert.Equal(t, "Dockerfile", client.BuildOptions[0].Dockerfile)
	assert.Equal(t, int64(-1), client.BuildOptions[0].MemorySwap)
	assert.Equal(t, true, client.BuildOptions[0].Remove)
	assert.Equal(t, int64(256*1024*1024), client.BuildOptions[0].ShmSize)
	assert.Equal(t, []string{"repo/reponame:abc123", "repo/reponame:master", "repo/reponame:latest"}, client.BuildOptions[0].Tags)
	logMock.Check(t, []string{"debug: Using CI <green>Gitlab</green>\n",
		"debug: Using registry <green>Dockerhub</green>\n",
		"debug: Authenticating against registry <green>Dockerhub</green>\n",
		"debug: Logged in\n",
		"debug: Using build variables commit <green>abc123</green> on branch <green>master</green>\n",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:abc123\n    - repo/reponame:master\n    - repo/reponame:latest\nsuppressoutput: false\nremotecontext: \"\"\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    CI_BRANCH: master\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:master\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: \"\"\nversion: \"\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful"})
}

func TestBuild_MainBranch(t *testing.T) {
	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "main")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	defer pkg.SetEnv("DOCKERHUB_USERNAME", "user")()
	defer pkg.SetEnv("DOCKERHUB_PASSWORD", "pass")()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	client := &docker.MockDocker{}
	buildContext, _ := archive.Generate("Dockerfile", "FROM scratch")
	code := build(client, name, ioutil.NopCloser(buildContext), Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		BuildArgs:  nil,
		NoLogin:    false,
		NoPull:     false,
	})

	assert.Equal(t, 0, code)
	assert.Equal(t, "Dockerfile", client.BuildOptions[0].Dockerfile)
	assert.Equal(t, int64(-1), client.BuildOptions[0].MemorySwap)
	assert.Equal(t, true, client.BuildOptions[0].Remove)
	assert.Equal(t, int64(256*1024*1024), client.BuildOptions[0].ShmSize)
	assert.Equal(t, []string{"repo/reponame:abc123", "repo/reponame:main", "repo/reponame:latest"}, client.BuildOptions[0].Tags)
	logMock.Check(t, []string{"debug: Using CI <green>Gitlab</green>\n",
		"debug: Using registry <green>Dockerhub</green>\n",
		"debug: Authenticating against registry <green>Dockerhub</green>\n",
		"debug: Logged in\n",
		"debug: Using build variables commit <green>abc123</green> on branch <green>main</green>\n",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:abc123\n    - repo/reponame:main\n    - repo/reponame:latest\nsuppressoutput: false\nremotecontext: \"\"\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    CI_BRANCH: main\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:main\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: \"\"\nversion: \"\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful"})
}

func TestBuild_ParseError(t *testing.T) {
	response := `{"errorDetail":{"code":1,"message":"The command '/bin/sh -c yarn install  --frozen-lockfile' returned a non-zero code: 1"},"error":"The command '/bin/sh -c yarn install  --frozen-lockfile' returned a non-zero code: 1"}`
	r := &responsetype{}

	err := json.Unmarshal([]byte(response), &r)

	assert.NoError(t, err)
}

func TestBuild_BadDockerHost(t *testing.T) {
	defer pkg.SetEnv("DOCKER_HOST", "abc-123")()
	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	exitCode := DoBuild(name, version.Info{})
	assert.Equal(t, -1, exitCode)
	logMock.Check(t, []string{"error: unable to parse docker host `abc-123`"})
}

func TestBuild_UnreadableDockerignore(t *testing.T) {
	filename := filepath.Join(name, ".dockerignore")
	_ = os.Mkdir(filename, 0777)
	defer func() { _ = os.RemoveAll(filename) }()
	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	exitCode := DoBuild(name, version.Info{})
	assert.Equal(t, -2, exitCode)
	logMock.Check(t, []string{
		fmt.Sprintf("error: read %s: is a directory", filename)})
}

func TestBuild_Unreadable_Dockerfile(t *testing.T) {
	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "master")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	defer pkg.SetEnv("DOCKERHUB_USERNAME", "user")()
	defer pkg.SetEnv("DOCKERHUB_PASSWORD", "pass")()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	client := &docker.MockDocker{}

	code := build(client, name, ioutil.NopCloser(&brokenReader{}), Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		BuildArgs:  nil,
		NoLogin:    false,
		NoPull:     false,
	})

	assert.Equal(t, -5, code)
	logMock.Check(t, []string{
		"debug: Using CI <green>Gitlab</green>\n",
		"debug: Using registry <green>Dockerhub</green>\n",
		"debug: Authenticating against registry <green>Dockerhub</green>\n",
		"debug: Logged in\n",
		"error: read error"})
}

func TestBuild_HandleCaching(t *testing.T) {
	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "master")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	defer pkg.SetEnv("DOCKERHUB_USERNAME", "user")()
	defer pkg.SetEnv("DOCKERHUB_PASSWORD", "pass")()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	client := &docker.MockDocker{}
	dockerfile := `
FROM scratch as build
RUN echo apa > file
FROM scratch as test
RUN echo cepa > file2
FROM scratch
COPY --from=build file .
COPY --from=test file2 .
`

	buildContext, _ := archive.Generate("Dockerfile", dockerfile)
	code := build(client, name, ioutil.NopCloser(buildContext), Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		BuildArgs:  nil,
		NoLogin:    false,
		NoPull:     false,
	})

	assert.Equal(t, 0, code)
	assert.Equal(t, "Dockerfile", client.BuildOptions[0].Dockerfile)
	assert.Equal(t, int64(-1), client.BuildOptions[0].MemorySwap)
	assert.Equal(t, true, client.BuildOptions[0].Remove)
	assert.Equal(t, int64(256*1024*1024), client.BuildOptions[0].ShmSize)
	assert.Equal(t, 3, len(client.BuildOptions))
	assert.Equal(t, []string{"repo/reponame:build"}, client.BuildOptions[0].Tags)
	assert.Equal(t, []string{"repo/reponame:test"}, client.BuildOptions[1].Tags)
	assert.Equal(t, []string{"repo/reponame:abc123", "repo/reponame:master", "repo/reponame:latest"}, client.BuildOptions[2].Tags)
	assert.Equal(t, []string{"repo/reponame:build"}, client.BuildOptions[0].CacheFrom)
	assert.Equal(t, []string{"repo/reponame:test", "repo/reponame:build"}, client.BuildOptions[1].CacheFrom)
	assert.Equal(t, []string{"repo/reponame:master", "repo/reponame:latest", "repo/reponame:test", "repo/reponame:build"}, client.BuildOptions[2].CacheFrom)
	logMock.Check(t, []string{"debug: Using CI <green>Gitlab</green>\n",
		"debug: Using registry <green>Dockerhub</green>\n",
		"debug: Authenticating against registry <green>Dockerhub</green>\n",
		"debug: Logged in\n",
		"debug: Using build variables commit <green>abc123</green> on branch <green>master</green>\n",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:build\nsuppressoutput: false\nremotecontext: \"\"\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    CI_BRANCH: master\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:build\nsecurityopt: []\nextrahosts: []\ntarget: build\nsessionid: \"\"\nplatform: \"\"\nversion: \"\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:test\nsuppressoutput: false\nremotecontext: \"\"\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    CI_BRANCH: master\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:test\n    - repo/reponame:build\nsecurityopt: []\nextrahosts: []\ntarget: test\nsessionid: \"\"\nplatform: \"\"\nversion: \"\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:abc123\n    - repo/reponame:master\n    - repo/reponame:latest\nsuppressoutput: false\nremotecontext: \"\"\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    CI_BRANCH: master\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:master\n    - repo/reponame:latest\n    - repo/reponame:test\n    - repo/reponame:build\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: \"\"\nversion: \"\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful"})
}

func TestBuild_BrokenStage(t *testing.T) {
	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "master")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	defer pkg.SetEnv("DOCKERHUB_USERNAME", "user")()
	defer pkg.SetEnv("DOCKERHUB_PASSWORD", "pass")()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	client := &docker.MockDocker{BuildError: []error{nil, errors.New("build error")}}
	dockerfile := `
FROM scratch as build
RUN echo apa > file
FROM scratch as test
RUN echo cepa > file2
FROM scratch
COPY --from=build file .
COPY --from=test file2 .
`

	buildContext, _ := archive.Generate("Dockerfile", dockerfile)
	code := build(client, name, ioutil.NopCloser(buildContext), Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		BuildArgs:  nil,
		NoLogin:    false,
		NoPull:     false,
	})

	assert.Equal(t, -7, code)
	logMock.Check(t, []string{
		"debug: Using CI <green>Gitlab</green>\n",
		"debug: Using registry <green>Dockerhub</green>\n",
		"debug: Authenticating against registry <green>Dockerhub</green>\n",
		"debug: Logged in\n",
		"debug: Using build variables commit <green>abc123</green> on branch <green>master</green>\n",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:build\nsuppressoutput: false\nremotecontext: \"\"\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    CI_BRANCH: master\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:build\nsecurityopt: []\nextrahosts: []\ntarget: build\nsessionid: \"\"\nplatform: \"\"\nversion: \"\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:test\nsuppressoutput: false\nremotecontext: \"\"\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    CI_BRANCH: master\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:test\n    - repo/reponame:build\nsecurityopt: []\nextrahosts: []\ntarget: test\nsessionid: \"\"\nplatform: \"\"\nversion: \"\"\nbuildid: \"\"\noutputs: []\n\n",
		"error: build error"})
}

type brokenReader struct{}

func (b brokenReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

var _ io.Reader = &brokenReader{}
