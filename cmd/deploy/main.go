package main

import (
	"fmt"
	"gitlab.com/sparetimecoders/build-tools/pkg/config"
	"gitlab.com/sparetimecoders/build-tools/pkg/deploy"
	"gitlab.com/sparetimecoders/build-tools/pkg/kubectl"
	"os"
	"time"
)

func main() {
	environment := os.Args[1]
	dir, _ := os.Getwd()

	if cfg, err := config.Load(dir); err != nil {
		fmt.Println(err.Error())
	} else {
		if env, err := cfg.CurrentEnvironment(environment); err != nil {
			fmt.Println(err.Error())
		} else {
			if ci, err := cfg.CurrentCI(); err != nil {
				fmt.Println(err.Error())
			} else {
				tstamp := time.Now().Format(time.RFC3339)
				if err := deploy.Deploy(dir, ci.Commit(), tstamp, kubectl.New(env)); err != nil {
					fmt.Println(err.Error())
				}
			}
		}
	}
}
