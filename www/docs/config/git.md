# Git

The `git` key in `.buildtools.yaml` defines the git configuration used for the project.
This will primarily be used for CI pipelines to push deployment descriptors images,
i.e the `promote` command.

TODO Use values from `~/.ssh/config`

|      Key              |                   Description       |
| :-------------------- | :---------------------------------- |
| `name`                | The name to use as author for the [commit] message |
| `email`               | The email to use as author for the [commit] message |
| `key`                 | Override the default ssh key (`~/.ssh/id_rsa`) |

[commit]: https://git-scm.com/docs/git-commit
