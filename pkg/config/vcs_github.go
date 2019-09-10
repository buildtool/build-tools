package config

type GithubVCS struct {
	git
}

func (v GithubVCS) Name() string {
	return "Github"
}

func (v GithubVCS) Scaffold(name string) string {
	return ""
}

var _ VCS = &GithubVCS{}
