package mhttp

import (
	"log"
)

func UpdateService() {
	for msg := range DirMsg {
		if msg.Update {
			err := UpdateMetas(CurrentUser.Username, msg.Dir)
			if err != nil {
				log.Println("update server meta ", err)
			}
		} else {
			err := DelDir(CurrentUser.Username, msg.Dir)
			if err != nil {
				log.Println("delete server meta ", err)
			}
		}

	}
}
