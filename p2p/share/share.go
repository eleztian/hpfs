package share

import (
	"sync"
	"path/filepath"

	"bs-2018/hpfs-client/errors"
	"bs-2018/hpfs-client/mhttp"
	"fmt"
)

var allShredDirs sync.Map

func init() {
	allShredDirs = sync.Map{}
}

func GetDirs() map[string]string {
	r := make(map[string]string)
	allShredDirs.Range(func(key, value interface{}) bool {
		r[key.(string)] = value.(string)
		return true
	})
	return r
}

func SetDirs(dir string, dirs ...string) error {
	d, f := filepath.Split(filepath.Clean(dir))
	if d == "" || f == "" {
		return errors.New("can not a valiable dir path")
	}
	allShredDirs.Store(f, d)
	mhttp.Update(true, dir)
	for _, dir = range dirs {
		d, f = filepath.Split(filepath.Clean(dir))
		if d == "" || f == "" {
			continue
		}
		allShredDirs.Store(f, d)
		mhttp.Update(true, dir)
	}
	return nil
}

func DelDir(dir string, dirs ...string) {
	name := GetNameByRelPath(dir)
	allShredDirs.Delete(name)
	mhttp.Update(false, dir)
	for _, dir = range dirs {
		name = GetNameByRelPath(dir)
		allShredDirs.Delete(name)
		mhttp.Update(false, dir)
	}
	fmt.Println("del.....", name)
}

func GetDir(name string) string {
	name = GetNameByPath(name)
	r, ok := allShredDirs.Load(name)
	if !ok {
		return ""
	}
	return filepath.Join(r.(string), name)
}
