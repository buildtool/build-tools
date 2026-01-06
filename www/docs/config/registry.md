# Registry

The `registry` key in `.buildtools.yaml` defines the docker registry used for the project.
This will primarily be used for CI pipelines to push built docker images, i.e the `push` command.

Locally it can be used to build images with correct tags,
making it possible to deploy locally built images.

Each supported registry has their own configuration keys, typically the setup looks like this:

````yaml
registry:
  <registry name>:
    <specific config>
````

## Supported registries
The following registries are supported:

| Config key                | Container registry                                                                                                   |
|:--------------------------|:---------------------------------------------------------------------------------------------------------------------|
| [`dockerhub`](#dockerhub) | [Docker hub](https://hub.docker.com/)                                                                                |
| [`acr`](#acr)             | [Azure Container Registry](https://azure.microsoft.com/en-us/products/container-registry)                            |
| [`ecr`](#ecr)             | [AWS Elastic Container Registry](https://docs.aws.amazon.com/ecr/index.html)                                         |
| [`github`](#github)       | [Github package registry](https://docs.github.com/en/packages/learn-github-packages/introduction-to-github-packages) |
| [`gitlab`](#gitlab)       | [Gitlab container registry](https://docs.gitlab.com/ee/user/packages/container_registry/)                            |
| [`gitea`](#gitea)         | [Gitea container registry](https://docs.gitea.com/usage/packages/container)                                          |
| [`quay`](#quay)           | [Quay docker registry](https://docs.quay.io/)                                                                        |
| [`gcr`](#gcr)             | [Google Container registry](https://cloud.google.com/container-registry)                                             |

### dockerhub

| Parameter         | Description                          | Env variable           |
| :---------------- | :----------------------------------- | :--------------------- |
| `namespace`       |  The namespace to publish to         | `DOCKERHUB_NAMESPACE`  |
| `username`        |  User to authenticate                | `DOCKERHUB_USERNAME`   |
| `password`        |  Password for `user` authentication  | `DOCKERHUB_PASSWORD`   |

### acr

A valid Azure CLI session must be available (via `az login`), read more [here](https://learn.microsoft.com/en-us/cli/azure/authenticate-azure-cli).

| Parameter  | Description                                               | Env variable    |
|:-----------|:----------------------------------------------------------|:----------------|
| `url`      | The ACR registry URL                                      | `ACR_URL`       |
| `tenantId` | The id of the Azure tenant where the registry belongs to) | `ACR_TENANT_ID` |

### ecr

AWS Credentials must be supplied as `ENV` variables, read more [here](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html).

| Parameter | Description                                                                                | Env variable           |
| :-------- | :----------------------------------------------------------------------------------------- | :--------------------- |
| `url`     | The ECR registry URL                                                                       | `ECR_URL`              |
| `region`  | Specify a region (if it's possible to derive from the `url` parameter it can be omitted)   | `ECR_REGION`           |

### github

To authenticate `token` or a combination of `username` and `password` must be provided.

| Parameter       | Description                                          | Env variable             |
| :-------------- | :--------------------------------------------------- | :----------------------- |
| `repository`    | The repository part of the docker image name         | `GITHUB_REPOSITORY`      |
| `username`      | User to authenticate                                 | `GITHUB_USERNAME`        |
| `password`      | Password for `user` authentication                   | `GITHUB_PASSWORD`        |
| `token`         | A personal access token to use for authentication    | `GITHUB_TOKEN`           |


### gitlab


| Parameter    | Description                                       | Env variable        |
|:-------------|:--------------------------------------------------|:--------------------|
| `registry`   | The registry part of the docker image name        | `CI_REGISTRY`       |
| `repository` | The repository part of the docker image name      | `CI_REGISTRY_IMAGE` |
| `user`       | User to authenticate                              | `CI_REGISTRY_USER`  |
| `token`      | A personal access token to use for authentication | `CI_JOB_TOKEN`      |

### gitea

Self-hosted [Gitea](https://gitea.com/) instances with container registry support (v1.17+).

Note: Gitea Actions does not yet support automatic token authentication for package registries. You must create a Personal Access Token with `package:write` scope and provide it via the `token` parameter or `GITEA_TOKEN` environment variable.

| Parameter    | Description                                       | Env variable        |
|:-------------|:--------------------------------------------------|:--------------------|
| `registry`   | The Gitea server hostname                         | `GITEA_REGISTRY`    |
| `repository` | The repository path (owner/repo format)           | `GITEA_REPOSITORY`  |
| `username`   | User to authenticate                              | `GITEA_USERNAME`    |
| `token`      | A personal access token to use for authentication | `GITEA_TOKEN`       |

Example configuration:

```yaml
registry:
  gitea:
    registry: gitea.example.com
    username: myuser
    token: ${GITEA_TOKEN}
    repository: myorg/myrepo
```

### quay


| Parameter       | Description                                          | Env variable         |
| :-------------- | :--------------------------------------------------- | :------------------- |
| `repository`    | The repository part of the docker image name         | `QUAY_REPOSITORY`    |
| `username`      | User to authenticate                                 | `QUAY_USERNAME`      |
| `password`      | Password for `user` authentication                   | `QUAY_PASSWORD`      |

### gcr

GCP Credentials must be supplied as [service account json key](https://cloud.google.com/container-registry/docs/advanced-authentication#json-key) (Base64 encoded)

| Parameter         | Description                       | Env variable           |
| :---------------- | :-------------------------------- | :--------------------- |
| `url`             | The GCR registry URL              | `GCR_URL`              |
| `keyfileContent`  | ServiceAccount keyfile content    | `GCR_KEYFILE_CONTENT`  |
