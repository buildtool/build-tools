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

package deploy

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/apex/log"
	"github.com/stretchr/testify/assert"
	mocks "gitlab.com/unboundsoftware/apex-mocks"

	"github.com/buildtool/build-tools/pkg/args"
	"github.com/buildtool/build-tools/pkg/kubectl"
)

func TestDeploy_MissingDeploymentFilesDir(t *testing.T) {
	client := &kubectl.MockKubectl{
		Responses: []error{nil},
	}
	defer client.Cleanup()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	err := Deploy(".", "abc123", "registryUrl", "20190513-17:22:36", client, Args{
		Globals:   args.Globals{},
		Target:    "",
		Context:   "",
		Namespace: "",
		Tag:       "",
		Timeout:   "2m",
	})

	assert.EqualError(t, err, "open k8s: no such file or directory")
	logMock.Check(t, []string{})
}

func TestDeploy_NoFiles(t *testing.T) {
	client := &kubectl.MockKubectl{
		Responses: []error{nil},
	}

	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	_ = os.Mkdir(filepath.Join(name, "k8s"), 0777)

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	err := Deploy(name, "image", "registryUrl", "20190513-17:22:36", client, Args{
		Globals:   args.Globals{},
		Target:    "",
		Context:   "",
		Namespace: "",
		Tag:       "abc123",
		Timeout:   "2m",
	})

	assert.NoError(t, err)
	assert.Equal(t, 0, len(client.Inputs))
	logMock.Check(t, []string{})
}

func TestDeploy_NoEnvSpecificFiles(t *testing.T) {
	client := &kubectl.MockKubectl{
		Responses: []error{nil},
	}

	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	_ = os.Mkdir(filepath.Join(name, "k8s"), 0777)
	yaml := `
apiVersion: v1
kind: Namespace
metadata:
  name: dummy
`
	deployFile := filepath.Join(name, "k8s", "deploy.yaml")
	_ = os.WriteFile(deployFile, []byte(yaml), 0777)

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	err := Deploy(name, "image", "registryUrl", "20190513-17:22:36", client, Args{
		Globals:   args.Globals{},
		Target:    "",
		Context:   "",
		Namespace: "",
		Tag:       "abc123",
		Timeout:   "2m",
	})

	assert.NoError(t, err)
	assert.Equal(t, 1, len(client.Inputs))
	assert.Equal(t, yaml, client.Inputs[0])
	logMock.Check(t, []string{
		"debug: considering file '<yellow>deploy.yaml</yellow>' for target: <green></green>\n",
		"debug: using file '<green>deploy.yaml</green>' for target: <green></green>\n",
		"debug: trying to apply: \n---\n\napiVersion: v1\nkind: Namespace\nmetadata:\n  name: dummy\n\n---\n",
	})
}

func TestDeploy_UnreadableFile(t *testing.T) {
	client := &kubectl.MockKubectl{
		Responses: []error{nil},
	}

	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	_ = os.MkdirAll(filepath.Join(name, "k8s", "deploy.yaml"), 0777)

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	err := Deploy(name, "image", "registryUrl", "20190513-17:22:36", client, Args{
		Globals:   args.Globals{},
		Target:    "",
		Context:   "",
		Namespace: "",
		Tag:       "abc123",
		Timeout:   "2m",
	})

	assert.EqualError(t, err, fmt.Sprintf("read %s/k8s/deploy.yaml: is a directory", name))
	logMock.Check(t, []string{
		"debug: considering file '<yellow>deploy.yaml</yellow>' for target: <green></green>\n",
		"debug: using file '<green>deploy.yaml</green>' for target: <green></green>\n"})
}

func TestDeploy_FileBrokenSymlink(t *testing.T) {
	client := &kubectl.MockKubectl{
		Responses: []error{nil},
	}

	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	_ = os.MkdirAll(filepath.Join(name, "k8s"), 0777)
	deployFile := filepath.Join(name, "k8s", "ns.yaml")
	_ = os.WriteFile(deployFile, []byte("test"), 0777)
	_ = os.Symlink(deployFile, filepath.Join(name, "k8s", "deploy.yaml"))
	_ = os.Remove(deployFile)

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	err := Deploy(name, "image", "registryUrl", "20190513-17:22:36", client, Args{
		Globals:   args.Globals{},
		Target:    "",
		Context:   "",
		Namespace: "",
		Tag:       "abc123",
		Timeout:   "2m",
	})

	assert.EqualError(t, err, fmt.Sprintf("open %s/k8s/deploy.yaml: no such file or directory", name))
	logMock.Check(t, []string{
		"debug: considering file '<yellow>deploy.yaml</yellow>' for target: <green></green>\n",
		"debug: using file '<green>deploy.yaml</green>' for target: <green></green>\n",
	})

}

func TestDeploy_EnvSpecificFilesWithSuffix(t *testing.T) {
	client := &kubectl.MockKubectl{
		Responses: []error{nil},
	}

	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	_ = os.Mkdir(filepath.Join(name, "k8s"), 0777)
	yaml := `
apiVersion: v1
kind: Namespace
metadata:
  name: dummy
`
	deployFile := filepath.Join(name, "k8s", "ns-dummy.yaml")
	_ = os.WriteFile(deployFile, []byte(yaml), 0777)

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	err := Deploy(name, "image", "registryUrl", "20190513-17:22:36", client, Args{
		Globals:   args.Globals{},
		Target:    "dummy",
		Context:   "",
		Namespace: "",
		Tag:       "abc123",
		Timeout:   "2m",
	})

	assert.NoError(t, err)
	assert.Equal(t, 1, len(client.Inputs))
	assert.Equal(t, yaml, client.Inputs[0])
	logMock.Check(t, []string{
		"debug: considering file '<yellow>ns-dummy.yaml</yellow>' for target: <green>dummy</green>\n",
		"debug: using file '<green>ns-dummy.yaml</green>' for target: <green>dummy</green>\n",
		"debug: trying to apply: \n---\n\napiVersion: v1\nkind: Namespace\nmetadata:\n  name: dummy\n\n---\n",
	})
}

func TestDeploy_EnvSpecificFiles(t *testing.T) {
	client := &kubectl.MockKubectl{
		Responses: []error{nil},
	}

	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	_ = os.MkdirAll(filepath.Join(name, "k8s", "prod"), 0777)
	deployFile := filepath.Join(name, "k8s", "ns-dummy.yaml")
	_ = os.WriteFile(deployFile, []byte("dummy yaml content"), 0777)
	_ = os.WriteFile(filepath.Join(name, "k8s", "ns-prod.yaml"), []byte("prod content"), 0777)
	_ = os.WriteFile(filepath.Join(name, "k8s", "other-dummy.sh"), []byte("dummy script content"), 0777)

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	err := Deploy(name, "image", "registryUrl", "20190513-17:22:36", client, Args{
		Globals:   args.Globals{},
		Target:    "prod",
		Context:   "",
		Namespace: "",
		Tag:       "abc123",
		Timeout:   "2m",
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(client.Inputs))
	assert.Equal(t, "prod content", client.Inputs[0])
	logMock.Check(t, []string{
		"debug: considering file '<yellow>ns-dummy.yaml</yellow>' for target: <green>prod</green>\n",
		"debug: not using file '<red>ns-dummy.yaml</red>' for target: <green>prod</green>\n",
		"debug: considering file '<yellow>ns-prod.yaml</yellow>' for target: <green>prod</green>\n",
		"debug: using file '<green>ns-prod.yaml</green>' for target: <green>prod</green>\n",
		"debug: considering script '<yellow>other-dummy.sh</yellow>' for target: <green>prod</green>\n",
		"debug: not using script '<red>other-dummy.sh</red>' for target: <green>prod</green>\n",
		"debug: trying to apply: \n---\nprod content\n---\n",
	})
}

func TestDeploy_IgnoreEmptyFiles(t *testing.T) {
	client := &kubectl.MockKubectl{
		Responses: []error{nil},
	}

	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	_ = os.MkdirAll(filepath.Join(name, "k8s", "prod"), 0777)
	deployFile := filepath.Join(name, "k8s", "ns-dummy.yaml")
	_ = os.WriteFile(deployFile, []byte("dummy yaml content"), 0777)
	_ = os.WriteFile(filepath.Join(name, "k8s", "ns-prod.yaml"), []byte(""), 0777)
	_ = os.WriteFile(filepath.Join(name, "k8s", "other-dummy.sh"), []byte("dummy script content"), 0777)

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	err := Deploy(name, "image", "registryUrl", "20190513-17:22:36", client, Args{
		Globals:   args.Globals{},
		Target:    "prod",
		Context:   "",
		Namespace: "",
		Tag:       "abc123",
		Timeout:   "2m",
	})
	assert.NoError(t, err)
	assert.Equal(t, 0, len(client.Inputs))
	logMock.Check(t, []string{
		"debug: considering file '<yellow>ns-dummy.yaml</yellow>' for target: <green>prod</green>\n",
		"debug: not using file '<red>ns-dummy.yaml</red>' for target: <green>prod</green>\n",
		"debug: considering file '<yellow>ns-prod.yaml</yellow>' for target: <green>prod</green>\n",
		"debug: using file '<green>ns-prod.yaml</green>' for target: <green>prod</green>\n",
		"debug: considering script '<yellow>other-dummy.sh</yellow>' for target: <green>prod</green>\n",
		"debug: not using script '<red>other-dummy.sh</red>' for target: <green>prod</green>\n",
		"debug: ignoring empty file '<yellow>ns-prod.yaml</yellow>'\n",
	})
}

func TestDeploy_ScriptExecution_NameSuffix(t *testing.T) {
	client := &kubectl.MockKubectl{
		Responses: []error{nil},
	}

	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	_ = os.MkdirAll(filepath.Join(name, "k8s", "prod"), 0777)
	script := `#!/usr/bin/env bash
echo "Prod-script with suffix"`
	_ = os.WriteFile(filepath.Join(name, "k8s", "setup-prod.sh"), []byte(script), 0777)

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	err := Deploy(name, "image", "registryUrl", "20190513-17:22:36", client, Args{
		Globals:   args.Globals{},
		Target:    "prod",
		Context:   "",
		Namespace: "",
		Tag:       "abc123",
		Timeout:   "2m",
	})
	assert.NoError(t, err)
	assert.Equal(t, 0, len(client.Inputs))
	logMock.Check(t, []string{
		"debug: considering script '<yellow>setup-prod.sh</yellow>' for target: <green>prod</green>\n",
		"debug: using script '<green>setup-prod.sh</green>' for target: <green>prod</green>\n",
		"info: Prod-script with suffix\n",
	})
}

func TestDeploy_ScriptExecution_No_Execute_Of_Common_If_Specific_Exists(t *testing.T) {
	client := &kubectl.MockKubectl{
		Responses: []error{nil},
	}

	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	_ = os.MkdirAll(filepath.Join(name, "k8s", "prod"), 0777)
	script := `#!/usr/bin/env bash
echo "Script without suffix should not execute"`
	_ = os.WriteFile(filepath.Join(name, "k8s", "setup.sh"), []byte(script), 0777)
	script2 := `#!/usr/bin/env bash
echo "Prod-script with suffix"`
	_ = os.WriteFile(filepath.Join(name, "k8s", "setup-prod.sh"), []byte(script2), 0777)

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	err := Deploy(name, "image", "registryUrl", "20190513-17:22:36", client, Args{
		Globals:   args.Globals{},
		Target:    "prod",
		Context:   "",
		Namespace: "",
		Tag:       "abc123",
		Timeout:   "2m",
	})
	assert.NoError(t, err)
	assert.Equal(t, 0, len(client.Inputs))
	logMock.Check(t, []string{
		"debug: considering script '<yellow>setup-prod.sh</yellow>' for target: <green>prod</green>\n",
		"debug: considering script '<yellow>setup.sh</yellow>' for target: <green>prod</green>\n",
		"debug: using script '<green>setup-prod.sh</green>' for target: <green>prod</green>\n",
		"debug: not using script '<red>setup.sh</red>' for target: <green>prod</green>\n",
		"info: Prod-script with suffix\n",
	})
}

func TestDeploy_ScriptExecution_NotExecutable(t *testing.T) {
	client := &kubectl.MockKubectl{
		Responses: []error{nil},
	}

	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	_ = os.MkdirAll(filepath.Join(name, "k8s", "prod"), 0777)
	script := `#!/usr/bin/env bash
echo "Prod-script with suffix"`
	scriptName := filepath.Join(name, "k8s", "setup-prod.sh")
	_ = os.WriteFile(scriptName, []byte(script), 0666)

	err := Deploy(name, "image", "registryUrl", "20190513-17:22:36", client, Args{
		Globals:   args.Globals{},
		Target:    "prod",
		Context:   "",
		Namespace: "",
		Tag:       "abc123",
		Timeout:   "2m",
	})
	assert.EqualError(t, err, fmt.Sprintf("fork/exec %s: permission denied", scriptName))
}

func TestDeploy_ErrorFromApply(t *testing.T) {
	client := &kubectl.MockKubectl{
		Responses: []error{errors.New("apply failed")},
	}

	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	_ = os.Mkdir(filepath.Join(name, "k8s"), 0777)
	yaml := `
apiVersion: v1
kind: Namespace
metadata:
  name: dummy
`
	deployFile := filepath.Join(name, "k8s", "deploy.yaml")
	_ = os.WriteFile(deployFile, []byte(yaml), 0777)

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	err := Deploy(name, "image", "registryUrl", "20190513-17:22:36", client, Args{
		Globals:   args.Globals{},
		Target:    "",
		Context:   "",
		Namespace: "",
		Tag:       "abc123",
		Timeout:   "2m",
	})

	assert.EqualError(t, err, "apply failed")
	logMock.Check(t, []string{
		"debug: considering file '<yellow>deploy.yaml</yellow>' for target: <green></green>\n",
		"debug: using file '<green>deploy.yaml</green>' for target: <green></green>\n",
		"debug: trying to apply: \n---\n\napiVersion: v1\nkind: Namespace\nmetadata:\n  name: dummy\n\n---\n",
	})
}

func TestDeploy_ReplacingCommitAndTimestampAndImage(t *testing.T) {
	client := &kubectl.MockKubectl{
		Responses: []error{nil},
	}

	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	_ = os.Mkdir(filepath.Join(name, "k8s"), 0777)
	yaml := `
apiVersion: v1
kind: Namespace
metadata:
  name: dummy
  commit: ${COMMIT}
  image: ${IMAGE}
  timestamp: ${TIMESTAMP}
`
	deployFile := filepath.Join(name, "k8s", "deploy.yaml")
	_ = os.WriteFile(deployFile, []byte(yaml), 0777)

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	err := Deploy(name, "registryUrl", "image", "2019-05-13T17:22:36Z01:00", client, Args{
		Globals:   args.Globals{},
		Target:    "test",
		Context:   "",
		Namespace: "",
		Tag:       "abc123",
		Timeout:   "2m",
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(client.Inputs))
	expectedInput := `
apiVersion: v1
kind: Namespace
metadata:
  name: dummy
  commit: abc123
  image: registryUrl/image:abc123
  timestamp: 2019-05-13T17:22:36Z01:00
`
	assert.Equal(t, expectedInput, client.Inputs[0])
	logMock.Check(t, []string{
		"debug: considering file '<yellow>deploy.yaml</yellow>' for target: <green>test</green>\n",
		"debug: using file '<green>deploy.yaml</green>' for target: <green>test</green>\n",
		"debug: trying to apply: \n---\n\napiVersion: v1\nkind: Namespace\nmetadata:\n  name: dummy\n  commit: abc123\n  image: registryUrl/image:abc123\n  timestamp: 2019-05-13T17:22:36Z01:00\n\n---\n",
	})
}

func TestDeploy_DeploymentExists(t *testing.T) {
	client := &kubectl.MockKubectl{
		Responses:  []error{nil},
		Deployment: true,
	}

	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	_ = os.Mkdir(filepath.Join(name, "k8s"), 0777)
	yaml := `
apiVersion: v1
kind: Namespace
metadata:
  name: dummy
`
	deployFile := filepath.Join(name, "k8s", "deploy.yaml")
	_ = os.WriteFile(deployFile, []byte(yaml), 0777)

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	err := Deploy(name, "image", "registryUrl", "20190513-17:22:36", client, Args{
		Globals:   args.Globals{},
		Target:    "",
		Context:   "",
		Namespace: "",
		Tag:       "abc123",
		Timeout:   "2m",
	})

	assert.EqualError(t, err, "failed to rollout")
	assert.Equal(t, 1, len(client.Inputs))
	assert.Equal(t, yaml, client.Inputs[0])
	logMock.Check(t, []string{
		"debug: considering file '<yellow>deploy.yaml</yellow>' for target: <green></green>\n",
		"debug: using file '<green>deploy.yaml</green>' for target: <green></green>\n",
		"debug: trying to apply: \n---\n\napiVersion: v1\nkind: Namespace\nmetadata:\n  name: dummy\n\n---\n",
		"error: Rollout failed. Fetching events.\n",
		"error: Deployment events",
		"error: Pod events",
	})
}

func TestDeploy_RolloutStatusFail(t *testing.T) {
	client := &kubectl.MockKubectl{
		Responses:  []error{nil},
		Deployment: true,
		Status:     false,
	}

	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	_ = os.Mkdir(filepath.Join(name, "k8s"), 0777)
	yaml := `
apiVersion: v1
kind: Namespace
metadata:
  name: dummy
`
	deployFile := filepath.Join(name, "k8s", "deploy.yaml")
	_ = os.WriteFile(deployFile, []byte(yaml), 0777)

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	err := Deploy(name, "image", "registryUrl", "20190513-17:22:36", client, Args{
		Globals:   args.Globals{},
		Target:    "",
		Context:   "",
		Namespace: "",
		Tag:       "abc123",
		Timeout:   "2m",
	})

	assert.EqualError(t, err, "failed to rollout")
	assert.Equal(t, 1, len(client.Inputs))
	assert.Equal(t, yaml, client.Inputs[0])
	logMock.Check(t, []string{"debug: considering file '<yellow>deploy.yaml</yellow>' for target: <green></green>\n",
		"debug: using file '<green>deploy.yaml</green>' for target: <green></green>\n",
		"debug: trying to apply: \n---\n\napiVersion: v1\nkind: Namespace\nmetadata:\n  name: dummy\n\n---\n",
		"error: Rollout failed. Fetching events.\n",
		"error: Deployment events",
		"error: Pod events",
	})
}
