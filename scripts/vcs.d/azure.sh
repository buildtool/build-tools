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
      | jq -r '"\(.sshUrl) \(.id)"'
  }

  vcs:azure:scaffold:policies() {
    local repositoryId=$1

    local policies=$(curl --silent \
      -u "${AZURE_USER}:${AZURE_TOKEN}" \
      "https://dev.azure.com/${AZURE_ORG}/${AZURE_PROJECT}/_apis/policy/types?api-version=5.0")

    local reviewersPolicyType=$(echo "$policies" | jq -r '.value[] | select(.displayName == "Minimum number of reviewers") | .id')
    local openCommentsPolicyType=$(echo "$policies" | jq -r '.value[] | select(.displayName == "Comment requirements") | .id')

    echo ${reviewersPolicyType}
    echo ${openCommentsPolicyType}

    curl --silent \
      -u "${AZURE_USER}:${AZURE_TOKEN}" \
      "https://dev.azure.com/${AZURE_ORG}/${AZURE_PROJECT}/_apis/policy/configurations/?api-version=5.0" \
      -H "Content-Type: application/json" \
      -d "
      {
        \"isEnabled\": true,
        \"isBlocking\": true,
        \"type\": {
          \"id\": \"${reviewersPolicyType}\"
        },
        \"settings\": {
          \"minimumApproverCount\": 1,
          \"creatorVoteCounts\": true,
          \"allowDownvotes\": false,
          \"resetOnSourcePush\": true,
          \"scope\": [
            {
              \"refName\": \"refs/heads/master\",
              \"matchKind\": \"Exact\",
              \"repositoryId\": \"${repositoryId}\"
            }
          ]
        }
      }" > /dev/null

    curl --silent \
      -u "${AZURE_USER}:${AZURE_TOKEN}" \
      "https://dev.azure.com/${AZURE_ORG}/${AZURE_PROJECT}/_apis/policy/configurations/?api-version=5.0" \
      -H "Content-Type: application/json" \
      -d "
      {
        \"isEnabled\": true,
        \"isBlocking\": true,
        \"type\": {
          \"id\": \"${openCommentsPolicyType}\"
        },
        \"settings\": {
          \"scope\": [
            {
              \"refName\": \"refs/heads/master\",
              \"matchKind\": \"Exact\",
              \"repositoryId\": \"${repositoryId}\"
            }
          ]
        }
      }" > /dev/null
  }

  vcs:azure:scaffold:local() {
    local url="$1"
    git clone "$url"
  }

  vcs:scaffold() {
    local projectname="$1"

    local projectid=$(vcs:azure:get_project_id)
    read url repositoryId < <(vcs:azure:scaffold:repo "$projectname" "${projectid}")
    (vcs:azure:scaffold:policies "$repositoryId")
    (vcs:azure:scaffold:local "$url") > /dev/null
    echo "$url"
  }

  vcs:webhook() {
    local projectname="$1"
    local webhook_url="$2"
  }
fi
