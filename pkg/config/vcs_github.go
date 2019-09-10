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

func (v GithubVCS) Webhook(name, url string) {
}

var _ VCS = &GithubVCS{}
