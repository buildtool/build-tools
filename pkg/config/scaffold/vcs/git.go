package vcs

import (
	"gopkg.in/src-d/go-git.v4"
	"io"
	"path/filepath"
)

type Git struct{}

func (g Git) Clone(dir, name, url string, out io.Writer) error {
	_, err := git.PlainClone(filepath.Join(dir, name), false, &git.CloneOptions{URL: url, Progress: out})
	return err
}
