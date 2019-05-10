package config

import (
	"bufio"
	"context"
	"docker.io/go-docker/api/types"
	"fmt"
	"gitlab.com/sparetimecoders/build-tools/pkg/docker"
)

type Registry interface {
	configured() bool
	Login(client docker.Client) error
	GetAuthInfo() string
	RegistryUrl() string
	Create(repository string) error
	PushImage(client docker.Client, auth, image string) error
	// TODO: Uncomment when implementing service-setup
	//Validate() bool
}

type dockerRegistry struct{}

func (dockerRegistry) PushImage(client docker.Client, auth, image string) error {
	if out, err := client.ImagePush(context.Background(), image, types.ImagePushOptions{All: true, RegistryAuth: auth}); err != nil {
		return err
	} else {
		scanner := bufio.NewScanner(out)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}

		return nil
	}
}
