package kubectl

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"k8s.io/kubectl/pkg/cmd/util"

	"github.com/buildtool/build-tools/pkg"
	"github.com/buildtool/build-tools/pkg/config"
)

func TestNew(t *testing.T) {
	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	k := New(&config.Target{Context: "missing", Namespace: "dev"}, out, eout)

	assert.Equal(t, "missing", k.(*kubectl).args["context"])
	assert.Equal(t, "dev", k.(*kubectl).args["namespace"])
	assert.Equal(t, "", out.String())
	assert.Equal(t, "", eout.String())
}

func TestNew_NoNamespace(t *testing.T) {
	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	calls = [][]string{}
	newKubectlCmd = mockCmd
	tempDir, _ := ioutil.TempDir(os.TempDir(), "build-tools")

	k := &kubectl{args: map[string]string{"context": "missing"}, tempDir: tempDir, out: out}

	err := k.Apply("")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(calls))
	assert.Equal(t, []string{"apply", "--context", "missing", "--file", fmt.Sprintf("%s/content.yaml", tempDir)}, calls[0])
	assert.Equal(t, "", out.String())
	assert.Equal(t, "", eout.String())
}
func TestNew_NoContext(t *testing.T) {
	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	calls = [][]string{}
	newKubectlCmd = mockCmd
	tempDir, _ := ioutil.TempDir(os.TempDir(), "build-tools")

	k := &kubectl{args: map[string]string{"namespace": "namespace"}, tempDir: tempDir, out: out}

	err := k.Apply("")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(calls))
	assert.Equal(t, []string{"apply", "--namespace", "namespace", "--file", fmt.Sprintf("%s/content.yaml", tempDir)}, calls[0])
	assert.Equal(t, "", out.String())
	assert.Equal(t, "", eout.String())
}

func TestKubectl_Apply(t *testing.T) {
	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	calls = [][]string{}
	newKubectlCmd = mockCmd
	tempDir, _ := ioutil.TempDir(os.TempDir(), "build-tools")

	k := &kubectl{args: map[string]string{"context": "missing", "namespace": "default"}, tempDir: tempDir, out: out}

	err := k.Apply("")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(calls))
	assert.Equal(t, []string{"apply", "--context", "missing", "--namespace", "default", "--file", fmt.Sprintf("%s/content.yaml", tempDir)}, calls[0])
	assert.Equal(t, "", out.String())
	assert.Equal(t, "", eout.String())
}

func TestKubectl_UnableToCreateTempDir(t *testing.T) {
	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	newKubectlCmd = mockCmd

	k := &kubectl{args: nil, tempDir: "/missing", out: out}

	err := k.Apply("")
	assert.EqualError(t, err, "open /missing/content.yaml: no such file or directory")
	assert.Equal(t, "", out.String())
	assert.Equal(t, "", eout.String())
}

func TestKubectl_Target(t *testing.T) {
	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	env := &config.Target{Context: "missing", Namespace: ""}
	k := New(env, out, eout)

	assert.Equal(t, "", k.(*kubectl).args["namespace"])
	assert.Equal(t, "", out.String())
	assert.Equal(t, "", eout.String())
}

func TestKubectl_DeploymentExistsTrue(t *testing.T) {
	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	calls = [][]string{}
	cmdError = nil
	o := `NAME          READY   UP-TO-DATE   AVAILABLE   AGE
api           1/1     1            1           2d11h
`
	cmdOut = &o
	newKubectlCmd = mockCmd

	k := New(&config.Target{Context: "missing", Namespace: "default"}, out, eout)

	result := k.DeploymentExists("image")
	assert.True(t, result)
	assert.Equal(t, 1, len(calls))
	assert.Equal(t, []string{"get", "deployment", "image", "--context", "missing", "--namespace", "default", "--ignore-not-found"}, calls[0])
	assert.Equal(t, "kubectl --context missing --namespace default get deployment image --ignore-not-found\n", out.String())
	assert.Equal(t, "", eout.String())
}

func TestKubectl_DeploymentExistsFalse(t *testing.T) {
	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	calls = [][]string{}
	e := "deployment not found"
	cmdError = &e
	newKubectlCmd = mockCmd

	k := New(&config.Target{Context: "missing", Namespace: "default"}, out, eout)

	result := k.DeploymentExists("image")
	assert.False(t, result)
	assert.Equal(t, 1, len(calls))
	assert.Equal(t, []string{"get", "deployment", "image", "--context", "missing", "--namespace", "default", "--ignore-not-found"}, calls[0])
	assert.Equal(t, "kubectl --context missing --namespace default get deployment image --ignore-not-found\n", out.String())
	assert.Equal(t, "", eout.String())
}

func TestKubectl_RolloutStatusSuccess(t *testing.T) {
	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	calls = [][]string{}
	cmdOut = nil
	cmdError = nil
	newKubectlCmd = mockCmd

	k := New(&config.Target{Context: "missing", Namespace: "other"}, out, eout)

	result := k.RolloutStatus("image", "2m")
	assert.True(t, result)
	assert.Equal(t, 1, len(calls))
	assert.Equal(t, []string{"rollout", "status", "deployment", "image", "--context", "missing", "--namespace", "other", "--timeout", "2m0s"}, calls[0])
	assert.Equal(t, "kubectl --context missing --namespace other rollout status deployment --timeout=2m image\n", out.String())
	assert.Equal(t, "", eout.String())
}

func TestKubectl_RolloutStatusFailure(t *testing.T) {
	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	calls = [][]string{}
	e := "rollout failed"
	cmdError = &e
	newKubectlCmd = mockCmd

	k := New(&config.Target{Context: "missing", Namespace: "default"}, out, eout)

	result := k.RolloutStatus("image", "2m")
	assert.False(t, result)
	assert.Equal(t, 1, len(calls))
	assert.Equal(t, []string{"rollout", "status", "deployment", "image", "--context", "missing", "--namespace", "default", "--timeout", "2m0s"}, calls[0])
	assert.Equal(t, "kubectl --context missing --namespace default rollout status deployment --timeout=2m image\n", out.String())
	assert.Equal(t, "", eout.String())
}

func TestKubectl_RolloutStatusFatal(t *testing.T) {
	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	calls = [][]string{}
	e := "rollout failed"
	cmdError = &e
	fatal = true
	defer func() { fatal = false }()

	newKubectlCmd = mockCmd

	k := New(&config.Target{Context: "missing", Namespace: "default"}, out, eout)

	result := k.RolloutStatus("image", "3m")
	assert.False(t, result)
	assert.Equal(t, 1, len(calls))
	assert.Equal(t, []string{"rollout", "status", "deployment", "image", "--context", "missing", "--namespace", "default", "--timeout", "3m0s"}, calls[0])
	assert.Equal(t, "kubectl --context missing --namespace default rollout status deployment --timeout=3m image\n", out.String())
	assert.Equal(t, "", eout.String())
}

func TestKubectl_KubeconfigSet(t *testing.T) {
	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	yaml := `contexts:
- context:
    cluster: k8s.prod
    user: user@example.org
`
	defer pkg.SetEnv(envKubeconfigContent, yaml)()
	k := New(&config.Target{}, out, eout)

	kubeconfigFile := filepath.Join(k.(*kubectl).tempDir, "kubeconfig")
	fileContent, err := ioutil.ReadFile(kubeconfigFile)
	assert.NoError(t, err)
	assert.Equal(t, "contexts:\n- context:\n    cluster: k8s.prod\n    user: user@example.org\n", string(fileContent))
	assert.Equal(t, kubeconfigFile, k.(*kubectl).args["kubeconfig"])
	k.Cleanup()
}

func TestKubectl_KubeconfigSetToEmptyValue(t *testing.T) {
	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	yaml := ``
	defer pkg.SetEnv(envKubeconfigContent, yaml)()
	k := New(&config.Target{}, out, eout)

	assert.Equal(t, "", k.(*kubectl).args["kubeconfig"])
	k.Cleanup()
}

func TestKubectl_KubeconfigBase64Set(t *testing.T) {
	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	yaml := `contexts:
- context:
    cluster: k8s.prod
    user: user@example.org
`
	defer pkg.SetEnv(envKubeconfigContent, base64.StdEncoding.EncodeToString([]byte(yaml)))()
	k := New(&config.Target{}, out, eout)

	kubeconfigFile := filepath.Join(k.(*kubectl).tempDir, "kubeconfig")
	fileContent, err := ioutil.ReadFile(kubeconfigFile)
	assert.NoError(t, err)
	assert.Equal(t, "contexts:\n- context:\n    cluster: k8s.prod\n    user: user@example.org\n", string(fileContent))
	assert.Equal(t, kubeconfigFile, k.(*kubectl).args["kubeconfig"])
	k.Cleanup()
}

func TestKubectl_KubeconfigExistingFile(t *testing.T) {

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	name, _ := ioutil.TempFile(os.TempDir(), "kubecontent")
	defer func() {
		_ = os.Remove(name.Name())
	}()

	k := New(&config.Target{Kubeconfig: name.Name()}, out, eout)
	assert.Equal(t, name.Name(), k.(*kubectl).args["kubeconfig"])
	k.Cleanup()
}

func TestKubectl_DeploymentEvents_Error(t *testing.T) {
	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	calls = [][]string{}
	newKubectlCmd = mockCmd
	e := "deployment not found"
	cmdError = &e

	k := New(&config.Target{Context: "missing", Namespace: "default"}, out, eout)

	result := k.DeploymentEvents("image")
	assert.Equal(t, "deployment not found", result)
	assert.Equal(t, 1, len(calls))
	assert.Equal(t, []string{"describe", "deployment", "image", "--context", "missing", "--namespace", "default", "--show-events", "true"}, calls[0])
	assert.Equal(t, "kubectl --context missing --namespace default describe deployment image --show-events=true\n", out.String())
	assert.Equal(t, "", eout.String())
}

func TestKubectl_DeploymentEvents_NoEvents(t *testing.T) {
	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	calls = [][]string{}
	cmdError = nil
	newKubectlCmd = mockCmd
	e := `
Name:               gpe-core
Namespace:          default
Events:          <none>
`
	cmdOut = &e

	k := New(&config.Target{Context: "missing", Namespace: "default"}, out, eout)

	result := k.DeploymentEvents("image")
	assert.Equal(t, "", result)
	assert.Equal(t, 1, len(calls))
	assert.Equal(t, []string{"describe", "deployment", "image", "--context", "missing", "--namespace", "default", "--show-events", "true"}, calls[0])
	assert.Equal(t, "kubectl --context missing --namespace default describe deployment image --show-events=true\n", out.String())
	assert.Equal(t, "", eout.String())
}

func TestKubectl_DeploymentEvents_SomeEvents(t *testing.T) {
	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	calls = [][]string{}
	cmdError = nil
	newKubectlCmd = mockCmd
	e := `
Name:               gpe-core
Namespace:          default
Events:
  Type    Reason             Age   From                   Message
  ----    ------             ----  ----                   -------
  Normal  ScalingReplicaSet  9m    deployment-controller  Scaled up replica set gpe-core-5cb459ff7d to 1
  Normal  ScalingReplicaSet  9m    deployment-controller  Scaled down replica set gpe-core-7fc44679dc to 0
  Normal  ScalingReplicaSet  61s   deployment-controller  Scaled up replica set gpe-core-c8798ff88 to 1
  Normal  ScalingReplicaSet  61s   deployment-controller  Scaled down replica set gpe-core-5cb459ff7d to 0
`
	cmdOut = &e

	k := New(&config.Target{Context: "missing", Namespace: "default"}, out, eout)

	result := k.DeploymentEvents("image")
	assert.Equal(t, "Events:\n  Type    Reason             Age   From                   Message\n  ----    ------             ----  ----                   -------\n  Normal  ScalingReplicaSet  9m    deployment-controller  Scaled up replica set gpe-core-5cb459ff7d to 1\n  Normal  ScalingReplicaSet  9m    deployment-controller  Scaled down replica set gpe-core-7fc44679dc to 0\n  Normal  ScalingReplicaSet  61s   deployment-controller  Scaled up replica set gpe-core-c8798ff88 to 1\n  Normal  ScalingReplicaSet  61s   deployment-controller  Scaled down replica set gpe-core-5cb459ff7d to 0\n", result)
	assert.Equal(t, 1, len(calls))
	assert.Equal(t, []string{"describe", "deployment", "image", "--context", "missing", "--namespace", "default", "--show-events", "true"}, calls[0])
	assert.Equal(t, "kubectl --context missing --namespace default describe deployment image --show-events=true\n", out.String())
	assert.Equal(t, "", eout.String())
}

func TestKubectl_PodEvents_Error(t *testing.T) {
	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	calls = [][]string{}
	newKubectlCmd = mockCmd
	e := "pod not found"
	cmdError = &e

	k := New(&config.Target{Context: "missing", Namespace: "default"}, out, eout)

	result := k.PodEvents("image")
	assert.Equal(t, "pod not found", result)
	assert.Equal(t, 1, len(calls))
	assert.Equal(t, []string{"describe", "pods", "--context", "missing", "--namespace", "default", "--show-events", "true", "--selector", "app=image"}, calls[0])
	assert.Equal(t, "kubectl --context missing --namespace default describe pods -l app=image --show-events=true\n", out.String())
	assert.Equal(t, "", eout.String())
}

func TestKubectl_PodEvents_NoEvents(t *testing.T) {
	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	calls = [][]string{}
	cmdError = nil
	newKubectlCmd = mockCmd
	e := `
Name:               gpe-core
Namespace:          default
Events:          <none>
`
	cmdOut = &e

	k := New(&config.Target{Context: "missing", Namespace: "default"}, out, eout)

	result := k.PodEvents("image")
	assert.Equal(t, "", result)
	assert.Equal(t, 1, len(calls))
	assert.Equal(t, []string{"describe", "pods", "--context", "missing", "--namespace", "default", "--show-events", "true", "--selector", "app=image"}, calls[0])
	assert.Equal(t, "kubectl --context missing --namespace default describe pods -l app=image --show-events=true\n", out.String())
	assert.Equal(t, "", eout.String())
}

func TestKubectl_PodEvents_SomeEvents(t *testing.T) {
	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	calls = [][]string{}
	cmdError = nil
	newKubectlCmd = mockCmd
	e := `
Events:
  Type     Reason     Age                From                                                 Message
  ----     ------     ----               ----                                                 -------
  Normal   Scheduled  61s                default-scheduler                                    Successfully assigned dev/gpe-core-c8798ff88-674tr to some-ip-somewhere
  Normal   Pulling    10s (x4 over 60s)  kubelet, some-ip-somewhere                           pulling image "quay.io/somewhere/gpe-core:9cdb0243e82b9bfdf037627d9d59cbfcbf55406c"
  Normal   Pulled     9s (x4 over 57s)   kubelet, some-ip-somewhere                           Successfully pulled image "quay.io/somewhere/gpe-core:9cdb0243e82b9bfdf037627d9d59cbfcbf55406c"
  Normal   Created    8s (x4 over 57s)   kubelet, some-ip-somewhere                           Created container
  Normal   Started    8s (x4 over 57s)   kubelet, some-ip-somewhere                           Started container
  Warning  BackOff    8s (x5 over 54s)   kubelet, some-ip-somewhere                           Back-off restarting failed container`
	cmdOut = &e

	k := New(&config.Target{Context: "missing", Namespace: "default"}, out, eout)

	result := k.PodEvents("image")
	assert.Equal(t, "Events:\n  Type     Reason     Age                From                                                 Message\n  ----     ------     ----               ----                                                 -------\n  Normal   Scheduled  61s                default-scheduler                                    Successfully assigned dev/gpe-core-c8798ff88-674tr to some-ip-somewhere\n  Normal   Pulling    10s (x4 over 60s)  kubelet, some-ip-somewhere                           pulling image \"quay.io/somewhere/gpe-core:9cdb0243e82b9bfdf037627d9d59cbfcbf55406c\"\n  Normal   Pulled     9s (x4 over 57s)   kubelet, some-ip-somewhere                           Successfully pulled image \"quay.io/somewhere/gpe-core:9cdb0243e82b9bfdf037627d9d59cbfcbf55406c\"\n  Normal   Created    8s (x4 over 57s)   kubelet, some-ip-somewhere                           Created container\n  Normal   Started    8s (x4 over 57s)   kubelet, some-ip-somewhere                           Started container\n  Warning  BackOff    8s (x5 over 54s)   kubelet, some-ip-somewhere                           Back-off restarting failed container\n", result)
	assert.Equal(t, 1, len(calls))
	assert.Equal(t, []string{"describe", "pods", "--context", "missing", "--namespace", "default", "--show-events", "true", "--selector", "app=image"}, calls[0])
	assert.Equal(t, "kubectl --context missing --namespace default describe pods -l app=image --show-events=true\n", out.String())
	assert.Equal(t, "", eout.String())
}

var calls [][]string
var cmdError *string
var cmdOut *string
var fatal = false

func mockCmd(_ io.Reader, out, _ io.Writer) *cobra.Command {
	var ctx, ns, file *string
	var timeout *time.Duration
	var showEvents *bool
	var ignoreNotFound *bool
	var selector *string
	var kubeconfig *string

	cmd := cobra.Command{
		Use: "kubectl",
		Args: func(cmd *cobra.Command, args []string) error {
			var call = args
			if *ctx != "" {
				call = append(call, "--context", *ctx)
			}
			if *kubeconfig != "" {
				call = append(call, "--kubeconfig", *kubeconfig)
			}
			if *ns != "" {
				call = append(call, "--namespace", *ns)
			}
			if *file != "" {
				call = append(call, "--file", *file)
			}
			if *timeout != time.Duration(0) {
				call = append(call, "--timeout", timeout.String())
			}
			if *showEvents {
				call = append(call, "--show-events", "true")
			}
			if *ignoreNotFound {
				call = append(call, "--ignore-not-found")
			}
			if *selector != "" {
				call = append(call, "--selector", fmt.Sprintf("%v", *selector))
			}
			calls = append(calls, call)
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if fatal {
				util.CheckErr(errors.New(*cmdError))
			}
			if cmdError != nil {
				return errors.New(*cmdError)
			}
			if cmdOut != nil {
				_, _ = out.Write([]byte(*cmdOut))
			}
			return nil
		},
	}

	ctx = cmd.Flags().StringP("context", "c", "", "")
	ns = cmd.Flags().StringP("namespace", "n", "", "")
	file = cmd.Flags().StringP("filename", "f", "", "")
	timeout = cmd.Flags().DurationP("timeout", "t", 0*time.Second, "")
	showEvents = cmd.Flags().BoolP("show-events", "", false, "")
	ignoreNotFound = cmd.Flags().BoolP("ignore-not-found", "", false, "")
	selector = cmd.Flags().StringP("selector", "l", "", "")
	kubeconfig = cmd.Flags().StringP("kubeconfig", "", "", "")

	return &cmd
}
