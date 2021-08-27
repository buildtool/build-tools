# Github Actions
Build-tools can also be used within our official build-tools actions through [GitHub Actions][actions]

You can create a workflow for pushing your releases by putting YAML configuration to `.github/workflows/build.yml`.

Below is a simple snippet to use the [build-action] and [push-action] in your workflow:

```yaml
name: Buildtool
on: [push]
jobs:
  build:
    runs-on: ubuntu-latest
    env:
      DOCKERHUB_USERNAME: sparetimecoders
      DOCKERHUB_PASSWORD: ${{ secrets.DOCKERHUB_PASSWORD }}
      DOCKERHUB_NAMESPACE: sparetimecoders
    steps:
      - name: Checkout
        uses: actions/checkout@v1
      - name: build
        uses: buildtool/build-action@v1
      - name: push
        uses: buildtool/push-action@v1
```

Read more about our actions here:
  - [build-action]
  - [push-action]
  - [deploy-action]

> For detailed intructions please follow GitHub Actions [workflow syntax][syntax].

[Github Actions]: https://github.com/features/actions
[build-action]: https://github.com/buildtool/build-action
[push-action]: https://github.com/buildtool/push-action
[deploy-action]: https://github.com/buildtool/deploy-action
[actions]: https://github.com/features/actions
[syntax]: https://help.github.com/en/articles/workflow-syntax-for-github-actions#About-yaml-syntax-for-workflows
