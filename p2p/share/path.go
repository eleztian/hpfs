package share

import (
	"strings"
	"path/filepath"
)

func GetNameByPath(path string) string {
	path = filepath.Clean(path)
	path = strings.Replace(path, "/", "\\", -1)
	ps := strings.Split(path, "\\")

	if len(ps) == 0 {
		return ""
	}
	if ps[0] == "" && len(ps) > 1 {
		return ps[1]
	}
	return ps[0]
}

func GetNameByRelPath(path string) string {
	_,f := filepath.Split(filepath.Clean(path))
	return f
}

func GetAbs(path string) string {
	path = filepath.Clean(path)
	name := GetNameByPath(path)
	p, ok := allShredDirs.Load(name)
	if !ok {
		return ""
	}
	return filepath.Join(p.(string), path)
}
