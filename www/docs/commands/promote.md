# promote

Templates deployment descriptors and promotes them to a Git-repository of choice.
Normal usage `promote <target>`, but additional flags can be used to override:

|      Flag             |                   Description                                                   |
| :-------------------- | :-------------------------------------------------------------------------------|
| `--namespace`, `-n`   | Use a different namespace than the one found in configuration                   |
| `--tag`               | Override the default tag to use (instead of the current commit tag or the value from CI) |
| `--url`               | override the URL to the Git repository where files will be generated |
| `--path`              | override the path in the Git repository where files will be generated |
| `--user`              | username for Git access, defaults to `git` |
| `--key`               | private key for Git access, defaults to `~/.ssh/id_rsa` |
| `--password`          | password for private key, defaults to `""` |
| `--out` , `-o`        | write output to specified file instead of committing and pushing to Git |


## Default usage, with `.buildtools.yaml` file
Only the `target` name has to be specified
```sh
$ promote local
```

### Overriding namespace from config:
```sh
$ promote --namespace test local
```

