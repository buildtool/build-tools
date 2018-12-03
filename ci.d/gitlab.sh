#!/usr/bin/env bash

if [ -n "${GITLAB_CI-}" ]; then
  echo "Running in GitlabCI"
  set +u
  CI_COMMIT="${CI_COMMIT_SHA}"
  CI_BUILD_NAME="${CI_PROJECT_NAME}"
  CI_BRANCH_NAME="${CI_COMMIT_REF_NAME}"
  set -u
fi
