package main

import (
	"encoding/json"
	"github.com/asticode/go-astilectron"
	"github.com/asticode/go-astilectron-bootstrap"

	"bs-2018/hpfs-client/config"
	"bs-2018/hpfs-client/mhttp"
	"bs-2018/hpfs-client/p2p"
	"bs-2018/hpfs-client/p2p/client"
	"bs-2018/hpfs-client/p2p/client/fs"
	"bs-2018/hpfs-client/p2p/git"
	"bs-2018/hpfs-client/p2p/share"
	"github.com/asticode/go-astilog"
	"github.com/pkg/errors"
)

// handleMessages handles messages
func handleMessages(w *astilectron.Window, m bootstrap.MessageIn) (payload interface{}, err error) {
	switch m.Name {
	case "exploreSharedDirs":
		payload = getSharedDirs()
	case "searchFiles":
		var c = struct {
			Filename string
			Filepath string
			Uid      string
		}{}
		err = json.Unmarshal(m.Payload, &c)
		if err != nil {
			return
		}
		payload, err = mhttp.GetMetas(c.Uid, c.Filename, c.Filepath)
	case "login":
		user := mhttp.User{}
		err = json.Unmarshal(m.Payload, &user)
		if err != nil {
			return
		}
		err = mhttp.Login(&user)
		if err != nil {
			return
		}
		p2p.SetSharedDirs(config.AppConfig.SharePaths...)
		go p2p.UpdateGateway(config.AppConfig.RPCPort)
		go share.Update()

	case "register":
		user := mhttp.User{}
		err = json.Unmarshal(m.Payload, &user)
		if err != nil {
			return
		}
		err = mhttp.Register(&user)

	case "showGitLog":
		var name string
		err = json.Unmarshal(m.Payload, &name)
		if err != nil {
			return
		}
		payload, err = getGitLog(name)
	case "mount":
		var msg = struct {
			Uid   string
			Point string
		}{}
		err = json.Unmarshal(m.Payload, &msg)
		if err != nil {
			return
		}
		url, err2 := mhttp.GetGateway(msg.Uid)
		if err2 != nil {
			err = err2
			return
		}
		err = client.Mount(msg.Point, url, msg.Uid, "tab", "tabzhang@gmail.com")
		if err != nil {
			return
		}
		payload = msg
	case "unmount":
		var name string
		err = json.Unmarshal(m.Payload, &name)
		if err != nil {
			return
		}
		err = client.UnMount(name)
	case "gitRest":
		var s = struct {
			Name string
			Sha1 string
		}{}
		err = json.Unmarshal(m.Payload, &s)
		if err != nil {
			return
		}
		err = fs.Reset(s.Name, s.Sha1)
	}

	return
}

func getSharedDirs() []string {
	r := []string{}
	for k := range share.GetDirs() {
		r = append(r, k)
	}
	return r
}

func getGitLog(name string) ([]*git.LogInfo, error) {
	d := share.GetDir(name)
	g, err := git.NewGit(d)
	if err != nil {
		return nil, err
	}
	return g.Log()
}

func updateShareDirs() {
	if err := bootstrap.SendMessage(w, "updateSharedDirs", getSharedDirs()); err != nil {
		astilog.Error(errors.Wrap(err, "sending updateSharedDirs event failed"))
	}
}
