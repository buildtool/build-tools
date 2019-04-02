#!/usr/bin/env bash

if [ -n "${QUAY_REPOSITORY:-}" ]; then
  DOCKER_REGISTRY_URL="quay.io/$QUAY_REPOSITORY"
  : ${QUAY_USERNAME:?"QUAY_USERNAME must be set"}
  : ${QUAY_PASSWORD:?"QUAY_PASSWORD must be set"}

  registry:login() {
    docker login -u ${QUAY_USERNAME} -p ${QUAY_PASSWORD} quay.io
  }

  registry:create() {
    true
  }
fi

if [ "${REGISTRY:-}" == "quay" ]; then
  echo "Will use Quay.io as container registry"
  DOCKER_REGISTRY_URL="quay.io/$QUAY_REPOSITORY"
fi
