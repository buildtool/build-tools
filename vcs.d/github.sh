#!/usr/bin/env bash

if [ "${VCS:-}" == "github" ]; then
  : ${GITHUB_ORG:?"GITHUB_ORG must be set"}
  : ${GITHUB_TOKEN:?"GITHUB_TOKEN must be set"}

  scaffold:git:github() {
    local projectname="$1"

    curl --silent -H "Authorization: token ${GITHUB_TOKEN}" https://api.github.com/orgs/${GITHUB_ORG}/repos \
      -o /dev/null \
      -d "
      {
        \"name\": \"${projectname}\",
        \"private\": true,
        \"auto_init\": true
      }"

    curl -X PUT -H "Authorization: token ${GITHUB_TOKEN}" https://api.github.com/repos/${GITHUB_ORG}/${projectname}/branches/master/protection \
      -H 'Accept: application/vnd.github.luke-cage-preview+json' \
      --silent  -o /dev/null \
      -d "
      {
        \"required_status_checks\": {
        \"strict\": true,
           \"contexts\": [
            \"buildkite/${projectname}/build\"
          ]
        },
          \"enforce_admins\": true,
          \"required_pull_request_reviews\": {
            \"dismiss_stale_reviews\": true,
            \"required_approving_review_count\": 1
          },
          \"restrictions\" :null
      }"
  }

  scaffold:git:local() {
    local clone_url="$1"
    git clone "$clone_url"
  }

  vcs:scaffold() {
    local projectname="$1"

    scaffold:git:github "$projectname"
    local clone_url="git@github.com:${GITHUB_ORG}/${projectname}.git"
    scaffold:git:local "$clone_url"
    echo "$clone_url"
  }

  vcs:webhook() {
    local projectname="$1"
    local webhook_url="$2"

    curl --silent -X POST -H "Authorization: token ${GITHUB_TOKEN}" https://api.github.com/repos/${GITHUB_ORG}/${projectname}/hooks \
      -o /dev/null -d "
      {
      \"name\": \"web\",
      \"active\": true,
      \"events\": [
        \"push\",
        \"pull_request\",
        \"deployment\"
      ],
      \"config\": {
        \"url\": \"${webhook_url}\",
        \"content_type\": \"json\"
      }"
  }
fi
