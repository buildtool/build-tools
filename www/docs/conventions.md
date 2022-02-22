# Conventions

- The project folder must be a [Git](https://git-scm.com/) repository, with a least one commit
- `Dockerfile` should be present in the root of the project directory
  (this can be overridden with [flags](/commands/build)).
  The `Dockerfile` will be used to build the project into a runnable docker image.
- Kubernetes descriptor files must be located in the `k8s` folder (only needed for `deploy` and `promote`)
- The `k8s` folder can also contain custom scripts that should be run during deployment
- The name of the directory will be used as the name of the application
    - If running in CI, `ENV` variables will be used to determine the name of the project being built
    - The name can also be overridden using the `IMAGE_NAME` environment variable
- The current commit id will be used as docker tag
- The current branch will be used as docker tag. If you're on the `master` or `main`
  branch the docker image will also be tagged `latest`. The `latest` tag will also be pushed in that case.
- [Targets](config/targets.md) (deployment targets) are configured in [`.buildtools.yaml` file(s)](config/config.md)
- [`.buildtools.yaml` file(s)](config/config.md) will be merged together hierarchically and can be used for multiple
  projects
- Use [Target](config/targets.md) names to use specific `k8s` files for different deployment targets

