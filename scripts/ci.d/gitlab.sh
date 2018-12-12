#!/usr/bin/env bash

if [[ -n "${GITLAB_CI:-}" ]]; then
  echo "Running in GitlabCI"
  set +u
  CI_COMMIT="${CI_COMMIT_SHA}"
  CI_BUILD_NAME="${CI_PROJECT_NAME}"
  CI_BRANCH_NAME="${CI_COMMIT_REF_NAME}"
  set -u
fi

if [[ "${CI:-}" == "gitlab" ]]; then
  : ${GITLAB_GROUP:?"GITLAB_GROUP must be set"}
  : ${GITLAB_TOKEN:?"GITLAB_TOKEN must be set"}

  echo "Will use Gitlab as CI"

  ci:gitlab:scaffold:file() {
    cat <<EOF > .gitlab-ci.yml
stages:
  - build
  - deploy-staging
  - deploy-prod

variables:
  DOCKER_HOST: tcp://docker:2375/

image: registry.gitlab.com/sparetimecoders/build-tools:master

build:
  stage: build
  services:
    - docker:dind
  script:
  - source \${BUILD_TOOLS_PATH}/docker.sh
  - docker:build
  - docker:push

deploy-to-staging:
  stage: deploy-staging
  when: on_success
  script:
    - echo Deploy to staging.
    - deploy staging
  environment:
    name: staging

deploy-to-prod:
  stage: deploy-prod
  when: on_success
  script:
    - echo Deploy to PROD.
    - deploy prod
  environment:
    name: prod
  only:
    - master
EOF
  }

  ci:validate() {
    local projectname="$1"
    local project=$(echo "$GITLAB_GROUP/$projectname" | sed 's/\//%2f/g')
    local status_code=$(curl --silent -I --header "PRIVATE-TOKEN: $GITLAB_TOKEN" "https://gitlab.com/api/v4/projects/$project" -w '%{http_code}' -o /dev/null)
    if [[ "200" -eq "$status_code" ]]; then
      echo "Gitlab project and pipeline already exists for ${projectname}"
      exit 1
    fi
  }

  ci:scaffold() {
    local projectname="$1"
    local repository="$2"
    (ci:gitlab:scaffold:file) >/dev/null
  }

  ci:badges() {
    local projectname="$1"
    local project=$(echo "$GITLAB_GROUP/$projectname" | sed 's/\//%2f/g')
    curl --silent --header "PRIVATE-TOKEN: $GITLAB_TOKEN" "https://gitlab.com/api/v4/projects/$project/badges" | jq '.[] | "[![](\(.rendered_image_url))](\(.rendered_link_url))"'
  }

  ci:webhook() {
    local projectname="$1"
  }
fi
