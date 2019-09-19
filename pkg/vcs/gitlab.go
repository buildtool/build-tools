package vcs

type Gitlab struct {
	Git
}

func (v *Gitlab) Name() string {
	return "Gitlab"
}

var _ VCS = &Gitlab{}
