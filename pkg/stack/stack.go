package stack

type Stack interface {
	Scaffold(name string) error
	Name() string
}

var Stacks = map[string]Stack{
	"none": &none{},
}
