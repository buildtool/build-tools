// MIT License
//
// Copyright (c) 2018 buildtool
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package push

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apex/log"

	"github.com/buildtool/build-tools/pkg/args"
	"github.com/buildtool/build-tools/pkg/version"

	"github.com/buildtool/build-tools/pkg/ci"
	"github.com/buildtool/build-tools/pkg/config"
	"github.com/buildtool/build-tools/pkg/docker"
)

type Args struct {
	args.Globals
	Dockerfile string `name:"file" short:"f" help:"name of the Dockerfile to use." default:"Dockerfile"`
}

var dockerClient = docker.DefaultClient

func Push(dir string, info version.Info, osArgs ...string) int {
	var pushArgs Args
	err := args.ParseArgs(dir, osArgs, info, &pushArgs)
	if err != nil {
		if err != args.ErrDone {
			return -1
		} else {
			return 0
		}
	}

	client, err := dockerClient()
	if err != nil {
		log.Error(fmt.Sprintf("<red>%s</red>", err.Error()))
		return -1
	}
	cfg, err := config.Load(dir)
	if err != nil {
		log.Error(fmt.Sprintf("<red>%s</red>", err.Error()))
		return -2
	}
	return doPush(client, cfg, dir, pushArgs.Dockerfile)
}

func doPush(client docker.Client, cfg *config.Config, dir, dockerfile string) int {
	// If BUILDKIT_HOST is set, images are pushed during build, so push is a no-op
	if os.Getenv("BUILDKIT_HOST") != "" {
		log.Info("BUILDKIT_HOST is set - images were pushed during build, skipping push")
		return 0
	}

	currentCI := cfg.CurrentCI()
	currentRegistry := cfg.CurrentRegistry()

	if err := currentRegistry.Login(client); err != nil {
		log.Error(fmt.Sprintf("<red>%s</red>", err.Error()))
		return -3
	}

	auth := currentRegistry.GetAuthInfo()

	if err := currentRegistry.Create(currentCI.BuildName()); err != nil {
		log.Error(fmt.Sprintf("<red>%s</red>", err.Error()))
		return -4
	}

	content, err := os.ReadFile(filepath.Join(dir, dockerfile))
	if err != nil {
		log.Error(fmt.Sprintf("<red>%s</red>", err.Error()))
		return -5
	}
	stages := docker.FindStages(string(content))

	var tags []string
	for _, stage := range stages {
		tags = append(tags, docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), stage))
	}

	if !ci.IsValid(currentCI) {
		log.Error("Commit and/or branch information is <red>missing</red>. Perhaps your not in a Git repository or forgot to set environment variables?")
		return -6
	}
	tags = append(tags,
		docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), currentCI.Commit()),
		docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), currentCI.BranchReplaceSlash()),
	)
	if currentCI.Branch() == "master" || currentCI.Branch() == "main" {
		tags = append(tags, docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), "latest"))
	}
	for _, tag := range tags {
		log.Info(fmt.Sprintf("Pushing tag '<green>%s</green>'\n", tag))
		if err := currentRegistry.PushImage(client, auth, tag); err != nil {
			log.Error(fmt.Sprintf("<red>%s</red>", err.Error()))
			return -7
		}
	}
	return 0
}
