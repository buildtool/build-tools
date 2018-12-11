#!/usr/bin/env bash

if [ -n "${CI_REGISTRY_IMAGE:-}" ]; then
  DOCKER_REGISTRY_URL=${CI_REGISTRY_IMAGE%/*}

  registry:login() {
    docker login -u gitlab-ci-token -p ${CI_BUILD_TOKEN} ${CI_REGISTRY}
  }

  registry:create() {
    true
  }
fi

if [ "$REGISTRY" == "gitlab" ]; then
  echo "Will use Gitlab as container registry"
  DOCKER_REGISTRY_URL="registry.gitlab.com/$GITLAB_GROUP"
fi
