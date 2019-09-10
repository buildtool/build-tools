package config

type AzureVCS struct {
	git
}

func (v AzureVCS) Name() string {
	return "Azure"
}

func (v AzureVCS) Scaffold(name string) (string, error) {
	return "", nil
}

func (v AzureVCS) Webhook(name, url string) {
}

var _ VCS = &AzureVCS{}
