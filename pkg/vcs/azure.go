package vcs

type Azure struct {
	Git
}

func (v Azure) Name() string {
	return "Azure"
}

var _ VCS = &Azure{}
