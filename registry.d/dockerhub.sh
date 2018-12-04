#!/usr/bin/env bash

if [ -n "${DOCKERHUB_USERNAME:-}" ]; then
  registry:login() {
    $(docker login -u ${DOCKERHUB_USERNAME} -p ${DOCKERHUB_PASSWORD})
  }

  registry:create() {
    true
  }
fi
