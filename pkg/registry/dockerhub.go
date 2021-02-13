package registry

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/buildtool/build-tools/pkg/docker"
	"github.com/docker/docker/api/types"
	"io"
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

func (r Dockerhub) Login(client docker.Client, out io.Writer) error {
	if ok, err := client.RegistryLogin(context.Background(), r.GetAuthConfig()); err == nil {
		_, _ = fmt.Fprintln(out, ok.Status)
		return nil
	} else {
		_, _ = fmt.Fprintln(out, "Unable to login")
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
