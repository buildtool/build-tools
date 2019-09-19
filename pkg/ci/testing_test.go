// +build !prod

package ci

import (
	"gitlab.com/sparetimecoders/build-tools/pkg/vcs"
	"io"
)

type mockVcs struct{}

func (m mockVcs) Identify(dir string, out io.Writer) bool {
	panic("implement me")
}

func (m mockVcs) Name() string {
	panic("implement me")
}

func (m mockVcs) Branch() string {
	return "fallback-branch"
}

func (m mockVcs) Commit() string {
	return "fallback-sha"
}

var _ vcs.VCS = &mockVcs{}
