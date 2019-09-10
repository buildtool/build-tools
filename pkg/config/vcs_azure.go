package config

type AzureVCS struct {
	git
}

func (v AzureVCS) Name() string {
	return "Azure"
}

func (v AzureVCS) Scaffold(name string) string {
	return ""
}

var _ VCS = &AzureVCS{}
