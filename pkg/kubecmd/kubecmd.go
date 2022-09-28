// MIT License
//
// Copyright (c) 2021 buildtool
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

package kubecmd

import (
	"fmt"

	"github.com/apex/log"

	"github.com/buildtool/build-tools/pkg/args"
	"github.com/buildtool/build-tools/pkg/config"
	"github.com/buildtool/build-tools/pkg/version"
)

type Args struct {
	args.Globals
	Target    string `arg:"" name:"target" help:"the target in the .buildtools.yaml"`
	Context   string `name:"context" short:"c" help:"override the context for default deployment target" default:""`
	Namespace string `name:"namespace" short:"n" help:"override the namespace for default deployment target" default:""`
}

func Kubecmd(dir string, info version.Info, osArgs ...string) *string {
	var kubeCmdArgs Args
	err := args.ParseArgs(dir, osArgs, info, &kubeCmdArgs)
	if err != nil {
		return nil
	}
	if cfg, err := config.Load(dir); err != nil {
		log.Error(err.Error())
	} else {
		if env, err := cfg.CurrentTarget(kubeCmdArgs.Target); err != nil {
			log.Error(err.Error())
		} else {
			if kubeCmdArgs.Context != "" {
				env.Context = kubeCmdArgs.Context
			}
			if kubeCmdArgs.Namespace != "" {
				env.Namespace = kubeCmdArgs.Namespace
			}

			if len(env.Namespace) == 0 {
				env.Namespace = "default"
			}

			cmd := fmt.Sprintf("kubectl --context %s --namespace %s", env.Context, env.Namespace)
			return &cmd
		}
	}

	return nil
}
