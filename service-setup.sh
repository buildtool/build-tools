#!/usr/bin/env bash

source "$( cd "$( dirname "${BASH_SOURCE-$0}" )" && pwd )/scripts/commons.sh"
sourceBuildToolsFiles
source "$( cd "$( dirname "${BASH_SOURCE-$0}" )" && pwd )/scripts/ci.sh"
source "$( cd "$( dirname "${BASH_SOURCE-$0}" )" && pwd )/scripts/registry.sh"
source "$( cd "$( dirname "${BASH_SOURCE-$0}" )" && pwd )/scripts/stack.sh"
source "$( cd "$( dirname "${BASH_SOURCE-$0}" )" && pwd )/scripts/scaffold.sh"

set -o errexit -o pipefail -o noclobber -o nounset

main() {
  OPTIONS=s:
  LONGOPTS=stack:
  ! PARSED=$(getopt --options=$OPTIONS --longoptions=$LONGOPTS --name "$0" -- "$@")
  if [[ ${PIPESTATUS[0]} -ne 0 ]]; then
    # e.g. return value is 1
    #  then getopt has complained about wrong arguments to stdout
    exit 2
  fi
  # read getopt’s output this way to handle the quoting right:
  eval set -- "$PARSED"

  local STACK=""

  while true; do
    case "$1" in
      -s|--stack)
        STACK="$2"
        shift 2
      ;;
      --)
        shift
        break
      ;;
      *)
        DOCKER_ADDITIONAL_ARGS="$DOCKER_ADDITIONAL_ARGS $1"
        shift
      ;;
    esac
  done

  # handle non-option arguments
  if [[ $# -ne 1 ]]; then
    echo "$0: A name is required."
    exit 4
  fi

  local projectname="$1"

  if [[ -d ${projectname} ]]
  then
    echo "${projectname} folder already exists at ${PWD}"
    exit 1
  fi

  if [[ -n "$STACK" ]]; then
    if [[ ! -f "$(cd "$(dirname "${BASH_SOURCE-$0}")" && pwd)/scripts/stack.d/$STACK.sh" ]]; then
      local stacks=$(ls -1 "$( cd "$( dirname "${BASH_SOURCE-$0}" )" && pwd )/scripts/stack.d/" | sed -e 's/.sh//' | paste -sd "," -)
      echo "Provided stack does not exist yet. Available stacks are: ($stacks)"
      exit 3
    else
      source "$( cd "$( dirname "${BASH_SOURCE-$0}" )" && pwd )/scripts/stack.d/$STACK.sh"
    fi
  fi

  scaffold:validate "$projectname"

  echo "Creating repository for $projectname"
  local repository=$(vcs:scaffold "$projectname")
  cd "$projectname"
  echo "Created repo $repository"
  scaffold:mkdirs
  echo "Creating build pipeline for $projectname"
  ci:scaffold "$projectname" "$repository"
  local webhook_url=$(ci:webhook "$projectname")
  vcs:webhook "$projectname" "$webhook_url"
  scaffold:dotfiles
  scaffold:create_readme "$projectname"
  deployment:scaffold "$projectname"
  stack:scaffold "$projectname"
}

main "$@"
