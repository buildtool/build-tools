package vcs

type no struct {
	CommonVCS
}

func (v no) Identify(dir string) bool {
	return true
}

func (v no) Name() string {
	return "none"
}

var _ VCS = &no{}
