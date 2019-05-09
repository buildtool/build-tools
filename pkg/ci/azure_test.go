package ci

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/sparetimecoders/build-tools/pkg/vcs"
	"os"
	"testing"
)

func TestIdentify_Azure(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("VSTS_PROCESS_LOOKUP_ID", "1")
	_ = os.Setenv("BUILD_SOURCEVERSION", "abc123")
	_ = os.Setenv("BUILD_REPOSITORY_NAME", "reponame")
	_ = os.Setenv("BUILD_SOURCEBRANCHNAME", "feature/first test")

	result, err := Identify(vcs.Identify("."))
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "abc123", result.Commit())
	assert.Equal(t, "reponame", result.BuildName())
	assert.Equal(t, "feature/first test", result.Branch())
	assert.Equal(t, "feature_first_test", result.BranchReplaceSlash())
}
