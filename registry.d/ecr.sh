#!/usr/bin/env bash

if [ -n "${ECR_URL:-}" ]; then
  registry:login() {
    $(aws ecr get-login --no-include-email --region ${ECR_REGION:-eu-west-1})
  }

  registry:create() {
    local IMAGE_NAME=$(ci:build_name)
    aws ecr create-repository --region ${ECR_REGION:-eu-west-1} --repository-name ${IMAGE_NAME} &> /dev/null || true

    aws ecr put-lifecycle-policy --repository-name ${IMAGE_NAME} \
   --cli-input-json '{    "lifecyclePolicyText": "{\"rules\":[{\"rulePriority\":10,\"description\":\"Only keep 20 images\",\"selection\":{\"tagStatus\":\"untagged\",\"countType\":\"imageCountMoreThan\",\"countNumber\":20},\"action\":{\"type\":\"expire\"}}]}"}' &> /dev/null || true
  }
fi
