package stack

type none struct{}

func (n *none) Scaffold(name string) error {
	return nil
}

func (n *none) Name() string {
	return "none"
}

var _ Stack = &none{}
