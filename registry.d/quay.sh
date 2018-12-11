#!/usr/bin/env bash

if [ -n "${QUAY_REPOSITORY:-}" ]; then
  DOCKER_REGISTRY_URL=${QUAY_REPOSITORY}

  registry:login() {
    docker login -u ${QUAY_USERNAME} -p ${QUAY_PASSWORD}
  }

  registry:create() {
    true
  }
fi

if [ "$REGISTRY" == "quay" ]; then
  echo "Will use Quay.io as container registry"
  DOCKER_REGISTRY_URL="$QUAY_REPOSITORY"
fi
