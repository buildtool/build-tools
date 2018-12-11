#!/usr/bin/env bash

if [ -n "${DOCKERHUB_REPOSITORY:-}" ]; then
  DOCKER_REGISTRY_URL=${DOCKERHUB_REPOSITORY}

  registry:login() {
    docker login -u ${DOCKERHUB_USERNAME} -p ${DOCKERHUB_PASSWORD}
  }

  registry:create() {
    true
  }
fi

if [ "$REGISTRY" == "dockerhub" ]; then
  echo "Will use Dockerhub as container registry"
  DOCKER_REGISTRY_URL="$DOCKERHUB_REPOSITORY"
fi
