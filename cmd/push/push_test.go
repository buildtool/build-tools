package main

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestPush(t *testing.T) {
	os.Clearenv()
	exitFunc = func(code int) {
		assert.Equal(t, -5, code)
	}

	oldPwd, _ := os.Getwd()
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()

	err := os.Chdir(name)
	assert.NoError(t, err)
	defer func() { _ = os.Chdir(oldPwd) }()

	os.Args = []string{"push"}
	main()
}
