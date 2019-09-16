package config

import (
	"gitlab.com/sparetimecoders/build-tools/pkg/vcs"
	"io"
)

var systems = []vcs.VCS{&vcs.Git{}}

func Identify(dir string, out io.Writer) vcs.VCS {
	for _, vcs := range systems {
		if vcs.Identify(dir, out) {
			return vcs
		}
	}

	no := &vcs.No{}
	no.Identify(dir, out)
	return no
}
