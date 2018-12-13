#!/usr/bin/env bash

# Returns the current branch (or the one set by CI in the variable `CI_BRANCH_NAME`)
vcs:getBranch() {
  echo "NO_BRANCH"
}

vcs:getCommit() {
  echo "NO_COMMIT"
}

vcs:scaffold() {
  echo "No VCS configured"
}

vcs:webhook() {
  echo "No VCS configured"
}

vcs:validate() {
  true
}

for FILE in $( cd "$( dirname "${BASH_SOURCE-$0}" )" && pwd )/vcs.d/*.sh; do
  source ${FILE}
done
