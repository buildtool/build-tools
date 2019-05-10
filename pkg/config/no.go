package config

type no struct {
	vcs
}

func (v no) identify(dir string) bool {
	v.commit = ""
	v.branch = ""

	return true
}
