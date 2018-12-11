#!/usr/bin/env bash

if [[ -n "${BUILDKITE_COMMIT:-}" ]]; then
  echo "Running in Buildkite"
  set +u
  CI_COMMIT="${BUILDKITE_COMMIT}"
  CI_BUILD_NAME="${BUILDKITE_PIPELINE_SLUG}"
  CI_BRANCH_NAME="${BUILDKITE_BRANCH_NAME}"
  set -u
fi

if [[ "${CI:-}" == "buildkite" ]]; then
  echo "Will use Buildkite as CI"

  : ${BUILDKITE_ORG:?"BUILDKITE_ORG must be set"}
  : ${BUILDKITE_TOKEN:?"BUILDKITE_TOKEN must be set"}

  scaffold:buildkite:create_pipeline() {
    local projectname="$1"
    local repository="$2"

    (curl --silent -X POST "https://api.buildkite.com/v2/organizations/${BUILDKITE_ORG}/pipelines" \
      -H "Authorization: Bearer ${BUILDKITE_TOKEN}" \
      -d "{
          \"name\": \"${projectname}\",
          \"repository\": \"$repository\",
          \"steps\": [
            {
              \"type\": \"script\",
              \"name\": \"Setup :package:\",
              \"command\": \"buildkite-agent pipeline upload\"
            }
          ],
          \"skip_queued_branch_builds\": true,
          \"cancel_running_branch_builds\": true
        }") >/dev/null

    (curl --silent -X PATCH "https://api.buildkite.com/v2/organizations/${BUILDKITE_ORG}/pipelines/${projectname}" \
      -H "Authorization: Bearer ${BUILDKITE_TOKEN}" \
      -d "{
            \"provider_settings\": {
            \"publish_commit_status\": true,
              \"build_pull_requests\": true,
              \"build_pull_request_forks\": false,
              \"build_tags\": false,
              \"publish_commit_status_per_step\": true,
              \"trigger_mode\": \"code\"
            }
          }") >/dev/null
  }

  scaffold:buildkite:file() {
    cat <<EOF > .buildkite/pipeline.yml
steps:
  - command: |-
      source \${BUILD_TOOLS_PATH}/docker.sh
      build
      push
    label: build

  - wait

  - command: |-
      \${BUILD_TOOLS_PATH}/deploy staging
    label: ":rocket: Deploy to STAGING"
    branches: "master"

  - block: ":rocket: Release PROD"
    branches: "master"

  - command: |-
      \${BUILD_TOOLS_PATH}/deploy prod
    label: ":rocket: Deploy to PROD"
    branches: "master"

EOF
}

  ci:scaffold:mkdirs() {
    mkdir -p .buildkite
  }

  ci:scaffold:dotfiles() {
    echo ".buildkite" >> .dockerignore
  }

  ci:validate() {
    local projectname="$1"

    local statuscode=$(curl --silent -I "https://api.buildkite.com/v2/user" \
      -H "Authorization: Bearer ${BUILDKITE_TOKEN}" \
      -w '%{http_code}' -o /dev/null)

    if [[ "403" -eq "$statuscode" ]]; then
      echo "Invalid Buildkite token"
      exit 1
    fi

    statuscode=$(curl --silent -I "https://api.buildkite.com/v2/organizations/${BUILDKITE_ORG}" \
    -H "Authorization: Bearer ${BUILDKITE_TOKEN}" \
    -w '%{http_code}' -o /dev/null)

    if [[ "404" -eq "$statuscode" ]]; then
      echo "Buildkite organization does not exist or token has no access"
      exit 1
    fi

    statuscode=$(curl --silent -I "https://api.buildkite.com/v2/organizations/${BUILDKITE_ORG}/pipelines/${projectname}" \
      -H "Authorization: Bearer ${BUILDKITE_TOKEN}" \
      -w '%{http_code}' -o /dev/null)

    if [[ "200" -eq "$statuscode" ]]; then
      echo "Buildkite pipeline already exists for ${projectname}"
      exit 1
    fi
  }

  ci:scaffold() {
    local projectname="$1"
    local repository="$2"
    scaffold:buildkite:create_pipeline "$projectname" "$repository"
    ci:scaffold:mkdirs
    scaffold:buildkite:file
  }

  ci:badges() {
    local projectname="$1"
    curl --silent "https://api.buildkite.com/v2/organizations/${BUILDKITE_ORG}/pipelines/${projectname}" \
      -H "Authorization: Bearer ${BUILDKITE_TOKEN}" \
      | jq -r '"[![](\(.badge_url))](\(.web_url))"'
  }

  ci:webhook() {
    local projectname="$1"
    curl --silent "https://api.buildkite.com/v2/organizations/${BUILDKITE_ORG}/pipelines/${projectname}" \
      -H "Authorization: Bearer ${BUILDKITE_TOKEN}" \
      | jq -r '.provider.webhook_url'
  }
fi
