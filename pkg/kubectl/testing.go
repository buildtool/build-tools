// +build !prod

package kubectl

type MockKubectl struct {
	Inputs     []string
	Responses  []error
	Deployment bool
	Status     bool
}

func (m *MockKubectl) Apply(input string) error {
	m.Inputs = append(m.Inputs, input)
	return m.Responses[len(m.Inputs)-1]
}

func (m *MockKubectl) Cleanup() {
}

func (m *MockKubectl) DeploymentExists(name string) bool {
	return m.Deployment
}

func (m *MockKubectl) RolloutStatus(name string) bool {
	return m.Status
}

func (m *MockKubectl) DeploymentEvents(name string) string {
	return "Deployment events"
}

func (m *MockKubectl) PodEvents(name string) string {
	return "Pod events"
}

var _ Kubectl = &MockKubectl{}
