package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"gopkg.in/ini.v1"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

type TokenReplaceConfig struct {
	GiteaRepositoriesDir string `json:"gitea_repositories_dir"`
	// regex :owner/:repo => token
	Rules map[string]string `json:"rules"`
}

func NewTokenReplaceConfig(f string) (*TokenReplaceConfig, error) {
	file, err := os.ReadFile(f)
	if err != nil {
		return nil, err
	}
	var cfg TokenReplaceConfig
	err = json.Unmarshal(file, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func main() {
	m := flag.String("m", "preview", "mode: preview or replace")
	c := flag.String("c", "config.json", "config file")
	h := flag.Bool("h", false, "help")
	flag.Parse()

	if *h {
		flag.Usage()
		return
	}

	conf, err := NewTokenReplaceConfig(*c)
	if err != nil {
		log.Fatalf("Fail to read config file: %v", err)
	}

	regexps := make(map[string]regexp.Regexp)
	for k, _ := range conf.Rules {
		regex, e := regexp.Compile(k)
		if e != nil {
			log.Fatalf("Fail to compile regex: %v", e)
		}
		regexps[k] = *regex
	}

	err = filepath.Walk(conf.GiteaRepositoriesDir, func(userPath string, user os.FileInfo, err error) error {
		// 第一层是用户目录
		if user.IsDir() {
			// 继续遍历用户目录下的仓库
			return filepath.Walk(userPath, func(repoPath string, repo os.FileInfo, err error) error {
				if !strings.HasSuffix(repo.Name(), ".git") {
					return nil
				}

				// 生成正则匹配的ID
				id := path.Join(user.Name(), repo.Name())
				id = strings.TrimRight(id, ".git")

				// 加载.git/config文件
				cfgPath := path.Join(repoPath, "config")
				if _, e := os.Stat(cfgPath); os.IsNotExist(e) {
					return nil
				}
				cfg, err := ini.Load(cfgPath)
				if err != nil {
					return err
				}

				// 获取 remote "origin" 的 url
				remote, err := cfg.GetSection("remote \"origin\"")
				if err != nil {
					return err
				}
				u, err := remote.GetKey("url")
				if err != nil {
					return err
				}

				// 根据 mode 执行操作
				switch *m {
				case "preview":
					fmt.Printf("ID: %s, URL: %s\n", id, u.String())
				case "replace":
					token := ""
					for k, v := range regexps {
						if v.MatchString(id) {
							token = conf.Rules[k]
							break
						}
					}
					if token == "" {
						return nil
					}
					// 替换 url 中的 password
					parseUrl, rErr := url.Parse(u.String())
					if rErr != nil {
						return rErr
					}
					parseUrl.User = url.UserPassword(parseUrl.User.Username(), token)
					u.SetValue(parseUrl.String())
					// 保存修改
					rErr = cfg.SaveTo(cfgPath)
					if rErr != nil {
						return rErr
					}

				}
				return nil
			})
		}
		return nil
	})

}
