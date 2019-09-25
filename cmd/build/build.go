package main

import (
	"fmt"
	"gitlab.com/sparetimecoders/build-tools/pkg/build"
	"os"
	"strings"
)

func main() {
	dir, _ := os.Getwd()

	fmt.Println(strings.Join(os.Environ(), "\n"))
	exitFunc(build.DoBuild(dir, os.Stdout, os.Stderr, os.Args[1:]...))
}

var exitFunc = os.Exit
