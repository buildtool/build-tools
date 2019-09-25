package main

import (
	"gitlab.com/sparetimecoders/build-tools/pkg/build"
	"os"
)

func main() {
	dir, _ := os.Getwd()

	exitFunc(build.DoBuild(dir, os.Stdout, os.Stderr, os.Args[1:]...))
}

var exitFunc = os.Exit
