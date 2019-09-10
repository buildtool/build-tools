package config

type GitlabVCS struct {
	git
}

func (v GitlabVCS) Name() string {
	return "Gitlab"
}

func (v GitlabVCS) Scaffold(name string) string {
	return ""
}

var _ VCS = &GitlabVCS{}
