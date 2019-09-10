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

func (v GitlabVCS) Webhook(name, url string) {
}

var _ VCS = &GitlabVCS{}
