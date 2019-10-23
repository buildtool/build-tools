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

type Quay struct {
	dockerRegistry
	Repository string `yaml:"repository" env:"QUAY_REPOSITORY"`
	Username   string `yaml:"username" env:"QUAY_USERNAME"`
	Password   string `yaml:"password" env:"QUAY_PASSWORD"`
}

var _ Registry = &Quay{}

func (r *Quay) Name() string {
	return "Quay.io"
}

func (r *Quay) Configured() bool {
	return len(r.Repository) > 0
}

func (r *Quay) Login(client docker.Client, out io.Writer) error {
	if ok, err := client.RegistryLogin(context.Background(), r.authConfig()); err == nil {
		_, _ = fmt.Fprintln(out, ok.Status)
		return nil
	} else {
		return err
	}
}

func (r Quay) GetAuthInfo() string {
	auth := r.authConfig()
	authBytes, _ := json.Marshal(auth)
	return base64.URLEncoding.EncodeToString(authBytes)
}

func (r Quay) authConfig() types.AuthConfig {
	return types.AuthConfig{Username: r.Username, Password: r.Password, ServerAddress: "quay.io"}
}

func (r Quay) RegistryUrl() string {
	return fmt.Sprintf("quay.io/%s", r.Repository)
}

func (r *Quay) Create(repository string) error {
	return nil
}
