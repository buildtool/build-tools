#!/usr/bin/env bash

source "$( cd "$( dirname "${BASH_SOURCE-$0}" )" && pwd )/commons.sh"
source "$( cd "$( dirname "${BASH_SOURCE-$0}" )" && pwd )/ci.sh"
source "$( cd "$( dirname "${BASH_SOURCE-$0}" )" && pwd )/vcs.sh"
source "$( cd "$( dirname "${BASH_SOURCE-$0}" )" && pwd )/registry.sh"

## CONFIG BLOCK
# The docker registry base url, used to name the image (and push).
# Define this in a .buildtools-file on a relevant level
#DOCKER_REGISTRY_URL=""
## END CONFIG BLOCK

declare -A valid_environments

sourceBuildToolsFiles

# Build and tags docker image.
# Tags:
# * The VCS branchname - using `vcs:getBranchReplaceSlash`
# * The vcs commit id
# * `latest` if the current branch being built is the `master` branch
#
# Arguments to this function are passed as additional parameters to docker build command
#
docker:build() {
  (
  registry:login

  OPTIONS=f:
  LONGOPTS=file:
  ! PARSED=$(getopt --options=$OPTIONS --longoptions=$LONGOPTS --name "$0" -- "$@")
  if [[ ${PIPESTATUS[0]} -ne 0 ]]; then
    # e.g. return value is 1
    #  then getopt has complained about wrong arguments to stdout
    exit 2
  fi
  # read getopt’s output this way to handle the quoting right:
  eval set -- "$PARSED"

  local DOCKERFILE="Dockerfile"
  local DOCKER_ADDITIONAL_ARGS=""

  while true; do
    case "$1" in
      -f|--file)
        DOCKERFILE="$2"
        shift 2
      ;;
      --)
        shift
        break
      ;;
      *)
        DOCKER_ADDITIONAL_ARGS="$DOCKER_ADDITIONAL_ARGS $1"
        shift
      ;;
    esac
  done

  local DOCKER_BUILD_PATH=.
  local DOCKER_TAG=$(ci:getBranchReplaceSlash)
  local IMAGE_NAME=$(ci:build_name)
  local COMMIT=$(ci:commit)

  local STAGES=$(grep -i "FROM .* AS .*" "$DOCKERFILE" | sed 's/^.* [aA][sS] \(.*\)$/\1/')
  local CACHES=""
  for STAGE in ${STAGES}; do
    echo "Trying to build docker image [${DOCKER_REGISTRY_URL}/${IMAGE_NAME}:${STAGE}]"
    docker pull ${DOCKER_REGISTRY_URL}/${IMAGE_NAME}:${STAGE} || true
    CACHES="--cache-from=${DOCKER_REGISTRY_URL}/${IMAGE_NAME}:${STAGE} ${CACHES}"
    try eval $(echo docker build --pull --shm-size 256m --memory=3g --memory-swap=-1 ${CACHES} --target ${STAGE} -t ${DOCKER_REGISTRY_URL}/${IMAGE_NAME}:${STAGE} ${DOCKER_BUILD_PATH} ${DOCKER_ADDITIONAL_ARGS})
  done
  docker pull ${DOCKER_REGISTRY_URL}/${IMAGE_NAME}:${DOCKER_TAG} || true
  docker pull ${DOCKER_REGISTRY_URL}/${IMAGE_NAME}:latest || true
  CACHES="--cache-from=${DOCKER_REGISTRY_URL}/${IMAGE_NAME}:${DOCKER_TAG} --cache-from=${DOCKER_REGISTRY_URL}/${IMAGE_NAME}:latest $CACHES"
  try eval $(echo docker build --pull --shm-size 256m --memory=3g --memory-swap=-1 ${CACHES} -t ${DOCKER_REGISTRY_URL}/${IMAGE_NAME}:${COMMIT} ${DOCKER_BUILD_PATH} ${DOCKER_ADDITIONAL_ARGS})

  if [[ "${DOCKER_TAG}" == "master" ]]; then
    docker:tag ${DOCKER_REGISTRY_URL}/${IMAGE_NAME} ${COMMIT} latest
  fi

  docker:tag ${DOCKER_REGISTRY_URL}/${IMAGE_NAME} ${COMMIT} ${DOCKER_TAG}
  )
}

docker:tag() {
  local DOCKER_NAME="$1"
  local COMMIT="$2"
  local DOCKER_TAG="$3"

  docker tag ${DOCKER_NAME}:${COMMIT} ${DOCKER_NAME}:${DOCKER_TAG}
  echo "Tagged docker image [${DOCKER_NAME}:${COMMIT}] with tag ${DOCKER_TAG}"
}

docker:push() {
  (
  registry:login
  registry:create

  OPTIONS=f:
  LONGOPTS=file:
  ! PARSED=$(getopt --options=$OPTIONS --longoptions=$LONGOPTS --name "$0" -- "$@")
  if [[ ${PIPESTATUS[0]} -ne 0 ]]; then
    # e.g. return value is 1
    #  then getopt has complained about wrong arguments to stdout
    exit 2
  fi
  # read getopt’s output this way to handle the quoting right:
  eval set -- "$PARSED"

  local DOCKERFILE="Dockerfile"

  while true; do
    case "$1" in
      -f|--file)
        DOCKERFILE="$2"
        shift 2
      ;;
      --)
        shift
        break
      ;;
      *)
        echo "Programming error"
        exit 3
      ;;
    esac
  done

  local DOCKER_TAG=$(ci:getBranchReplaceSlash)
  local IMAGE_NAME=$(ci:build_name)
  local COMMIT=$(ci:commit)
  local STAGES=$(grep -i "FROM .* AS .*" "$DOCKERFILE" | sed 's/^.* [aA][sS] \(.*\)$/\1/')
  for STAGE in ${STAGES}; do
    docker push ${DOCKER_REGISTRY_URL}/${IMAGE_NAME}:${STAGE}
  done

  if [[ "${DOCKER_TAG}" == "master" ]];
  then
    docker push ${DOCKER_REGISTRY_URL}/${IMAGE_NAME}:latest
  fi
  docker push ${DOCKER_REGISTRY_URL}/${IMAGE_NAME}:${DOCKER_TAG}
  docker push ${DOCKER_REGISTRY_URL}/${IMAGE_NAME}:${COMMIT}
)
}
