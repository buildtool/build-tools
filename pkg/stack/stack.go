package stack

type Stack interface {
	Scaffold(dir, name string, data TemplateData) error
	Name() string
}

var Stacks = map[string]Stack{
	"none":  &None{},
	"go":    &Go{},
	"scala": &Scala{},
}

type TemplateData struct {
	ProjectName   string
	Badges        string
	Organisation  string
	RepositoryUrl string
}
