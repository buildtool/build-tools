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

package registry

import (
	"fmt"

	"github.com/apex/log"
	"github.com/docker/docker/api/types"

	"github.com/buildtool/build-tools/pkg/docker"
)

type NoDockerRegistry struct{}

func (n NoDockerRegistry) Configured() bool {
	return true
}

func (n NoDockerRegistry) Name() string {
	return "No docker registry"
}

func (n NoDockerRegistry) Login(client docker.Client) error {
	log.Debugf("Authentication <yellow>not supported</yellow> for registry <green>%s</green>\n", n.Name())
	return nil
}

func (n NoDockerRegistry) GetAuthConfig() types.AuthConfig {
	return types.AuthConfig{}
}

func (n NoDockerRegistry) GetAuthInfo() string {
	return ""
}

func (n NoDockerRegistry) RegistryUrl() string {
	return "noregistry"
}

func (n NoDockerRegistry) Create(repository string) error {
	return nil
}

func (n NoDockerRegistry) PushImage(client docker.Client, auth, image string) error {
	return fmt.Errorf("push not supported by registry")
}

var _ Registry = &NoDockerRegistry{}
