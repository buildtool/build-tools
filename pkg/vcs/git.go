package vcs

import (
	"github.com/apex/log"
	git2 "github.com/go-git/go-git/v5"
)

type git struct {
	CommonVCS
	repo *git2.Repository
}

func (v *git) Identify(dir string) bool {
	options := &git2.PlainOpenOptions{DetectDotGit: true}
	repo, err := git2.PlainOpenWithOptions(dir, options)
	if err != nil {
		return false
	}
	v.repo = repo
	ref, err := repo.Head()
	if err != nil {
		log.Debugf("Unable to fetch head: %s\n", err)
		return false
	}
	v.CurrentCommit = ref.Hash().String()
	v.CurrentBranch = ref.Name().Short()

	return true
}

func (v *git) Name() string {
	return "Git"
}

var _ VCS = &git{}
