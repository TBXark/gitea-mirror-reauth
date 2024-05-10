# gitea-mirror-reauth

As Gitea does not provide a simple way to modify the authorization information of mirrored repositories, I wrote this tool to solve this problem.

### Installation

```bash
go get -u github.com/tbxark/gitea-mirror-reauth@latest
```

### Usage

```
Usage of gitea-mirror-reauth:
  -c string
        config file
  -d string
        gitea repositories dir (default "/home/git/data/gitea-repositories")
  -h    help
  -m string
        mode: preview or replace (default "preview")
  -r string
        replace mode: auto or manual (default "manual")
```

### Config file

```json
{
  "regex": "NEW_TOKEN",
  "tbxark/private_repo_name": "NEW_PRIVATE_SCOPE_TOKEN",
  "tbxark/.*": "NEW_PUBLIC_SCOPE_TOKEN"
}
```

The configuration file is a JSON file, the key is regular expression, and the value is the new token. The regular expression is used to match the repository path, and the new token is used to replace the old token.

### License

**gitea-mirror-reauth** is released under the MIT license. See [LICENSE](LICENSE) for details.
