// MIT License
//
// Copyright (c) 2026 buildtool
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
	"strings"

	"github.com/apex/log"
	"github.com/docker/docker/api/types/registry"

	"github.com/buildtool/build-tools/pkg/docker"
)

type Gitea struct {
	dockerRegistry `yaml:"-"`
	Registry       string `yaml:"registry" env:"GITEA_REGISTRY"`
	Username       string `yaml:"username" env:"GITEA_USERNAME"`
	Token          string `yaml:"token,omitempty" env:"GITEA_TOKEN"`
	Repository     string `yaml:"repository" env:"GITEA_REPOSITORY"`
}

var _ Registry = &Gitea{}

func (r Gitea) Name() string {
	return "Gitea"
}

func (r Gitea) Configured() bool {
	return len(r.Repository) > 0 || len(r.Registry) > 0
}

func (r Gitea) Login(client docker.Client) error {
	tokenInfo := maskToken(r.Token)
	log.Debugf("Gitea login: registry=%s, username=%s, token=%s\n", r.Registry, r.Username, tokenInfo)
	authConfig := r.GetAuthConfig()
	ok, err := client.RegistryLogin(context.Background(), authConfig)
	if err != nil {
		log.Errorf("Gitea login failed: %v\n", err)
		return fmt.Errorf("gitea registry login to %s failed: %w", r.Registry, err)
	}
	log.Debugf("Gitea login successful: %s\n", ok.Status)
	return nil
}

// maskToken returns a masked representation of the token showing length and first/last 4 chars.
func maskToken(token string) string {
	if len(token) == 0 {
		return "<empty>"
	}
	if len(token) <= 8 {
		return fmt.Sprintf("[len=%d]", len(token))
	}
	return fmt.Sprintf("%s...%s[len=%d]", token[:4], token[len(token)-4:], len(token))
}

func (r Gitea) GetAuthConfig() registry.AuthConfig {
	return registry.AuthConfig{Username: r.Username, Password: r.Token, ServerAddress: r.Registry}
}

func (r Gitea) GetAuthInfo() string {
	authBytes, _ := json.Marshal(r.GetAuthConfig())
	return base64.URLEncoding.EncodeToString(authBytes)
}

func (r Gitea) RegistryUrl() string {
	if len(r.Repository) != 0 {
		if strings.Contains(r.Repository, "/") {
			return fmt.Sprintf("%s/%s", r.Registry, r.Repository[:strings.LastIndex(r.Repository, "/")])
		}
		return fmt.Sprintf("%s/%s", r.Registry, r.Repository)
	}

	return r.Registry
}

func (r *Gitea) Create(repository string) error {
	return nil
}
