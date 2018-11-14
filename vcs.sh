#!/usr/bin/env bash

# Returns the current branch (or the one set by CI in the variable `CI_BRANCH_NAME`)
vcs:getBranch() {
  set -u
  echo ${CI_BRANCH_NAME:-$(git rev-parse --abbrev-ref HEAD)}
}

vcs:getBranchReplaceSlash() {
  vcs:getBranch | sed 's^/^_^g' | sed 's/ /_/g'
}

vcs:getCommit() {
  git rev-parse HEAD
}