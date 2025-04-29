# Gitlab CI

[Gitlab CI] is configured with a `.gitlab-ci.yaml` file in your project.

````yaml
stages:
  - build
  - deploy-staging
  - deploy-prod

variables:
  DOCKER_HOST: tcp://docker:2375/

image: buildtool/build-tools:latest

build:
  stage: build
  services:
    - docker:dind
  script:
  - build
  - push

deploy-to-staging:
  stage: deploy-staging
  when: on_success
  script:
    - echo Deploy to staging.
    - deploy staging
  environment:
    name: staging

deploy-to-prod:
  stage: deploy-prod
  when: on_success
  script:
    - echo Deploy to PROD.
    - deploy prod
  environment:
    name: prod
  only:
    - main
````

[Gitlab CI]: https://docs.gitlab.com/ci
