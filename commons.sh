#!/usr/bin/env bash
#set -euo pipefail

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