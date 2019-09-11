package stack

import (
	"gitlab.com/sparetimecoders/build-tools/pkg/file"
	"path/filepath"
)

type Go struct{}

func (g Go) Scaffold(dir, name string, data TemplateData) error {
	editorconfig := `
[*.go]
indent_style = tab
indent_size = 4
`
	return file.Append(filepath.Join(dir, ".editorconfig"), editorconfig)
}

func (g Go) Name() string {
	return "go"
}

var _ Stack = &Go{}

var goMod = `
module {{.RepositoryUrl}}

go 1.12
`
