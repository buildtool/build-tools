package registry

import (
  "context"
  "docker.io/go-docker/api/types"
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

func (r dockerhub) Login(client DockerClient) bool {
  if ok, err := client.RegistryLogin(context.Background(), types.AuthConfig{Username: r.username, Password: r.password}); err == nil {
    log.Println(ok.Status)
    return true
  } else {
    panic(err)
    return false
  }
}

func (r dockerhub) RegistryUrl() string {
  return r.repository
}

func (r dockerhub) Create() bool {
  panic("implement me")
}

func (r dockerhub) Validate() bool {
  panic("implement me")
}
