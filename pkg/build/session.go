package build

import (
	"context"

	"github.com/moby/buildkit/session"
)

type Session interface {
	Allow(a session.Attachable)
	ID() string
	Run(ctx context.Context, dialer session.Dialer) error
	Close() error
}
