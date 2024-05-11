package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {

	if len(os.Args) < 2 {
		handleMissingSubcommand()
	}

	previewCmd := flag.NewFlagSet("preview", flag.ExitOnError)
	autoReplaceCmd := flag.NewFlagSet("auto-replace", flag.ExitOnError)
	tokenReplaceCmd := flag.NewFlagSet("token-replace", flag.ExitOnError)

	var giteaDir string
	var configFilePath string
	var confirm bool
	var help bool

	subCommands := []*flag.FlagSet{previewCmd, autoReplaceCmd, tokenReplaceCmd}

	autoReplaceCmd.StringVar(&configFilePath, "config", "", "config file path")
	autoReplaceCmd.BoolVar(&confirm, "confirm", false, "confirm")

	for _, cmd := range subCommands {
		cmd.StringVar(&giteaDir, "gitea-dir", "/home/git/data/gitea-repositories", "gitea repositories dir")
		cmd.BoolVar(&help, "help", false, "help")
		if os.Args[1] == cmd.Name() {
			err := cmd.Parse(os.Args[2:])
			if err != nil {
				panic(err)
			}
			if help {
				cmd.Usage()
				os.Exit(0)
			}
			break
		}
	}

	switch os.Args[1] {
	case previewCmd.Name():
		handlePreview(giteaDir)
	case autoReplaceCmd.Name():
		handleAutoReplace(giteaDir, configFilePath, confirm)
	case tokenReplaceCmd.Name():
		handleTokenReplace(giteaDir)
	default:
		handleMissingSubcommand()
	}

}

func handleMissingSubcommand() {
	fmt.Println("gitea-mirror-reauth")
	fmt.Println("expected 'preview', 'auto-replace' or 'token-replace' subcommands")
	fmt.Println("Usage:")
	fmt.Println("  preview       preview all gitea repositories")
	fmt.Println("  auto-replace  auto replace gitea repositories token by config")
	fmt.Println("  token-replace replace gitea repositories token manually")
	os.Exit(1)
}

func getRemoteOriginURL(dir string) (string, error) {
	// git --git-dir=$dir config --get remote.origin.url
	cmd := exec.Command("git", "--git-dir="+dir, "config", "--get", "remote.origin.url")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func setRemoteOriginURL(dir string, url string) error {
	// git --git-dir=/path/to/repo/.git remote set-url origin $url
	cmd := exec.Command("git", "--git-dir="+dir, "remote", "set-url", "origin", url)
	return cmd.Run()
}

func getAllRemoteBranches(dir string) ([]string, error) {
	cmd := exec.Command("git", "--git-dir="+dir, "branch", "-r")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	branches := strings.Split(string(out), "\n")
	res := make([]string, 0)
	for _, branch := range branches {
		branch = strings.TrimSpace(branch)
		if branch == "" {
			continue
		}
		res = append(res, branch)
	}
	return res, nil
}

type giteaRepo struct {
	ID  string `json:"id"`
	Dir string `json:"dir"`
	URL string `json:"url"`
}

func loadGiteaRepos(dir string) ([]giteaRepo, error) {
	res := make([]giteaRepo, 0)
	err := filepath.Walk(dir, func(userPath string, user os.FileInfo, err error) error {
		// 第一层是用户目录
		if user.IsDir() {
			// 继续遍历用户目录下的仓库
			return filepath.Walk(userPath, func(repoPath string, repo os.FileInfo, err error) error {
				if !strings.HasSuffix(repoPath, ".git") {
					return nil
				}
				// 生成正则匹配的ID
				// /home/git/data/gitea-repositories/tbxark/gitea.git => tbxark/gitea
				split := strings.Split(repoPath, "/")
				id := strings.Join(split[len(split)-2:], "/")
				id = strings.TrimSuffix(id, ".git")

				// 获取 remote "origin" 的 url
				remoteUrl, err := getRemoteOriginURL(repoPath)
				if err != nil {
					return nil
				}
				res = append(res, giteaRepo{
					ID:  id,
					Dir: repoPath,
					URL: remoteUrl,
				})
				return nil

			})
		}
		return nil
	})
	return res, err
}

func handlePreview(giteaDir string) {
	repos, err := loadGiteaRepos(giteaDir)
	if err != nil {
		panic(err)
	}
	for _, repo := range repos {
		fmt.Printf("%s\n\t%s\n\t%s\n\n", repo.ID, repo.Dir, repo.URL)
		branches, e := getAllRemoteBranches(repo.Dir)
		if e == nil && len(branches) > 0 {
			fmt.Printf("\tBranches:\n")
			for _, branch := range branches {
				fmt.Printf("\t\t%s\n", branch)
			}
		}
	}
}

func handleAutoReplace(giteaDir string, configFilePath string, confirm bool) {
	configFileRaw, err := os.ReadFile(configFilePath)
	if err != nil {
		panic(err)
	}
	var config map[string]string
	err = json.Unmarshal(configFileRaw, &config)
	if err != nil {
		panic(err)
	}
	regexps := make(map[string]*regexp.Regexp)
	for k, v := range config {
		regexps[k] = regexp.MustCompile(v)
	}
	repos, err := loadGiteaRepos(giteaDir)
	if err != nil {
		panic(err)
	}
	for _, repo := range repos {
		for k, v := range regexps {
			if v.MatchString(repo.ID) {
				urlParse, re := url.Parse(repo.URL)
				if re != nil {
					continue
				}
				if urlParse.User == nil {
					continue
				}
				urlParse.User = url.UserPassword(urlParse.User.Username(), config[k])
				newURL := urlParse.String()
				if confirm {
					fmt.Printf("Replace %s\n", repo.ID)
					fmt.Printf("\t%s\n", repo.URL)
					fmt.Printf("\t%s\n", newURL)
					fmt.Print("Confirm? [y/N]: ")
					var confirmStr string
					_, e := fmt.Scanln(&confirmStr)
					if e != nil || confirmStr != "y" {
						continue
					}
				}
				re = setRemoteOriginURL(repo.Dir, newURL)
				if re != nil {
					fmt.Printf("Error: %s\n", re)
				} else {
					fmt.Printf("Success: %s\n", newURL)
				}
			}
		}
	}
}

func handleTokenReplace(giteaDir string) {
	repos, err := loadGiteaRepos(giteaDir)
	if err != nil {
		panic(err)
	}
	tokens := make(map[string][]giteaRepo)
	for _, repo := range repos {
		urlParse, e := url.Parse(repo.URL)
		if e != nil {
			continue
		}
		if urlParse.User == nil {
			continue
		}
		token, ok := urlParse.User.Password()
		if !ok {
			continue
		}
		tokens[token] = append(tokens[token], repo)
	}
	fmt.Printf("Found %d tokens\n", len(tokens))
	for token, list := range tokens {
		fmt.Printf("Token: %s\n", token)
		var newToken string
		fmt.Print("New Token: ")
		_, e := fmt.Scanln(&newToken)
		if e != nil {
			continue
		}
		for _, repo := range list {
			fmt.Printf("Replace %s\n", repo.ID)
			// 替换URL
			urlParse, re := url.Parse(repo.URL)
			if re != nil {
				continue
			}
			urlParse.User = url.UserPassword(urlParse.User.Username(), newToken)
			newURL := urlParse.String()
			re = setRemoteOriginURL(repo.Dir, newURL)
			if re != nil {
				fmt.Printf("Error: %s\n", re)
			} else {
				fmt.Printf("Success: %s\n", newURL)
			}
		}

	}
}
