package ci

type Buildkite struct {
	*Common
	CICommit     string `env:"BUILDKITE_COMMIT"`
	CIBuildName  string `env:"BUILDKITE_PIPELINE_SLUG"`
	CIBranchName string `env:"BUILDKITE_BRANCH_NAME"`
}

var _ CI = &Buildkite{}

func (c *Buildkite) Name() string {
	return "Buildkite"
}

func (c *Buildkite) BranchReplaceSlash() string {
	return branchReplaceSlash(c.Branch())
}

func (c *Buildkite) BuildName() string {
	return c.Common.BuildName(c.CIBuildName)
}

func (c *Buildkite) Branch() string {
	return c.Common.Branch(c.CIBranchName)
}

func (c *Buildkite) Commit() string {
	return c.Common.Commit(c.CICommit)
}

func (c *Buildkite) Configured() bool {
	return c.CIBuildName != ""
}
