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

func main() {
	d := flag.String("d", "/home/git/data/gitea-repositories", "gitea repositories dir")
	m := flag.String("m", "preview", "mode: preview or replace")
	c := flag.String("c", "", "config file")
	h := flag.Bool("h", false, "help")
	flag.Parse()

	if *h {
		flag.Usage()
		return
	}

	regexps := make(map[string]regexp.Regexp)
	tokens := make(map[string]string)
	for *c != "" {
		file, err := os.ReadFile(*c)
		if err != nil {
			break
		}
		var conf map[string]string
		err = json.Unmarshal(file, &conf)
		if err != nil {
			break
		}
		for k, v := range conf {
			regex, e := regexp.Compile(k)
			if e != nil {
				log.Fatalf("Fail to compile regex %s: %v", k, e)
			}
			regexps[k] = *regex
			tokens[k] = v
		}
		break
	}
	err := filepath.Walk(*d, func(userPath string, user os.FileInfo, err error) error {
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
							token = tokens[k]
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
	if err != nil {
		fmt.Printf("Fail to walk: %v", err)
	}
}
