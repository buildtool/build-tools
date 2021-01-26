package config

import (
	"encoding/base64"
	"fmt"
	"github.com/buildtool/build-tools/pkg/ci"
	"github.com/buildtool/build-tools/pkg/registry"
	"github.com/buildtool/build-tools/pkg/vcs"
	"github.com/caarlos0/env"
	"github.com/imdario/mergo"
	"github.com/liamg/tml"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

type Config struct {
	VCS                 *VCSConfig        `yaml:"vcs"`
	CI                  *CIConfig         `yaml:"ci"`
	Registry            *RegistryConfig   `yaml:"registry"`
	Targets             map[string]Target `yaml:"targets"`
	AvailableCI         []ci.CI
	AvailableRegistries []registry.Registry
}

type VCSConfig struct {
	VCS vcs.VCS
}

type CIConfig struct {
	Azure     *ci.Azure     `yaml:"azure"`
	Buildkite *ci.Buildkite `yaml:"buildkite"`
	Gitlab    *ci.Gitlab    `yaml:"gitlab"`
	Github    *ci.Github    `yaml:"github"`
	TeamCity  *ci.TeamCity  `yaml:"teamcity"`
}

type RegistryConfig struct {
	Dockerhub *registry.Dockerhub `yaml:"dockerhub"`
	ECR       *registry.ECR       `yaml:"ecr"`
	Github    *registry.Github    `yaml:"github"`
	Gitlab    *registry.Gitlab    `yaml:"gitlab"`
	Quay      *registry.Quay      `yaml:"quay"`
	GCR       *registry.GCR       `yaml:"gcr"`
}

type Target struct {
	Context    string `yaml:"context"`
	Namespace  string `yaml:"namespace"`
	Kubeconfig string `yaml:"kubeconfig"`
}

func Load(dir string, out io.Writer) (*Config, error) {
	cfg := InitEmptyConfig()

	if content, ok := os.LookupEnv("BUILDTOOLS_CONTENT"); ok {
		_, _ = fmt.Fprintln(out, "Parsing config from env: BUILDTOOLS_CONTENT")
		if decoded, err := base64.StdEncoding.DecodeString(content); err != nil {
			return nil, errors.Wrap(err, "Failed to decode content")
		} else {
			if err := parseConfig(decoded, cfg); err != nil {
				if strings.Contains(string(decoded), "environments:") {
					_, _ = fmt.Fprintln(out, "BUILDTOOLS_CONTENT contains deprecated 'environments' tag, please change to 'targets'")
				}
				err = parseOldConfig(decoded, cfg)
				return cfg, err
			}
		}
	} else {
		err := parseConfigFiles(dir, out, func(dir string) error {
			return parseConfigFile(out, dir, cfg)
		})
		if err != nil {
			return cfg, err
		}
	}

	err := env.Parse(cfg)

	identifiedVcs := vcs.Identify(dir, out)
	cfg.VCS.VCS = identifiedVcs

	// TODO: Validate and clean config

	return cfg, err
}

func InitEmptyConfig() *Config {
	c := &Config{
		VCS: &VCSConfig{},
		CI: &CIConfig{
			Azure:     &ci.Azure{Common: &ci.Common{}},
			Buildkite: &ci.Buildkite{Common: &ci.Common{}},
			Gitlab:    &ci.Gitlab{Common: &ci.Common{}},
			Github:    &ci.Github{Common: &ci.Common{}},
			TeamCity:  &ci.TeamCity{Common: &ci.Common{}},
		},
		Registry: &RegistryConfig{
			Dockerhub: &registry.Dockerhub{},
			ECR:       &registry.ECR{},
			Github:    &registry.Github{},
			Gitlab:    &registry.Gitlab{},
			Quay:      &registry.Quay{},
			GCR:       &registry.GCR{},
		},
	}
	c.AvailableCI = []ci.CI{c.CI.Azure, c.CI.Buildkite, c.CI.Gitlab, c.CI.TeamCity, c.CI.Github}
	c.AvailableRegistries = []registry.Registry{c.Registry.Dockerhub, c.Registry.ECR, c.Registry.Github, c.Registry.Gitlab, c.Registry.Quay, c.Registry.GCR}
	return c
}

func (c *Config) CurrentVCS() vcs.VCS {
	return c.VCS.VCS
}

func (c *Config) CurrentCI() ci.CI {
	for _, x := range c.AvailableCI {
		if x.Configured() {
			x.SetVCS(c.CurrentVCS())
			return x
		}
	}
	x := &ci.No{Common: &ci.Common{}}
	x.SetVCS(c.CurrentVCS())
	return x
}

func (c *Config) CurrentRegistry() registry.Registry {
	for _, reg := range c.AvailableRegistries {
		if reg.Configured() {
			return reg
		}
	}
	return registry.NoDockerRegistry{}
}

func (c *Config) CurrentTarget(target string) (*Target, error) {
	if e, exists := c.Targets[target]; exists {
		return &e, nil
	}
	return nil, fmt.Errorf("no target matching %s found", target)
}

var abs = filepath.Abs

func parseConfigFiles(dir string, out io.Writer, fn func(string) error) error {
	parent, err := abs(dir)
	if err != nil {
		return err
	}
	var files []string
	for !strings.HasSuffix(filepath.Clean(parent), string(os.PathSeparator)) {
		filename := filepath.Join(parent, ".buildtools.yaml")
		if _, err := os.Stat(filename); !os.IsNotExist(err) {
			files = append(files, filename)
		}

		parent = filepath.Dir(parent)
	}
	for i, file := range files {
		if i == 0 {
			_, _ = fmt.Fprintln(out, tml.Sprintf("Parsing config from file: <green>'%s'</green>", file))
		} else {
			_, _ = fmt.Fprintln(out, tml.Sprintf("Merging with config from file: <green>'%s'</green>", file))
		}
		if err := fn(file); err != nil {
			return err
		}
	}

	return nil
}

func parseConfigFile(out io.Writer, filename string, cfg *Config) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	if strings.Contains(string(data), "environments:") {
		_, _ = fmt.Fprintln(out, tml.Sprintf("file: <green>'%s'</green> <red>contains deprecated 'environments' tag, please change to 'targets'</red>", filename))
		return parseOldConfig(data, cfg)
	}

	return parseConfig(data, cfg)
}

func parseOldConfig(content []byte, config *Config) error {
	oldConfig := &struct {
		VCS      *VCSConfig        `yaml:"vcs"`
		CI       *CIConfig         `yaml:"ci"`
		Registry *RegistryConfig   `yaml:"registry"`
		Targets  map[string]Target `yaml:"environments"`
	}{}
	if err := yaml.UnmarshalStrict(content, oldConfig); err != nil {
		return err
	} else {
		if err := mergo.Merge(config, Config{
			Registry: oldConfig.Registry,
			Targets:  oldConfig.Targets,
			CI:       oldConfig.CI,
			VCS:      oldConfig.VCS,
		}); err != nil {
			return err
		}
		return validate(config)
	}
}
func parseConfig(content []byte, config *Config) error {
	temp := &Config{}
	if err := yaml.UnmarshalStrict(content, temp); err != nil {
		return err
	} else {
		if err := mergo.Merge(config, temp); err != nil {
			return err
		}
		return validate(config)
	}
}

func validate(config *Config) error {
	elem := reflect.ValueOf(config.Registry).Elem()
	found := false
	for i := 0; i < elem.NumField(); i++ {
		f := elem.Field(i)
		if f.Interface().(registry.Registry).Configured() {
			if found {
				return fmt.Errorf("registry already defined, please check configuration")
			}
			found = true
		}
	}

	return nil
}
