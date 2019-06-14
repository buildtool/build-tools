package docker

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestParseDockerignore_FileMissing(t *testing.T) {
	oldPwd, _ := os.Getwd()
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer os.RemoveAll(name)
	err := os.Chdir(name)
	assert.NoError(t, err)
	defer os.Chdir(oldPwd)

	var empty []string
	result, err := ParseDockerignore()
	assert.NoError(t, err)
	assert.Equal(t, empty, result)
}

func TestParseDockerignore_EmptyFile(t *testing.T) {
	oldPwd, _ := os.Getwd()
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer os.RemoveAll(name)
	err := os.Chdir(name)
	assert.NoError(t, err)
	defer os.Chdir(oldPwd)

	content := ``
	_ = ioutil.WriteFile(filepath.Join(name, ".dockerignore"), []byte(content), 0777)

	var empty []string
	result, err := ParseDockerignore()
	assert.NoError(t, err)
	assert.Equal(t, empty, result)
}

func TestParseDockerignore_UnreadableFile(t *testing.T) {
	oldPwd, _ := os.Getwd()
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer os.RemoveAll(name)
	err := os.Chdir(name)
	assert.NoError(t, err)
	defer os.Chdir(oldPwd)
	filename := filepath.Join(name, ".dockerignore")
	_ = os.Mkdir(filename, 0777)

	_, err = ParseDockerignore()
	assert.EqualError(t, err, "read .dockerignore: is a directory")
}

func TestParseDockerignore(t *testing.T) {
	oldPwd, _ := os.Getwd()
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer os.RemoveAll(name)
	err := os.Chdir(name)
	assert.NoError(t, err)
	defer os.Chdir(oldPwd)

	content := `
node_modules
*.swp`
	_ = ioutil.WriteFile(filepath.Join(name, ".dockerignore"), []byte(content), 0777)

	result, err := ParseDockerignore()
	assert.NoError(t, err)
	assert.Equal(t, []string{"node_modules", "*.swp"}, result)
}
