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

package registry

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/apex/log"
	"github.com/docker/docker/api/types"

	"github.com/buildtool/build-tools/pkg/docker"
)

type Dockerhub struct {
	dockerRegistry
	Namespace string `yaml:"namespace" env:"DOCKERHUB_NAMESPACE"`
	Username  string `yaml:"username" env:"DOCKERHUB_USERNAME"`
	Password  string `yaml:"password" env:"DOCKERHUB_PASSWORD"`
}

var _ Registry = &Dockerhub{}

func (r Dockerhub) Name() string {
	return "Dockerhub"
}

func (r Dockerhub) Configured() bool {
	return len(r.Namespace) > 0
}

func (r Dockerhub) Login(client docker.Client) error {
	if ok, err := client.RegistryLogin(context.Background(), r.GetAuthConfig()); err == nil {
		log.Debugf("%s\n", ok.Status)
		return nil
	} else {
		log.Errorf("%s", "Unable to login\n")
		return err
	}
}

func (r Dockerhub) GetAuthConfig() types.AuthConfig {
	return types.AuthConfig{Username: r.Username, Password: r.Password}
}

func (r Dockerhub) GetAuthInfo() string {
	authBytes, _ := json.Marshal(r.GetAuthConfig())
	return base64.URLEncoding.EncodeToString(authBytes)
}

func (r Dockerhub) RegistryUrl() string {
	return r.Namespace
}

func (r *Dockerhub) Create(repository string) error {
	return nil
}
