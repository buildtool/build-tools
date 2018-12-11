#!/usr/bin/env bash

if [ "${VCS:-}" == "azure" ]; then
  : ${AZURE_ORG:?"AZURE_ORG must be set"}
  : ${AZURE_USER:?"AZURE_USER must be set"}
  : ${AZURE_TOKEN:?"AZURE_TOKEN must be set"}
  : ${AZURE_PROJECT:?"AZURE_PROJECT must be set"}

  echo "Will use Azure as VCS"

  vcs:azure:get_project_id() {
    curl --silent -u "${AZURE_USER}:${AZURE_TOKEN}" "https://dev.azure.com/${AZURE_ORG}/_apis/projects/${AZURE_PROJECT}?api-version=4.0" | jq -r '.id'
  }

  vcs:azure:scaffold:repo() {
    local projectname="$1"
    local projectid="$2"
    curl --silent \
      -u "${AZURE_USER}:${AZURE_TOKEN}" \
      "https://dev.azure.com/${AZURE_ORG}/${AZURE_PROJECT}/_apis/git/repositories?api-version=4.1" \
      -H "Content-Type: application/json" \
      -d "
      {
        \"name\": \"${projectname}\",
        \"project\": {
          \"id\": \"${projectid}\"
        }
      }" \
      | jq -r '.sshUrl'
  }

  vcs:azure:scaffold:local() {
    local url="$1"
    git clone "$url"
  }

  vcs:scaffold() {
    local projectname="$1"

    local projectid=$(vcs:azure:get_project_id)
    local url=$(vcs:azure:scaffold:repo "$projectname" "${projectid}")
    (vcs:azure:scaffold:local "$url") > /dev/null
    echo "$url"
  }

  vcs:webhook() {
    local projectname="$1"
    local webhook_url="$2"
  }
fi
