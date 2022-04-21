package registry

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/apex/log"
	"github.com/docker/docker/api/types"

	"github.com/buildtool/build-tools/pkg/docker"
)

type Gitlab struct {
	dockerRegistry `yaml:"-"`
	Registry       string `yaml:"registry" env:"CI_REGISTRY"`
	User           string `yaml:"user" env:"CI_REGISTRY_USER"`
	Repository     string `yaml:"repository" env:"CI_REGISTRY_IMAGE"`
	Token          string `yaml:"token,omitempty" env:"CI_JOB_TOKEN"`
}

var _ Registry = &Gitlab{}

func (r Gitlab) Name() string {
	return "Gitlab"
}

func (r Gitlab) Configured() bool {
	return len(r.Repository) > 0 || len(r.Registry) > 0
}

func (r Gitlab) Login(client docker.Client) error {
	if ok, err := client.RegistryLogin(context.Background(), r.GetAuthConfig()); err == nil {
		log.Debugf("%s\n", ok.Status)
		return nil
	} else {
		return err
	}
}

func (r Gitlab) GetAuthConfig() types.AuthConfig {
	return types.AuthConfig{Username: r.User, Password: r.Token, ServerAddress: r.Registry}
}

func (r Gitlab) GetAuthInfo() string {
	authBytes, _ := json.Marshal(r.GetAuthConfig())
	return base64.URLEncoding.EncodeToString(authBytes)
}

func (r Gitlab) RegistryUrl() string {
	if len(r.Repository) != 0 {
		if strings.Contains(r.Repository, "/") {
			return r.Repository[:strings.LastIndex(r.Repository, "/")]
		}
		return r.Repository
	}

	return r.Registry
}

func (r *Gitlab) Create(repository string) error {
	return nil
}
