# Azure Devops

[Azure Devops] is configured with a `azure-pipelines.yml` file in your project.

````yaml
resources:
  containers:
  - container: build-tools
    image: buildtool/build-tools:latest

jobs:
- job: build_and_deploy
  pool:
    vmImage: 'Ubuntu 16.04'
  container: build-tools
  steps:
  - script: |
      build
      push
    name: build
    env:
      QUAY_PASSWORD: $(QUAY_PASSWORD)
  - script: deploy staging
    name: deploy_staging
    condition: succeeded()
````

[azure devops]: https://azure.microsoft.com/en-us/services/devops/pipelines/
