package main

import (
	"os"
	"testing"

	"github.com/apex/log"
	"github.com/stretchr/testify/assert"
	mocks "gitlab.com/unboundsoftware/apex-mocks"
)

func TestVersion(t *testing.T) {
	logMock := mocks.New()
	handler = logMock
	log.SetLevel(log.DebugLevel)
	version = "1.0.0"
	exitFunc = func(code int) {
		assert.Equal(t, 0, code)
	}
	os.Args = []string{"promote", "--version"}
	main()

	logMock.Check(t, []string{"info: Version: 1.0.0, commit none, built at unknown\n"})
}
