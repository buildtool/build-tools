#!/usr/bin/env bash

if [ -d .git ]; then
  vcs:getBranch() {
    git rev-parse --abbrev-ref HEAD
  }

  vcs:getCommit() {
    git rev-parse HEAD
  }
fi
