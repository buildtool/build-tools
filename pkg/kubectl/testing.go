// +build !prod

package kubectl

import (
	"gitlab.com/sparetimecoders/build-tools/pkg/config"
	"io"
	"io/ioutil"
)

type MockKubectl struct {
	Inputs    []string
	Calls     [][]string
	Responses []error
}

func (m *MockKubectl) Apply(input io.Reader, args ...string) error {
	bytes, _ := ioutil.ReadAll(input)
	m.Inputs = append(m.Inputs, string(bytes))
	m.Calls = append(m.Calls, args)
	return m.Responses[len(m.Calls)-1]
}

func (m *MockKubectl) Environment() *config.Environment {
	return &config.Environment{Name: "dummy", Context: "dummy", Namespace: "default"}
}

var _ Kubectl = &MockKubectl{}
