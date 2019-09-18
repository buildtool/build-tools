package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestBuild(t *testing.T) {
	exitFunc = func(code int) {
		assert.Equal(t, -7, code)
	}
	os.Args = []string{"build"}
	main()
}
