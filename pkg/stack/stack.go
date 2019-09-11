package stack

import "gitlab.com/sparetimecoders/build-tools/pkg/templating"

type Stack interface {
	Scaffold(dir, name string, data templating.TemplateData) error
	Name() string
}

var Stacks = map[string]Stack{
	"none":  &None{},
	"go":    &Go{},
	"scala": &Scala{},
}
