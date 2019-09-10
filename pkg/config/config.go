package config

import (
	"errors"
	"fmt"
	"github.com/caarlos0/env"
	"github.com/liamg/tml"
	"gitlab.com/sparetimecoders/build-tools/pkg/stack"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Config struct {
	VCS          *VCSConfig      `yaml:"vcs"`
	CI           *CIConfig       `yaml:"ci"`
	Registry     *RegistryConfig `yaml:"registry"`
	Environments []Environment   `yaml:"environments"`
}

type VCSConfig struct {
	Selected string     `yaml:"selected" env:"VCS"`
	Azure    *AzureVCS  `yaml:"azure"`
	Github   *GithubVCS `yaml:"github"`
	Gitlab   *GitlabVCS `yaml:"gitlab"`
	VCS      *VCS
}

type CIConfig struct {
	Selected  string       `yaml:"selected" env:"CI"`
	Azure     *AzureCI     `yaml:"azure"`
	Buildkite *BuildkiteCI `yaml:"buildkite"`
	Gitlab    *GitlabCI    `yaml:"gitlab"`
}

type RegistryConfig struct {
	Selected  string             `yaml:"selected" env:"REGISTRY"`
	Dockerhub *DockerhubRegistry `yaml:"dockerhub"`
	ECR       *ECRRegistry       `yaml:"ecr"`
	Gitlab    *GitlabRegistry    `yaml:"gitlab"`
	Quay      *QuayRegistry      `yaml:"quay"`
}

type Environment struct {
	Name      string `yaml:"name"`
	Context   string `yaml:"context"`
	Namespace string `yaml:"namespace"`
}

func Load(dir string, out io.Writer) (*Config, error) {
	cfg := initEmptyConfig()

	err := parseConfigFiles(dir, out, func(dir string) error {
		return parseConfigFile(dir, cfg)
	})
	if err != nil {
		return cfg, err
	}

	err = env.Parse(cfg)

	vcs := Identify(dir, out)
	cfg.VCS.VCS = &vcs

	// TODO: Validate and clean config

	return cfg, err
}

func initEmptyConfig() *Config {
	return &Config{
		VCS: &VCSConfig{
			Azure:  &AzureVCS{},
			Github: &GithubVCS{},
			Gitlab: &GitlabVCS{},
		},
		CI: &CIConfig{
			Azure:     &AzureCI{ci: &ci{}},
			Buildkite: &BuildkiteCI{ci: &ci{}},
			Gitlab:    &GitlabCI{ci: &ci{}},
		},
		Registry: &RegistryConfig{
			Dockerhub: &DockerhubRegistry{},
			ECR:       &ECRRegistry{},
			Gitlab:    &GitlabRegistry{},
			Quay:      &QuayRegistry{},
		},
	}
}

func (c *Config) CurrentVCS() VCS {
	switch c.VCS.Selected {
	case "azure":
		return c.VCS.Azure
	case "github":
		return c.VCS.Github
	case "gitlab":
		return c.VCS.Gitlab
	}
	return *c.VCS.VCS
}

func (c *Config) CurrentCI() CI {
	switch c.CI.Selected {
	case "azure":
		c.CI.Azure.setVCS(*c)
		return c.CI.Azure
	case "buildkite":
		c.CI.Buildkite.setVCS(*c)
		return c.CI.Buildkite
	case "gitlab":
		c.CI.Gitlab.setVCS(*c)
		return c.CI.Gitlab
	case "":
		vals := []CI{c.CI.Azure, c.CI.Buildkite, c.CI.Gitlab}
		for _, ci := range vals {
			if ci.configured() {
				ci.setVCS(*c)
				return ci
			}
		}
	}
	ci := &noOpCI{ci: &ci{}}
	ci.setVCS(*c)
	return ci
}

func (c *Config) CurrentRegistry() (Registry, error) {
	switch c.Registry.Selected {
	case "dockerhub":
		c.Registry.Dockerhub.setVCS(*c)
		return c.Registry.Dockerhub, nil
	case "ecr":
		c.Registry.ECR.setVCS(*c)
		return c.Registry.ECR, nil
	case "gitlab":
		c.Registry.Gitlab.setVCS(*c)
		return c.Registry.Gitlab, nil
	case "quay":
		c.Registry.Quay.setVCS(*c)
		return c.Registry.Quay, nil
	case "":
		vals := []Registry{c.Registry.Dockerhub, c.Registry.ECR, c.Registry.Gitlab, c.Registry.Quay}
		for _, reg := range vals {
			if reg.configured() {
				return reg, nil
			}
		}
	}
	return nil, errors.New("no Docker registry found")
}

func (c *Config) CurrentEnvironment(environment string) (*Environment, error) {
	for _, e := range c.Environments {
		if e.Name == environment {
			return &e, nil
		}
	}
	return nil, fmt.Errorf("no environment matching %s found", environment)
}

func (c *Config) Scaffold(name string, stack stack.Stack, out io.Writer, exit func(code int)) {
	vcs := c.CurrentVCS()
	ci := c.CurrentCI()
	if registry, err := c.CurrentRegistry(); err != nil {
		_, _ = fmt.Fprintln(out, tml.Sprintf("<red>%s</red>", err.Error()))
		exit(-2)
	} else {
		if err := validate(); err != nil {
			_, _ = fmt.Fprintln(out, tml.Sprintf("<red>%s</red>", err.Error()))
			exit(-3)
		} else {
			_, _ = fmt.Fprint(out, tml.Sprintf("<lightblue>Creating new service </lightblue><white><bold>'%s'</bold></white> <lightblue>using stack </lightblue><white><bold>'%s'</bold></white>\n", name, stack.Name()))
			_, _ = fmt.Fprint(out, tml.Sprintf("<lightblue>Creating repository at </lightblue><white><bold>'%s'</bold></white>\n", vcs.Name()))
			repository, _ := vcs.Scaffold(name)
			_, _ = fmt.Fprint(out, tml.Sprintf("<green>Created repository </green><white><bold>'%s'</bold></white>\n", repository))
			if err := vcs.Clone(name, repository, out); err != nil {
				_, _ = fmt.Fprintln(out, tml.Sprintf("<red>%s</red>", err.Error()))
				exit(-4)
			}
			createDirectories()
			_, _ = fmt.Fprint(out, tml.Sprintf("<lightblue>Creating build pipeline for </lightblue><white><bold>'%s'</bold></white>\n", name))
			if webhook := ci.Scaffold(name, repository); webhook != nil {
				vcs.Webhook(name, *webhook)
			}
			createDotfiles()
			createReadme(name)
			createDeployment(name, registry)
			if err := stack.Scaffold(name); err != nil {
				_, _ = fmt.Fprintln(out, tml.Sprintf("<red>%s</red>", err.Error()))
				exit(-5)
			}
		}
	}
}

func validate() error {
	return nil
}

func createDirectories() {

}

func createDotfiles() {}

func createReadme(name string) {}

func createDeployment(name string, registry Registry) {}

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
	for i := len(files) - 1; i >= 0; i-- {
		_, _ = fmt.Fprintf(out, "Parsing config from file: '%s'\n", files[i])
		if err := fn(files[i]); err != nil {
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
	if err := yaml.UnmarshalStrict(content, &config); err != nil {
		return err
	} else {
		//var missingFields []string
		//
		//if err = handleDefaultValues(reflect.ValueOf(config).Elem(), &missingFields, ""); err != nil {
		//	panic(err)
		//}

		//if len(missingFields) != 0 {
		//	return errors.New(fmt.Sprintf("Missing required value for field(s): '%v'\n", missingFields))
		//}
		return nil
	}
}

//func handleDefaultValues(t reflect.Value, missingFields *[]string, prefix string) error {
//	refType := t.Type()
//	for i := 0; i < refType.NumField(); i++ {
//		name := strings.TrimPrefix(fmt.Sprintf("%s.%s", prefix, refType.Field(i).Name), ".")
//		value := t.Field(i)
//		defaultValue := refType.Field(i).Tag.Get("default")
//		mandatory := refType.Field(i).Tag.Get("optional") != "true"
//		if value.Kind() == reflect.Struct {
//			if err := handleDefaultValues(value, missingFields, name); err != nil {
//				return err
//			}
//		} else if value.Kind() == reflect.Ptr && !value.IsNil() {
//			if err := handleDefaultValues(value.Elem(), missingFields, name); err != nil {
//				return err
//			}
//		} else if isZeroOfUnderlyingType(value) && mandatory {
//			if defaultValue == "" {
//				*missingFields = append(*missingFields, name)
//			} else {
//				log.Printf("Setting default value for field '%s' = '%s'", name, defaultValue)
//				if err := set(value, name, defaultValue, missingFields); err != nil {
//					return err
//				}
//			}
//		}
//	}
//
//	return nil
//}

//func isZeroOfUnderlyingType(x reflect.Value) bool {
//	return reflect.DeepEqual(x.Interface(), reflect.Zero(x.Type()).Interface())
//}

//func set(field reflect.Value, name string, value string, missingFields *[]string) error {
//	switch field.Kind() {
//	case reflect.Slice:
//		arr := strings.Split(value, ",")
//		field.Set(reflect.ValueOf(arr))
//	case reflect.Struct:
//		s := reflect.New(field.Type())
//		if err := handleDefaultValues(s.Elem(), missingFields, name); err != nil {
//			return err
//		}
//		field.Set(s.Elem())
//	case reflect.String:
//		field.SetString(value)
//	case reflect.Bool:
//		bvalue, err := strconv.ParseBool(value)
//		if err != nil {
//			return err
//		}
//		field.SetBool(bvalue)
//	case reflect.Int:
//		intValue, err := strconv.ParseInt(value, 10, 32)
//		if err != nil {
//			return err
//		}
//		field.SetInt(intValue)
//	}
//	return nil
//}
