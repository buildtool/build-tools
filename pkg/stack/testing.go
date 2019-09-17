// +build !prod

package stack

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func assertFileContent(t *testing.T, file, expectedContent string) {
	bytes, err := ioutil.ReadFile(file)
	assert.NoError(t, err)
	assert.Equal(t, expectedContent, string(bytes))
}
