package fs

import (
	"bs-2018/hpfs-client/p2p/client/fuse"
	"bs-2018/hpfs-client/p2p/rpc"
	"bs-2018/hpfs-client/p2p/server"
	"bytes"
	"crypto/md5"
	"encoding/gob"
	"fmt"
	"github.com/billziss-gh/cgofuse/examples/shared"
	"github.com/silenceper/pool"
	"gopkg.in/src-d/go-git.v4/utils/binary"
	"log"
	"net"
	"sync"
	"bs-2018/hpfs-client/p2p/git"
	"bs-2018/hpfs-client/p2p/share"
)

func trace(vals ...interface{}) func(vals ...interface{}) {
	uid, gid, _ := fuse.Getcontext()
	return shared.Trace(1, fmt.Sprintf("[uid=%v,gid=%v]", uid, gid), vals...)
}

func init() {
	gob.Register(server.ReadArgs{})
	gob.Register(server.WriteArgs{})
	gob.Register(server.TruncateArgs{})
	gob.Register(server.RenameArgs{})
}

type node_t struct {
	stat fuse.Stat_t
}
type Rfs struct {
	fuse.FileSystemBase
	User struct {
		Username string
		Email    string
		Password string
	}
	RPCPool pool.Pool

	openmap map[uint64]*node_t
	ino     uint64
	root    *node_t
	lock    sync.Mutex
}

func NewRfs() *Rfs {
	rfs := Rfs{}
	defer rfs.synchronize()()
	rfs.ino++
	rfs.root = newNode(0, rfs.ino, fuse.S_IFDIR|00777, 0, 0)
	rfs.openmap = map[uint64]*node_t{}
	return &rfs
}

func (rfs *Rfs) synchronize() func() {
	rfs.lock.Lock()
	return func() {
		rfs.lock.Unlock()
	}
}

func newNode(dev uint64, ino uint64, mode uint32, uid uint32, gid uint32) *node_t {
	tmsp := fuse.Now()
	rfs := node_t{
		stat: fuse.Stat_t{
			Dev:      dev,
			Ino:      ino,
			Mode:     mode,
			Nlink:    1,
			Uid:      uid,
			Gid:      gid,
			Atim:     tmsp,
			Mtim:     tmsp,
			Ctim:     tmsp,
			Birthtim: tmsp,
			Flags:    0,
		},
	}
	return &rfs
}

func (rfs *Rfs) Log(funcname, path string, args ...interface{}) {
	log.Printf("[%s] %s : %v\n", funcname, path, args)
}

func (rfs *Rfs) lookupNode(path string) (node *node_t) {
	result := server.FileInfo{}
	if err := rfs.RPC("File.GetAttr",
		path, &result); err != nil {
		rfs.Log("Getaddr", path, err)
		return nil
	}

	uid, gid, _ := fuse.Getcontext()
	node = newNode(0, rfs.ino, getMode(result.IsDir), uid, gid)
	node.stat.Ino = INo(path)
	node.stat.Size = result.Size
	node.stat.Atim = fuse.Now()
	node.stat.Mtim = fuse.NewTimespec(result.MTime)
	return
}

func (rfs *Rfs) openNode(path string, dir bool) (int, uint64) {
	node := rfs.lookupNode(path)
	if node == nil {
		return -fuse.ENOENT, ^uint64(0)
	}
	rfs.openmap[node.stat.Ino] = node

	return 0, node.stat.Ino
}

func (rfs *Rfs) getNode(path string, fh uint64) *node_t {
	if ^uint64(0) == fh {
		node := rfs.lookupNode(path)
		return node
	} else {
		n, ok := rfs.openmap[fh]
		if !ok {
			n = rfs.lookupNode(path)
			if n != nil {
				rfs.openmap[INo(path)] = n
			}
		}
		return n
	}
}

func (rfs *Rfs) closeNode(fh uint64) (errc int) {
	node, ok := rfs.openmap[fh]
	if ok {
		delete(rfs.openmap, node.stat.Ino)
	}
	return 0
}

func (rfs *Rfs) makeNode(path string, mode uint32) int {
	if fuse.S_IFDIR == mode&fuse.S_IFDIR {
		if err := rfs.RPC("File.MkDir",
			path, nil); err != nil {
			rfs.Log("MkDir", path, err)
			return -fuse.ENOENT
		}
	} else {
		if err := rfs.RPC("File.MkNod",
			path, nil); err != nil {
			rfs.Log("Mknod", path, err)
			return -fuse.ENOENT
		}
	}
	return 0
}

func (rfs *Rfs) Destroy() {
	defer trace()()
	rfs.RPCPool.Release()
	rfs.RPCPool = nil
	rfs.openmap = nil
	rfs.root = nil
}

func (rfs *Rfs) Open(path string, flags int) (errc int, fh uint64) {
	defer trace(path, flags)(&errc, &fh)
	defer rfs.synchronize()()
	return rfs.openNode(path, false)
}

func (rfs *Rfs) Opendir(path string) (errc int, fh uint64) {
	defer trace(path)(&errc, &fh)
	defer rfs.synchronize()()
	return rfs.openNode(path, true)
}

func (rfs *Rfs) Release(path string, fh uint64) (errc int) {
	defer trace(path, fh)(&errc)
	defer rfs.synchronize()()
	return rfs.closeNode(fh)
}

func (rfs *Rfs) Releasedir(path string, fh uint64) (errc int) {
	defer trace(path, fh)(&errc)
	defer rfs.synchronize()()
	return rfs.closeNode(fh)
}

func (rfs *Rfs) Getattr(path string, stat *fuse.Stat_t, fh uint64) (errc int) {
	defer trace(path, fh)(&errc, stat)
	defer rfs.synchronize()()
	node := rfs.getNode(path, fh)
	if nil == node {
		return -fuse.ENOENT
	}
	*stat = node.stat
	return 0
}

func (rfs *Rfs) Readdir(path string,
	fill func(name string, stat *fuse.Stat_t, ofst int64) bool,
	ofst int64,
	fh uint64) (errc int) {

	defer trace(path, fill, ofst, fh)(&errc)
	defer rfs.synchronize()()
	node, ok := rfs.openmap[fh]
	if ok {
		fill(".", &node.stat, 0)
	} else {
		fill(".", nil, 0)
	}
	fill("..", nil, 0)

	result := []server.FileInfo{}

	if err := rfs.RPC("File.ReadDir",
		path, &result); err != nil {
		rfs.Log("ReadDir", path, err)
		return -fuse.ENOENT
	}
	for _, v := range result {
		if !fill(v.Name, nil, 0) {
			break
		}
	}
	return 0
}

func (rfs *Rfs) Mknod(path string, mode uint32, dev uint64) (errc int) {
	defer trace(path, mode, dev)(&errc)
	defer rfs.synchronize()()
	return rfs.makeNode(path, mode)
}

func (rfs *Rfs) Mkdir(path string, mode uint32) (errc int) {
	defer trace(path, mode)(&errc)
	defer rfs.synchronize()()
	return rfs.makeNode(path, fuse.S_IFDIR|(mode&07777))
}

func (rfs *Rfs) Read(path string, buff []byte, ofst int64, fh uint64) (n int) {
	defer trace(path, buff, ofst, fh)(&n)
	defer rfs.synchronize()()
	result := make([]byte, len(buff))
	if err := rfs.RPC("File.Read",
		server.ReadArgs{Path: path, Offset: ofst, N: len(buff)},
		&result); err != nil {
		rfs.Log("Read", path, err)
		return -fuse.EACCES
	}
	return copy(buff, result)
}
func (rfs *Rfs) Write(path string, buff []byte, ofst int64, fh uint64) (n int) {
	defer trace(path, buff, ofst, fh)(&n)
	defer rfs.synchronize()()
	var result int
	if err := rfs.RPC("File.Write",
		server.WriteArgs{
			Path:   path,
			Offset: ofst,
			Ctx:    buff}, &result); err != nil {
		rfs.Log("Write", path, err)
		return -fuse.ENOENT
	}

	return result
}
func (rfs *Rfs) Truncate(path string, size int64, fh uint64) (errc int) {
	defer trace(path, size, fh)(&errc)
	defer rfs.synchronize()()
	// path = rfs.getPath(path)
	if err := rfs.RPC("File.Truncate",
		server.TruncateArgs{
			Path: path,
			Size: size}, nil); err != nil {
		rfs.Log("Truncate", path, err)
		return -fuse.ENOENT
	}
	return 0
}

func (rfs *Rfs) Unlink(path string) (errc int) {
	defer trace(path)(&errc)
	defer rfs.synchronize()()
	if err := rfs.RPC("File.Rm",
		path, nil); err != nil {
		rfs.Log("Rm", path, err)
		return -fuse.ENOENT
	}
	return 0
}

func (rfs *Rfs) Rename(oldPath string, newPath string) (errc int) {
	defer trace(oldPath, newPath)(&errc)
	defer rfs.synchronize()()
	if "" == newPath {
		return -fuse.EINVAL
	}
	if err := rfs.RPC("File.Rename",
		server.RenameArgs{
			OldPath: oldPath,
			NewPath: newPath}, nil); err != nil {
		rfs.Log("Rename", oldPath+"->"+newPath, err)
		return -fuse.ENOENT
	}
	return 0
}

func (rfs *Rfs) Rmdir(path string) (errc int) {
	defer trace(path)(&errc)
	defer rfs.synchronize()()
	if err := rfs.RPC("File.Rm",
		path, nil); err != nil {
		rfs.Log("Rm", path, err)
		return -fuse.ENOENT
	}
	return 0
}

func (rfs *Rfs) Utimens(path string, tmsp []fuse.Timespec) (errc int) {
	// defer trace(path, tmsp)(&errc)
	// defer rfs.synchronize()()
	// node := rfs.lookupNode(path)
	// if nil == node {
	// 	return -fuse.ENOENT
	// }
	// node.stat.Ctim = fuse.Now()
	// if nil == tmsp {
	// 	tmsp0 := node.stat.Ctim
	// 	tmsa := [2]fuse.Timespec{tmsp0, tmsp0}
	// 	tmsp = tmsa[:]
	// }
	// node.stat.Atim = tmsp[0]
	// node.stat.Mtim = tmsp[1]
	return 0
}

func (rfs *Rfs) Chmod(path string, mode uint32) (errc int) {
	defer trace(path, mode)(&errc)
	defer rfs.synchronize()()
	node := rfs.lookupNode(path)
	if node == nil {
		return -fuse.ENOENT
	}
	node.stat.Mode = (node.stat.Mode & fuse.S_IFMT) | mode&07777
	node.stat.Ctim = fuse.Now()
	return 0
}

func (rfs *Rfs) Chown(path string, uid uint32, gid uint32) (errc int) {
	defer trace(path, uid, gid)(&errc)
	defer rfs.synchronize()()
	node := rfs.lookupNode(path)
	if node == nil {
		return -fuse.ENOENT
	}
	if ^uint32(0) != uid {
		node.stat.Uid = uid
	}
	if ^uint32(0) != gid {
		node.stat.Gid = gid
	}
	node.stat.Ctim = fuse.Now()
	return 0
}

func (rfs *Rfs) RPC(funcName string, args, replay interface{}) error {
	conn, err := rfs.RPCPool.Get()
	if err != nil {
		return err
	}
	err = rpc.Call(conn.(net.Conn), funcName,
		rpc.Args{
			User: rfs.User,
			Args: args,
		}, replay)
	if err != nil {
		return err
	}
	return err
}

func INo(path string) uint64 {
	m := md5.New()
	m.Write([]byte(path))
	b := bytes.NewReader(m.Sum(nil))
	r, _ := binary.ReadUint64(b)
	return r
}

func Reset(name, sha1 string) error {
	g, err := git.NewGit(share.GetDir(name))
	if err != nil {
		return err
	}
	return g.Reset(git.SHA1(sha1))
}