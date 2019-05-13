package kubectl

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/sparetimecoders/build-tools/pkg/config"
	"strings"
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
	k := New(&config.Environment{Context: "missing", Namespace: "default", Name: "dummy"})

	err := k.Apply(strings.NewReader(""), "xxx", "yyy")
	assert.EqualError(t, err, "unknown command \"xxx\" for \"kubectl\"")
}

func TestKubectl_Environment(t *testing.T) {
	env := &config.Environment{Context: "missing", Namespace: "default", Name: "dummy"}
	k := New(env)

	assert.Equal(t, env, k.Environment())
}
