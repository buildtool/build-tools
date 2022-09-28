// MIT License
//
// Copyright (c) 2021 buildtool
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

package docker

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDockerignore_FileMissing(t *testing.T) {
	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()

	defaultIgnores := []string{"k8s"}
	result, err := ParseDockerignore(name, "Dockerfile")
	assert.NoError(t, err)
	assert.Equal(t, defaultIgnores, result)
}

func TestParseDockerignore_EmptyFile(t *testing.T) {
	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()

	content := ``
	_ = os.WriteFile(filepath.Join(name, ".dockerignore"), []byte(content), 0777)

	defaultIgnores := []string{"k8s"}
	result, err := ParseDockerignore(name, "Dockerfile")
	assert.NoError(t, err)
	assert.Equal(t, defaultIgnores, result)
}

func TestParseDockerignore_UnreadableFile(t *testing.T) {
	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	filename := filepath.Join(name, ".dockerignore")
	_ = os.Mkdir(filename, 0777)

	_, err := ParseDockerignore(name, "Dockerfile")
	assert.EqualError(t, err, fmt.Sprintf("read %s: is a directory", filename))
}

func TestParseDockerignore(t *testing.T) {
	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()

	content := `
node_modules
*.swp`
	_ = os.WriteFile(filepath.Join(name, ".dockerignore"), []byte(content), 0777)

	result, err := ParseDockerignore(name, "Dockerfile")
	assert.NoError(t, err)
	assert.Equal(t, []string{"k8s", "node_modules", "*.swp"}, result)
}

func TestParseDockerignore_Dockerfile_Ignored(t *testing.T) {
	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()

	content := `
node_modules
Dockerfile
*.swp`
	_ = os.WriteFile(filepath.Join(name, ".dockerignore"), []byte(content), 0777)

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
