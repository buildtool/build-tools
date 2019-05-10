package config

type VCS interface {
	identify(dir string) bool
	Branch() string
	Commit() string
}

var systems = []VCS{&git{}}

func Identify(dir string) VCS {
	for _, vcs := range systems {
		if vcs.identify(dir) {
			return vcs
		}
	}

	no := &no{}
	no.identify(dir)
	return no
}

type vcs struct {
	branch string
	commit string
}

func (v vcs) Branch() string {
	return v.branch
}

func (v vcs) Commit() string {
	return v.commit
}
