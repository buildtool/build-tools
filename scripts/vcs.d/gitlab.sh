#!/usr/bin/env bash

if [ "${VCS:-}" == "gitlab" ]; then
  : ${GITLAB_GROUP:?"GITLAB_GROUP must be set"}
  : ${GITLAB_TOKEN:?"GITLAB_TOKEN must be set"}

  echo "Will use Gitlab as VCS"

  vcs:gitlab:namespace() {
    curl --silent -H "PRIVATE-TOKEN: $GITLAB_TOKEN" "https://gitlab.com/api/v4/namespaces" | jq -r --arg group "$GITLAB_GROUP" '.[] | select(.full_path == $group) | .id'
  }

  vcs:gitlab:scaffold:repo() {
    local projectname="$1"
    local namespace="$2"

    (curl --silent \
      -H "PRIVATE-TOKEN: $GITLAB_TOKEN" \
      -H "Content-Type: application/json" \
      "https://gitlab.com/api/v4/projects" \
      -d "
      {
        \"name\":\"$projectname\",
        \"namespace_id\":$namespace,
        \"issues_enabled\":true,
        \"merge_requests_enabled\":true,
        \"jobs_enabled\":true,
        \"wiki_enabled\":true,
        \"snippets_enabled\":true,
        \"resolve_outdated_diff_discussions\":true,
        \"container_registry_enabled\":true,
        \"shared_runners_enabled\":true,
        \"visibility\":\"internal\",
        \"only_allow_merge_if_pipeline_succeeds\":true,
        \"lfs_enabled\":true,
        \"printing_merge_request_link_enabled\":true,
        \"ci_config_path\":\".gitlab-ci.yml\"
      }") >/dev/null
  }

  vcs:git:scaffold:local() {
    local clone_url="$1"
    git clone "$clone_url"
  }

  vcs:scaffold() {
    local projectname="$1"

    local namespace=$(vcs:gitlab:namespace)
    (vcs:gitlab:scaffold:repo "$projectname" "$namespace")
    local clone_url="git@gitlab.com:${GITLAB_GROUP}/${projectname}.git"
    vcs:git:scaffold:local "$clone_url"
    echo "$clone_url"
  }

  vcs:webhook() {
    local projectname="$1"
    local webhook_url="$2"
  }
fi
