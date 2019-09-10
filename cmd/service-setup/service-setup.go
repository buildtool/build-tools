package main

import (
	service "gitlab.com/sparetimecoders/build-tools/pkg/service-setup"
	"os"
)

func main() {
	dir, _ := os.Getwd()

	service.Setup(dir, os.Stdout, os.Exit, os.Args[1:]...)
}
