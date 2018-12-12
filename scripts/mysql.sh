#!/usr/bin/env bash
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source ${SCRIPT_DIR}/kubernetes.sh

# Creates a user for a service in a mysql database running inside kubernetes
# A secret with the credentials will also be created.
# Intended for local setup.
# Parameters
# 1 - the service name for which to create the database user, username and password will be the same as well
# 2 - the environment to use to access Kubernetes
mysql:create_database_user() {
  local SERVICE_NAME="${1}"
  local ENVIRONMENT="${2}"
  local KUBECTL_CMD=$(kubernetes:get_command ${ENVIRONMENT})
  db_pod_name=$(${KUBECTL_CMD} get pods --selector 'app=mysql' --output jsonpath={.items..metadata.name})

  ${KUBECTL_CMD} exec -it ${db_pod_name} -- bash -c "echo \"CREATE USER IF NOT EXISTS ${SERVICE_NAME} IDENTIFIED BY '${SERVICE_NAME}';CREATE DATABASE IF NOT EXISTS ${SERVICE_NAME}; GRANT ALL ON ${SERVICE_NAME}.* TO ${SERVICE_NAME};\" | mysql -u root -p\"password\""

  local SECRET_NAME="${SERVICE_NAME}-db"
  ${KUBECTL_CMD} delete secret ${SECRET_NAME} &> /dev/null || true
  ${KUBECTL_CMD} create secret generic ${SECRET_NAME} \
 --from-literal=USERNAME="${SERVICE_NAME}" \
 --from-literal=PASSWORD="${SERVICE_NAME}"
}

# Loads a dump-file into the database for a service
# Intended for local setup.
# Parameters
# 1 - the service name for which to load data
# 2 - the environment to use to access Kubernetes
# 3 - path to the file to load
mysql:load_data() {
  local SERVICE_NAME="$1"
  local ENVIRONMENT="$2"
  local FILE="$3"

  local KUBECTL_CMD=$(kubernetes:get_command ${ENVIRONMENT})
  db_pod_name=$(${KUBECTL_CMD} get pods --selector 'app=mysql' --output jsonpath={.items..metadata.name})

  cat "$FILE" | ${KUBECTL_CMD} exec -it ${db_pod_name} -- bash -c "mysql -u $SERVICE_NAME -p\"$SERVICE_NAME\" $SERVICE_NAME"
  ${KUBECTL_CMD} exec -it ${db_pod_name} -- bash -c "mysqlcheck -u root --password=\"password\" --optimize \"${SERVICE_NAME}\""
}
