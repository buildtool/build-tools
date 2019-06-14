package config

import (
	"bufio"
	"context"
	"docker.io/go-docker/api/types"
	"fmt"
	"gitlab.com/sparetimecoders/build-tools/pkg/docker"
	"io"
)

type Registry interface {
	configured() bool
	Login(client docker.Client, out io.Writer) error
	GetAuthInfo() string
	RegistryUrl() string
	Create(repository string) error
	PushImage(client docker.Client, auth, image string, out io.Writer) error
	// TODO: Uncomment when implementing service-setup
	//Validate() bool
}

type dockerRegistry struct {
	CI CI
}

func (r *dockerRegistry) setVCS(cfg Config) {
	r.CI = cfg.CurrentCI()
}

func (dockerRegistry) PushImage(client docker.Client, auth, image string, ow io.Writer) error {
	if out, err := client.ImagePush(context.Background(), image, types.ImagePushOptions{All: true, RegistryAuth: auth}); err != nil {
		return err
	} else {
		scanner := bufio.NewScanner(out)
		for scanner.Scan() {
			_, _ = fmt.Fprintln(ow, scanner.Text())
		}

		return nil
	}
}
