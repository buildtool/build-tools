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
