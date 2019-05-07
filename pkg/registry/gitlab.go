package registry

import (
	"context"
	"docker.io/go-docker/api/types"
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

func (r gitlab) Login(client docker.Client) bool {
	if ok, err := client.RegistryLogin(context.Background(), types.AuthConfig{Username: "gitlab-ci-token", Password: r.token, ServerAddress: r.url}); err == nil {
		log.Println(ok.Status)
		return true
	} else {
		panic(err)
	}
}

func (r gitlab) RegistryUrl() string {
	return r.url
}

func (r gitlab) Create() bool {
	panic("implement me")
}

func (r gitlab) Validate() bool {
	panic("implement me")
}
