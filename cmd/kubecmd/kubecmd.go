package main

import (
	"fmt"
	"gitlab.com/sparetimecoders/build-tools/pkg/kubecmd"
	"os"
)

func main() {
	dir, _ := os.Getwd()

	if cmd := kubecmd.Kubecmd(dir, os.Stderr, os.Args[1:]...); cmd != nil {
		fmt.Println(cmd)
	}
}
