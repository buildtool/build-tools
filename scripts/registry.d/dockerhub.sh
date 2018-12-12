#!/usr/bin/env bash

if [[ -n "${DOCKERHUB_REPOSITORY:-}" ]]; then
  DOCKER_REGISTRY_URL=${DOCKERHUB_REPOSITORY}
  : ${DOCKERHUB_USERNAME:?"DOCKERHUB_USERNAME must be set"}
  : ${DOCKERHUB_PASSWORD:?"DOCKERHUB_PASSWORD must be set"}

  registry:login() {
    docker login -u ${DOCKERHUB_USERNAME} -p ${DOCKERHUB_PASSWORD}
  }

  registry:create() {
    true
  }
fi

if [[ "${REGISTRY:-}" == "dockerhub" ]]; then
  echo "Will use Dockerhub as container registry"
  DOCKER_REGISTRY_URL="$DOCKERHUB_REPOSITORY"
fi
