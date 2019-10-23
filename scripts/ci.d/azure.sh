#!/usr/bin/env bash

if [[ -n "${VSTS_PROCESS_LOOKUP_ID-}" ]]; then
  echo "Running in Azure"
  set +u
  CI_COMMIT="${BUILD_SOURCEVERSION}"
  CI_BUILD_NAME="${BUILD_REPOSITORY_NAME}"
  CI_BRANCH_NAME="${BUILD_SOURCEBRANCHNAME}"
  set -u
fi

if [[ "${CI:-}" == "azure" ]]; then
  echo "Will use Azure as CI"

  : ${AZURE_USER:?"AZURE_USER must be set"}
  : ${AZURE_PROJECT:?"AZURE_PROJECT must be set"}
  : ${AZURE_ORG:?"AZURE_ORG must be set"}
  : ${AZURE_TOKEN:?"AZURE_TOKEN must be set"}

  ci:azure:get_project_id() {
    curl --silent -u "${AZURE_USER}:${AZURE_TOKEN}" "https://dev.azure.com/${AZURE_ORG}/_apis/projects/${AZURE_PROJECT}?api-version=4.0" | jq -r '.id'
  }

  ci:azure:get_ubuntu_queue() {
    curl --silent \
      -u "${AZURE_USER}:${AZURE_TOKEN}" \
      "https://dev.azure.com/${AZURE_ORG}/${AZURE_PROJECT}/_apis/distributedtask/queues?api-version=5.0-preview.1" \
      | jq -r '.value[] | select(.name == "Hosted Ubuntu 1604")'
  }

  ci:azure:get_definition_id() {
    local projectname="$1"
    curl --silent \
      -u "${AZURE_USER}:${AZURE_TOKEN}" \
      "https://dev.azure.com/${AZURE_ORG}/${AZURE_PROJECT}/_apis/build/definitions?api-version=4.1&name=$projectname" \
      | jq '.value[0].id'
  }

  ci:azure:get_definition() {
    local projectname="$1"
    local definitionId=$(ci:azure:get_definition_id "$projectname")
    curl --silent \
      -u "${AZURE_USER}:${AZURE_TOKEN}" \
      "https://dev.azure.com/${AZURE_ORG}/${AZURE_PROJECT}/_apis/build/definitions/$definitionId?api-version=4.1"
  }

  ci:azure:variable() {
    local name="$1"
    local value="$2"
    local secret="false"
    if [[ "$value" == secret:* ]]; then
      secret="true"
      value="${value#secret:}"
    fi

    echo "\"$name\": { \"value\": \"$value\", \"isSecret\":$secret}"
  }

  ci:azure:scaffold:pipeline() {
    local projectname="$1"
    local projectid="$2"
    local queue="$3"

    local variables=""
    for v in ${!PIPELINE_VARIABLES[@]}; do
      local json=$(ci:azure:variable "$v" "${PIPELINE_VARIABLES[${v}]}")
      if [[ -n "$variables" ]]; then
        variables="$variables,"
      fi
      variables="${variables}${json}"
    done

    curl --silent \
      -u "${AZURE_USER}:${AZURE_TOKEN}" \
      "https://dev.azure.com/${AZURE_ORG}/${AZURE_PROJECT}/_apis/build/definitions?api-version=4.1" \
      -H "Content-Type: application/json" \
      -d "{
        \"options\": [
        ],
        \"triggers\": [
          {
            \"branchFilters\": [],
            \"pathFilters\": [],
            \"settingsSourceType\": 2,
            \"batchChanges\": false,
            \"maxConcurrentBuildsPerBranch\": 1,
            \"triggerType\": \"continuousIntegration\"
          }
        ],
        \"variables\": {$variables},
        \"properties\": {},
        \"tags\": [],
        \"jobAuthorizationScope\": \"projectCollection\",
        \"jobTimeoutInMinutes\": 60,
        \"jobCancelTimeoutInMinutes\": 5,
        \"process\": {
          \"yamlFilename\": \"/azure-pipelines.yml\",
          \"type\": 2
        },
        \"repository\": {
          \"properties\": {
            \"cloneUrl\": \"https://${AZURE_ORG}@dev.azure.com/${AZURE_ORG}/${AZURE_PROJECT}/_git/${projectname}\",
            \"fullName\": \"test\",
            \"defaultBranch\": \"refs/heads/master\",
            \"isFork\": \"False\",
            \"reportBuildStatus\": \"true\",
            \"cleanOptions\": \"0\",
            \"fetchDepth\": \"0\",
            \"gitLfsSupport\": \"false\",
            \"skipSyncSource\": \"false\",
            \"checkoutNestedSubmodules\": \"false\",
            \"labelSources\": \"0\",
            \"labelSourcesFormat\": \"\$(build.buildNumber)\"
          },
          \"type\": \"TfsGit\",
          \"name\": \"${projectname}\",
          \"url\": \"https://${AZURE_ORG}@dev.azure.com/${AZURE_ORG}/${AZURE_PROJECT}/_git/${projectname}\",
          \"defaultBranch\": \"refs/heads/master\",
          \"clean\": null,
          \"checkoutSubmodules\": false
        },
        \"processParameters\": {},
        \"quality\": \"definition\",
        \"drafts\": [],
        \"queue\": ${queue},
        \"name\": \"${projectname}\",
        \"type\": \"build\",
        \"queueStatus\": \"enabled\",
        \"project\": {
          \"id\": \"${projectid}\",
        }
      }"
  }

  ci:azure:scaffold:file() {
    cat <<EOF > azure-pipelines.yml
resources:
  containers:
  - container: build-tools
    image: registry.github.com/sparetimecoders/build-tools:master

jobs:
- job: build_and_deploy
  pool:
    vmImage: 'Ubuntu 16.04'
  container: build-tools
  steps:
  - script: |
      set -e -o pipefail
      build
      push
    name: build
    env:
      DOCKERHUB_PASSWORD: \$(DOCKERHUB_PASSWORD)
      QUAY_PASSWORD: \$(QUAY_PASSWORD)
  - script: deploy staging
    name: deploy_staging
    condition: succeeded()
  - script: deploy prod
    name: deploy_prod
    condition: succeeded()
EOF
  }

  ci:validate() {
    local projectname="$1"

    local statuscode=$(curl --silent -I -u "${AZURE_USER}:${AZURE_TOKEN}" -X GET "https://dev.azure.com/${AZURE_ORG}/_apis/projects?api-version=4.0" -w '%{http_code}' -o /dev/null)
    if [[ "404" -eq "$statuscode" ]]; then
      echo "Invalid Azure user, token or organization"
      exit 1
    fi
    local projectId=$(ci:azure:get_project_id)
    local definitionId=$(ci:azure:get_definition_id "$projectname")
  }

  ci:scaffold() {
    local projectname="$1"
    local repository="$2"
    local projectid=$(ci:azure:get_project_id)
    local queue=$(ci:azure:get_ubuntu_queue)
    (ci:azure:scaffold:pipeline "$projectname" "$projectid" "$queue") >/dev/null
    (ci:azure:scaffold:file) >/dev/null
  }

  ci:badges() {
    local projectname="$1"
    local response=$(ci:azure:get_definition "$projectname")
    local id=$(echo "$response" | jq -r '.id')
    local build_url="https://dev.azure.com/$AZURE_ORG/$AZURE_PROJECT/_build/latest?definitionId=$id"
    (echo "$response" | jq --arg web "$build_url" -r '"[![](\(._links.badge.href))](\($web))"')
  }

  ci:webhook() {
    local projectname="$1"
  }
fi
