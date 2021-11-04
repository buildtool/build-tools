package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/apex/log"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	mocks "gitlab.com/unboundsoftware/apex-mocks"

	"github.com/buildtool/build-tools/pkg"
)

func TestBuild(t *testing.T) {
	logMock := mocks.New()
	handler = logMock
	log.SetLevel(log.InfoLevel)
	defer pkg.SetEnv("DOCKER_HOST", "abc-123")()
	exitFunc = func(code int) {
		assert.Equal(t, -1, code)
	}
	os.Args = []string{"build"}
	main()
	logMock.Check(t, []string{"error: unable to parse docker host `abc-123`"})
}

func TestVersion(t *testing.T) {
	logMock := mocks.New()
	handler = logMock
	log.SetLevel(log.DebugLevel)
	exitFunc = func(code int) {
		assert.Equal(t, 0, code)
	}
	os.Args = []string{"build", "--version"}
	main()

	logMock.Check(t, []string{"info: Version: dev, commit none, built at unknown\n"})
}

func TestArguments(t *testing.T) {
	logMock := mocks.New()
	handler = logMock
	log.SetLevel(log.DebugLevel)
	exitFunc = func(code int) {
		assert.Equal(t, -1, code)
	}
	os.Args = []string{"build", "--unknown"}
	main()

	logMock.Check(t, []string{
		"info: build: error: unknown flag --unknown\n",
	})
}

func TestSuccess(t *testing.T) {
	logMock := mocks.New()
	handler = logMock
	log.SetLevel(log.DebugLevel)
	exitFunc = func(code int) {
		assert.Equal(t, 0, code)
	}
	dir := t.TempDir()
	err := os.Chdir(dir)
	assert.NoError(t, err)
	repo, err := git.PlainInit(dir, false)
	assert.NoError(t, err)
	dockerFile := []byte(`
FROM scratch as build
ADD . /test
FROM scratch as export
COPY --from=build /test/testfile testfile
   `)
	err = ioutil.WriteFile(filepath.Join(dir, "testfile"), []byte("test"), 0666)
	assert.NoError(t, err)
	err = ioutil.WriteFile(filepath.Join(dir, "Dockerfile"), dockerFile, 0666)

	assert.NoError(t, err)
	tree, err := repo.Worktree()
	assert.NoError(t, err)
	_, err = tree.Add("Dockerfile")
	assert.NoError(t, err)
	commit, err := tree.Commit("initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "John Doe",
			Email: "john@example.com",
			When:  time.Now(),
		}})
	assert.NoError(t, err)
	os.Args = []string{"build"}
	main()
	base := filepath.Base(dir)
	hash := commit.String()
	logMock.Check(t, []string{
		"debug: Using CI <green>none</green>\n",
		"debug: Using registry <green>No docker registry</green>\n",
		"debug: Authenticating against registry <green>No docker registry</green>\n",
		"debug: Authentication <yellow>not supported</yellow> for registry <green>No docker registry</green>\n",
		fmt.Sprintf("debug: Using build variables commit <green>%s</green> on branch <green>master</green>\n", hash),
		fmt.Sprintf("debug: performing docker build with options (auths removed):\ntags:\n    - noregistry/%[1]s:build\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    CI_BRANCH: master\n    CI_COMMIT: %[2]s\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - noregistry/%[1]s:build\nsecurityopt: []\nextrahosts: []\ntarget: build\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n", base, hash),
		"info: Successfully tagged noregistry/001:build\n",
		fmt.Sprintf("debug: performing docker build with options (auths removed):\ntags:\n    - noregistry/%[1]s:export\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    CI_BRANCH: master\n    CI_COMMIT: %[2]s\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - noregistry/%[1]s:export\n    - noregistry/%[1]s:build\nsecurityopt: []\nextrahosts: []\ntarget: export\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs:\n    - type: local\n      attrs: {}\n\n", base, hash),
		"info: ",
		fmt.Sprintf("debug: performing docker build with options (auths removed):\ntags:\n    - noregistry/%[1]s:%[2]s\n    - noregistry/%[1]s:master\n    - noregistry/%[1]s:latest\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    CI_BRANCH: master\n    CI_COMMIT: %[2]s\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - noregistry/%[1]s:master\n    - noregistry/%[1]s:latest\n    - noregistry/001:export\n    - noregistry/001:build\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n", base, hash),
		fmt.Sprintf("info: Successfully tagged noregistry/%[1]s:%[2]s\nSuccessfully tagged noregistry/%[1]s:master\nSuccessfully tagged noregistry/%[1]s:latest\n", base, hash),
	})
}
