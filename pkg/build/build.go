package build

import (
	"bufio"
	"context"
	dkr "docker.io/go-docker"
	"docker.io/go-docker/api/types"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/docker/docker/pkg/archive"
	"github.com/pkg/errors"
	"gitlab.com/sparetimecoders/build-tools/pkg/config"
	"gitlab.com/sparetimecoders/build-tools/pkg/docker"
	"io"
)

type responsetype struct {
	Stream      string `json:"stream"`
	ErrorDetail *struct {
		Code    int64  `json:"code"`
		Message string `json:"message"`
	} `json:"errorDetail"`
	Error string `json:"error"`
}

func DoBuild(dir string, out, eout io.Writer, exit func(code int), args ...string) int {
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
		exit(-1)
	} else {
		if buildContext, err := createBuildContext(dir); err != nil {
			_, _ = fmt.Fprintln(out, err.Error())
			exit(-2)
		} else {
			err = build(client, dir, buildContext, dockerfile, out, eout)
			if err != nil {
				fmt.Println(err.Error())
			}
		}
	}

	return 0
}

func createBuildContext(dir string) (io.ReadCloser, error) {
	if ignored, err := docker.ParseDockerignore(dir); err != nil {
		return nil, err
	} else {
		return archive.TarWithOptions(".", &archive.TarOptions{ExcludePatterns: ignored})
	}
}

func build(client docker.Client, dir string, buildContext io.ReadCloser, dockerfile string, out, eout io.Writer) error {
	cfg, err := config.Load(dir, out)
	if err != nil {
		return err
	}
	currentCI := cfg.CurrentCI()
	currentRegistry, err := cfg.CurrentRegistry()
	if err != nil {
		return err
	}

	if err := currentRegistry.Login(client, out); err != nil {
		return err
	}

	commit := currentCI.Commit()
	branch := currentCI.BranchReplaceSlash()
	tags := []string{
		docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), commit),
		docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), branch),
	}
	if currentCI.Branch() == "master" {
		tags = append(tags, docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), "latest"))
	}
	args := map[string]*string{
		"CI_COMMIT": &commit,
		"CI_BRANCH": &branch,
	}
	// TODO: Parse Dockerfile and build and tag each stage for caching?
	response, err := client.ImageBuild(context.Background(), buildContext, types.ImageBuildOptions{
		BuildArgs:  args,
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
