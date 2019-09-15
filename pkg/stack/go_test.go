package stack

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/sparetimecoders/build-tools/pkg/templating"
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

func TestTemplating_Badges(t *testing.T) {
	template := `{{range .Badges}}[![{{.Title}}]({{.ImageUrl}})]({{.LinkUrl}}){{end}}`
	result, err := templating.Execute(template, templating.TemplateData{
		ProjectName: "test",
		Badges: []templating.Badge{
			{"Title1", "https://img1", "https://link1"},
			{"Title2", "https://img2", "https://link2"},
		},
		Organisation:   "org.example",
		RepositoryUrl:  "git@github.com/org/example",
		RepositoryHost: "github.com",
		RepositoryPath: "/org/example",
	})
	assert.NoError(t, err)
	assert.Equal(t, "[![Title1](https://img1)](https://link1)[![Title2](https://img2)](https://link2)", result)
}
