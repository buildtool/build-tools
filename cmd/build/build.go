package main

import (
	"gitlab.com/sparetimecoders/build-tools/pkg/build"
	"os"
)

func main() {
	dir, _ := os.Getwd()

	build.DoBuild(dir, os.Stdout, os.Stderr, os.Exit, os.Args[1:]...)
}
