package common

import (
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/BurntSushi/toml"
)

func LoadConfig(config any, configdir, app string) {
	// 读区business.form.toml,并发送http://configserver/config/render?app=$app
	// 读取文件
	// 使用本地配置文件覆盖远程文件， 开发人员自己保证

	file, err := os.Open(path.Join(configdir, "business.form.toml"))
	if err == nil {
		resp, err := http.Post("http://configserver/config/render?app="+app, "text/text", file)
		if err == nil {
			defer resp.Body.Close()
			_, err = toml.NewDecoder(resp.Body).Decode(config)
			if err != nil {
				panic(err)
			}
		} else {
			fmt.Printf("failed to post business.form.toml to configserver reason: %s\n", err.Error())
		}
	} else {
		fmt.Printf("failed to read business.form.toml reason: %s\n", err.Error())
	}
	LoadConfigFile(config, path.Join(configdir, "business.public.toml"))
	LoadConfigFile(config, path.Join(configdir, "business.private.toml"))
}
func LoadConfigFile(config any, file string) bool {
	buf, err := os.ReadFile(file)
	if err == nil {
		_, err = toml.Decode(string(buf), config)
		if err != nil {
			panic(err)
		}
		return true
	}
	return false
}
