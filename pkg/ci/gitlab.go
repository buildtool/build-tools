package ci

type Gitlab struct {
	*Common
	CICommit     string `env:"CI_COMMIT_SHA"`
	CIBuildName  string `env:"CI_PROJECT_NAME"`
	CIBranchName string `env:"CI_COMMIT_REF_NAME"`
}

var _ CI = &Gitlab{}

func (c *Gitlab) Name() string {
	return "Gitlab"
}

func (c *Gitlab) BranchReplaceSlash() string {
	return branchReplaceSlash(c.Branch())
}

func (c *Gitlab) BuildName() string {
	return c.Common.BuildName(c.CIBuildName)
}

func (c *Gitlab) Branch() string {
	return c.Common.Branch(c.CIBranchName)
}

func (c *Gitlab) Commit() string {
	return c.Common.Commit(c.CICommit)
}

func (c *Gitlab) Configured() bool {
	return c.CIBuildName != ""
}
