#!/usr/bin/env bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source ${SCRIPT_DIR}/environment.sh
source ${SCRIPT_DIR}/commons.sh
source ${SCRIPT_DIR}/ci.sh

DEPLOYMENT_FILES_PATH="deployment_files"
export KOPS_STATE_STORE=s3://adfenix-k8s.adfenix.com-kops-storage

kubernetes:get_command() {
  local ENVIRONMENT="${1}"
  echo "kubectl $(environment:get_context_for_environment ${ENVIRONMENT})"
}

kubernetes:local_setup() {
  if [[ -f ${DEPLOYMENT_FILES_PATH}/setup-local.sh ]]
  then
    echo "Setting up local development environment"
    ${DEPLOYMENT_FILES_PATH}/setup-local.sh
  fi
}

# Deploy deployment_files to kubernetes.
# Parameters:
# 1 - the environment to deploy to - must be in environments:valid_environments
# rest - override the default environment deployment target by passing the variables to `kubectl` directly.
# For example `deploy prod --context test-cluster --namespace test` would deploy to namsepace `test` in the `test-cluster` but assuming to use the `prod` configuration files (if present).
kubernetes:deploy() {
  local ENVIRONMENT="${1}"
  local KUBE_OVERRIDES="${@:2}"
  if [[ ! -z ${KUBE_OVERRIDES} ]]; then
    local KUBECTL_CMD="kubectl ${KUBE_OVERRIDES}"
  else
    local KUBECTL_CMD=$(kubernetes:get_command ${ENVIRONMENT})
  fi

  local IMAGE_NAME=$(ci:build_name)

  if [[ -z ${KUBECTL_CMD} ]]
  then
    die "Invalid kubectl command string. Environment: $1, ctx: $2"
  fi
  if [[ -z ${IMAGE_NAME} ]]
  then
    die "Missing image name"
  fi
  if [[ ! -s ${DEPLOYMENT_FILES_PATH}/deploy.yaml ]]
  then
    die "Missing ${DEPLOYMENT_FILES_PATH}/deploy.yaml file"
  fi

  echo "Deploying '${IMAGE_NAME}' using '${KUBECTL_CMD}'"
  [[ "${ENVIRONMENT}" == "local" ]] && kubernetes:local_setup

  shopt -s nullglob
  for FILE in ${DEPLOYMENT_FILES_PATH}/{.,$ENVIRONMENT}/*.yaml; do
    COMMIT=${COMMIT} TIMESTAMP=$(date +%Y%m%d-%H:%M:%S) envsubst < ${FILE} | ${KUBECTL_CMD} apply --record=false -f -
  done

  if [[ $(${KUBECTL_CMD} get deployment ${IMAGE_NAME} 2> /dev/null) ]]; then
    ${KUBECTL_CMD} rollout status deployment ${IMAGE_NAME}
  fi
}


# TODO Move this elsewhere
kubernetes:create_database_user() {
  local SERVICE_NAME="${1}"
  local KUBECTL_CMD=$(kubernetes:get_command)
  db_pod_name=$(${KUBECTL_CMD} get pods --selector 'app=postgres-postgresql' --output jsonpath={.items..metadata.name})

  ${KUBECTL_CMD} exec -it ${db_pod_name} -- bash -c "echo \"CREATE USER ${SERVICE_NAME} WITH PASSWORD '${SERVICE_NAME}';CREATE DATABASE ${SERVICE_NAME} WITH OWNER ${SERVICE_NAME} ENCODING utf8\" | psql -f -"

  local SECRET_NAME="${SERVICE_NAME}-db"
  ${KUBECTL_CMD} delete secret ${SECRET_NAME} &> /dev/null || true
  ${KUBECTL_CMD} create secret generic ${SECRET_NAME} \
 --from-literal=USERNAME="${SERVICE_NAME}" \
 --from-literal=PASSWORD="${SERVICE_NAME}"
}
