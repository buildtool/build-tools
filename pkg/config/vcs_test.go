package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/apex/log"
	"github.com/stretchr/testify/assert"
	mocks "gitlab.com/unboundsoftware/apex-mocks"

	"github.com/buildtool/build-tools/pkg/vcs"
)

func TestIdentify(t *testing.T) {
	os.Clearenv()

	dir, _ := os.MkdirTemp("", "build-tools")
	defer func() { _ = os.RemoveAll(dir) }()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	result := vcs.Identify(dir)
	assert.NotNil(t, result)
	assert.Equal(t, "none", result.Name())
	assert.Equal(t, "", result.Commit())
	assert.Equal(t, "", result.Branch())
	logMock.Check(t, []string{})
}

func TestGit_Identify(t *testing.T) {
	dir, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(dir) }()

	hash, _ := InitRepoWithCommit(dir)

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	result := vcs.Identify(dir)
	assert.NotNil(t, result)
	assert.Equal(t, "Git", result.Name())
	assert.Equal(t, hash.String(), result.Commit())
	assert.Equal(t, "master", result.Branch())
	logMock.Check(t, []string{})
}

func TestGit_Identify_Subdirectory(t *testing.T) {
	dir, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(dir) }()

	hash, _ := InitRepoWithCommit(dir)

	subdir := filepath.Join(dir, "subdir")
	_ = os.Mkdir(subdir, 0777)

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	result := vcs.Identify(subdir)
	assert.NotNil(t, result)
	assert.Equal(t, "Git", result.Name())
	assert.Equal(t, hash.String(), result.Commit())
	assert.Equal(t, "master", result.Branch())
	logMock.Check(t, []string{})
}

func TestGit_MissingRepo(t *testing.T) {
	dir, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(dir) }()

	_ = os.Mkdir(filepath.Join(dir, ".git"), 0777)

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	result := vcs.Identify(dir)
	assert.NotNil(t, result)
	assert.Equal(t, "", result.Commit())
	assert.Equal(t, "", result.Branch())
	logMock.Check(t, []string{})
}

func TestGit_NoCommit(t *testing.T) {
	dir, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(dir) }()

	InitRepo(dir)

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	result := vcs.Identify(dir)
	assert.NotNil(t, result)
	assert.Equal(t, "", result.Commit())
	assert.Equal(t, "", result.Branch())
	logMock.Check(t, []string{"debug: Unable to fetch head: reference not found\n"})
}
