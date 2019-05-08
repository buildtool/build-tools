package docker

import (
	"context"
	"docker.io/go-docker"
	"docker.io/go-docker/api/types"
	"docker.io/go-docker/api/types/registry"
	"fmt"
	"io"
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
