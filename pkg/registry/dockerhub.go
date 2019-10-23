package registry

import (
	"context"
	"docker.io/go-docker/api/types"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/sparetimecoders/build-tools/pkg/docker"
	"io"
)

type Dockerhub struct {
	dockerRegistry
	Repository string `yaml:"repository" env:"DOCKERHUB_REPOSITORY"`
	Username   string `yaml:"username" env:"DOCKERHUB_USERNAME"`
	Password   string `yaml:"password" env:"DOCKERHUB_PASSWORD"`
}

var _ Registry = &Dockerhub{}

func (r Dockerhub) Name() string {
	return "Dockerhub"
}

func (r Dockerhub) Configured() bool {
	return len(r.Repository) > 0
}

func (r Dockerhub) Login(client docker.Client, out io.Writer) error {
	if ok, err := client.RegistryLogin(context.Background(), r.authConfig()); err == nil {
		_, _ = fmt.Fprintln(out, ok.Status)
		return nil
	} else {
		_, _ = fmt.Fprintln(out, "Unable to login")
		return err
	}
}

func (r Dockerhub) GetAuthInfo() string {
	authBytes, _ := json.Marshal(r.authConfig())
	return base64.URLEncoding.EncodeToString(authBytes)
}

func (r Dockerhub) authConfig() types.AuthConfig {
	return types.AuthConfig{Username: r.Username, Password: r.Password}
}

func (r Dockerhub) RegistryUrl() string {
	return r.Repository
}

func (r *Dockerhub) Create(repository string) error {
	return nil
}
