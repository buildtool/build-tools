package main

import (
	"bytes"
	"github.com/sparetimecoders/build-tools/pkg"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func Test(t *testing.T) {
	oldPwd, _ := os.Getwd()
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer os.RemoveAll(name)
	err := os.Chdir(name)
	assert.NoError(t, err)
	defer os.Chdir(oldPwd)

	defer pkg.SetEnv("REGISTRY", "dockerhub")()

	exitFunc = func(code int) {
		assert.Equal(t, -1, code)
	}
	os.Args = []string{"service-setup"}
	main()
}

func TestVersion(t *testing.T) {
	out = &bytes.Buffer{}
	version = "1.0.0"
	commit = "67d2fcf276fcd9cf743ad4be9a9ef5828adc082f"
	date = "2006-01-02T15:04:05Z07:00"
	exitFunc = func(code int) {
		assert.Equal(t, 0, code)
	}
	os.Args = []string{"build", "-version"}
	main()

	assert.Equal(t, "Version: 1.0.0, commit 67d2fcf276fcd9cf743ad4be9a9ef5828adc082f, built at 2006-01-02T15:04:05Z07:00\n", out.(*bytes.Buffer).String())
}
