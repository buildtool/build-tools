#!/usr/bin/env bash

source "$( cd "$( dirname "${BASH_SOURCE-$0}" )" && pwd )/vcs.sh"

# CI build engine configuration
# Map CI specific environment variables to the ones used by these tools

ci:scaffold() {
  echo "No CI engine configured"
}

ci:validate() {
  true
}

ci:scaffold:mkdirs() {
  true
}

ci:scaffold:dotfiles() {
  true
}

for FILE in $( cd "$( dirname "${BASH_SOURCE-$0}" )" && pwd )/ci.d/*.sh; do
  source ${FILE}
done

ci:build_name() {
  local BUILD_NAME=${CI_BUILD_NAME:-$(basename $PWD)}
  : ${BUILD_NAME:?"BUILD_NAME cannot be determined"}
  echo $BUILD_NAME
}

ci:branch() {
  local BRANCH=${CI_BRANCH_NAME:-$(vcs:getBranch)}
  : ${BRANCH:?"BRANCH cannot be determined"}
  echo $BRANCH
}

ci:getBranchReplaceSlash() {
  ci:branch | sed 's^/^_^g' | sed 's/ /_/g'
}

ci:commit() {
  local COMMIT=${CI_COMMIT:-$(vcs:getCommit)}
  : ${COMMIT:?"COMMIT cannot be determined"}
  echo $COMMIT
}
