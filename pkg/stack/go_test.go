package stack

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"gitlab.com/sparetimecoders/build-tools/pkg/templating"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestTemplating(t *testing.T) {
	result, err := templating.Execute(goMod, templating.TemplateData{
		ProjectName:    "test",
		Badges:         nil,
		Organisation:   "org.example",
		RepositoryUrl:  "git@github.com/org/example",
		RepositoryHost: "github.com",
		RepositoryPath: "/org/example",
	})
	assert.NoError(t, err)
	assert.Equal(t, "\nmodule github.com/org/example\n\ngo 1.12\n", result)
}

func TestGo_Scaffold_Error(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	filename := filepath.Join(name, "test")
	_ = ioutil.WriteFile(filename, []byte("abc"), 0666)

	stack := &Go{}

	err := stack.Scaffold(filename, templating.TemplateData{
		ProjectName:    "test",
		Badges:         nil,
		Organisation:   "org.example",
		RepositoryUrl:  "git@github.com/org/example",
		RepositoryHost: "github.com",
		RepositoryPath: "/org/example",
	})

	assert.EqualError(t, err, fmt.Sprintf("mkdir %s: not a directory", filename))
}

func TestGo_Scaffold(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	editorconfig := filepath.Join(name, ".editorconfig")
	gomod := filepath.Join(name, "go.mod")
	_ = ioutil.WriteFile(editorconfig, []byte("root=true"), 0666)

	stack := &Go{}

	err := stack.Scaffold(name, templating.TemplateData{
		ProjectName:    "test",
		Badges:         nil,
		Organisation:   "org.example",
		RepositoryUrl:  "git@github.com/org/example",
		RepositoryHost: "github.com",
		RepositoryPath: "/org/example",
	})

	assert.NoError(t, err)
	assertFileContent(t, editorconfig, expectedEditorConfig)
	assertFileContent(t, gomod, expectedGoMod)
}

func TestGo_Name(t *testing.T) {
	stack := &Go{}

	assert.Equal(t, "go", stack.Name())
}

var expectedEditorConfig = `root=true

[*.go]
indent_style = tab
indent_size = 4

`

var expectedGoMod = `module github.com/org/example

go 1.12
`
