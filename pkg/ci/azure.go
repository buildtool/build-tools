package ci

type Azure struct {
	*Common
	CICommit     string `env:"BUILD_SOURCEVERSION"`
	CIBuildName  string `env:"BUILD_REPOSITORY_NAME"`
	CIBranchName string `env:"BUILD_SOURCEBRANCHNAME"`
}

var _ CI = &Azure{}

func (c Azure) Name() string {
	return "Azure"
}

func (c Azure) BranchReplaceSlash() string {
	return branchReplaceSlash(c.Branch())
}

func (c Azure) BuildName() string {
	return c.Common.BuildName(c.CIBuildName)
}

func (c Azure) Branch() string {
	return c.Common.Branch(c.CIBranchName)
}

func (c Azure) Commit() string {
	return c.Common.Commit(c.CICommit)
}

func (c Azure) Configured() bool {
	return c.CIBuildName != ""
}
