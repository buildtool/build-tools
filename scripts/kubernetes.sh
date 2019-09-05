#!/usr/bin/env bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source ${SCRIPT_DIR}/environment.sh

kubernetes:get_command() {
  local ENVIRONMENT="${1}"
  echo "kubectl $(environment:get_context_for_environment ${ENVIRONMENT})"
}
