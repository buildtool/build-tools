package vcs

type AzureVCS struct {
	Git
}

func (v AzureVCS) Name() string {
	return "Azure"
}

func (v AzureVCS) Scaffold(name string) (*RepositoryInfo, error) {
	return nil, nil
}

func (v AzureVCS) Webhook(name, url string) error {
	return nil
}

func (v AzureVCS) Validate(name string) error {
	return nil
}

func (v AzureVCS) Configure() {}

var _ VCS = &AzureVCS{}
