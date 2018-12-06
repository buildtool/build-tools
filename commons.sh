#!/usr/bin/env bash

yell() {
  echo "$0: $*" >&2;
}

die() {
  yell "$*"
  exit 1
}

try() {
  "$@" || die "cannot $*";
}

upfind:local() {
  CURRDIR=$1
  FILE=$2
  while [[ ${CURRDIR} != / ]]; do
    find "${CURRDIR}" -maxdepth 1 -mindepth 1 -name "$FILE"
    CURRDIR="$(readlink -f "${CURRDIR}"/..)"
  done
}

upfind() {
  upfind:local "$1" "$2" | tac
}

sourceBuildToolsFiles() {
  if [ -n "${BUILDTOOLS_CONTENT:-}" ]; then
    echo "Found buildtools content, creating .buildtools-file"
    echo "${BUILDTOOLS_CONTENT}" | base64 -d >! .buildtools
  fi

  for CONFIG in $(upfind "${PWD}" ".buildtools"); do
    echo "Sourcing ${CONFIG}"
    source ${CONFIG}
  done
}
