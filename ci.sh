#!/usr/bin/env bash

source "$( cd "$( dirname "${BASH_SOURCE-$0}" )" && pwd )/vcs.sh"

# CI build engine configuration
# Map CI specific environment variables to the ones used by these tools

for CI in $( cd "$( dirname "${BASH_SOURCE-$0}" )" && pwd )/ci.d/*.sh; do
  source ${CI}
done

ci:build_name() {
  local BUILD_NAME=${CI_BUILD_NAME:-$(basename $PWD)}
  : ${BUILD_NAME:?"BUILD_NAME cannot be determined"}
  echo $BUILD_NAME
}

ci:commit() {
  local COMMIT=${CI_COMMIT:-$(vcs:getCommit)}
  : ${COMMIT:?"COMMIT cannot be determined"}
  echo $COMMIT
}
