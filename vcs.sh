#!/usr/bin/env bash

# Returns the current branch (or the one set by CI in the variable `CI_BRANCH_NAME`)
vcs:getBranch() {
  echo "NO_BRANCH"
}

vcs:getCommit() {
  echo "NO_COMMIT"
}

for VCS in $( cd "$( dirname "${BASH_SOURCE-$0}" )" && pwd )/vcs.d/*.sh; do
  source ${VCS}
done
