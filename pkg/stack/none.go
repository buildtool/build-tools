package stack

type none struct{}

func (n *none) Scaffold(name string) error {
	return nil
}

var _ Stack = &none{}
