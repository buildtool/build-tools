package stack

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/sparetimecoders/build-tools/pkg/templating"
	"testing"
)

func TestTemplating(t *testing.T) {
	result, err := templating.Execute(goMod, templating.TemplateData{
		ProjectName:    "test",
		Badges:         "",
		Organisation:   "org.example",
		RepositoryUrl:  "git@github.com/org/example",
		RepositoryHost: "github.com",
		RepositoryPath: "/org/example",
	})
	assert.NoError(t, err)
	assert.Equal(t, "\nmodule github.com/org/example\n\ngo 1.12\n", result)
}
