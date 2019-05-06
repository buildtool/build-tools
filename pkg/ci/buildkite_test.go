package ci

import (
  "github.com/stretchr/testify/assert"
  "os"
  "testing"
)

func TestIdentify_Buildkite(t *testing.T) {
  os.Clearenv()
  _ = os.Setenv("BUILDKITE_COMMIT", "abc123")
  _ = os.Setenv("BUILDKITE_PIPELINE_SLUG", "reponame")
  _ = os.Setenv("BUILDKITE_BRANCH_NAME", "feature/first test")

  result := Identify()
  assert.NotNil(t, result)
  assert.Equal(t, "abc123", result.Commit())
  assert.Equal(t, "reponame", result.BuildName())
  assert.Equal(t, "feature/first test", result.Branch())
  assert.Equal(t, "feature_first_test", result.BranchReplaceSlash())
}
