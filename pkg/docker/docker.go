package docker

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/apex/log"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
)

type Client interface {
	RegistryLogin(ctx context.Context, auth types.AuthConfig) (registry.AuthenticateOKBody, error)
	ImageBuild(ctx context.Context, buildContext io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error)
	ImagePush(ctx context.Context, image string, options types.ImagePushOptions) (io.ReadCloser, error)
	DialHijack(ctx context.Context, url, proto string, meta map[string][]string) (net.Conn, error)
	BuildCancel(ctx context.Context, id string) error
}

var _ Client = &client.Client{}

func Tag(registry, image, tag string) string {
	slug := SlugifyTag(tag)
	if slug != tag {
		log.Debugf("<yellow>Warning: tag was changed from '%s' to '%s' due to Dockers rules.</yellow>\n", tag, slug)
	}
	return fmt.Sprintf("%s/%s:%s", registry, image, slug)
}

func SlugifyTag(tag string) string {
	validChars := regexp.MustCompile(`(?i)[^a-zA-Z0-9.\-_]`)
	temp := validChars.ReplaceAllString(tag, "")
	leading := regexp.MustCompile(`^([.-]*)([a-zA-Z0-9.\-_]*)$`)
	result := leading.FindStringSubmatch(temp)[2]
	if len(result) > 128 {
		return result[:128]
	}
	return result
}

func ParseDockerignore(dir, dockerfile string) ([]string, error) {
	var defaultIgnore = []string{"k8s"}
	filePath := filepath.Join(dir, ".dockerignore")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return defaultIgnore, nil
	}

	if file, err := ioutil.ReadFile(filePath); err != nil {
		return defaultIgnore, err
	} else {
		var result = defaultIgnore
		scanner := bufio.NewScanner(bytes.NewReader(file))
		for scanner.Scan() {
			text := scanner.Text()
			if len(text) > 0 && text != dockerfile {
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
