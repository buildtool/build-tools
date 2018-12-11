#!/usr/bin/env bash

registry:login() {
  true
}

registry:create() {
  true
}

for FILE in $( cd "$( dirname "${BASH_SOURCE-$0}" )" && pwd )/registry.d/*.sh; do
  source ${FILE}
done
