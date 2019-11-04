package file

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var name string

func TestMain(m *testing.M) {
	tempDir := setup()
	code := m.Run()
	teardown(tempDir)
	os.Exit(code)
}

func setup() string {
	name, _ = ioutil.TempDir(os.TempDir(), "build-tools")

	return name
}

func teardown(tempDir string) {
	_ = os.RemoveAll(tempDir)
}

func TestWrite_Creates_All_Parent_Directories(t *testing.T) {
	fileName := filepath.Join(name, "missing", "path", "file")
	defer func() { _ = os.RemoveAll(fileName) }()
	err := Write(name, filepath.Join("missing", "path", "file"), "abc")
	assert.NoError(t, err)
	bytes, err := ioutil.ReadFile(fileName)
	assert.NoError(t, err)
	assert.Equal(t, "abc\n", string(bytes))
}
