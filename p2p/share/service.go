package share

import (
	"time"
	"bs-2018/hpfs-client/mhttp"
	"path/filepath"
)

func Update() {
	timer := time.NewTimer(10*time.Minute)

	for {
		select {
		case <-timer.C:
			allShredDirs.Range(func(key, value interface{}) bool {
				mhttp.Update(true, filepath.Join(value.(string), key.(string)))
				return true
			})
			timer.Reset(10*time.Minute)
		}
	}
}
