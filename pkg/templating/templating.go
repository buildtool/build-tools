package templating

import (
	"bytes"
	"text/template"
)

type TemplateData struct {
	ProjectName    string
	Badges         []Badge
	Organisation   string
	RegistryUrl    string
	RepositoryUrl  string
	RepositoryHost string
	RepositoryPath string
}

type Badge struct {
	Title    string
	ImageUrl string
	LinkUrl  string
}

func Execute(content string, data TemplateData) (string, error) {
	if t, err := template.New("content").Parse(content); err != nil {
		return "", err
	} else {
		buf := bytes.Buffer{}
		if err := t.Execute(&buf, data); err != nil {
			return "", err
		}
		return buf.String(), nil
	}
}
