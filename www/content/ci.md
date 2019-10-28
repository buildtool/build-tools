---
title: Continuous Integration
menu: true
weight: 9
---

Command recognize which CI/CD environment they are executed in based on which environment variables are present.

| Environment variable      | CI/CD                 | Scaffolding  |
| :-------------------------| :-------------------- | :------------|
| `BUILDKITE_PIPELINE_SLUG` | [buildkite]           | ✅ |
| `CI_PROJECT_NAME`         | [gitlab ci]           | ✅ |
| `RUNNER_WORKSPACE`        | [github actions]      | |
| `TEAMCITY_PROJECT_NAME`   | [teamcity]            | |
| `BUILD_REPOSITORY_NAME`   | [azure devops]        | |



[buildkite]: https://buildkite.com
[gitlab ci]: https://docs.gitlab.com/ee/ci
[github actions]: https://github.com/features/actions
[teamcity]: https://www.jetbrains.com/teamcity
[azure devops]: https://azure.microsoft.com/en-us/services/devops/pipelines/