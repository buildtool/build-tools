package registry

import (
	"context"
	"docker.io/go-docker/api/types"
	"encoding/base64"
	"encoding/json"
	"gitlab.com/sparetimecoders/build-tools/pkg/docker"
	"log"
	"os"
	"strings"
)

type gitlab struct {
	url   string
	token string
}

var _ Registry = &gitlab{}

func (r *gitlab) identify() bool {
	if _, exists := os.LookupEnv("CI_REGISTRY_IMAGE"); exists {
		log.Println("Will use Gitlab as container registry")
		registry := os.Getenv("CI_REGISTRY")
		r.url = registry[:strings.LastIndex(registry, "/")]
		r.token = os.Getenv("CI_BUILD_TOKEN")
		return true
	}
	return false
}

func (r gitlab) Login(client docker.Client) error {
	if ok, err := client.RegistryLogin(context.Background(), types.AuthConfig{Username: "gitlab-ci-token", Password: r.token, ServerAddress: r.url}); err == nil {
		log.Println(ok.Status)
		return nil
	} else {
		return err
	}
}

func (r gitlab) GetAuthInfo() string {
	auth := types.AuthConfig{Username: "gitlab-ci-token", Password: r.token, ServerAddress: r.url}
	authBytes, _ := json.Marshal(auth)
	return base64.URLEncoding.EncodeToString(authBytes)
}

func (r gitlab) RegistryUrl() string {
	return r.url
}

func (r *gitlab) Create() error {
	return nil
}
