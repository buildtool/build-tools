package config

import (
	"fmt"
	"github.com/caarlos0/env"
	"github.com/imdario/mergo"
	"github.com/liamg/tml"
	"github.com/sparetimecoders/build-tools/pkg/ci"
	"github.com/sparetimecoders/build-tools/pkg/config/scaffold"
	"github.com/sparetimecoders/build-tools/pkg/registry"
	"github.com/sparetimecoders/build-tools/pkg/vcs"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Config struct {
	VCS                 *VCSConfig             `yaml:"vcs"`
	CI                  *CIConfig              `yaml:"ci"`
	Registry            *RegistryConfig        `yaml:"registry"`
	Environments        map[string]Environment `yaml:"environments"`
	Scaffold            *scaffold.Config       `yaml:"scaffold"`
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
}

type Environment struct {
	Context    string `yaml:"context"`
	Namespace  string `yaml:"namespace"`
	Kubeconfig string `yaml:"kubeconfig"`
}

func Load(dir string, out io.Writer) (*Config, error) {
	cfg := InitEmptyConfig()

	if content, ok := os.LookupEnv("BUILDTOOLS_CONTENT"); ok {
		_, _ = fmt.Fprintln(out, "Parsing config from env: BUILDTOOLS_CONTENT")
		if err := parseConfig([]byte(content), cfg); err != nil {
			return cfg, err
		}
	} else {
		err := parseConfigFiles(dir, out, func(dir string) error {
			return parseConfigFile(dir, cfg)
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
		},
		Scaffold: scaffold.InitEmptyConfig(),
	}
	c.AvailableCI = []ci.CI{c.CI.Azure, c.CI.Buildkite, c.CI.Gitlab, c.CI.TeamCity, c.CI.Github}
	c.AvailableRegistries = []registry.Registry{c.Registry.Dockerhub, c.Registry.ECR, c.Registry.Github, c.Registry.Gitlab, c.Registry.Quay}
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

func (c *Config) CurrentEnvironment(environment string) (*Environment, error) {
	if e, exists := c.Environments[environment]; exists {
		return &e, nil
	}
	return nil, fmt.Errorf("no environment matching %s found", environment)
}

var abs = filepath.Abs

func parseConfigFiles(dir string, out io.Writer, fn func(string) error) error {
	parent, err := abs(dir)
	if err != nil {
		return err
	}
	var files []string
	for parent != "/" {
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

func parseConfigFile(filename string, cfg *Config) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	return parseConfig(data, cfg)
}

func parseConfig(content []byte, config *Config) error {
	temp := &Config{}
	if err := yaml.UnmarshalStrict(content, temp); err != nil {
		return err
	} else {
		if err := mergo.Merge(config, temp); err != nil {
			return err
		}
		return nil
	}
}
