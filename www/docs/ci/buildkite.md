# Buildkite

[Buildkite] is configured with `.buildkite/pipeline.yml` file in your project.

## Buildkite plugin
Using the Buildkite [plugin] for buildtools is probably the simplest way

```yaml
steps:
  - command: build
    label: build
    plugins:
      - buildtool/buildtools:
  - wait

  - command: push
    label: push
    plugins:
      - buildtool/buildtools:

  - block: ":rocket: Release PROD"
    branches: "main"

  - command: deploy prod
    label: Deploy PROD
    branches: "main"
    plugins:
      - buildtool/buildtools:
          config: s3://my-buildkite-secrets/configs/myapp/env
```

## Docker plugin
build-tools can also be used with the Buildkite [docker-plugin]

```yaml
steps:
  - command: |-
      build
      push
    label: build
    plugins:
      - docker#v3.3.0:
          image: buildtool/build-tools
          volumes:
            - "/var/run/docker.sock:/var/run/docker.sock"
          propagate-environment: true
  - wait

  - block: ":rocket: Release PROD"
    branches: "main"

  - command: |-
      deploy prod
    label: Deploy PROD
    branches: "main"
    plugins:
      - docker#v3.3.0:
          image: buildtool/build-tools
          volumes:
            - "/var/run/docker.sock:/var/run/docker.sock"
          propagate-environment: true
```

[Buildkite]: https://buildkite.com
[plugin]: https://github.com/buildtool/buildtools-buildkite-plugin
[docker-plugin]: https://github.com/buildkite-plugins/docker-buildkite-plugin
