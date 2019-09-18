package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestPush(t *testing.T) {
	exitFunc = func(code int) {
		assert.Equal(t, -2, code)
	}

	os.Args = []string{"push"}
	main()
}
