package build

import (
	"bufio"
	"context"
	"docker.io/go-docker/api/types"
	"fmt"
	"gitlab.com/sparetimecoders/build-tools/pkg/ci"
	"gitlab.com/sparetimecoders/build-tools/pkg/docker"
	"gitlab.com/sparetimecoders/build-tools/pkg/registry"
	"gitlab.com/sparetimecoders/build-tools/pkg/vcs"
	"io"
	"os"
)

func Build(client docker.Client, buildContext io.ReadCloser, dockerfile string) error {
	dir, _ := os.Getwd()
	currentVCS := vcs.Identify(dir)
	currentCI, err := ci.Identify(currentVCS)
	if err != nil {
		return err
	}
	currentRegistry := registry.Identify()
	if currentRegistry == nil {
		return fmt.Errorf("no Docker registry found")
	}

	if err := currentRegistry.Login(client); err != nil {
		return err
	}

	tags := []string{
		docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), currentCI.Commit()),
		docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), currentCI.BranchReplaceSlash()),
	}
	if currentCI.Branch() == "master" {
		tags = append(tags, docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), "latest"))
	}
	// TODO: Parse Dockerfile and build and tag each stage for caching?
	response, err := client.ImageBuild(context.Background(), buildContext, types.ImageBuildOptions{
		Dockerfile: dockerfile,
		Memory:     3 * 1024 * 1024 * 1024,
		MemorySwap: -1,
		Remove:     true,
		ShmSize:    256 * 1024 * 1024,
		Tags:       tags,
	})

	if err != nil {
		return err
	} else {
		scanner := bufio.NewScanner(response.Body)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}

	return nil
}
