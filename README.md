# gitea-mirror-reauth

As Gitea does not provide a simple way to modify the authorization information of mirrored repositories, I wrote this tool to solve this problem.

## Installation

```bash
go install github.com/TBXark/gitea-mirror-reauth@latest
```

## Usage

```
gitea-mirror-reauth

Usage:
  preview       --gitea-dir /path/to/gitea-repositories
  auto-replace  --gitea-dir /path/to/gitea-repositories --config /path/to/config.json --confirm
  token-replace --gitea-dir /path/to/gitea-repositories
```

### preview
Check all the repositories in gitea-repositories and output the repository information
```
Usage of preview:
  -gitea-dir string
        gitea repositories dir (default "/home/git/data/gitea-repositories")
  -help
        help
```

### auto-replace
Replace the token in the repository according to the configuration file
```
Usage of auto-replace:
  -config string
        config file path
  -confirm
        confirm
  -gitea-dir string
        gitea repositories dir (default "/home/git/data/gitea-repositories")
  -help
        help
```
Configuration file is json format, key is regular expression, value is new token.
```json
{
  "regex": "NEW_TOKEN",
  "tbxark/private_repo_name": "NEW_PRIVATE_SCOPE_TOKEN",
  "tbxark/.*": "NEW_PUBLIC_SCOPE_TOKEN"
}
```

### token-replace
Find all the tokens in the repository and replace them
```
Usage of token-replace:
  -gitea-dir string
        gitea repositories dir (default "/home/git/data/gitea-repositories")
  -help
        help
```


The configuration file is a JSON file, the key is regular expression, and the value is the new token. The regular expression is used to match the repository path, and the new token is used to replace the old token.

## License

**gitea-mirror-reauth** is released under the MIT license. See [LICENSE](LICENSE) for details.
