#!/usr/bin/env bash

registry:login() {
  true
}

registry:create() {
  true
}

for REGISTRY in $( cd "$( dirname "${BASH_SOURCE-$0}" )" && pwd )/registry.d/*.sh; do
  source ${REGISTRY}
done
