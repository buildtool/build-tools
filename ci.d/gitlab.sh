#!/usr/bin/env bash

if [ -n "${GITLAB_CI-}" ]; then
  echo "Running in GitlabCI"
  set +u
  CI_COMMIT="${CI_COMMIT_SHA}"
  CI_BUILD_NAME="${CI_PROJECT_NAME}"
  CI_BRANCH_NAME="${CI_COMMIT_REF_NAME}"
  set -u
fi

if [ "${CI:-}" == "gitlab" ]; then
  echo "Will use Gitlab as CI"

  ci:scaffold() {
    local projectname="$1"
    local repository="$2"
  }
fi
