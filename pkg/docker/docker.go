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
	"path/filepath"
	"regexp"
	"strings"
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

func ParseDockerignore(dir string) ([]string, error) {
	var defaultIgnore = []string{"k8s"}
	filePath := filepath.Join(dir, ".dockerignore")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return defaultIgnore, nil
	}

	if file, err := ioutil.ReadFile(filePath); err != nil {
		return defaultIgnore, err
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

func FindStages(content string) []string {
	var stages []string

	re := regexp.MustCompile(`(?i)^FROM .* AS (.*)$`)
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		text := scanner.Text()
		matches := re.FindStringSubmatch(text)
		if len(matches) != 0 {
			stages = append(stages, matches[1])
		}
	}
	return stages
}
