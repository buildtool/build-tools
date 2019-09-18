package vcs

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAzureVCS_Name(t *testing.T) {
	vcs := &AzureVCS{}

	assert.Equal(t, "Azure", vcs.Name())
}

func TestAzureVCS_Scaffold(t *testing.T) {
	vcs := &AzureVCS{}

	repositoryInfo, err := vcs.Scaffold("project")
	assert.NoError(t, err)
	assert.Nil(t, repositoryInfo)
}

func TestAzureVCS_Webhook(t *testing.T) {
	vcs := &AzureVCS{}

	err := vcs.Webhook("project", "url")
	assert.NoError(t, err)
}

func TestAzureVCS_Validate(t *testing.T) {
	vcs := &AzureVCS{}

	err := vcs.Validate("project")
	assert.NoError(t, err)
}

func TestAzureVCS_Configure(t *testing.T) {
	vcs := &AzureVCS{}

	vcs.Configure()
}
