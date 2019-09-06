package docker

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestParseDockerignore_FileMissing(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()

	var empty []string
	result, err := ParseDockerignore(name)
	assert.NoError(t, err)
	assert.Equal(t, empty, result)
}

func TestParseDockerignore_EmptyFile(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()

	content := ``
	_ = ioutil.WriteFile(filepath.Join(name, ".dockerignore"), []byte(content), 0777)

	var empty []string
	result, err := ParseDockerignore(name)
	assert.NoError(t, err)
	assert.Equal(t, empty, result)
}

func TestParseDockerignore_UnreadableFile(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	filename := filepath.Join(name, ".dockerignore")
	_ = os.Mkdir(filename, 0777)

	_, err := ParseDockerignore(name)
	assert.EqualError(t, err, fmt.Sprintf("read %s: is a directory", filename))
}

func TestParseDockerignore(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()

	content := `
node_modules
*.swp`
	_ = ioutil.WriteFile(filepath.Join(name, ".dockerignore"), []byte(content), 0777)

	result, err := ParseDockerignore(name)
	assert.NoError(t, err)
	assert.Equal(t, []string{"node_modules", "*.swp"}, result)
}
