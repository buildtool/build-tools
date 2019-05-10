package config

import (
	"context"
	"docker.io/go-docker/api/types"
	"encoding/base64"
	"encoding/json"
	"gitlab.com/sparetimecoders/build-tools/pkg/docker"
	"log"
	"strings"
)

type GitlabRegistry struct {
	dockerRegistry
	Registry   string `yaml:"registry" env:"CI_REGISTRY"`
	Repository string `yaml:"repository" env:"CI_REGISTRY_IMAGE"`
	Token      string `yaml:"token" env:"CI_TOKEN"`
}

var _ Registry = &GitlabRegistry{}

func (r GitlabRegistry) configured() bool {
	return len(r.Repository) > 0
}

func (r GitlabRegistry) Login(client docker.Client) error {
	if ok, err := client.RegistryLogin(context.Background(), types.AuthConfig{Username: "gitlab-ci-token", Password: r.Token, ServerAddress: r.Registry}); err == nil {
		log.Println(ok.Status)
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
	return r.Repository[:strings.LastIndex(r.Repository, "/")]
}

func (r *GitlabRegistry) Create(repository string) error {
	return nil
}
