package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/apex/log"
	"github.com/stretchr/testify/assert"
	mocks "gitlab.com/unboundsoftware/apex-mocks"
)

func TestPush(t *testing.T) {
	logMock := mocks.New()
	handler = logMock
	os.Clearenv()
	exitFunc = func(code int) {
		assert.Equal(t, -5, code)
	}

	oldPwd, _ := os.Getwd()
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()

	err := os.Chdir(name)
	assert.NoError(t, err)
	workDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldPwd) }()

	os.Args = []string{"push"}
	main()
	logMock.Check(t, []string{fmt.Sprintf("error: <red>open %s/Dockerfile: no such file or directory</red>", workDir)})
}

func TestVersion(t *testing.T) {
	logMock := mocks.New()
	handler = logMock
	log.SetLevel(log.DebugLevel)
	version = "1.0.0"
	commit = "67d2fcf276fcd9cf743ad4be9a9ef5828adc082f"
	date = "2006-01-02T15:04:05Z07:00"
	exitFunc = func(code int) {
		assert.Equal(t, 0, code)
	}
	os.Args = []string{"push", "--version"}
	main()

	logMock.Check(t, []string{"info: Version: 1.0.0, commit 67d2fcf276fcd9cf743ad4be9a9ef5828adc082f, built at 2006-01-02T15:04:05Z07:00\n"})
}
