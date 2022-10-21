# Github Actions
Build-tools can also be used within our official build-tools actions through [GitHub Actions]

You can create a workflow by putting YAML configuration to `.github/workflows/build.yml`.

Below is a simple snippet to use the [setup-buildtools-action] in your workflow:

```yaml
name: Buildtool
on: [push]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v1
      - uses: buildtool/setup-buildtools-action@v1
      - run: builds
```

Read more about available [commands](/commands/build):

> For detailed intructions please follow GitHub Actions [syntax].

[Github Actions]: https://github.com/features/actions
[setup-buildtools-action]: https://github.com/buildtool/setup-buildtools-action
[syntax]: https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#About-yaml-syntax-for-workflows
