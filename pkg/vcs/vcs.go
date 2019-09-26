package vcs

import "io"

// VCS represent the VersionControlSystem used
type VCS interface {
	// Identify returns true if it is the expected VCS type (based on information found in dir)
	Identify(dir string, out io.Writer) bool
	// Name returns the name of the VCS
	Name() string
	// Branch returns the current branch
	Branch() string
	// Commit returns the current commit
	Commit() string
}

// CommonVCS contains functions shared by all VCSs
type CommonVCS struct {
	CurrentBranch string
	CurrentCommit string
}

// Branch returns the current branch
func (v CommonVCS) Branch() string {
	return v.CurrentBranch
}

// Commit returns the current commit
func (v CommonVCS) Commit() string {
	return v.CurrentCommit
}

var systems = []VCS{&git{}}

// Identify tries to identify the actual VCS
func Identify(dir string, out io.Writer) VCS {
	for _, vcs := range systems {
		if vcs.Identify(dir, out) {
			return vcs
		}
	}

	no := &no{CommonVCS: CommonVCS{}}
	return no
}
