package client

import (
	"bs-2018/hpfs-client/errors"
	"bs-2018/hpfs-client/p2p/client/fs"
	"bs-2018/hpfs-client/p2p/client/fuse"
	"fmt"
	"github.com/silenceper/pool"
	"net"
	"time"
)

var mounted = make(map[string]*struct {
	Fs   *fs.Rfs
	Host *fuse.FileSystemHost
})

func Mount(mountpoint, address, username, password, email string) error {
	factory := func() (interface{}, error) { return net.DialTimeout("tcp", address, 5*time.Second) }

	close := func(v interface{}) error { return v.(net.Conn).Close() }

	poolConfig := &pool.PoolConfig{
		InitialCap:  5,
		MaxCap:      30,
		Factory:     factory,
		Close:       close,
		IdleTimeout: 15 * time.Second,
	}
	p, err := pool.NewChannelPool(poolConfig)
	if err != nil {
		return fmt.Errorf("Create Pool Err=" + err.Error())
	}

	f := fs.NewRfs()
	f.RPCPool = p
	f.User.Username = username
	f.User.Password = password
	f.User.Email = email

	host := fuse.NewFileSystemHost(f)
	r := true
	go func(re *bool) {
		b := host.Mount("", []string{mountpoint})
		if re != nil {
			*re = b
		}
	}(&r)

	time.Sleep(100*time.Millisecond)
	if !r {
		p.Release()
		return errors.New("mount failed")
	}
	mounted[username] = &struct {
		Fs   *fs.Rfs
		Host *fuse.FileSystemHost
	}{f, host}
	return nil
}

func UnMount(username string) error {
	s, ok := mounted[username]
	if !ok {
		return errors.New("not found")
	}
	ok = s.Host.Unmount()
	if !ok {
		return errors.New("unMount failed")
	}
	// s.Fs.Destroy()
	s.Fs = nil
	delete(mounted, username)
	return nil
}
