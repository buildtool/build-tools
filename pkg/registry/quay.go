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
	"fmt"

	"github.com/apex/log"
	"github.com/docker/docker/api/types"

	"github.com/buildtool/build-tools/pkg/docker"
)

type Quay struct {
	dockerRegistry `yaml:"-"`
	Repository     string `yaml:"repository" env:"QUAY_REPOSITORY"`
	Username       string `yaml:"username" env:"QUAY_USERNAME"`
	Password       string `yaml:"password" env:"QUAY_PASSWORD"`
}

var _ Registry = &Quay{}

func (r *Quay) Name() string {
	return "Quay.io"
}

func (r *Quay) Configured() bool {
	return len(r.Repository) > 0
}

func (r *Quay) Login(client docker.Client) error {
	if ok, err := client.RegistryLogin(context.Background(), r.GetAuthConfig()); err == nil {
		log.Debugf("%s\n", ok.Status)
		return nil
	} else {
		return err
	}
}

func (r Quay) GetAuthConfig() types.AuthConfig {
	return types.AuthConfig{Username: r.Username, Password: r.Password, ServerAddress: "quay.io"}
}

func (r Quay) GetAuthInfo() string {
	authBytes, _ := json.Marshal(r.GetAuthConfig())
	return base64.URLEncoding.EncodeToString(authBytes)
}

func (r Quay) RegistryUrl() string {
	return fmt.Sprintf("quay.io/%s", r.Repository)
}

func (r *Quay) Create(repository string) error {
	return nil
}
