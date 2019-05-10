package config

import (
	"errors"
	"fmt"
	"github.com/caarlos0/env"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Config struct {
	VCS      *VCSConfig      `yaml:"vcs"`
	CI       *CIConfig       `yaml:"ci"`
	Registry *RegistryConfig `yaml:"registry"`
}

type VCSConfig struct {
	Selected string `yaml:"selected"`
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

func Load(dir string) (*Config, error) {
	cfg := initEmptyConfig()

	err := parseConfigFiles(dir, func(dir string) error {
		return parseConfigFile(dir, cfg)
	})
	if err != nil {
		return cfg, err
	}

	err = env.Parse(cfg)

	vcs := Identify(dir)
	cfg.VCS.VCS = &vcs

	// TODO: Validate and clean config

	return cfg, err
}

func initEmptyConfig() *Config {
	return &Config{
		VCS: &VCSConfig{},
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

func (c *Config) CurrentVCS() (VCS, error) {
	return *c.VCS.VCS, nil
}

func (c *Config) CurrentCI() (CI, error) {
	switch c.CI.Selected {
	case "azure":
		c.CI.Azure.setVCS(*c)
		return c.CI.Azure, nil
	case "buildkite":
		c.CI.Buildkite.setVCS(*c)
		return c.CI.Buildkite, nil
	case "gitlab":
		c.CI.Gitlab.setVCS(*c)
		return c.CI.Gitlab, nil
	case "":
		vals := []CI{c.CI.Azure, c.CI.Buildkite, c.CI.Gitlab}
		for _, ci := range vals {
			if len(ci.BuildName()) > 0 {
				ci.setVCS(*c)
				return ci, nil
			}
		}
	}
	return nil, errors.New("no CI found")
}

func (c *Config) CurrentRegistry() (Registry, error) {
	switch c.Registry.Selected {
	//case "dockerhub":
	//case "ecr":
	//case "gitlab":
	//case "quay":
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

var abs = filepath.Abs

func parseConfigFiles(dir string, fn func(string) error) error {
	parent, err := abs(dir)
	if err != nil {
		return err
	}
	var files []string
	for parent != "/" {
		filename := filepath.Join(parent, "buildtools.yaml")
		if _, err := os.Stat(filename); !os.IsNotExist(err) {
			files = append(files, filename)
		}

		parent = filepath.Dir(parent)
	}
	for i := len(files) - 1; i >= 0; i-- {
		fmt.Printf("Parsing config from file: '%s'\n", files[i])
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
