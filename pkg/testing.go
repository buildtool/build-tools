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

//go:build !prod
// +build !prod

package pkg

import (
	"os"
	"strings"
)

func SetEnv(key, value string) func() {
	_ = os.Setenv(key, value)
	return func() { _ = os.Unsetenv(key) }
}

func UnsetGithubEnvironment() func() {
	envs := make(map[string]string)
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "GITHUB_") {
			parts := strings.Split(e, "=")
			envs[parts[0]] = os.Getenv(parts[1])
			_ = os.Unsetenv(parts[0])
		}
	}
	envs["RUNNER_WORKSPACE"] = os.Getenv("RUNNER_WORKSPACE")
	_ = os.Unsetenv("RUNNER_WORKSPACE")
	return func() {
		for k, v := range envs {
			_ = os.Setenv(k, v)
		}
	}
}
