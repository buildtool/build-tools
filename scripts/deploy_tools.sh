#!/usr/bin/env bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${SCRIPT_DIR}/commons.sh
source ${SCRIPT_DIR}/environment.sh
source ${SCRIPT_DIR}/kubernetes.sh

deploy:main() {
  environment:check_args "$@"
  kubernetes:deploy "$@"
}
