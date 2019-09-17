package build

import (
	"bufio"
	"bytes"
	"context"
	dkr "docker.io/go-docker"
	"docker.io/go-docker/api/types"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/docker/docker/pkg/archive"
	"github.com/liamg/tml"
	"github.com/pkg/errors"
	"gitlab.com/sparetimecoders/build-tools/pkg/config"
	"gitlab.com/sparetimecoders/build-tools/pkg/docker"
	"gitlab.com/sparetimecoders/build-tools/pkg/tar"
	"io"
	"regexp"
	"strings"
)

type responsetype struct {
	Stream      string `json:"stream"`
	ErrorDetail *struct {
		Code    int64  `json:"code"`
		Message string `json:"message"`
	} `json:"errorDetail"`
	Error string `json:"error"`
}

func DoBuild(dir string, out, eout io.Writer, args ...string) int {
	var dockerfile string
	const (
		defaultDockerfile = "Dockerfile"
		usage             = "name of the Dockerfile to use"
	)

	set := flag.NewFlagSet("build", flag.ExitOnError)
	set.StringVar(&dockerfile, "file", defaultDockerfile, usage)
	set.StringVar(&dockerfile, "f", defaultDockerfile, usage+" (shorthand)")
	_ = set.Parse(args)

	if client, err := dkr.NewEnvClient(); err != nil {
		_, _ = fmt.Fprintln(out, err.Error())
		return -1
	} else {
		if buildContext, err := createBuildContext(dir); err != nil {
			_, _ = fmt.Fprintln(out, err.Error())
			return -2
		} else {
			return build(client, dir, buildContext, dockerfile, out, eout)
		}
	}
}

func createBuildContext(dir string) (io.ReadCloser, error) {
	if ignored, err := docker.ParseDockerignore(dir); err != nil {
		return nil, err
	} else {
		return archive.TarWithOptions(dir, &archive.TarOptions{ExcludePatterns: ignored})
	}
}

func build(client docker.Client, dir string, buildContext io.ReadCloser, dockerfile string, out, eout io.Writer) int {
	cfg, err := config.Load(dir, out)
	if err != nil {
		_, _ = fmt.Fprintln(eout, err.Error())

		return -3
	}
	currentCI, err := cfg.CurrentCI()
	if err != nil {
		_, _ = fmt.Fprintln(eout, err.Error())
		return -4
	}
	_, _ = fmt.Fprintln(out, tml.Sprintf("Using CI <green>%s</green>\n", currentCI.Name()))

	currentRegistry, err := cfg.CurrentRegistry()
	if err != nil {
		_, _ = fmt.Fprintln(eout, err.Error())
		return -5
	} else {
		_, _ = fmt.Fprintln(out, tml.Sprintf("Using registry <green>%s</green>\n", currentRegistry.Name()))
		if err := currentRegistry.Login(client, out); err != nil {
			_, _ = fmt.Fprintln(eout, err.Error())
			return -6
		}
	}

	var buf bytes.Buffer
	tee := io.TeeReader(buildContext, &buf)
	stages, err := findStages(tee, dockerfile)
	if err != nil {
		_, _ = fmt.Fprintln(eout, err.Error())
		return -7
	}

	commit := currentCI.Commit()
	branch := currentCI.BranchReplaceSlash()
	_, _ = fmt.Fprintln(out, tml.Sprintf("Using build variables commit <green>%s</green> on branch <green>%s</green>\n", commit, branch))
	var caches []string

	args := map[string]*string{
		"CI_COMMIT": &commit,
		"CI_BRANCH": &branch,
	}
	for _, stage := range stages {
		tag := docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), stage)
		caches = append([]string{tag}, caches...)
		if err := doBuild(client, bytes.NewBuffer(buf.Bytes()), dockerfile, args, []string{tag}, caches, stage, out, eout); err != nil {
			return -8
		}
	}

	branchTag := docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), branch)
	latestTag := docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), "latest")
	tags := []string{
		docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), commit),
		branchTag,
	}
	if currentCI.Branch() == "master" {
		tags = append(tags, latestTag)
	}

	caches = append([]string{branchTag, latestTag}, caches...)
	if err := doBuild(client, bytes.NewBuffer(buf.Bytes()), dockerfile, args, tags, caches, "", out, eout); err != nil {
		return -9
	}

	return 0
}

func doBuild(client docker.Client, buildContext io.Reader, dockerfile string, args map[string]*string, tags, caches []string, target string, out, eout io.Writer) error {
	response, err := client.ImageBuild(context.Background(), buildContext, types.ImageBuildOptions{
		BuildArgs:  args,
		CacheFrom:  caches,
		Dockerfile: dockerfile,
		Memory:     3 * 1024 * 1024 * 1024,
		MemorySwap: -1,
		Remove:     true,
		ShmSize:    256 * 1024 * 1024,
		Tags:       tags,
		Target:     target,
	})

	if err != nil {
		return err
	} else {
		scanner := bufio.NewScanner(response.Body)
		for scanner.Scan() {
			r := &responsetype{}
			response := scanner.Bytes()
			if err := json.Unmarshal(response, &r); err != nil {
				_, _ = fmt.Fprintf(eout, "Unable to parse response: %s, Error: %v\n", string(response), err)
				return err
			} else {
				if r.ErrorDetail != nil {
					_, _ = fmt.Fprintf(eout, "Code: %v Message: %v\n", r.ErrorDetail.Code, r.ErrorDetail.Message)
					return errors.New(r.ErrorDetail.Message)
				} else {
					_, _ = fmt.Fprint(out, r.Stream)
				}
			}
		}
	}

	return nil
}

func findStages(buildContext io.Reader, dockerfile string) ([]string, error) {
	content, err := tar.ExtractFileContent(buildContext, dockerfile)
	if err != nil {
		return nil, err
	}
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
	return stages, nil
}
