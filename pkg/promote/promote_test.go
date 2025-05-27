// MIT License
//
// Copyright (c) 2018 buildtool
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package promote

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/go-git/go-git/v5"

	"github.com/buildtool/build-tools/pkg"
	"github.com/buildtool/build-tools/pkg/config"

	"github.com/apex/log"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	mocks "gitlab.com/unboundsoftware/apex-mocks"

	"github.com/buildtool/build-tools/pkg/version"
)

func TestDoPromote(t *testing.T) {
	tests := []struct {
		name              string
		config            string
		descriptor        string
		args              []string
		env               map[string]string
		want              int
		wantLogged        []string
		wantCommitMessage *string
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
				"info: Usage:  \\[<target>\\] \\[flags\\]\n",
				"info: \n",
				"info: Arguments:\n",
				"info:   \\[<target>\\]    the target in the .buildtools.yaml\n",
				"info: \n",
				"info: Flags:\n",
				"info:   -h, --help           Show context-sensitive help.\n",
				"info:       --version        Print args information and exit\n",
				"info:   -v, --verbose        Enable verbose mode\n",
				"info:       --config         Print parsed config and exit\n",
				"info:       --tag=\"\"         override the tag to deploy, not using the CI or VCS\n",
				"info:                        evaluated value\n",
				"info:       --url=\"\"         override the URL to the Git repository where files will\n",
				"info:                        be generated\n",
				"info:       --path=\"\"        override the path in the Git repository where files will\n",
				"info:                        be generated\n",
				"info:       --user=\"git\"     username for Git access\n",
				"info:       --key=\"\"         private key for Git access \\(defaults to ~/.ssh/id_rsa\\)\n",
				"info:       --password=\"\"    password for private key\n",
				"info:   -o, --out=\"\"         write output to specified file instead of committing and\n",
				"info:                        pushing to Git\n",
			},
		},
		{
			name: "broken config",
			config: `ci: []
`,
			args:       []string{"dummy"},
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
			name: "no deployment descriptors",
			config: `
gitops:
  target:
    url: "{{.repo}}"
`,
			args: []string{"target"},
			env: map[string]string{
				"CI_COMMIT_SHA":      "abc123",
				"CI_PROJECT_NAME":    "dummy",
				"CI_COMMIT_REF_NAME": "master",
			},
			want:       -4,
			wantLogged: []string{"error: no deployment descriptors found in k8s directory"},
		},
		{
			name: "generation successful",
			config: `
git:
  name: Some User
  email: some.user@example.org
gitops:
  target:
    url: "{{.repo}}"
`,
			descriptor: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: test
data:
  BASE_URL: https://example.org
`,
			args: []string{"target"},
			env: map[string]string{
				"CI_COMMIT_SHA":      "abc12345678",
				"CI_PROJECT_NAME":    "dummy",
				"CI_COMMIT_REF_NAME": "master",
			},
			want: 0,
			wantLogged: []string{
				"info: generating...\n",
				"^info: pushing commit [0-9a-f]+ to .*\n$",
			},
			wantCommitMessage: strPointer("ci: promoting dummy to target, commit abc1234"),
		},
		{
			name: "build name is normalized",
			config: `
git:
  name: Some User
  email: some.user@example.org
gitops:
  target:
    url: "{{.repo}}"
`,
			descriptor: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: test
data:
  BASE_URL: https://example.org
`,
			args: []string{"target"},
			env: map[string]string{
				"CI_COMMIT_SHA":      "abc123",
				"CI_PROJECT_NAME":    "dummy_repo",
				"CI_COMMIT_REF_NAME": "master",
			},
			want: 0,
			wantLogged: []string{
				"info: generating...",
				"^info: pushing commit [0-9a-f]+ to .*git-repo.*\\/dummy-repo\n$",
			},
			wantCommitMessage: strPointer("ci: promoting dummy-repo to target, commit abc123"),
		},
		{
			name: "other repo, path and tag",
			config: `
gitops:
  target:
    url: "{{.repo}}"
`,
			descriptor: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: test
data:
  BASE_URL: https://example.org
`,
			args: []string{"target", "--url", "{{.other}}", "--path", "test/path", "--tag", "testing"},
			env: map[string]string{
				"CI_COMMIT_SHA":      "abc123",
				"CI_PROJECT_NAME":    "dummy",
				"CI_COMMIT_REF_NAME": "master",
			},
			want: 0,
			wantLogged: []string{
				"info: Using passed tag <green>testing</green> to promote\n",
				"info: generating...\n",
				"^info: pushing commit [0-9a-f]+ to .*other-repo.*\\/test\\/path/dummy\n$",
			},
		},
		{
			name: "other ssh key from config",
			config: `
gitops:
  target:
    url: "{{.repo}}"
git:
  key: ~/other/id_rsa
`,
			descriptor: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: test
data:
  BASE_URL: https://example.org
`,
			args: []string{"target", "--tag", "providedlongtag"},
			env: map[string]string{
				"CI_COMMIT_SHA":      "abc123",
				"CI_PROJECT_NAME":    "dummy",
				"CI_COMMIT_REF_NAME": "master",
			},
			want: 0,
			wantLogged: []string{
				"info: Using passed tag <green>providedlongtag</green> to promote\n",
				"info: generating...",
				"^info: pushing commit [0-9a-f]+ to .*git-repo.*\\/dummy\n$",
			},
			wantCommitMessage: strPointer("ci: promoting dummy to target, commit providedlongtag"),
		},
		{
			name: "clone error",
			config: `
gitops:
  target:
    url: "{{.repo}}"
`,
			descriptor: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: test
data:
  BASE_URL: https://example.org
`,
			args: []string{"target", "--url", "/missing/repo"},
			env: map[string]string{
				"CI_COMMIT_SHA":      "abc123",
				"CI_PROJECT_NAME":    "dummy",
				"CI_COMMIT_REF_NAME": "master",
			},
			want: -4,
			wantLogged: []string{
				"info: generating...",
				"error: repository not found",
			},
		},
		{
			name: "missing SSH key",
			config: `
gitops:
  target:
    url: "{{.repo}}"
`,
			descriptor: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: test
data:
  BASE_URL: https://example.org
`,
			args: []string{"target", "--key", "/missing/key"},
			env: map[string]string{
				"CI_COMMIT_SHA":      "abc123",
				"CI_PROJECT_NAME":    "dummy",
				"CI_COMMIT_REF_NAME": "master",
			},
			want: -4,
			wantLogged: []string{
				"info: generating...",
				"error: ssh key: open /missing/key: no such file or directory",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer pkg.UnsetGithubEnvironment()()
			logMock := mocks.New()
			log.SetHandler(logMock)
			oldPwd, _ := os.Getwd()
			home, _ := os.MkdirTemp(os.TempDir(), "home")
			defer func() { _ = os.RemoveAll(home) }()
			_ = os.Setenv("HOME", home)
			defer func() {
				_ = os.Unsetenv("HOME")
			}()
			sshPath := filepath.Join(home, ".ssh")
			generateSSHKey(t, sshPath)
			otherSshPath := filepath.Join(home, "other")
			generateSSHKey(t, otherSshPath)
			name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
			defer func() { _ = os.RemoveAll(name) }()
			repo, _ := InitRepo(t, "git-repo", true)
			defer func() { _ = os.RemoveAll(repo) }()
			otherrepo, _ := InitRepo(t, "other-repo", true)
			defer func() { _ = os.RemoveAll(otherrepo) }()

			buff := Template(t, tt.config, repo, otherrepo)
			err := os.WriteFile(filepath.Join(name, ".buildtools.yaml"), buff.Bytes(), 0o777)
			assert.NoError(t, err)

			if tt.descriptor != "" {
				k8s := filepath.Join(name, "k8s")
				err = os.MkdirAll(k8s, 0o777)
				assert.NoError(t, err)
				err = os.WriteFile(filepath.Join(k8s, "descriptor.yaml"), []byte(tt.descriptor), 0o666)
				assert.NoError(t, err)
			}
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
			args := make([]string, len(tt.args))
			for i, a := range tt.args {
				args[i] = Template(t, a, repo, otherrepo).String()
			}
			if got := DoPromote(name, version.Info{}, args...); got != tt.want {
				t.Errorf("DoPromote() = %v, want %v", got, tt.want)
			}
			CheckLogged(t, tt.wantLogged, logMock.Logged)

			commits := GetCommits(t, repo)
			if tt.wantCommitMessage != nil {
				assert.Equal(t, *tt.wantCommitMessage, commits[0].Message)
			} else {
				assert.Equal(t, 1, len(commits))
			}
		})
	}
}

func TestPromote_OutParam(t *testing.T) {
	type args struct {
		target *config.Gitops
		args   Args
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantLogged []string
	}{
		{
			name: "error writing file",
			args: args{
				args: Args{Out: "non-existing-dir/output.yaml"},
			},
			wantErr:    true,
			wantLogged: []string{"info: generating..."},
		},
		{
			name: "success writing file",
			args: args{
				args: Args{Out: filepath.Join(os.TempDir(), "output.yaml")},
			},
			wantErr:    false,
			wantLogged: []string{"info: generating..."},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logMock := mocks.New()
			log.SetHandler(logMock)
			name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
			defer func() { _ = os.RemoveAll(name) }()
			err := os.MkdirAll(filepath.Join(name, "k8s"), 0o777)
			assert.NoError(t, err)
			err = os.WriteFile(filepath.Join(name, "k8s", "deploy.yaml"), []byte(`some data`), 0o666)
			assert.NoError(t, err)
			cfg := config.InitEmptyConfig()

			if err := Promote(name, "dummy", "", tt.args.target, tt.args.args, cfg); (err != nil) != tt.wantErr {
				t.Errorf("Promote() error = %v, wantErr %v", err, tt.wantErr)
			}
			CheckLogged(t, tt.wantLogged, logMock.Logged)
		})
	}
}

func generateSSHKey(t *testing.T, dir string) {
	err := os.MkdirAll(dir, 0o777)
	assert.NoError(t, err)
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(key)
	pemFile, err := os.Create(filepath.Join(dir, "id_rsa"))
	assert.NoError(t, err)
	keyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}
	err = pem.Encode(pemFile, keyBlock)
	assert.NoError(t, err)
}

func CheckLogged(t testing.TB, wantLogged []string, gotLogged []string) {
	t.Helper()
	if len(gotLogged) != 0 || len(wantLogged) != 0 {
		if len(gotLogged) != len(wantLogged) {
			assert.Equal(t, wantLogged, gotLogged)
		}
		for i, got := range gotLogged {
			assert.Regexp(t, wantLogged[i], got)
		}
	}
}

func InitRepo(t *testing.T, prefix string, bare bool) (string, plumbing.Hash) {
	repo, err := os.MkdirTemp(os.TempDir(), prefix)
	assert.NoError(t, err)
	gitrepo, err := git.PlainInit(repo, false)
	assert.NoError(t, err)
	tree, err := gitrepo.Worktree()
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(repo, "file"), []byte("test"), 0o666)
	assert.NoError(t, err)
	_, err = tree.Add("file")
	assert.NoError(t, err)
	hash, err := tree.Commit("Test", &git.CommitOptions{Author: &object.Signature{Email: "test@example.com"}})
	assert.NoError(t, err)
	if bare {
		repo2, err := os.MkdirTemp(os.TempDir(), prefix)
		assert.NoError(t, err)
		_, err = git.PlainClone(repo2, true, &git.CloneOptions{URL: fmt.Sprintf("file://%s", repo)})
		assert.NoError(t, err)
		err = os.RemoveAll(repo)
		assert.NoError(t, err)
		return repo2, hash
	}
	return repo, hash
}

func GetCommits(t *testing.T, repo string) []*object.Commit {
	gitrepo, err := git.PlainOpen(repo)
	assert.NoError(t, err)
	iter, err := gitrepo.Log(&git.LogOptions{})
	assert.NoError(t, err)
	var result []*object.Commit
	for {
		commit, err := iter.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			assert.NoError(t, err)
		}
		result = append(result, commit)
	}
	return result
}

func CountCommits(t *testing.T, repo string) int {
	gitrepo, err := git.PlainOpen(repo)
	assert.NoError(t, err)
	iter, err := gitrepo.Log(&git.LogOptions{})
	assert.NoError(t, err)
	gotCommits := -1
	for {
		_, err := iter.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			assert.NoError(t, err)
		}
		gotCommits++
	}
	return gotCommits
}

func Template(t *testing.T, text, repo, otherrepo string) *bytes.Buffer {
	tpl, err := template.New("config").Parse(text)
	assert.NoError(t, err)
	buff := &bytes.Buffer{}
	err = tpl.Execute(buff, map[string]string{
		"repo":  repo,
		"other": otherrepo,
	})
	assert.NoError(t, err)
	return buff
}

func strPointer(s string) *string {
	return &s
}
