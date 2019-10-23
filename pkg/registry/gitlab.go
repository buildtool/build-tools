package registry

import (
	"context"
	"docker.io/go-docker/api/types"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/sparetimecoders/build-tools/pkg/docker"
	"io"
	"strings"
)

type Gitlab struct {
	dockerRegistry
	Registry   string `yaml:"registry" env:"CI_REGISTRY"`
	Repository string `yaml:"repository" env:"CI_REGISTRY_IMAGE"`
	Token      string `yaml:"token" env:"CI_TOKEN"`
}

var _ Registry = &Gitlab{}

func (r Gitlab) Name() string {
	return "Gitlab"
}

func (r Gitlab) Configured() bool {
	return len(r.Repository) > 0 || len(r.Registry) > 0
}

func (r Gitlab) Login(client docker.Client, out io.Writer) error {
	if ok, err := client.RegistryLogin(context.Background(), types.AuthConfig{Username: "gitlab-ci-token", Password: r.Token, ServerAddress: r.Registry}); err == nil {
		_, _ = fmt.Fprintln(out, ok.Status)
		return nil
	} else {
		return err
	}
}

func (r Gitlab) GetAuthInfo() string {
	auth := types.AuthConfig{Username: "gitlab-ci-token", Password: r.Token, ServerAddress: r.Registry}
	authBytes, _ := json.Marshal(auth)
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
