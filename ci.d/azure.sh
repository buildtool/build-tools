#!/usr/bin/env bash

if [ -n "${VSTS_PROCESS_LOOKUP_ID-}" ]; then
  echo "Running in Azure"
  set +u
  CI_COMMIT="${BUILD_SOURCEVERSION}"
  CI_BUILD_NAME="${BUILD_REPOSITORY_NAME}"
  CI_BRANCH_NAME="${BUILD_SOURCEBRANCHNAME}"
  set -u
fi
