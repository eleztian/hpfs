package server

import (
	"bs-2018/hpfs-client/p2p/rpc"
	"bs-2018/hpfs-client/p2p/share"
)

type User struct {
}

// var allToken = make(map[string]*Token)
// var allTimer = make(map[string]*time.Timer)

type Token struct {
	Token    string
	Username string
}

func (*User) VerifyUP(args rpc.Args, reply *bool) error {
	r := args.Auth()
	*reply = r
	return nil
}

// func (*User) VerifyToken(args *Args, reply *bool) error {
// 	t, ok := allToken[args.Username]
// 	if ok {
// 		if args.Token == t.Token {
// 			*reply = true
// 		}
// 		return nil
// 	}
// 	*reply = false
// 	return nil
// }

func (*User) ListShareDirs(_ rpc.Args, replay *[]string) error {
	r := []string{}
	for k := range share.GetDirs() {
		r = append(r, k)
	}
	*replay = r
	return nil
}

// func newToken(username, password string, exprice time.Duration) Token {
// 	hash := md5.New()
// 	hash.Write([]byte(username + password + time.Now().String()))
// 	token := fmt.Sprintf("%x", hash.Sum(nil))
// 	timer := time.AfterFunc(exprice, func() {
// 		delete(allToken, username)
// 	})
// 	t := Token{
// 		Token:    token,
// 		Username: username,
// 	}
// 	timerOld, ok := allTimer[username]
// 	if ok {
// 		timerOld.Stop()
// 	}
// 	allTimer[username] = timer
// 	allToken[username] = &t
// 	return t
// }
