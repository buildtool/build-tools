package config

import (
	"context"
	"docker.io/go-docker/api/types"
	"encoding/base64"
	"encoding/json"
	"gitlab.com/sparetimecoders/build-tools/pkg/docker"
	"log"
)

type DockerhubRegistry struct {
	dockerRegistry
	Repository string `yaml:"repository" env:"DOCKERHUB_REPOSITORY"`
	Username   string `yaml:"username" env:"DOCKERHUB_USERNAME"`
	Password   string `yaml:"password" env:"DOCKERHUB_PASSWORD"`
}

var _ Registry = &DockerhubRegistry{}

func (r DockerhubRegistry) configured() bool {
	return len(r.Repository) > 0
}

func (r DockerhubRegistry) Login(client docker.Client) error {
	if ok, err := client.RegistryLogin(context.Background(), r.authConfig()); err == nil {
		log.Println(ok.Status)
		return nil
	} else {
		log.Println("Unable to login")
		return err
	}
}

func (r DockerhubRegistry) GetAuthInfo() string {
	authBytes, _ := json.Marshal(r.authConfig())
	return base64.URLEncoding.EncodeToString(authBytes)
}

func (r DockerhubRegistry) authConfig() types.AuthConfig {
	return types.AuthConfig{Username: r.Username, Password: r.Password}
}

func (r DockerhubRegistry) RegistryUrl() string {
	return r.Repository
}

func (r *DockerhubRegistry) Create(repository string) error {
	return nil
}
