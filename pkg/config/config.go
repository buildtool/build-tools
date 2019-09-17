package config

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/caarlos0/env"
	"github.com/liamg/tml"
	"gitlab.com/sparetimecoders/build-tools/pkg/ci"
	"gitlab.com/sparetimecoders/build-tools/pkg/file"
	"gitlab.com/sparetimecoders/build-tools/pkg/registry"
	stck "gitlab.com/sparetimecoders/build-tools/pkg/stack"
	"gitlab.com/sparetimecoders/build-tools/pkg/templating"
	"gitlab.com/sparetimecoders/build-tools/pkg/vcs"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	url2 "net/url"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type Config struct {
	VCS          *VCSConfig      `yaml:"vcs"`
	CI           *CIConfig       `yaml:"ci"`
	Registry     *RegistryConfig `yaml:"registry"`
	Environments []Environment   `yaml:"environments"`
	Organisation string          `yaml:"organisation"`
	availableCI  []ci.CI
}

type VCSConfig struct {
	Selected string         `yaml:"selected" env:"VCS"`
	Azure    *vcs.AzureVCS  `yaml:"azure"`
	Github   *vcs.GithubVCS `yaml:"github"`
	Gitlab   *vcs.GitlabVCS `yaml:"gitlab"`
	VCS      vcs.VCS
}

type CIConfig struct {
	Selected  string          `yaml:"selected" env:"CI"`
	Azure     *ci.AzureCI     `yaml:"azure"`
	Buildkite *ci.BuildkiteCI `yaml:"buildkite"`
	Gitlab    *ci.GitlabCI    `yaml:"gitlab"`
	TeamCity  *ci.TeamCityCI  `yaml:"teamcity"`
}

type RegistryConfig struct {
	Selected  string                      `yaml:"selected" env:"REGISTRY"`
	Dockerhub *registry.DockerhubRegistry `yaml:"dockerhub"`
	ECR       *registry.ECRRegistry       `yaml:"ecr"`
	Gitlab    *registry.GitlabRegistry    `yaml:"gitlab"`
	Quay      *registry.QuayRegistry      `yaml:"quay"`
}

type Environment struct {
	Name       string `yaml:"name"`
	Context    string `yaml:"context"`
	Namespace  string `yaml:"namespace"`
	Kubeconfig string `yaml:"kubeconfig" env:"KUBECONFIG_CONTENT"`
}

func Load(dir string, out io.Writer) (*Config, error) {
	cfg := initEmptyConfig()

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

func initEmptyConfig() *Config {
	c := &Config{
		VCS: &VCSConfig{
			Azure:  &vcs.AzureVCS{},
			Github: &vcs.GithubVCS{},
			Gitlab: &vcs.GitlabVCS{},
		},
		CI: &CIConfig{
			Azure:     &ci.AzureCI{CommonCI: &ci.CommonCI{}},
			Buildkite: &ci.BuildkiteCI{CommonCI: &ci.CommonCI{}},
			Gitlab:    &ci.GitlabCI{CommonCI: &ci.CommonCI{}},
			TeamCity:  &ci.TeamCityCI{CommonCI: &ci.CommonCI{}},
		},
		Registry: &RegistryConfig{
			Dockerhub: &registry.DockerhubRegistry{},
			ECR:       &registry.ECRRegistry{},
			Gitlab:    &registry.GitlabRegistry{},
			Quay:      &registry.QuayRegistry{},
		},
	}
	c.availableCI = []ci.CI{c.CI.Azure, c.CI.Buildkite, c.CI.Gitlab, c.CI.TeamCity}
	return c
}

func (c *Config) CurrentVCS() vcs.VCS {
	switch c.VCS.Selected {
	case "azure":
		c.VCS.Azure.Configure()
		return c.VCS.Azure
	case "github":
		c.VCS.Github.Configure()
		return c.VCS.Github
	case "gitlab":
		c.VCS.Gitlab.Configure()
		return c.VCS.Gitlab
	}
	return c.VCS.VCS
}

func (c *Config) CurrentCI() (ci.CI, error) {
	switch c.CI.Selected {
	case "azure":
		c.CI.Azure.SetVCS(c.CurrentVCS())
		if err := c.CI.Azure.Configure(); err != nil {
			return nil, err
		}
		return c.CI.Azure, nil
	case "buildkite":
		c.CI.Buildkite.SetVCS(c.CurrentVCS())
		if err := c.CI.Buildkite.Configure(); err != nil {
			return nil, err
		}
		return c.CI.Buildkite, nil
	case "gitlab":
		c.CI.Gitlab.SetVCS(c.CurrentVCS())
		if err := c.CI.Gitlab.Configure(); err != nil {
			return nil, err
		}
		return c.CI.Gitlab, nil
	case "":
		for _, ci := range c.availableCI {
			if ci.Configured() {
				ci.SetVCS(c.CurrentVCS())
				return ci, nil
			}
		}
	}
	ci := &ci.NoOpCI{CommonCI: &ci.CommonCI{}}
	ci.SetVCS(c.CurrentVCS())
	if err := ci.Configure(); err != nil {
		return nil, err
	}
	return ci, nil
}

func (c *Config) CurrentRegistry() (registry.Registry, error) {
	switch c.Registry.Selected {
	case "dockerhub":
		return c.Registry.Dockerhub, nil
	case "ecr":
		return c.Registry.ECR, nil
	case "gitlab":
		return c.Registry.Gitlab, nil
	case "quay":
		return c.Registry.Quay, nil
	case "":
		vals := []registry.Registry{c.Registry.Dockerhub, c.Registry.ECR, c.Registry.Gitlab, c.Registry.Quay}
		for _, reg := range vals {
			if reg.Configured() {
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

func (c *Config) Scaffold(dir, name string, stack stck.Stack, out io.Writer) int {
	currentVcs := c.CurrentVCS()
	ci, err := c.CurrentCI()
	if err != nil {
		_, _ = fmt.Fprintln(out, tml.Sprintf("<red>%s</red>", err.Error()))
		return -2
	}
	registry, err := c.CurrentRegistry()
	if err != nil {
		_, _ = fmt.Fprintln(out, tml.Sprintf("<red>%s</red>", err.Error()))
		return -3
	}
	if err := validate(name, currentVcs, ci); err != nil {
		_, _ = fmt.Fprintln(out, tml.Sprintf("<red>%s</red>", err.Error()))
		return -4
	}
	_, _ = fmt.Fprint(out, tml.Sprintf("<lightblue>Creating new service </lightblue><white><bold>'%s'</bold></white> <lightblue>using stack </lightblue><white><bold>'%s'</bold></white>\n", name, stack.Name()))
	_, _ = fmt.Fprint(out, tml.Sprintf("<lightblue>Creating repository at </lightblue><white><bold>'%s'</bold></white>\n", currentVcs.Name()))
	repository, err := currentVcs.Scaffold(name)
	if err != nil {
		_, _ = fmt.Fprintln(out, tml.Sprintf("<red>%s</red>", err.Error()))
		return -5
	}
	_, _ = fmt.Fprint(out, tml.Sprintf("<green>Created repository </green><white><bold>'%s'</bold></white>\n", repository.SSHURL))
	if err := currentVcs.Clone(dir, name, repository.SSHURL, out); err != nil {
		_, _ = fmt.Fprintln(out, tml.Sprintf("<red>%s</red>", err.Error()))
		return -6
	}
	projectDir := filepath.Join(dir, name)
	if err := createDirectories(projectDir); err != nil {
		_, _ = fmt.Fprintln(out, tml.Sprintf("<red>%s</red>", err.Error()))
		return -7
	}
	_, _ = fmt.Fprint(out, tml.Sprintf("<lightblue>Creating build pipeline for </lightblue><white><bold>'%s'</bold></white>\n", name))
	badges, err := ci.Badges(name)
	if err != nil {
		_, _ = fmt.Fprintln(out, tml.Sprintf("<red>%s</red>", err.Error()))
		return -8
	}
	url, err := url2.Parse(repository.HTTPURL)
	if err != nil {
		_, _ = fmt.Fprintln(out, tml.Sprintf("<red>%s</red>", err.Error()))
		return -9
	}
	data := templating.TemplateData{
		ProjectName:    name,
		Badges:         badges,
		Organisation:   c.Organisation,
		RegistryUrl:    registry.RegistryUrl(),
		RepositoryUrl:  repository.SSHURL,
		RepositoryHost: url.Host,
		RepositoryPath: strings.Replace(url.Path, ".git", "", 1),
	}
	webhook, err := ci.Scaffold(dir, data)
	if err != nil {
		_, _ = fmt.Fprintln(out, tml.Sprintf("<red>%s</red>", err.Error()))
		return -10
	}
	if err := addWebhook(name, webhook, currentVcs); err != nil {
		_, _ = fmt.Fprintln(out, tml.Sprintf("<red>%s</red>", err.Error()))
		return -11
	}
	if err := createDotfiles(projectDir); err != nil {
		_, _ = fmt.Fprintln(out, tml.Sprintf("<red>%s</red>", err.Error()))
		return -12
	}
	if err := createReadme(projectDir, data); err != nil {
		_, _ = fmt.Fprintln(out, tml.Sprintf("<red>%s</red>", err.Error()))
		return -13
	}
	if err := createDeployment(projectDir, data); err != nil {
		_, _ = fmt.Fprintln(out, tml.Sprintf("<red>%s</red>", err.Error()))
		return -14
	}
	if err := stack.Scaffold(projectDir, data); err != nil {
		_, _ = fmt.Fprintln(out, tml.Sprintf("<red>%s</red>", err.Error()))
		return -15
	}
	return 0
}

func validate(name string, vcs vcs.VCS, ci ci.CI) error {
	if err := vcs.Validate(name); err != nil {
		return err
	}
	return ci.Validate(name)
}

func createDirectories(dir string) error {
	return os.Mkdir(filepath.Join(dir, "deployment_files"), 0777)
}

func addWebhook(name string, url *string, vcs vcs.VCS) error {
	if url != nil {
		return vcs.Webhook(name, *url)
	}
	return nil
}

func createDotfiles(dir string) error {
	if err := file.Write(dir, ".gitignore", ""); err != nil {
		return err
	}
	editorconfig := `
root = true

[*]
end_of_line = lf
insert_final_newline = true
charset = utf-8
trim_trailing_whitespace = true
`
	if err := file.Write(dir, ".editorconfig", editorconfig); err != nil {
		return err
	}
	dockerignore := `
.git
.editorconfig
Dockerfile
README.md
`
	if err := file.Write(dir, ".dockerignore", dockerignore); err != nil {
		return err
	}
	return nil
}

func createReadme(dir string, data templating.TemplateData) error {
	content := `
| README.md
# {{.ProjectName}}
{{range .Badges}}[![{{.Title}}]({{.ImageUrl}})]({{.LinkUrl}}){{end}}
`
	tpl, err := template.New("readme").Parse(content)
	if err != nil {
		return err
	}
	buff := bytes.Buffer{}
	if err = tpl.Execute(&buff, data); err != nil {
		return err
	}
	return file.Write(dir, "README.md", buff.String())
}

func createDeployment(dir string, data templating.TemplateData) error {
	return file.WriteTemplated(dir, filepath.Join("deployment_files", "deploy.yaml"), deployment, data)
}

var deployment = `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: {{ .ProjectName }}
  name: {{ .ProjectName }}
  annotations:
    kubernetes.io/change-cause: "${TIMESTAMP} Deployed commit id: ${COMMIT}"
spec:
  replicas: 2
  selector:
    matchLabels:
      app: {{ .ProjectName }}
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 1
    type: RollingUpdate
  template:
    metadata:
      labels:
        app: {{ .ProjectName }}
    spec:
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              labelSelector:
                matchExpressions:
                - key: "app"
                  operator: In
                  values:
                  - {{ .ProjectName }}
              topologyKey: kubernetes.io/hostname
      containers:
      - name: {{ .ProjectName }}
        readinessProbe:
          httpGet:
            path: /
            port: 80
          initialDelaySeconds: 5
          periodSeconds: 5
          timeoutSeconds: 5
        imagePullPolicy: Always
        image: {{ .RegistryUrl }}/{{ .ProjectName }}:${COMMIT}
        ports:
        - containerPort: 80
      restartPolicy: Always
---

apiVersion: v1
kind: Service
metadata:
  name: {{ .ProjectName }}
spec:
  ports:
  - port: 80
    protocol: TCP
    targetPort: 80
  selector:
    app: {{ .ProjectName }}
  type: ClusterIP
`
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
		return nil
	}
}
