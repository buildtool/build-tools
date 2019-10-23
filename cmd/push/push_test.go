package main

import (
	"bytes"
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

func TestVersion(t *testing.T) {
	out = &bytes.Buffer{}
	version = "1.0.0"
	commit = "67d2fcf276fcd9cf743ad4be9a9ef5828adc082f"
	date = "2006-01-02T15:04:05Z07:00"
	exitFunc = func(code int) {
		assert.Equal(t, 0, code)
	}
	os.Args = []string{"push", "-version"}
	main()

	assert.Equal(t, "Version: 1.0.0, commit 67d2fcf276fcd9cf743ad4be9a9ef5828adc082f, built at 2006-01-02T15:04:05Z07:00\n", out.(*bytes.Buffer).String())
}
