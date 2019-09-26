package ci

type No struct {
	*Common
}

var _ CI = &No{}

func (c No) Name() string {
	return "none"
}

func (c No) BranchReplaceSlash() string {
	return branchReplaceSlash(c.Branch())
}

func (c No) BuildName() string {
	return c.Common.BuildName("")
}

func (c No) Branch() string {
	return c.VCS.Branch()
}

func (c No) Commit() string {
	return c.VCS.Commit()
}

func (c No) Configured() bool {
	return false
}
