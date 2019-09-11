package stack

type None struct{}

func (n *None) Scaffold(dir, name string, data TemplateData) error {
	return nil
}

func (n *None) Name() string {
	return "none"
}

var _ Stack = &None{}
