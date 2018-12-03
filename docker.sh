#!/usr/bin/env bash

source "$( cd "$( dirname "${BASH_SOURCE-$0}" )" && pwd )/commons.sh"
source "$( cd "$( dirname "${BASH_SOURCE-$0}" )" && pwd )/ci.sh"
source "$( cd "$( dirname "${BASH_SOURCE-$0}" )" && pwd )/vcs.sh"

## CONFIG BLOCK
# The docker registry base url, used to name the image (and push).
# Define this in a .buildtools-file on a relevant level
#DOCKER_REGISTRY_URL=""
## END CONFIG BLOCK

declare -A valid_environments

sourceBuildToolsFiles

## The following methods are ECR specific
# Logs in to the docker repository
docker:login() {
  $(aws ecr get-login --no-include-email --region eu-west-1)
}

docker:ecr_create() {
  local IMAGE_NAME=$(ci:build_name)
  aws ecr create-repository --region eu-west-1 --repository-name ${IMAGE_NAME} &> /dev/null || true

  aws ecr put-lifecycle-policy --repository-name ${IMAGE_NAME} \
 --cli-input-json '{    "lifecyclePolicyText": "{\"rules\":[{\"rulePriority\":10,\"description\":\"Only keep 20 images\",\"selection\":{\"tagStatus\":\"untagged\",\"countType\":\"imageCountMoreThan\",\"countNumber\":20},\"action\":{\"type\":\"expire\"}}]}"}' &> /dev/null || true
}

## End ECR specific

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
  local DOCKER_BUILD_PATH=.
  local DOCKER_ADDITIONAL_ARGS="${@}"
  local DOCKER_TAG=$(vcs:getBranchReplaceSlash)
  local IMAGE_NAME=$(ci:build_name)
  local COMMIT=$(ci:commit)
  echo "Trying to build docker image [${DOCKER_REGISTRY_URL}/${IMAGE_NAME}]"
  try eval $(echo docker build --pull --shm-size 256m --memory=3g --memory-swap=-1 -t ${DOCKER_REGISTRY_URL}/${IMAGE_NAME}:${COMMIT} ${DOCKER_BUILD_PATH} ${DOCKER_ADDITIONAL_ARGS})
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
  docker:ecr_create
  local DOCKER_TAG=$(vcs:getBranchReplaceSlash)
  local IMAGE_NAME=$(ci:build_name)
  local COMMIT=$(ci:commit)

  if [[ "${DOCKER_TAG}" == "master" ]];
  then
    docker push ${DOCKER_REGISTRY_URL}/${IMAGE_NAME}:latest
  fi
  docker push ${DOCKER_REGISTRY_URL}/${IMAGE_NAME}:${DOCKER_TAG}
  docker push ${DOCKER_REGISTRY_URL}/${IMAGE_NAME}:${COMMIT}
)
}
