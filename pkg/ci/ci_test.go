package ci

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/sparetimecoders/build-tools/pkg/vcs"
	"os"
	"testing"
)

func TestIdentify(t *testing.T) {
	os.Clearenv()

	_, err := Identify(vcs.Identify("."))
	assert.EqualError(t, err, "no CI found")
}

func TestCi_Branch_VCS_Fallback(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("GITLAB_CI", "1")

	result, err := Identify(vcs.Identify("."))
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "", result.Branch())
}

func TestCi_Commit_VCS_Fallback(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("GITLAB_CI", "1")

	result, err := Identify(vcs.Identify("."))
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "", result.Commit())
}
