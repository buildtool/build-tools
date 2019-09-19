package main

import (
	"gitlab.com/sparetimecoders/build-tools/pkg/push"
	"os"
)

func main() {
	dir, _ := os.Getwd()
	exitFunc(push.Push(dir, os.Stdout, os.Stderr, os.Args[1:]...))
}

var exitFunc = os.Exit
