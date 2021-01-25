# Continuous Integration

Commands recognize which CI/CD environment they are executed in based on which environment variables are present.
The following variables are checked to determine which CI/CD tool that we run under.
Feel free to add support for more by supplying a Pull Request :).

| Environment variable      | CI/CD                             |
| :-------------------------| :-------------------------------- |
| `BUILDKITE_PIPELINE_SLUG` | [Buildkite]           |
| `CI_PROJECT_NAME`         | [Gitlab CI]           |
| `RUNNER_WORKSPACE`        | [Github Actions] |
| `TEAMCITY_PROJECT_NAME`   | [TeamCity]             |
| `BUILD_REPOSITORY_NAME`   | [Azure Devops]     |


## Configuration file from environment variables
A `.buildtools.yaml` file can be created by defining an environment variable in the build pipeline named `BUILDTOOLS_CONTENT`.
The value should be a base64-encoded string. This can be useful when setting up CI/CD pipelines where the file system is not
easily accessible.

On MacOS the value can be created and copied to the clipboard using the following snippet:

```sh
$ cat - <<EOF | base64 -w0 | pbcopy
targets:
  local-test:
    context: docker-desktop
)
EOF
```

[Buildkite]: https://buildkite.com
[Gitlab CI]: https://docs.gitlab.com/ce/ci
[Github Actions]: https://github.com/features/actions
[teamcity]: https://www.jetbrains.com/teamcity
[azure devops]: https://azure.microsoft.com/en-us/services/devops/pipelines/

