package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestPush(t *testing.T) {
	os.Clearenv()
	exitFunc = func(code int) {
		assert.Equal(t, -5, code)
	}

	os.Args = []string{"push"}
	main()
}
