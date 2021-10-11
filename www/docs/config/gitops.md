# $Gitops

`gitops` is the equivalent to [targets](./targets.md), but determines where the generated files from
`promote` will end up.

Each `<name>` must be unique and contains the [git push url](https://git-scm.com/docs/git-push) and
the path inside the repository where the files will be stored.

```yaml
gitops:
  <name>:
    url:
    path:
```

| Parameter     |  Description                                           |
| :------ |  :---------------------------------------------------  |
| `url`   | The git URL (for example `git@github.com:buildtool/build-tools.git`) |
| `path`  | Root path in the repository, files will be put under `$path/$name`, defaults to `/`         |

## Examples

````yaml
gitops:
  local:
    url: git@github.com:buildtool/build-tools.git
    path: local
  prod:
    url: git@github.com:buildtool/build-tools.git
    path: prod
````
