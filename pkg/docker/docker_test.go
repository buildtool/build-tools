package docker

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDockerignore_FileMissing(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()

	defaultIgnores := []string{"k8s"}
	result, err := ParseDockerignore(name, "Dockerfile")
	assert.NoError(t, err)
	assert.Equal(t, defaultIgnores, result)
}

func TestParseDockerignore_EmptyFile(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()

	content := ``
	_ = ioutil.WriteFile(filepath.Join(name, ".dockerignore"), []byte(content), 0777)

	defaultIgnores := []string{"k8s"}
	result, err := ParseDockerignore(name, "Dockerfile")
	assert.NoError(t, err)
	assert.Equal(t, defaultIgnores, result)
}

func TestParseDockerignore_UnreadableFile(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	filename := filepath.Join(name, ".dockerignore")
	_ = os.Mkdir(filename, 0777)

	_, err := ParseDockerignore(name, "Dockerfile")
	assert.EqualError(t, err, fmt.Sprintf("read %s: is a directory", filename))
}

func TestParseDockerignore(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()

	content := `
node_modules
*.swp`
	_ = ioutil.WriteFile(filepath.Join(name, ".dockerignore"), []byte(content), 0777)

	result, err := ParseDockerignore(name, "Dockerfile")
	assert.NoError(t, err)
	assert.Equal(t, []string{"k8s", "node_modules", "*.swp"}, result)
}

func TestParseDockerignore_Dockerfile_Ignored(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()

	content := `
node_modules
Dockerfile
*.swp`
	_ = ioutil.WriteFile(filepath.Join(name, ".dockerignore"), []byte(content), 0777)

	result, err := ParseDockerignore(name, "Dockerfile")
	assert.NoError(t, err)
	assert.Equal(t, []string{"k8s", "node_modules", "*.swp"}, result)
}

func Test_slugify(t *testing.T) {
	type args struct {
		tag string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "valid tags are left intact",
			args: args{tag: "valid-tag"},
			want: "valid-tag",
		},
		{
			name: "leading .'s are removed",
			args: args{tag: "....leading"},
			want: "leading",
		},
		{
			name: "leading -'s are removed",
			args: args{tag: "---leading"},
			want: "leading",
		},
		{
			name: "invalid characters are removed",
			args: args{tag: "ab!€#%&/()=?`´'^¨:;<>§°"},
			want: "ab",
		},
		{
			name: "tags longer than 128 chars are truncated",
			args: args{tag: "abcdefghijklmnopqrstuvwxyz0123456789.-_abcdefghijklmnopqrstuvwxyz0123456789.-_abcdefghijklmnopqrstuvwxyz0123456789.-_abcdefghijklmnopqrstuvwxyz0123456789.-_"},
			want: "abcdefghijklmnopqrstuvwxyz0123456789.-_abcdefghijklmnopqrstuvwxyz0123456789.-_abcdefghijklmnopqrstuvwxyz0123456789.-_abcdefghijk",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SlugifyTag(tt.args.tag); got != tt.want {
				t.Errorf("SlugifyTag() = %v, want %v", got, tt.want)
			}
		})
	}
}
