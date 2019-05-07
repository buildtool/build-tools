package registry

import (
  "context"
  "docker.io/go-docker/api/types"
  "gitlab.com/sparetimecoders/build-tools/pkg/docker"
  "log"
  "os"
)

type dockerhub struct {
  repository string
  username   string
  password   string
}

var _ Registry = &dockerhub{}

func (r *dockerhub) identify() bool {
  if repository, exists := os.LookupEnv("DOCKERHUB_REPOSITORY"); exists {
    log.Println("Will use Dockerhub as container registry")
    r.repository = repository
    r.username = os.Getenv("DOCKERHUB_USERNAME")
    r.password = os.Getenv("DOCKERHUB_PASSWORD")
    return true
  }
  return false
}

func (r dockerhub) Login(client docker.Client) error {
  if ok, err := client.RegistryLogin(context.Background(), types.AuthConfig{Username: r.username, Password: r.password}); err == nil {
    log.Println(ok.Status)
    return nil
  } else {
    log.Println("Unable to login")
    return err
  }
}

func (r dockerhub) RegistryUrl() string {
  return r.repository
}
