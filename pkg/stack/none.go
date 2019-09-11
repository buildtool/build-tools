package stack

import "gitlab.com/sparetimecoders/build-tools/pkg/templating"

type None struct{}

func (n *None) Scaffold(dir, name string, data templating.TemplateData) error {
	return nil
}

func (n *None) Name() string {
	return "none"
}

var _ Stack = &None{}
