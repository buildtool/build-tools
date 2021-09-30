package prepare

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/apex/log"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	mocks "gitlab.com/unboundsoftware/apex-mocks"

	"github.com/buildtool/build-tools/pkg/version"
)

func TestDoPrepare(t *testing.T) {
	tests := []struct {
		name       string
		config     string
		args       []string
		env        map[string]string
		want       int
		wantLogged []string
	}{
		{
			name:       "invalid argument",
			config:     "",
			args:       []string{"--unknown"},
			want:       -1,
			wantLogged: []string{"info: error: unknown flag --unknown\n"},
		},
		{
			name: "help",
			config: `ci: []
`,
			args: []string{"--help"},
			want: 0,
			wantLogged: []string{
				"info: Usage:  [<target>]\n",
				"info: \n",
				"info: Arguments:\n",
				"info:   [<target>]    the target in the .buildtools.yaml\n",
				"info: \n",
				"info: Flags:\n",
				"info:   -h, --help                   Show context-sensitive help.\n",
				"info:       --version                Print args information and exit\n",
				"info:   -v, --verbose                Enable verbose mode\n",
				"info:       --config                 Print parsed config and exit\n",
				"info:   -n, --namespace=STRING       override the namespace for default deployment\n",
				"info:                                target\n",
				"info:       --tag=STRING             override the tag to deploy, not using the CI or\n",
				"info:                                VCS evaluated value\n",
				"info:       --url=STRING             override the URL to the Git repository where\n",
				"info:                                files will be generated\n",
				"info:       --path=STRING            override the path in the Git repository where\n",
				"info:                                files will be generated\n",
				"info:       --user=\"git\"             username for Git access\n",
				"info:       --key=\"~/.ssh/id_rsa\"    private key for Git access\n",
				"info:       --password=STRING        password for private key\n",
			},
		},
		{
			name: "broken config",
			config: `ci: []
`,
			args:       []string{},
			want:       -1,
			wantLogged: []string{"error: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!seq into config.CIConfig"},
		},
		{
			name:       "missing target",
			config:     "",
			args:       []string{"dummy"},
			want:       -2,
			wantLogged: []string{"error: no gitops matching dummy found"},
		},
		{
			name: "no CI",
			config: `
gitops:
  dummy: {}
`,
			args:       []string{"dummy"},
			want:       -3,
			wantLogged: []string{"error: Commit and/or branch information is <red>missing</red>. Perhaps your not in a Git repository or forgot to set environment variables?"},
		},
		{
			name: "no env",
			config: `
gitops:
  dummy: {}
`,
			env: map[string]string{
				"CI_COMMIT_SHA":      "abc123",
				"CI_PROJECT_NAME":    "dummy",
				"CI_COMMIT_REF_NAME": "master",
			},
			args:       []string{},
			want:       -2,
			wantLogged: []string{"error: no gitops matching local found"},
		},
		{
			name: "no options",
			config: `
gitops:
  dummy:
    url: "{{.repo}}"
`,
			args: []string{"dummy"},
			env: map[string]string{
				"CI_COMMIT_SHA":      "abc123",
				"CI_PROJECT_NAME":    "dummy",
				"CI_COMMIT_REF_NAME": "master",
			},
			want:       -4,
			wantLogged: []string{"error: no deployment descriptors found in k8s directory"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logMock := mocks.New()
			log.SetHandler(logMock)
			oldPwd, _ := os.Getwd()
			name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
			defer func() { _ = os.RemoveAll(name) }()
			repo, _ := ioutil.TempDir(os.TempDir(), "git-repo")
			defer func() { _ = os.RemoveAll(repo) }()
			gitrepo, err := git.PlainInit(repo, false)
			assert.NoError(t, err)
			tree, _ := gitrepo.Worktree()
			file, _ := ioutil.TempFile(repo, "file")
			_, _ = file.WriteString("test")
			_ = file.Close()
			_, _ = tree.Add(file.Name())
			_, _ = tree.Commit("Test", &git.CommitOptions{Author: &object.Signature{Email: "test@example.com"}})

			tpl, err := template.New("config").Parse(tt.config)
			assert.NoError(t, err)
			buff := &bytes.Buffer{}
			err = tpl.Execute(buff, map[string]string{
				"repo": repo,
			})
			assert.NoError(t, err)
			_ = ioutil.WriteFile(filepath.Join(name, ".buildtools.yaml"), buff.Bytes(), 0777)

			err = os.Chdir(name)
			assert.NoError(t, err)
			defer func() { _ = os.Chdir(oldPwd) }()

			for k, v := range tt.env {
				err := os.Setenv(k, v)
				assert.NoError(t, err)
			}
			defer func() {
				for k := range tt.env {
					err := os.Unsetenv(k)
					assert.NoError(t, err)
				}
			}()
			if got := DoPrepare(name, version.Info{}, tt.args...); got != tt.want {
				t.Errorf("DoPrepare() = %v, want %v", got, tt.want)
			}
			logMock.Check(t, tt.wantLogged)
		})
	}
}
