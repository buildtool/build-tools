package stack

type Stack interface {
	Scaffold(name string) error
}

var Stacks = map[string]Stack{
	"none": &none{},
}
