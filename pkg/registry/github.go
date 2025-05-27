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
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/apex/log"
	"github.com/docker/docker/api/types/registry"

	"github.com/buildtool/build-tools/pkg/docker"
)

type Github struct {
	dockerRegistry `yaml:"-"`
	Username       string `yaml:"username" env:"GITHUB_USERNAME"`
	Password       string `yaml:"password" env:"GITHUB_PASSWORD"`
	Token          string `yaml:"token" env:"GITHUB_TOKEN"`
	Repository     string `yaml:"repository" env:"GITHUB_REPOSITORY_OWNER"`
}

var _ Registry = &Github{}

func (r Github) Name() string {
	return "Github"
}

func (r Github) Configured() bool {
	return len(r.Repository) > 0
}

func (r Github) Login(client docker.Client) error {
	if ok, err := client.RegistryLogin(context.Background(), registry.AuthConfig{Username: r.Username, Password: r.password(), ServerAddress: "ghcr.io"}); err == nil {
		log.Debugf("%s\n", ok.Status)
		return nil
	} else {
		return err
	}
}

func (r Github) password() string {
	if len(r.Token) > 0 {
		return r.Token
	}
	return r.Password
}

func (r Github) GetAuthConfig() registry.AuthConfig {
	return registry.AuthConfig{Username: r.Username, Password: r.password(), ServerAddress: "ghcr.io"}
}

func (r Github) GetAuthInfo() string {
	authBytes, _ := json.Marshal(r.GetAuthConfig())
	return base64.URLEncoding.EncodeToString(authBytes)
}

func (r Github) RegistryUrl() string {
	return fmt.Sprintf("ghcr.io/%s", r.Repository)
}

func (r *Github) Create(repository string) error {
	return nil
}
