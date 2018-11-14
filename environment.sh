#!/usr/bin/env bash
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${SCRIPT_DIR}/commons.sh

declare -A valid_environments
valid_environments=(
["local"]="--context docker-for-desktop --namespace default"
["staging"]="--context docker-for-desktop --namespace staging"
["prod"]="--context docker-for-desktop --namespace prod"
)

environment:check_args() {
  : ${1:?"Usage: $0 ENVIRONMENT"}
  local environment=$(echo "$1" | tr '[:upper:]' '[:lower:]')
  if [[ ${valid_environments[$environment]+abc} != abc ]];then
    die "Wrong environment ${environment} not in (${!valid_environments[@]})"
  fi
}

environment:get_context_for_environment() {
  local ENVIRONMENT="${1}"
  echo "${valid_environments[${ENVIRONMENT}]}"
}