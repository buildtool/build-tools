// +build !prod

package kubectl

import (
	"gitlab.com/sparetimecoders/build-tools/pkg/config"
)

type MockKubectl struct {
	Inputs    []string
	Responses []error
}

func (m *MockKubectl) Apply(input string) error {
	m.Inputs = append(m.Inputs, input)
	return m.Responses[len(m.Inputs)-1]
}

func (m *MockKubectl) Environment() *config.Environment {
	return &config.Environment{Name: "dummy", Context: "dummy", Namespace: "default"}
}

func (m *MockKubectl) Cleanup() {
}

var _ Kubectl = &MockKubectl{}
