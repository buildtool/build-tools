package ci

type CI interface {
  identify() bool
  Validate() bool
  Scaffold() error
  BuildName() string
  Branch() string
  BranchReplaceSlash() string
  Commit() string
}

var cis = []CI{&azure{}, &buildkite{}, &gitlab{}}

func Identify() CI {
  for _, ci := range cis {
    if ci.identify() {
      return ci
    }
  }

  return nil
}
