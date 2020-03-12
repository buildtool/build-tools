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

	defaultIgnores := []string{"k8s"}
	result, err := ParseDockerignore(name, "Dockerfile")
	assert.NoError(t, err)
	assert.Equal(t, defaultIgnores, result)
}

func TestParseDockerignore_EmptyFile(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()

	content := ``
	_ = ioutil.WriteFile(filepath.Join(name, ".dockerignore"), []byte(content), 0777)

	defaultIgnores := []string{"k8s"}
	result, err := ParseDockerignore(name, "Dockerfile")
	assert.NoError(t, err)
	assert.Equal(t, defaultIgnores, result)
}

func TestParseDockerignore_UnreadableFile(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	filename := filepath.Join(name, ".dockerignore")
	_ = os.Mkdir(filename, 0777)

	_, err := ParseDockerignore(name, "Dockerfile")
	assert.EqualError(t, err, fmt.Sprintf("read %s: is a directory", filename))
}

func TestParseDockerignore(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()

	content := `
node_modules
*.swp`
	_ = ioutil.WriteFile(filepath.Join(name, ".dockerignore"), []byte(content), 0777)

	result, err := ParseDockerignore(name, "Dockerfile")
	assert.NoError(t, err)
	assert.Equal(t, []string{"k8s", "node_modules", "*.swp"}, result)
}

func TestParseDockerignore_Dockerfile_Ignored(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()

	content := `
node_modules
Dockerfile
*.swp`
	_ = ioutil.WriteFile(filepath.Join(name, ".dockerignore"), []byte(content), 0777)

	result, err := ParseDockerignore(name, "Dockerfile")
	assert.NoError(t, err)
	assert.Equal(t, []string{"k8s", "node_modules", "*.swp"}, result)
}
