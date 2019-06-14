package docker

import (
	"bufio"
	"bytes"
	"context"
	"docker.io/go-docker"
	"docker.io/go-docker/api/types"
	"docker.io/go-docker/api/types/registry"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

type Client interface {
	RegistryLogin(ctx context.Context, auth types.AuthConfig) (registry.AuthenticateOKBody, error)
	ImageBuild(ctx context.Context, buildContext io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error)
	ImagePush(ctx context.Context, image string, options types.ImagePushOptions) (io.ReadCloser, error)
}

var _ Client = &docker.Client{}

func Tag(registry, image, tag string) string {
	return fmt.Sprintf("%s/%s:%s", registry, image, tag)
}

func ParseDockerignore() ([]string, error) {
	var empty []string
	if _, err := os.Stat(".dockerignore"); os.IsNotExist(err) {
		return empty, nil
	}
	if file, err := ioutil.ReadFile(".dockerignore"); err != nil {
		return empty, err
	} else {
		var result []string
		scanner := bufio.NewScanner(bytes.NewReader(file))
		for scanner.Scan() {
			text := scanner.Text()
			if len(text) > 0 {
				result = append(result, text)
			}
		}
		return result, nil
	}
}
