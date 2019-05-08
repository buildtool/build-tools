package registry

import (
	"bufio"
	"context"
	"docker.io/go-docker/api/types"
	"fmt"
	"gitlab.com/sparetimecoders/build-tools/pkg/docker"
)

type Registry interface {
	identify() bool
	Login(client docker.Client) error
	GetAuthInfo() string
	RegistryUrl() string
	Create(repository string) error
	// TODO: Uncomment when implementing service-setup
	//Validate() bool
}

var registries = []Registry{&dockerhub{}, &ecr{}, &gitlab{}, &quay{}}

func Identify() Registry {
	for _, reg := range registries {
		if reg.identify() {
			return reg
		}
	}
	return nil
}

func PushImage(client docker.Client, auth, image string) error {
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
