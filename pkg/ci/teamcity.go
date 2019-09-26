package ci

type TeamCity struct {
	*Common
	CICommit     string `env:"BUILD_VCS_NUMBER"`
	CIBuildName  string `env:"TEAMCITY_PROJECT_NAME"`
	CIBranchName string `env:"BUILD_VCS_BRANCH"`
}

var _ CI = &TeamCity{}

func (c TeamCity) Name() string {
	return "TeamCity"
}

func (c TeamCity) BranchReplaceSlash() string {
	return branchReplaceSlash(c.Branch())
}

func (c TeamCity) BuildName() string {
	return c.Common.BuildName(c.CIBuildName)
}

func (c TeamCity) Branch() string {
	return c.Common.Branch(c.CIBranchName)
}

func (c TeamCity) Commit() string {
	return c.Common.Commit(c.CICommit)
}

func (c TeamCity) Configured() bool {
	return c.CIBuildName != ""
}
