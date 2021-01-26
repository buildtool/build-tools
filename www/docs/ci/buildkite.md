# Buildkite

[Buildkite] is configured with `.buildkite/pipeline.yml` file in your project.

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
