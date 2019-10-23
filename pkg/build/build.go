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
	"github.com/sparetimecoders/build-tools/pkg"
	"github.com/sparetimecoders/build-tools/pkg/ci"
	"github.com/sparetimecoders/build-tools/pkg/config"
	"github.com/sparetimecoders/build-tools/pkg/docker"
	"github.com/sparetimecoders/build-tools/pkg/tar"
	"io"
	"os"
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
	if client, err := dockerClient(); err != nil {
		_, _ = fmt.Fprintln(out, err.Error())
		return -1
	} else {
		if buildContext, err := createBuildContext(dir); err != nil {
			_, _ = fmt.Fprintln(out, err.Error())
			return -2
		} else {
			return build(client, dir, buildContext, out, eout, args...)
		}
	}
}

var dockerClient = func() (docker.Client, error) {
	return dkr.NewEnvClient()
}

func createBuildContext(dir string) (io.ReadCloser, error) {
	if ignored, err := docker.ParseDockerignore(dir); err != nil {
		return nil, err
	} else {
		return archive.TarWithOptions(dir, &archive.TarOptions{ExcludePatterns: ignored})
	}
}

func build(client docker.Client, dir string, buildContext io.ReadCloser, out, eout io.Writer, args ...string) int {
	var dockerfile string
	var buildArgsFlags arrayFlags
	var skipLogin bool
	const (
		defaultDockerfile = "Dockerfile"
		usage             = "name of the Dockerfile to use"
	)

	set := flag.NewFlagSet("build", flag.ExitOnError)
	set.StringVar(&dockerfile, "file", defaultDockerfile, usage)
	set.StringVar(&dockerfile, "f", defaultDockerfile, usage+" (shorthand)")
	set.Var(&buildArgsFlags, "build-arg", "")
	set.BoolVar(&skipLogin, "skiplogin", false, "disable login to docker registry")

	_ = set.Parse(args)
	cfg, err := config.Load(dir, out)
	if err != nil {
		_, _ = fmt.Fprintln(eout, err.Error())
		return -3
	}
	currentCI := cfg.CurrentCI()
	_, _ = fmt.Fprintln(out, tml.Sprintf("Using CI <green>%s</green>", currentCI.Name()))

	currentRegistry := cfg.CurrentRegistry()
	_, _ = fmt.Fprintln(out, tml.Sprintf("Using registry <green>%s</green>", currentRegistry.Name()))
	if skipLogin {
		_, _ = fmt.Fprintln(out, tml.Sprintf("Login <yellow>disabled</yellow>"))
	} else {
		_, _ = fmt.Fprintln(out, tml.Sprintf("Authenticating against registry <green>%s</green>", currentRegistry.Name()))
		if err := currentRegistry.Login(client, out); err != nil {
			_, _ = fmt.Fprintln(eout, err.Error())
			return -4
		}
	}

	var buf bytes.Buffer
	tee := io.TeeReader(buildContext, &buf)
	stages, err := findStages(tee, dockerfile)
	if err != nil {
		_, _ = fmt.Fprintln(eout, err.Error())
		return -5
	}
	dockerTagOverride := os.Getenv("DOCKER_TAG")

	if !ci.IsValid(currentCI) && len(dockerTagOverride) == 0 {
		_, _ = fmt.Fprintln(eout, tml.Sprintf("Commit and/or branch information is <red>missing</red>. Perhaps your not in a Git repository or forgot to set environment variables?"))
		return -6
	}

	commit := currentCI.Commit()
	branch := currentCI.BranchReplaceSlash()
	_, _ = fmt.Fprintln(out, tml.Sprintf("Using build variables commit <green>%s</green> on branch <green>%s</green>", commit, branch))
	var caches []string

	buildArgs := map[string]*string{
		"CI_COMMIT": &commit,
		"CI_BRANCH": &branch,
	}
	for _, arg := range buildArgsFlags {
		split := strings.Split(arg, "=")
		key := split[0]
		value := strings.Join(split[1:], "=")
		if len(split) > 1 && len(value) > 0 {
			buildArgs[key] = pkg.String(value)
		} else {
			_, _ = fmt.Fprintf(out, "ignoring build-arg %s\n", key)
		}
	}
	for _, stage := range stages {
		tag := docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), stage)
		caches = append([]string{tag}, caches...)
		if err := doBuild(client, bytes.NewBuffer(buf.Bytes()), dockerfile, buildArgs, []string{tag}, caches, stage, out, eout); err != nil {
			_, _ = fmt.Fprintln(eout, err.Error())
			return -7
		}
	}

	var tags []string
	if len(dockerTagOverride) > 0 {
		tag := docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), dockerTagOverride)
		caches = append([]string{tag}, caches...)
		tags = append(tags, tag)
		_, _ = fmt.Fprintf(out, "overriding docker tags with value from env DOCKER_TAG %s\n", dockerTagOverride)
	} else {
		branchTag := docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), branch)
		latestTag := docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), "latest")
		tags = append(tags, []string{
			docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), commit),
			branchTag,
		}...)
		if currentCI.Branch() == "master" {
			tags = append(tags, latestTag)
		}

		caches = append([]string{branchTag, latestTag}, caches...)
	}
	if err := doBuild(client, bytes.NewBuffer(buf.Bytes()), dockerfile, buildArgs, tags, caches, "", out, eout); err != nil {
		_, _ = fmt.Fprintln(eout, err.Error())
		return -7
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
				return fmt.Errorf("unable to parse response: %s, Error: %v", string(response), err)
			} else {
				if r.ErrorDetail != nil {
					return fmt.Errorf("error Code: %v Message: %v", r.ErrorDetail.Code, r.ErrorDetail.Message)
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
	stages := docker.FindStages(content)

	return stages, nil
}

type arrayFlags []string

func (i *arrayFlags) String() string {
	// change this, this is just can example to satisfy the interface
	return "my string representation"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, strings.TrimSpace(value))
	return nil
}
