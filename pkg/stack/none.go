package stack

type none struct{}

func (n *none) Scaffold(name string) error {
	panic("implement me")
}

var _ Stack = &none{}
