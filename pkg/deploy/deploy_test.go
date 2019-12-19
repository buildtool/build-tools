package deploy

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/buildtool/build-tools/pkg/kubectl"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestDeploy_MissingDeploymentFilesDir(t *testing.T) {
	client := &kubectl.MockKubectl{
		Responses: []error{nil},
	}
	defer client.Cleanup()

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	err := Deploy(".", "abc123", "image", "20190513-17:22:36", "test", client, out, eout, "2m")

	assert.EqualError(t, err, "open k8s: no such file or directory")
	assert.Equal(t, "", out.String())
	assert.Equal(t, "", eout.String())
}

func TestDeploy_NoFiles(t *testing.T) {
	client := &kubectl.MockKubectl{
		Responses: []error{nil},
	}

	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	_ = os.Mkdir(filepath.Join(name, "k8s"), 0777)

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	err := Deploy(name, "abc123", "image", "20190513-17:22:36", "test", client, out, eout, "2m")

	assert.NoError(t, err)
	assert.Equal(t, 0, len(client.Inputs))
	assert.Equal(t, "", out.String())
	assert.Equal(t, "", eout.String())
}

func TestDeploy_NoEnvSpecificFiles(t *testing.T) {
	client := &kubectl.MockKubectl{
		Responses: []error{nil},
	}

	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	_ = os.Mkdir(filepath.Join(name, "k8s"), 0777)
	yaml := `
apiVersion: v1
kind: Namespace
metadata:
  name: dummy
`
	deployFile := filepath.Join(name, "k8s", "deploy.yaml")
	_ = ioutil.WriteFile(deployFile, []byte(yaml), 0777)

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	err := Deploy(name, "abc123", "image", "20190513-17:22:36", "test", client, out, eout, "2m")

	assert.NoError(t, err)
	assert.Equal(t, 1, len(client.Inputs))
	assert.Equal(t, yaml, client.Inputs[0])
	assert.Equal(t, "", out.String())
	assert.Equal(t, "", eout.String())
}

func TestDeploy_UnreadableFile(t *testing.T) {
	client := &kubectl.MockKubectl{
		Responses: []error{nil},
	}

	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	_ = os.MkdirAll(filepath.Join(name, "k8s", "deploy.yaml"), 0777)

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	err := Deploy(name, "abc123", "image", "20190513-17:22:36", "test", client, out, eout, "2m")

	assert.EqualError(t, err, fmt.Sprintf("read %s/k8s/deploy.yaml: is a directory", name))
	assert.Equal(t, "", out.String())
	assert.Equal(t, "", eout.String())
}

func TestDeploy_FileBrokenSymlink(t *testing.T) {
	client := &kubectl.MockKubectl{
		Responses: []error{nil},
	}

	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	_ = os.MkdirAll(filepath.Join(name, "k8s"), 0777)
	deployFile := filepath.Join(name, "k8s", "ns.yaml")
	_ = ioutil.WriteFile(deployFile, []byte("test"), 0777)
	_ = os.Symlink(deployFile, filepath.Join(name, "k8s", "deploy.yaml"))
	_ = os.Remove(deployFile)

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	err := Deploy(name, "abc123", "image", "20190513-17:22:36", "test", client, out, eout, "2m")

	assert.EqualError(t, err, fmt.Sprintf("open %s/k8s/deploy.yaml: no such file or directory", name))
	assert.Equal(t, "", out.String())
	assert.Equal(t, "", eout.String())
}

func TestDeploy_EnvSpecificFilesWithSuffix(t *testing.T) {
	client := &kubectl.MockKubectl{
		Responses: []error{nil},
	}

	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	_ = os.Mkdir(filepath.Join(name, "k8s"), 0777)
	yaml := `
apiVersion: v1
kind: Namespace
metadata:
  name: dummy
`
	deployFile := filepath.Join(name, "k8s", "ns-dummy.yaml")
	_ = ioutil.WriteFile(deployFile, []byte(yaml), 0777)

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	err := Deploy(name, "abc123", "image", "20190513-17:22:36", "dummy", client, out, eout, "2m")

	assert.NoError(t, err)
	assert.Equal(t, 1, len(client.Inputs))
	assert.Equal(t, yaml, client.Inputs[0])
	assert.Equal(t, "", out.String())
	assert.Equal(t, "", eout.String())
}

func TestDeploy_EnvSpecificFiles(t *testing.T) {
	client := &kubectl.MockKubectl{
		Responses: []error{nil},
	}

	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	_ = os.MkdirAll(filepath.Join(name, "k8s", "prod"), 0777)
	deployFile := filepath.Join(name, "k8s", "ns-dummy.yaml")
	_ = ioutil.WriteFile(deployFile, []byte("dummy yaml content"), 0777)
	_ = ioutil.WriteFile(filepath.Join(name, "k8s", "ns-prod.yaml"), []byte("prod content"), 0777)
	_ = ioutil.WriteFile(filepath.Join(name, "k8s", "other-dummy.sh"), []byte("dummy script content"), 0777)

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	err := Deploy(name, "abc123", "image", "20190513-17:22:36", "prod", client, out, eout, "2m")

	assert.NoError(t, err)
	assert.Equal(t, 1, len(client.Inputs))
	assert.Equal(t, "prod content", client.Inputs[0])
	assert.Equal(t, "", out.String())
	assert.Equal(t, "", eout.String())
}

func TestDeploy_ScriptExecution_NameSuffix(t *testing.T) {
	client := &kubectl.MockKubectl{
		Responses: []error{nil},
	}

	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	_ = os.MkdirAll(filepath.Join(name, "k8s", "prod"), 0777)
	script := `#!/usr/bin/env bash
echo "Prod-script with suffix"`
	_ = ioutil.WriteFile(filepath.Join(name, "k8s", "setup-prod.sh"), []byte(script), 0777)

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	err := Deploy(name, "abc123", "image", "20190513-17:22:36", "prod", client, out, eout, "2m")

	assert.NoError(t, err)
	assert.Equal(t, 0, len(client.Inputs))
	assert.Equal(t, "Prod-script with suffix\n", out.String())
	assert.Equal(t, "", eout.String())
}

func TestDeploy_ScriptExecution_NoSuffixInK8s(t *testing.T) {
	client := &kubectl.MockKubectl{
		Responses: []error{nil},
	}

	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	_ = os.MkdirAll(filepath.Join(name, "k8s", "prod"), 0777)
	script := `#!/usr/bin/env bash
echo "Script without suffix should not execute"`
	_ = ioutil.WriteFile(filepath.Join(name, "k8s", "setup.sh"), []byte(script), 0777)

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	err := Deploy(name, "abc123", "image", "20190513-17:22:36", "prod", client, out, eout, "2m")

	assert.NoError(t, err)
	assert.Equal(t, 0, len(client.Inputs))
	assert.Equal(t, "", out.String())
	assert.Equal(t, "", eout.String())
}

func TestDeploy_ScriptExecution_NotExecutable(t *testing.T) {
	client := &kubectl.MockKubectl{
		Responses: []error{nil},
	}

	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	_ = os.MkdirAll(filepath.Join(name, "k8s", "prod"), 0777)
	script := `#!/usr/bin/env bash
echo "Prod-script with suffix"`
	scriptName := filepath.Join(name, "k8s", "setup-prod.sh")
	_ = ioutil.WriteFile(scriptName, []byte(script), 0666)

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	err := Deploy(name, "abc123", "image", "20190513-17:22:36", "prod", client, out, eout, "2m")

	assert.EqualError(t, err, fmt.Sprintf("fork/exec %s: permission denied", scriptName))
}

func TestDeploy_ErrorFromApply(t *testing.T) {
	client := &kubectl.MockKubectl{
		Responses: []error{errors.New("apply failed")},
	}

	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	_ = os.Mkdir(filepath.Join(name, "k8s"), 0777)
	yaml := `
apiVersion: v1
kind: Namespace
metadata:
  name: dummy
`
	deployFile := filepath.Join(name, "k8s", "deploy.yaml")
	_ = ioutil.WriteFile(deployFile, []byte(yaml), 0777)

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	err := Deploy(name, "abc123", "image", "20190513-17:22:36", "test", client, out, eout, "2m")

	assert.EqualError(t, err, "apply failed")
	assert.Equal(t, "", out.String())
	assert.Equal(t, "", eout.String())
}

func TestDeploy_ReplacingCommitAndTimestamp(t *testing.T) {
	client := &kubectl.MockKubectl{
		Responses: []error{nil},
	}

	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	_ = os.Mkdir(filepath.Join(name, "k8s"), 0777)
	yaml := `
apiVersion: v1
kind: Namespace
metadata:
  name: dummy
  commit: ${COMMIT}
  timestamp: ${TIMESTAMP}
`
	deployFile := filepath.Join(name, "k8s", "deploy.yaml")
	_ = ioutil.WriteFile(deployFile, []byte(yaml), 0777)

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	err := Deploy(name, "abc123", "image", "2019-05-13T17:22:36Z01:00", "test", client, out, eout, "2m")

	assert.NoError(t, err)
	assert.Equal(t, 1, len(client.Inputs))
	expectedInput := `
apiVersion: v1
kind: Namespace
metadata:
  name: dummy
  commit: abc123
  timestamp: 2019-05-13T17:22:36Z01:00
`
	assert.Equal(t, expectedInput, client.Inputs[0])
	assert.Equal(t, "", out.String())
	assert.Equal(t, "", eout.String())
}

func TestDeploy_DeploymentExists(t *testing.T) {
	client := &kubectl.MockKubectl{
		Responses:  []error{nil},
		Deployment: true,
	}

	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	_ = os.Mkdir(filepath.Join(name, "k8s"), 0777)
	yaml := `
apiVersion: v1
kind: Namespace
metadata:
  name: dummy
`
	deployFile := filepath.Join(name, "k8s", "deploy.yaml")
	_ = ioutil.WriteFile(deployFile, []byte(yaml), 0777)

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	err := Deploy(name, "abc123", "image", "20190513-17:22:36", "test", client, out, eout, "2m")

	assert.EqualError(t, err, "failed to rollout")
	assert.Equal(t, 1, len(client.Inputs))
	assert.Equal(t, yaml, client.Inputs[0])
	assert.Equal(t, "Rollout failed. Fetching events.Deployment eventsPod events", out.String())
	assert.Equal(t, "", eout.String())
}

func TestDeploy_RolloutStatusFail(t *testing.T) {
	client := &kubectl.MockKubectl{
		Responses:  []error{nil},
		Deployment: true,
		Status:     false,
	}

	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	_ = os.Mkdir(filepath.Join(name, "k8s"), 0777)
	yaml := `
apiVersion: v1
kind: Namespace
metadata:
  name: dummy
`
	deployFile := filepath.Join(name, "k8s", "deploy.yaml")
	_ = ioutil.WriteFile(deployFile, []byte(yaml), 0777)

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	err := Deploy(name, "abc123", "image", "20190513-17:22:36", "test", client, out, eout, "2m")

	assert.EqualError(t, err, "failed to rollout")
	assert.Equal(t, 1, len(client.Inputs))
	assert.Equal(t, yaml, client.Inputs[0])
	assert.Equal(t, "Rollout failed. Fetching events.Deployment eventsPod events", out.String())
	assert.Equal(t, "", eout.String())
}
