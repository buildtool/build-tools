package vcs

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestIdentify(t *testing.T) {
	os.Clearenv()

	dir, _ := ioutil.TempDir("", "build-tools")
	defer os.RemoveAll(dir)

	result := Identify(dir)
	assert.NotNil(t, result)
	assert.Equal(t, "", result.Commit())
	assert.Equal(t, "", result.Branch())
}
