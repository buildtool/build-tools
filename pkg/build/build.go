package build

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	dkr "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/liamg/tml"

	"github.com/buildtool/build-tools/pkg/args"
	"github.com/buildtool/build-tools/pkg/ci"
	"github.com/buildtool/build-tools/pkg/config"
	"github.com/buildtool/build-tools/pkg/docker"
	"github.com/buildtool/build-tools/pkg/tar"
	"github.com/buildtool/build-tools/pkg/version"
)

type responsetype struct {
	Stream      string `json:"stream"`
	ErrorDetail *struct {
		Code    int64  `json:"code"`
		Message string `json:"message"`
	} `json:"errorDetail"`
	Error string `json:"error"`
}

type Args struct {
	args.Globals
	Dockerfile string   `name:"file" short:"f" help:"name of the Dockerfile to use." default:"Dockerfile"`
	BuildArgs  []string `name:"build-arg" type:"list" help:"additional docker build-args to use, see https://docs.docker.com/engine/reference/commandline/build/ for more information."`
	NoLogin    bool     `help:"disable login to docker registry" default:"false" `
	NoPull     bool     `help:"disable pulling latest from docker registry" default:"false"`
}

func DoBuild(dir string, out, eout io.Writer, info version.Info, osArgs ...string) int {
	var buildArgs Args
	err := args.ParseArgs(out,
		eout,
		osArgs,
		info,
		&buildArgs)
	if err != nil {
		if err != args.Done {
			return -1
		} else {
			return 0
		}
	}

	if client, err := dockerClient(); err != nil {
		_, _ = fmt.Fprintln(out, err.Error())
		return -1
	} else {
		if buildContext, err := createBuildContext(dir, buildArgs.Dockerfile); err != nil {
			_, _ = fmt.Fprintln(out, err.Error())
			return -2
		} else {
			return build(client, dir, buildContext, out, eout, buildArgs)
		}
	}
}

var dockerClient = func() (docker.Client, error) {
	return dkr.NewClientWithOpts(dkr.FromEnv)
}

func createBuildContext(dir, dockerfile string) (io.ReadCloser, error) {
	if ignored, err := docker.ParseDockerignore(dir, dockerfile); err != nil {
		return nil, err
	} else {
		return archive.TarWithOptions(dir, &archive.TarOptions{ExcludePatterns: ignored})
	}
}

func build(client docker.Client, dir string, buildContext io.ReadCloser, out, eout io.Writer, buildVars Args) int {
	cfg, err := config.Load(dir, out)
	if err != nil {
		_, _ = fmt.Fprintln(eout, err.Error())
		return -3
	}
	currentCI := cfg.CurrentCI()
	_, _ = fmt.Fprintln(out, tml.Sprintf("Using CI <green>%s</green>", currentCI.Name()))

	currentRegistry := cfg.CurrentRegistry()
	_, _ = fmt.Fprintln(out, tml.Sprintf("Using registry <green>%s</green>", currentRegistry.Name()))
	authConfigs := make(map[string]types.AuthConfig)
	if buildVars.NoLogin {
		_, _ = fmt.Fprintln(out, tml.Sprintf("Login <yellow>disabled</yellow>"))
	} else {
		_, _ = fmt.Fprintln(out, tml.Sprintf("Authenticating against registry <green>%s</green>", currentRegistry.Name()))
		if err := currentRegistry.Login(client, out); err != nil {
			_, _ = fmt.Fprintln(eout, err.Error())
			return -4
		}
		authConfigs[currentRegistry.RegistryUrl()] = currentRegistry.GetAuthConfig()
	}

	var buf bytes.Buffer
	tee := io.TeeReader(buildContext, &buf)
	stages, err := findStages(tee, buildVars.Dockerfile)
	if err != nil {
		_, _ = fmt.Fprintln(eout, err.Error())
		return -5
	}
	if !ci.IsValid(currentCI) {
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
	for _, arg := range buildVars.BuildArgs {
		split := strings.Split(arg, "=")
		key := split[0]
		value := strings.Join(split[1:], "=")
		if len(split) > 1 && len(value) > 0 {
			buildArgs[key] = &value
		} else {
			if env, exists := os.LookupEnv(key); exists {
				buildArgs[key] = &env
			} else {
				_, _ = fmt.Fprintf(out, "ignoring build-arg %s\n", key)
			}
		}
	}
	for _, stage := range stages {
		tag := docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), stage, eout)
		caches = append([]string{tag}, caches...)
		if err := doBuild(client, bytes.NewBuffer(buf.Bytes()), buildVars.Dockerfile, buildArgs, []string{tag}, caches, stage, authConfigs, out, !buildVars.NoPull); err != nil {
			_, _ = fmt.Fprintln(eout, err.Error())
			return -7
		}
	}

	var tags []string
	branchTag := docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), branch, eout)
	latestTag := docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), "latest", eout)
	tags = append(tags, []string{
		docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), commit, eout),
		branchTag,
	}...)
	if currentCI.Branch() == "master" || currentCI.Branch() == "main" {
		tags = append(tags, latestTag)
	}

	caches = append([]string{branchTag, latestTag}, caches...)
	if err := doBuild(client, bytes.NewBuffer(buf.Bytes()), buildVars.Dockerfile, buildArgs, tags, caches, "", authConfigs, out, !buildVars.NoPull); err != nil {
		_, _ = fmt.Fprintln(eout, err.Error())
		return -7
	}

	return 0
}

func doBuild(client docker.Client, buildContext io.Reader, dockerfile string, args map[string]*string, tags, caches []string, target string, authConfigs map[string]types.AuthConfig, out io.Writer, pullParent bool) error {
	response, err := client.ImageBuild(context.Background(), buildContext, types.ImageBuildOptions{
		AuthConfigs: authConfigs,
		BuildArgs:   args,
		CacheFrom:   caches,
		Dockerfile:  dockerfile,
		PullParent:  pullParent,
		MemorySwap:  -1,
		Remove:      true,
		ShmSize:     256 * 1024 * 1024,
		Tags:        tags,
		Target:      target,
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
