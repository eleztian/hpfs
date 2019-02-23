package rpc

import (
	"log"
	"net/rpc"
	"context"
)

var (
	usernameC string
	passwordC string
)

func SetUP(username, password string) {
	usernameC, passwordC = username, password
}

var Logger *log.Logger

func SetLogger(l *log.Logger) {
	Logger = l
}

type User struct {
	Username string
	Email    string
	Password string
}

func (u *User) Auth() bool {
	if u.Username == usernameC && u.Password == passwordC {
		return true
	}
	return false
}

type Args struct {
	User
	Args interface{}
}

// func Call(client *rpc.Client, funcName string, args interface{}, replay interface{}) error {
// 	// conn, err := net.Dial("tcp", serverAddress)
// 	// if err != nil {
// 	// 	return err
// 	// }
// 	err := client.Call(funcName, args, replay)
// 	if err != nil {
// 		fmt.Println(err)
// 		return err
// 	}
// 	fmt.Println("Over")
// 	return nil
// }

func StartRpc(ctx context.Context, address string, recv ...interface{}) error {
	// defer func() {
	// 	if err := recover(); err != nil {
	//
	// 		log.Println(err)
	// 	}
	// }()
	err := register(recv...)
	if err != nil {
		return err
	}
	ListenRPC(ctx, address)
	return nil
}

func register(recv ...interface{}) error {
	for _, i := range recv {
		err := rpc.Register(i)
		if err != nil {
			return err
		}
	}
	return nil
}
