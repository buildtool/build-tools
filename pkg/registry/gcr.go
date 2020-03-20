package registry

import (
	"context"
	"docker.io/go-docker/api/types"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/buildtool/build-tools/pkg/docker"
	"io"
)

type GCR struct {
	dockerRegistry
	Url            string `yaml:"url" env:"GCR_URL"`
	KeyFileContent string `yaml:"keyfileContent" env:"GCR_KEYFILE_CONTENT"`
}

var _ Registry = &GCR{}

func (r *GCR) Name() string {
	return "GCR"
}

func (r *GCR) Configured() bool {
	if len(r.Url) > 0 {
		return true
	}
	return false
}

func (r *GCR) Login(client docker.Client, out io.Writer) error {
	auth := r.GetAuthConfig()
	auth.ServerAddress = r.Url
	if ok, err := client.RegistryLogin(context.Background(), auth); err == nil {
		_, _ = fmt.Fprintln(out, ok.Status)
		return nil
	} else {
		return err
	}
}

func (r *GCR) GetAuthConfig() types.AuthConfig {
	decoded, err := base64.StdEncoding.DecodeString(r.KeyFileContent)
	if err != nil {
		return types.AuthConfig{}
	}
	return types.AuthConfig{Username: "_json_key", Password: string(decoded)}
}

func (r *GCR) GetAuthInfo() string {
	authBytes, _ := json.Marshal(r.GetAuthConfig())
	return base64.URLEncoding.EncodeToString(authBytes)
}

func (r GCR) RegistryUrl() string {
	return r.Url
}

func (r GCR) Create(repository string) error {
	return nil
}
