package ci

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/sparetimecoders/build-tools/pkg/templating"
	"io/ioutil"
	"os"
	"testing"
)

func TestNoOpCI_Scaffold(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()

	ci := &NoOpCI{}

	hookUrl, err := ci.Scaffold(name, templating.TemplateData{})
	assert.Nil(t, hookUrl)
	assert.NoError(t, err)
}

func TestNoOpCI_Badges(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()

	ci := &NoOpCI{}

	badges, err := ci.Badges(name)
	assert.Nil(t, badges)
	assert.NoError(t, err)
}
