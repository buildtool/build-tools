package config

import (
	"context"
	"docker.io/go-docker/api/types"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"gitlab.com/sparetimecoders/build-tools/pkg/docker"
	"log"
)

type QuayRegistry struct {
	dockerRegistry
	Repository string `yaml:"repository" env:"QUAY_REPOSITORY"`
	Username   string `yaml:"username" env:"QUAY_USERNAME"`
	Password   string `yaml:"password" env:"QUAY_PASSWORD"`
}

var _ Registry = &QuayRegistry{}

func (r *QuayRegistry) configured() bool {
	return len(r.Repository) > 0
}

func (r *QuayRegistry) Login(client docker.Client) error {
	if ok, err := client.RegistryLogin(context.Background(), r.authConfig()); err == nil {
		log.Println(ok.Status)
		return nil
	} else {
		return err
	}
}

func (r QuayRegistry) GetAuthInfo() string {
	auth := r.authConfig()
	authBytes, _ := json.Marshal(auth)
	return base64.URLEncoding.EncodeToString(authBytes)
}

func (r QuayRegistry) authConfig() types.AuthConfig {
	return types.AuthConfig{Username: r.Username, Password: r.Password, ServerAddress: "quay.io"}
}

func (r QuayRegistry) RegistryUrl() string {
	return fmt.Sprintf("quay.io/%s", r.Repository)
}

func (r *QuayRegistry) Create(repository string) error {
	return nil
}
