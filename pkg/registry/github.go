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

type Github struct {
	dockerRegistry `yaml:"-"`
	Username       string `yaml:"username" env:"GITHUB_USERNAME"`
	Password       string `yaml:"password" env:"GITHUB_PASSWORD"`
	Token          string `yaml:"token" env:"GITHUB_TOKEN"`
	Repository     string `yaml:"repository" env:"GITHUB_REPOSITORY_OWNER"`
}

var _ Registry = &Github{}

func (r Github) Name() string {
	return "Github"
}

func (r Github) Configured() bool {
	return len(r.Repository) > 0
}

func (r Github) Login(client docker.Client) error {
	if ok, err := client.RegistryLogin(context.Background(), types.AuthConfig{Username: r.Username, Password: r.password(), ServerAddress: "ghcr.io"}); err == nil {
		log.Debugf("%s\n", ok.Status)
		return nil
	} else {
		return err
	}
}

func (r Github) password() string {
	if len(r.Token) > 0 {
		return r.Token
	}
	return r.Password
}
func (r Github) GetAuthConfig() types.AuthConfig {
	return types.AuthConfig{Username: r.Username, Password: r.password(), ServerAddress: "ghcr.io"}
}

func (r Github) GetAuthInfo() string {
	authBytes, _ := json.Marshal(r.GetAuthConfig())
	return base64.URLEncoding.EncodeToString(authBytes)
}

func (r Github) RegistryUrl() string {
	return fmt.Sprintf("ghcr.io/%s", r.Repository)
}

func (r *Github) Create(repository string) error {
	return nil
}
