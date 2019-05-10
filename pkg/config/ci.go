package config

type CI interface {
	// TODO: Uncomment when implementing service-setup
	//Validate() bool
	//Scaffold() error
	BuildName() string
	Branch() string
	BranchReplaceSlash() string
	Commit() string
	setVCS(cfg Config)
}

type ci struct {
	VCS VCS
}

func (c *ci) setVCS(cfg Config) {
	if v, e := cfg.CurrentVCS(); e == nil {
		c.VCS = v
	}
}
