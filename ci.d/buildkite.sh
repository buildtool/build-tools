#!/usr/bin/env bash

if [ -n "${BUILDKITE_COMMIT-}" ]; then
  echo "Running in Buildkite"
  set +u
  CI_COMMIT="${BUILDKITE_COMMIT}"
  CI_BUILD_NAME="${BUILDKITE_PIPELINE_SLUG}"
  CI_BRANCH_NAME="${BUILDKITE_BRANCH_NAME}"
  set -u
fi
