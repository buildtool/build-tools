package main

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/sparetimecoders/build-tools/pkg"
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
