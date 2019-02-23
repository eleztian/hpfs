package mhttp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var CurrentUser = struct {
	Username string
	Token    string
	Expire   time.Time
}{}

type User struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	Description string `json:"description, omitempty"`
}

var DirMsg = make(chan struct {
	Update bool
	Dir    string
}, 10)

func init() {
	go UpdateService()
}

func Update(update bool, dirs ...string) {
	if CurrentUser.Token == "" {
		return
	}
	for _, dir := range dirs {
		DirMsg <- struct {
			Update bool
			Dir    string
		}{update, dir}
	}
}

func Login(u *User) error {
	b, err := json.Marshal(u)
	if err != nil {
		return err
	}
	rsp, err := Post("/v1/login", b, nil, "")

	if err != nil {
		return err
	}
	err = map2struct(rsp.(map[string]interface{}), &CurrentUser)
	if err != nil {
		return err
	}
	CurrentUser.Username = u.Username

	return nil
}

func Register(u *User) error {
	b, err := json.Marshal(u)
	if err != nil {
		return err
	}
	_, err = Post("/v1/register", b, nil, "")
	if err != nil {
		return err
	}
	return nil
}

type Meta struct {
	UID        string `json:"uid"`
	Path       string `json:"path"`
	Name       string `json:"name"`
	Size       int64  `json:"size"`
	Type       bool   `json:"type"`
	MTime      string `json:"m_time"`
	CTime      string `json:"c_time"`
	Permission int    `json:"permission"`
}

func GetMetas(uid, filename, filePath string) ([]*Meta, error) {
	p := map[string]string{}
	if uid != "" {
		p["uid"] = uid
	}
	if filePath != "" {
		p["path"] = filePath
	}
	if filename != "" {
		p["filename"] = filename
	}
	rsp, err := Get("/v1/meta", p, CurrentUser.Token)
	if err != nil {
		return nil, err
	}
	result := make([]*Meta, 0)
	r := rsp.([]interface{})
	for _, v := range r {
		ri := Meta{}
		err = map2struct(v.(map[string]interface{}), &ri)
		if err != nil {
			return nil, err
		}

		result = append(result, &ri)
	}
	return result, err

}

func UpdateMetas(uid, path string) error {
	metas := getFileInfo(uid, path)
	b, err := json.Marshal(metas)
	if err != nil {
		return err
	}
	_, err = Post("/v1/meta", b, nil, CurrentUser.Token)
	return err
}

func DelDir(uid, path string) error {
	_, err := Delete("/v1/meta", []byte(path), nil, CurrentUser.Token)
	return err
}

func getFileInfo(uid, path string) []*Meta {
	metas := make([]*Meta, 0)

	dir, _ := filepath.Split(path)
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.Contains(path, ".git") {
			return nil
		}
		m := &Meta{
			UID:   uid,
			Size:  info.Size(),
			Name:  info.Name(),
			MTime: info.ModTime().Format("2006-01-02 15:04:05"),
			CTime: info.ModTime().Format("2006-01-02 15:04:05"),
			Type:  info.IsDir(),
		}
		m.Path = strings.Replace(path[len(dir):], "\\", "/", -1)
		metas = append(metas, m)
		return nil
	})
	return metas
}

func GetGateway(uid string) (string, error) {
	p := map[string]string{}
	if uid != "" {
		p["uid"] = uid
	}
	rsp, err := Get("/v1/gateway", p, CurrentUser.Token)
	if err != nil {
		return "", err
	}

	return rsp.(string), nil
}

func UpdateGateway(url string) error {
	_, err := Post("/v1/gateway", []byte(url), nil, CurrentUser.Token)
	return err
}

func map2struct(m map[string]interface{}, to interface{}) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, to)
}
