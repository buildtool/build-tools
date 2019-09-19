package main

import (
	service "gitlab.com/sparetimecoders/build-tools/pkg/service-setup"
	"os"
)

func main() {
	dir, _ := os.Getwd()

	exitFunc(service.Setup(dir, os.Stdout, os.Args[1:]...))
}

var exitFunc = os.Exit
