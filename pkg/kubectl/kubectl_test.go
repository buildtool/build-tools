package kubectl

import (
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"gitlab.com/sparetimecoders/build-tools/pkg/config"
	"io"
	"testing"
)

func TestNew(t *testing.T) {
	k := New(&config.Environment{Context: "missing", Namespace: "dev", Name: "dummy"})

	assert.Equal(t, "missing", k.(*kubectl).context)
	assert.Equal(t, "dev", k.(*kubectl).namespace)
}

func TestNew_NoNamespace(t *testing.T) {
	k := New(&config.Environment{Context: "missing", Namespace: "", Name: "dummy"})

	assert.Equal(t, "missing", k.(*kubectl).context)
	assert.Equal(t, "default", k.(*kubectl).namespace)
}

func TestKubectl_Apply(t *testing.T) {
	newKubectlCmd = mockCmd

	k := New(&config.Environment{Context: "missing", Namespace: "default", Name: "dummy"})

	err := k.Apply("")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(calls))
	assert.Equal(t, []string{}, calls[0])
}

func TestKubectl_UnableToCreateTempDir(t *testing.T) {
	newKubectlCmd = mockCmd

	k := &kubectl{context: "missing", namespace: "default", environment: &config.Environment{}, tempDir: "/missing"}

	err := k.Apply("")
	assert.EqualError(t, err, "open /missing/content.yaml: no such file or directory")
}

func TestKubectl_Environment(t *testing.T) {
	env := &config.Environment{Context: "missing", Namespace: "default", Name: "dummy"}
	k := New(env)

	assert.Equal(t, env, k.Environment())
}

var calls [][]string

func mockCmd(in io.Reader, out, err io.Writer) *cobra.Command {
	cmd := cobra.Command{
		Use: "kubectl",
	}

	apply := cobra.Command{
		Use: "apply",
		Run: func(cmd *cobra.Command, args []string) {
			calls = append(calls, args)
		},
	}
	apply.Flags().StringP("context", "c", "", "")
	apply.Flags().StringP("namespace", "n", "", "")
	apply.Flags().StringP("filename", "f", "", "")
	cmd.AddCommand(&apply)

	return &cmd
}
