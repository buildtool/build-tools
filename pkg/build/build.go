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

	"github.com/apex/log"
	"github.com/docker/docker/api/types"
	dkr "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"

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

func DoBuild(dir string, info version.Info, osArgs ...string) int {
	var buildArgs Args
	// TODO See if we can move this "up" one level to remove out,eout completely
	err := args.ParseArgs(dir, osArgs, info, &buildArgs)
	if err != nil {
		if err != args.Done {
			return -1
		} else {
			return 0
		}
	}

	if client, err := dockerClient(); err != nil {
		log.Error(err.Error())
		return -1
	} else {
		if buildContext, err := createBuildContext(dir, buildArgs.Dockerfile); err != nil {
			log.Error(err.Error())
			return -2
		} else {
			return build(client, dir, buildContext, buildArgs)
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

func build(client docker.Client, dir string, buildContext io.ReadCloser, buildVars Args) int {
	cfg, err := config.Load(dir)
	if err != nil {
		log.Error(err.Error())
		return -3
	}
	currentCI := cfg.CurrentCI()
	log.Debugf("Using CI <green>%s</green>\n", currentCI.Name())

	currentRegistry := cfg.CurrentRegistry()
	log.Debugf("Using registry <green>%s</green>\n", currentRegistry.Name())
	authConfigs := make(map[string]types.AuthConfig)
	if buildVars.NoLogin {
		log.Debugf("Login <yellow>disabled</yellow>\n")
	} else {
		log.Debugf("Authenticating against registry <green>%s</green>\n", currentRegistry.Name())
		if err := currentRegistry.Login(client); err != nil {
			log.Error(err.Error())
			return -4
		}
		authConfigs[currentRegistry.RegistryUrl()] = currentRegistry.GetAuthConfig()
	}

	var buf bytes.Buffer
	tee := io.TeeReader(buildContext, &buf)
	stages, err := findStages(tee, buildVars.Dockerfile)
	if err != nil {
		log.Error(err.Error())
		return -5
	}
	if !ci.IsValid(currentCI) {
		log.Debugf("Commit and/or branch information is <red>missing</red>. Perhaps your not in a Git repository or forgot to set environment variables?\n")
		return -6
	}

	commit := currentCI.Commit()
	branch := currentCI.BranchReplaceSlash()
	log.Debugf("Using build variables commit <green>%s</green> on branch <green>%s</green>\n", commit, branch)
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
				log.Debugf("ignoring build-arg %s\n", key)
			}
		}
	}
	for _, stage := range stages {
		tag := docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), stage)
		caches = append([]string{tag}, caches...)
		if err := doBuild(client, bytes.NewBuffer(buf.Bytes()), buildVars.Dockerfile, buildArgs, []string{tag}, caches, stage, authConfigs, !buildVars.NoPull); err != nil {
			log.Error(err.Error())
			return -7
		}
	}

	var tags []string
	branchTag := docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), branch)
	latestTag := docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), "latest")
	tags = append(tags, []string{
		docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), commit),
		branchTag,
	}...)
	if currentCI.Branch() == "master" || currentCI.Branch() == "main" {
		tags = append(tags, latestTag)
	}

	caches = append([]string{branchTag, latestTag}, caches...)
	if err := doBuild(client, bytes.NewBuffer(buf.Bytes()), buildVars.Dockerfile, buildArgs, tags, caches, "", authConfigs, !buildVars.NoPull); err != nil {
		log.Error(err.Error())
		return -7
	}

	return 0
}

func doBuild(client docker.Client, buildContext io.Reader, dockerfile string, args map[string]*string, tags, caches []string, target string, authConfigs map[string]types.AuthConfig, pullParent bool) error {
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
					log.Info(r.Stream)
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
