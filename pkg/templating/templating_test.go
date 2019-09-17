package templating

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTemplating_Badges(t *testing.T) {
	template := `{{range .Badges}}[![{{.Title}}]({{.ImageUrl}})]({{.LinkUrl}}){{end}}`
	result, err := Execute(template, TemplateData{
		ProjectName: "test",
		Badges: []Badge{
			{Title: "Title1", ImageUrl: "https://img1", LinkUrl: "https://link1"},
			{Title: "Title2", ImageUrl: "https://img2", LinkUrl: "https://link2"},
		},
		Organisation:   "org.example",
		RepositoryUrl:  "git@github.com/org/example",
		RepositoryHost: "github.com",
		RepositoryPath: "/org/example",
	})
	assert.NoError(t, err)
	assert.Equal(t, "[![Title1](https://img1)](https://link1)[![Title2](https://img2)](https://link2)", result)
}

func TestTemplating_Error(t *testing.T) {
	template := `{{range .ProjectName}}{{.Fluff}}{{end}}`
	_, err := Execute(template, TemplateData{
		ProjectName: "test",
	})
	assert.EqualError(t, err, "template: content:1:8: executing \"content\" at <.ProjectName>: range can't iterate over test")
}
