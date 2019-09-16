package vcs

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNo_Configure(t *testing.T) {
	vcs := &No{}

	vcs.Configure()
}

func TestNo_Scaffold(t *testing.T) {
	vcs := &No{}

	repositoryInfo, err := vcs.Scaffold("project")
	assert.NoError(t, err)
	assert.Equal(t, repositoryInfo, &RepositoryInfo{})
}

func TestNo_Webhook(t *testing.T) {
	vcs := &No{}

	err := vcs.Webhook("project", "url")
	assert.NoError(t, err)
}

func TestNo_Validate(t *testing.T) {
	vcs := &No{}

	err := vcs.Validate("project")
	assert.NoError(t, err)
}
