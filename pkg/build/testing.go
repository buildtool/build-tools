package build

import (
	"context"

	"github.com/moby/buildkit/session"
)

type MockSession struct {
}

func (m *MockSession) Allow(a session.Attachable) {
}

func (m *MockSession) ID() string {
	return "random-id"
}

func (m *MockSession) Run(ctx context.Context, dialer session.Dialer) error {
	return nil
}

func (m *MockSession) Close() error {
	return nil
}

var _ Session = &MockSession{}
