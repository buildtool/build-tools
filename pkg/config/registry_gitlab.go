package config

import (
	"context"
	"docker.io/go-docker/api/types"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"gitlab.com/sparetimecoders/build-tools/pkg/docker"
	"io"
	"strings"
)

type GitlabRegistry struct {
	dockerRegistry
	Registry   string `yaml:"registry" env:"CI_REGISTRY"`
	Repository string `yaml:"repository" env:"CI_REGISTRY_IMAGE"`
	Token      string `yaml:"token" env:"CI_TOKEN"`
}

var _ Registry = &GitlabRegistry{}

func (r GitlabRegistry) Name() string {
	return "Gitlab"
}

func (r GitlabRegistry) configured() bool {
	return len(r.Repository) > 0
}

func (r GitlabRegistry) Login(client docker.Client, out io.Writer) error {
	if ok, err := client.RegistryLogin(context.Background(), types.AuthConfig{Username: "gitlab-ci-token", Password: r.Token, ServerAddress: r.Registry}); err == nil {
		_, _ = fmt.Fprintln(out, ok.Status)
		return nil
	} else {
		return err
	}
}

func (r GitlabRegistry) GetAuthInfo() string {
	auth := types.AuthConfig{Username: "gitlab-ci-token", Password: r.Token, ServerAddress: r.Registry}
	authBytes, _ := json.Marshal(auth)
	return base64.URLEncoding.EncodeToString(authBytes)
}

func (r GitlabRegistry) RegistryUrl() string {
	if len(r.Repository) != 0 {
		if strings.Index(r.Repository, "/") != -1 {
			return r.Repository[:strings.LastIndex(r.Repository, "/")]
		}
		return r.Repository
	}

	return r.Registry
}

func (r *GitlabRegistry) Create(repository string) error {
	return nil
}
