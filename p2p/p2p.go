/*
	package p2p is used to connect user.every client also is a server.
*/
package p2p

import (
	"bs-2018/hpfs-client/mhttp"
	"bs-2018/hpfs-client/p2p/rpc"
	"bs-2018/hpfs-client/p2p/server"
	"log"
	"os"
	"bs-2018/hpfs-client/p2p/share"
	"context"
	"time"
)

var RPCCfunc context.CancelFunc

func StartService(ctx context.Context, port string) error {
	path := "0.0.0.0:" + port
	log.Println("RPC server start at ", path)
	rpc.SetLogger(log.New(os.Stdout, "[RPC] ", log.Ldate|log.Ltime|log.Lshortfile))
	if RPCCfunc != nil {
		RPCCfunc()
		time.Sleep(1*time.Second)
	}
	c, f := context.WithCancel(ctx)
	RPCCfunc = f
	return rpc.StartRpc(c, path, new(server.User), new(server.File))
}

func SetServerUP(username, password string) {
	rpc.SetUP(username, password)
}

func SetSharedDirs(dirs ...string) {
	if len(dirs) == 0 {
		return
	}
	share.SetDirs(dirs[0], dirs[1:]...)
}

func SetHost(url string) {
	mhttp.SetHost(url)
}

func UpdateGateway(port string) {
	mhttp.UpdateGateway(port)
}

