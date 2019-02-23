/*
	package config is using to read config from config file.
*/
package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"fmt"
	"net/http"
	"bs-2018/hpfs-client/p2p"
	"bs-2018/hpfs-client/p2p/share"
)

var (
	AppConfig *Config
)

type Config struct {
	GitPath    string   `json:"git_path"`
	HomeDir    string   `json:"-"`
	HostUrl    string   `json:"host_url"`
	SharePaths []string `json:"share_paths"`

	User struct{
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"user"`

	RPCPort string `json:"rpc_port"`
}

func init() {
	u, err := user.Current()
	if err != nil {
		panic(err)
	}
	path := filepath.Join(u.HomeDir, ".hpfs", "config.json")
	f, err := ReadConfig(path)
	if err != nil {
		panic(err)
	}
	if f == nil {
		panic("not config json")
	}
	defer configServer(path)
	AppConfig = f
	fmt.Println(AppConfig)
	AppConfig.HomeDir = path


}

// ReadConfig read and unmarshal by json from filename.
// if config file is not exit it will create it by default config.
func ReadConfig(filename string) (*Config, error) {
	logrus.Println("open config file ", filename)
	f, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) { // 配置文件不存在使用默认配置文件
			err = createDefaultFile(filename)
			if err != nil {
				panic(err)
			}
		}
		return nil, err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	config := &Config{}
	err = json.Unmarshal(b, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func configServer(path string) {
	http.HandleFunc("/config", func(writer http.ResponseWriter, request *http.Request) {
		s , _ := json.Marshal(AppConfig)
		writer.Write(s)
	})
	http.HandleFunc("/config/update", func(writer http.ResponseWriter, request *http.Request) {
		b, err := ioutil.ReadAll(request.Body)
		if err != nil {
			panic(err)
		}
		newConfig := Config{}
		err = json.Unmarshal(b, &newConfig)
		if err != nil {
			panic(err)
		}
		for _, d := range AppConfig.SharePaths {
			flag := false
			for _, d2 := range newConfig.SharePaths {
				if d == d2 {
					flag = true
					break
				}
			}
			if !flag {
				share.DelDir(d)
			}
		}

		f, err := os.OpenFile(path, os.O_RDWR|os.O_TRUNC, os.ModePerm)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		f.Write(b)
		AppConfig = &newConfig

		p2p.SetHost(AppConfig.HostUrl)
		p2p.SetServerUP(AppConfig.User.Username, AppConfig.User.Password)
		p2p.SetSharedDirs(AppConfig.SharePaths...)
		p2p.UpdateGateway(AppConfig.RPCPort)
	})
 	go http.ListenAndServe(":8080", nil)

}