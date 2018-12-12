#!/usr/bin/env bash

if [ -n "${CI_REGISTRY_IMAGE:-}" ]; then
  DOCKER_REGISTRY_URL=${CI_REGISTRY_IMAGE%/*}

  : ${CI_BUILD_TOKEN:?"CI_BUILD_TOKEN must be set"}
  : ${CI_REGISTRY:?"CI_REGISTRY must be set"}

  registry:login() {
    docker login -u gitlab-ci-token -p ${CI_BUILD_TOKEN} ${CI_REGISTRY}
  }

  registry:create() {
    true
  }
fi

if [ "${REGISTRY:-}" == "gitlab" ]; then
  echo "Will use Gitlab as container registry"
  : ${GITLAB_GROUP:?"GITLAB_GROUP must be set"}

  DOCKER_REGISTRY_URL="registry.gitlab.com/$GITLAB_GROUP"
fi
