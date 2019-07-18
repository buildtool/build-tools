package push

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"gitlab.com/sparetimecoders/build-tools/pkg/docker"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var name string

func TestMain(m *testing.M) {
	oldPwd, tempDir := setup()
	code := m.Run()
	teardown(oldPwd, tempDir)
	os.Exit(code)
}

func setup() (string, string) {
	oldPwd, _ := os.Getwd()
	name, _ = ioutil.TempDir(os.TempDir(), "build-tools")
	_ = os.Chdir(name)

	return oldPwd, name
}

func teardown(oldPwd, tempDir string) {
	_ = os.RemoveAll(tempDir)
	_ = os.Chdir(oldPwd)
}

func TestPush_BrokenConfig(t *testing.T) {
	os.Clearenv()
	yaml := `ci: [] `
	file := filepath.Join(name, ".buildtools.yaml")
	_ = ioutil.WriteFile(file, []byte(yaml), 0777)
	defer os.Remove(file)

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	client := &docker.MockDocker{}
	err := Push(client, "Dockerfile", out, eout)

	cwd, _ := os.Getwd()
	absPath, _ := filepath.Abs(filepath.Join(cwd, ".buildtools.yaml"))
	assert.NotNil(t, err)
	assert.EqualError(t, err, "yaml: unmarshal errors:\n  line 1: cannot unmarshal !!seq into config.CIConfig")
	assert.Equal(t, fmt.Sprintf("Parsing config from file: '%s'\n", absPath), out.String())
	assert.Equal(t, "", eout.String())
}

func TestPush_NoRegistry(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("GITLAB_CI", "1")
	_ = os.Setenv("CI_COMMIT_SHA", "abc123")
	_ = os.Setenv("CI_PROJECT_NAME", "reponame")
	_ = os.Setenv("CI_COMMIT_REF_NAME", "feature1")

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	client := &docker.MockDocker{}
	err := Push(client, "Dockerfile", out, eout)

	assert.NotNil(t, err)
	assert.EqualError(t, err, "no Docker registry found")
	assert.Equal(t, "", out.String())
	assert.Equal(t, "", eout.String())
}

func TestPush_LoginFailure(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("GITLAB_CI", "1")
	_ = os.Setenv("CI_COMMIT_SHA", "abc123")
	_ = os.Setenv("CI_PROJECT_NAME", "reponame")
	_ = os.Setenv("CI_COMMIT_REF_NAME", "feature1")
	_ = os.Setenv("ECR_URL", "ecr_url")

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	client := &docker.MockDocker{}
	err := Push(client, "Dockerfile", out, eout)

	assert.NotNil(t, err)
	assert.EqualError(t, err, "MissingRegion: could not find region configuration")
	assert.Equal(t, "", out.String())
	assert.Equal(t, "", eout.String())
}

func TestPush_PushError(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("GITLAB_CI", "1")
	_ = os.Setenv("CI_COMMIT_SHA", "abc123")
	_ = os.Setenv("CI_PROJECT_NAME", "reponame")
	_ = os.Setenv("CI_COMMIT_REF_NAME", "feature1")
	_ = os.Setenv("DOCKERHUB_REPOSITORY", "repo")
	_ = os.Setenv("DOCKERHUB_USERNAME", "user")
	_ = os.Setenv("DOCKERHUB_PASSWORD", "pass")

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	client := &docker.MockDocker{PushError: fmt.Errorf("unable to push layer")}
	err := Push(client, "Dockerfile", out, eout)

	assert.NotNil(t, err)
	assert.EqualError(t, err, "unable to push layer")
	assert.Equal(t, "Logged in\n", out.String())
	assert.Equal(t, "", eout.String())
}

func TestPush_PushFeatureBranch(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("GITLAB_CI", "1")
	_ = os.Setenv("CI_COMMIT_SHA", "abc123")
	_ = os.Setenv("CI_PROJECT_NAME", "reponame")
	_ = os.Setenv("CI_COMMIT_REF_NAME", "feature1")
	_ = os.Setenv("DOCKERHUB_REPOSITORY", "repo")
	_ = os.Setenv("DOCKERHUB_USERNAME", "user")
	_ = os.Setenv("DOCKERHUB_PASSWORD", "pass")

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	pushOut := `{"status":"Push successful"}`
	client := &docker.MockDocker{PushOutput: &pushOut}
	err := Push(client, "Dockerfile", out, eout)

	assert.Nil(t, err)
	assert.Equal(t, []string{"repo/reponame:abc123", "repo/reponame:feature1"}, client.Images)
	assert.Equal(t, "Logged in\nPush successful\nPush successful\n", out.String())
	assert.Equal(t, "", eout.String())
}

func TestPush_PushMasterBranch(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("GITLAB_CI", "1")
	_ = os.Setenv("CI_COMMIT_SHA", "abc123")
	_ = os.Setenv("CI_PROJECT_NAME", "reponame")
	_ = os.Setenv("CI_COMMIT_REF_NAME", "master")
	_ = os.Setenv("DOCKERHUB_REPOSITORY", "repo")
	_ = os.Setenv("DOCKERHUB_USERNAME", "user")
	_ = os.Setenv("DOCKERHUB_PASSWORD", "pass")

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	pushOut := `{"status":"Push successful"}`
	client := &docker.MockDocker{PushOutput: &pushOut}
	err := Push(client, "Dockerfile", out, eout)

	assert.Nil(t, err)
	assert.Equal(t, []string{"repo/reponame:abc123", "repo/reponame:master", "repo/reponame:latest"}, client.Images)
	assert.Equal(t, "Logged in\nPush successful\nPush successful\nPush successful\n", out.String())
	assert.Equal(t, "", eout.String())
}

func TestPush_Output(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("GITLAB_CI", "1")
	_ = os.Setenv("CI_COMMIT_SHA", "abc123")
	_ = os.Setenv("CI_PROJECT_NAME", "reponame")
	_ = os.Setenv("CI_COMMIT_REF_NAME", "master")
	_ = os.Setenv("DOCKERHUB_REPOSITORY", "repo")
	_ = os.Setenv("DOCKERHUB_USERNAME", "user")
	_ = os.Setenv("DOCKERHUB_PASSWORD", "pass")

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
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
	err := Push(client, "Dockerfile", out, eout)

	assert.Nil(t, err)
	assert.Equal(t, []string{"repo/reponame:abc123", "repo/reponame:master", "repo/reponame:latest"}, client.Images)
	assert.Equal(t, "Logged in\nThe push refers to repository [registry.gitlab.com/project/image]\nc49bda176134: Preparing\ncb13bd9b95b6: Preparing\n5905e8d02856: Preparing\ne3ef84c7b541: Preparing\n6096558c3d50: Preparing\n3b12aae5d4ca: Preparing\nac7b6b272904: Preparing\n5b1304247ae3: Preparing\n75e70aa52609: Preparing\ndda151859818: Preparing\nfbd2732ad777: Preparing\nba9de9d8475e: Preparing\ndda151859818: Waiting\n3b12aae5d4ca: Waiting\nac7b6b272904: Waiting\nba9de9d8475e: Waiting\n5b1304247ae3: Waiting\n75e70aa52609: Waiting\nfbd2732ad777: Waiting\n6096558c3d50: Layer already exists\nc49bda176134: Layer already exists\ne3ef84c7b541: Layer already exists\ncb13bd9b95b6: Pushing [=>                                                 ]     512B/13.09kB\ncb13bd9b95b6: Pushing [==================================================>]   16.9kB\n5905e8d02856: Pushing [=======>                                           ]     512B/3.511kB\n5905e8d02856: Pushing [==================================================>]  6.144kB\nac7b6b272904: Layer already exists\n3b12aae5d4ca: Layer already exists\n5b1304247ae3: Layer already exists\n75e70aa52609: Layer already exists\ndda151859818: Layer already exists\nfbd2732ad777: Layer already exists\nba9de9d8475e: Layer already exists\n5905e8d02856: Pushed\ncb13bd9b95b6: Pushed\ncd38b8b25e3e62d05589ad6b4639e2e222086604: digest: sha256:af534ee896ce2ac80f3413318329e45e3b3e74b89eb337b9364b8ac1e83498b7 size: 2828\nThe push refers to repository [registry.gitlab.com/project/image]\nc49bda176134: Preparing\ncb13bd9b95b6: Preparing\n5905e8d02856: Preparing\ne3ef84c7b541: Preparing\n6096558c3d50: Preparing\n3b12aae5d4ca: Preparing\nac7b6b272904: Preparing\n5b1304247ae3: Preparing\n75e70aa52609: Preparing\ndda151859818: Preparing\nfbd2732ad777: Preparing\nba9de9d8475e: Preparing\ndda151859818: Waiting\n3b12aae5d4ca: Waiting\nac7b6b272904: Waiting\nba9de9d8475e: Waiting\n5b1304247ae3: Waiting\n75e70aa52609: Waiting\nfbd2732ad777: Waiting\n6096558c3d50: Layer already exists\nc49bda176134: Layer already exists\ne3ef84c7b541: Layer already exists\ncb13bd9b95b6: Pushing [=>                                                 ]     512B/13.09kB\ncb13bd9b95b6: Pushing [==================================================>]   16.9kB\n5905e8d02856: Pushing [=======>                                           ]     512B/3.511kB\n5905e8d02856: Pushing [==================================================>]  6.144kB\nac7b6b272904: Layer already exists\n3b12aae5d4ca: Layer already exists\n5b1304247ae3: Layer already exists\n75e70aa52609: Layer already exists\ndda151859818: Layer already exists\nfbd2732ad777: Layer already exists\nba9de9d8475e: Layer already exists\n5905e8d02856: Pushed\ncb13bd9b95b6: Pushed\ncd38b8b25e3e62d05589ad6b4639e2e222086604: digest: sha256:af534ee896ce2ac80f3413318329e45e3b3e74b89eb337b9364b8ac1e83498b7 size: 2828\nThe push refers to repository [registry.gitlab.com/project/image]\nc49bda176134: Preparing\ncb13bd9b95b6: Preparing\n5905e8d02856: Preparing\ne3ef84c7b541: Preparing\n6096558c3d50: Preparing\n3b12aae5d4ca: Preparing\nac7b6b272904: Preparing\n5b1304247ae3: Preparing\n75e70aa52609: Preparing\ndda151859818: Preparing\nfbd2732ad777: Preparing\nba9de9d8475e: Preparing\ndda151859818: Waiting\n3b12aae5d4ca: Waiting\nac7b6b272904: Waiting\nba9de9d8475e: Waiting\n5b1304247ae3: Waiting\n75e70aa52609: Waiting\nfbd2732ad777: Waiting\n6096558c3d50: Layer already exists\nc49bda176134: Layer already exists\ne3ef84c7b541: Layer already exists\ncb13bd9b95b6: Pushing [=>                                                 ]     512B/13.09kB\ncb13bd9b95b6: Pushing [==================================================>]   16.9kB\n5905e8d02856: Pushing [=======>                                           ]     512B/3.511kB\n5905e8d02856: Pushing [==================================================>]  6.144kB\nac7b6b272904: Layer already exists\n3b12aae5d4ca: Layer already exists\n5b1304247ae3: Layer already exists\n75e70aa52609: Layer already exists\ndda151859818: Layer already exists\nfbd2732ad777: Layer already exists\nba9de9d8475e: Layer already exists\n5905e8d02856: Pushed\ncb13bd9b95b6: Pushed\ncd38b8b25e3e62d05589ad6b4639e2e222086604: digest: sha256:af534ee896ce2ac80f3413318329e45e3b3e74b89eb337b9364b8ac1e83498b7 size: 2828\n", out.String())
	assert.Equal(t, "", eout.String())
}

func TestPush_BrokenOutput(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("GITLAB_CI", "1")
	_ = os.Setenv("CI_COMMIT_SHA", "abc123")
	_ = os.Setenv("CI_PROJECT_NAME", "reponame")
	_ = os.Setenv("CI_COMMIT_REF_NAME", "master")
	_ = os.Setenv("DOCKERHUB_REPOSITORY", "repo")
	_ = os.Setenv("DOCKERHUB_USERNAME", "user")
	_ = os.Setenv("DOCKERHUB_PASSWORD", "pass")

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	pushOut := `Broken output`
	client := &docker.MockDocker{PushOutput: &pushOut}
	err := Push(client, "Dockerfile", out, eout)

	assert.EqualError(t, err, "invalid character 'B' looking for beginning of value")
}
