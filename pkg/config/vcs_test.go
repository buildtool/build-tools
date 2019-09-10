package config

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestIdentify(t *testing.T) {
	os.Clearenv()

	dir, _ := ioutil.TempDir("", "build-tools")
	defer os.RemoveAll(dir)

	out := &bytes.Buffer{}
	result := Identify(dir, out)
	assert.NotNil(t, result)
	assert.Equal(t, "none", result.Name())
	assert.Equal(t, "", result.Commit())
	assert.Equal(t, "", result.Branch())
	assert.Equal(t, "", out.String())
}
