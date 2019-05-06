package registry

import (
  "github.com/stretchr/testify/assert"
  "os"
  "testing"
)

func TestIdentify_Dockerhub(t *testing.T) {
  os.Clearenv()
  _ = os.Setenv("DOCKERHUB_REPOSITORY", "repo")
  _ = os.Setenv("DOCKERHUB_USERNAME", "user")
  _ = os.Setenv("DOCKERHUB_PASSWORD", "pass")

  docker := &MockDocker{}
  result := Identify()
  assert.NotNil(t, result)
  assert.Equal(t, "repo", result.RegistryUrl())
  result.Login(docker)
  assert.Equal(t, "user", docker.Username)
  assert.Equal(t, "pass", docker.Password)
  assert.Equal(t, "", docker.ServerAddress)
}
