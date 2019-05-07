package registry

import (
	"context"
	"docker.io/go-docker/api/types"
	"fmt"
	"gitlab.com/sparetimecoders/build-tools/pkg/docker"
	"log"
	"os"
)

type quay struct {
	url      string
	username string
	password string
}

var _ Registry = &quay{}

func (r *quay) identify() bool {
	if url, exists := os.LookupEnv("QUAY_REPOSITORY"); exists {
		log.Println("Will use Quay.io as container registry")
		r.url = fmt.Sprintf("quay.io/%s", url)
		r.username = os.Getenv("QUAY_USERNAME")
		r.password = os.Getenv("QUAY_PASSWORD")
		return true
	}
	return false
}

func (r quay) Login(client docker.Client) bool {
	if ok, err := client.RegistryLogin(context.Background(), types.AuthConfig{Username: r.username, Password: r.password, ServerAddress: "quay.io"}); err == nil {
		log.Println(ok.Status)
		return true
	} else {
		panic(err)
	}
}

func (r quay) RegistryUrl() string {
	return r.url
}

func (r quay) Create() bool {
	panic("implement me")
}

func (r quay) Validate() bool {
	panic("implement me")
}
