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

upfind() {
  IFS=/; dn=($1); ct=${#dn[@]}
  for((i=0; i<ct; i++)); do
    subd+=/"${dn[i]}"
    dots=$(for((j=ct-i; j>1; j--)); do printf "../"; done)
    find "$subd" -maxdepth 1 -type f -name "$2" -printf "$dots%f\n"
  done
}
