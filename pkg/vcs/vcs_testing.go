package vcs

type mockVcs struct {
	branch string
	commit string
}

// NewMockVcs returns a mockVcs with default commit and branch name
func NewMockVcs() VCS {
	return &mockVcs{
		branch: "fallback-branch",
		commit: "fallback-sha",
	}
}

// NewMockVcsWithBranch returns a mockVcs with a specific branch name
func NewMockVcsWithBranch(branch string) VCS {
	return &mockVcs{
		branch: branch,
		commit: "fallback-sha",
	}
}
func (m mockVcs) Identify(dir string) bool {
	panic("implement me")
}

func (m mockVcs) Name() string {
	panic("implement me")
}

func (m mockVcs) Branch() string {
	return m.branch
}

func (m mockVcs) Commit() string {
	return m.commit
}

var _ VCS = mockVcs{}
