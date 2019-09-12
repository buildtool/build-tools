package config

type GitlabVCS struct {
	git
}

func (v GitlabVCS) Name() string {
	return "Gitlab"
}

func (v GitlabVCS) Scaffold(name string) (*RepositoryInfo, error) {
	return nil, nil
}

func (v GitlabVCS) Webhook(name, url string) error {
	return nil
}

func (v GitlabVCS) Validate() error {
	return nil
}

var _ VCS = &GitlabVCS{}
