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

	"github.com/apex/log"
	"github.com/docker/docker/api/types"

	"github.com/buildtool/build-tools/pkg/docker"
)

type GCR struct {
	dockerRegistry `yaml:"-"`
	Url            string `yaml:"url" env:"GCR_URL"`
	KeyFileContent string `yaml:"keyfileContent,omitempty" env:"GCR_KEYFILE_CONTENT"`
}

var _ Registry = &GCR{}

func (r *GCR) Name() string {
	return "GCR"
}

func (r *GCR) Configured() bool {
	if len(r.Url) <= 0 || len(r.KeyFileContent) <= 0 {
		return false
	}
	return r.GetAuthConfig() != types.AuthConfig{}
}

func (r *GCR) Login(client docker.Client) error {
	auth := r.GetAuthConfig()
	auth.ServerAddress = r.Url
	if ok, err := client.RegistryLogin(context.Background(), auth); err == nil {
		log.Debugf("%s\n", ok.Status)
		return nil
	} else {
		return err
	}
}

func (r *GCR) GetAuthConfig() types.AuthConfig {
	decoded, err := base64.StdEncoding.DecodeString(r.KeyFileContent)
	if err != nil {
		return types.AuthConfig{}
	}
	return types.AuthConfig{Username: "_json_key", Password: string(decoded)}
}

func (r *GCR) GetAuthInfo() string {
	authBytes, _ := json.Marshal(r.GetAuthConfig())
	return base64.URLEncoding.EncodeToString(authBytes)
}

func (r GCR) RegistryUrl() string {
	return r.Url
}

func (r GCR) Create(repository string) error {
	return nil
}
