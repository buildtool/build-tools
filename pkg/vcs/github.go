package vcs

type Github struct {
	Git
}

func (v *Github) Name() string {
	return "Github"
}

var _ VCS = &Github{}
