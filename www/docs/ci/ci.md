# Continuous Integration

Commands recognize which CI/CD environment they are executed in based on which environment variables are present.
The following variables are checked to determine which CI/CD tool that we run under.

| Environment variable      | CI/CD                             |
| :-------------------------| :-------------------------------- |
| `BUILDKITE_PIPELINE_SLUG` | [Buildkite]           |
| `CI_PROJECT_NAME`         | [Gitlab CI]           |
| `RUNNER_WORKSPACE`        | [Github Actions] |
| `TEAMCITY_PROJECT_NAME`   | [TeamCity]             |
| `BUILD_REPOSITORY_NAME`   | [Azure Devops]     |


[Buildkite]: https://buildkite.com
[Gitlab CI]: https://docs.gitlab.com/ci/
[Github Actions]: https://github.com/features/actions
[TeamCity]: https://www.jetbrains.com/teamcity
[azure devops]: https://azure.microsoft.com/en-us/services/devops/pipelines/
