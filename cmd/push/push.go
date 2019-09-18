package main

import (
	"gitlab.com/sparetimecoders/build-tools/pkg/push"
	"os"
)

func main() {
	dir, _ := os.Getwd()
	exitFunc(push.Push(dir))
}

var exitFunc = os.Exit
