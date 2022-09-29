// MIT License
//
// Copyright (c) 2018 buildtool
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

//go:build !prod
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

func (m *MockKubectl) RolloutStatus(name, timeout string) bool {
	return m.Status
}

func (m *MockKubectl) DeploymentEvents(name string) string {
	return "Deployment events"
}

func (m *MockKubectl) PodEvents(name string) string {
	return "Pod events"
}

var _ Kubectl = &MockKubectl{}
