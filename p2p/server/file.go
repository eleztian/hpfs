package server

import (
	"bs-2018/hpfs-client/errors"
	"bs-2018/hpfs-client/p2p/rpc"
	"bs-2018/hpfs-client/p2p/share"
	"context"
	"crypto/md5"
	"fmt"
	"github.com/theckman/go-flock"
	"io"
	"os"
	"path/filepath"
	"time"
	"bs-2018/hpfs-client/p2p/git"
)

func synchronize(path string) func() {
	m := md5.New()
	m.Write([]byte(path))
	path = filepath.Join(os.TempDir(), fmt.Sprintf("%x.lock", m.Sum(nil)))
	fileLock := flock.NewFlock(path)
	lockCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	locked, err := fileLock.TryLockContext(lockCtx, 678*time.Millisecond)
	if err != nil {
		panic(err)
	}
	if locked {
		return func() {
			if err != fileLock.Unlock() {
				os.Remove(path)
			}
		}
	}
	return func() {}
}

func commit(workPath string, user rpc.User, note string) error {
	g, err := git.NewGit(share.GetDir(workPath))
	if err != nil {
		panic(err)
		return err
	}
	err = g.Add(".")
	if err != nil {
		return err
	}
	err = g.Commit(user.Username, user.Email, note)
	if err != nil {
		return err
	}
	return nil
}

type File struct{}

type FileInfo struct {
	Name  string
	Size  int64
	Mode  os.FileMode
	MTime time.Time
	IsDir bool
}

// Ls return all the dir's files information.
// dirName should is a relative path of shared path.
// path should start with '/'
func (*File) ReadDir(args rpc.Args, result *[]FileInfo) error {
	path, ok := args.Args.(string)
	if !ok {
		return errors.New("args should is a string")
	}
	r := make([]FileInfo, 0)
	if path == "/" || path == "\\" {
		for k,v := range share.GetDirs() {
			f, err := os.Stat(filepath.Join(v, k))
			if err != nil {
				if err == os.ErrNotExist {
					continue
				}
				return err
			}
			t := FileInfo{
				Name:  f.Name(),
				IsDir: f.IsDir(),
				Size:f.Size(),
				MTime:f.ModTime(),
			}
			r = append(r, t)
		}
		*result = r
		return nil
	}

	f, err := os.Open(share.GetAbs(path))
	if err != nil {
		return err
	}
	defer f.Close()
	fileInfo, err := f.Readdir(-1)
	if err != nil {
		return err
	}

	for _, st := range fileInfo {
		if st.Name() == ".git" {
			continue
		}
		t := FileInfo{
			Name:  st.Name(),
			Size:  st.Size(),
			Mode:  st.Mode(),
			MTime: st.ModTime(),
			IsDir: st.IsDir(),
		}
		r = append(r, t)
	}
	*result = r
	return nil
}

func (*File) GetAttr(args rpc.Args, replay *FileInfo) error {
	path, ok := args.Args.(string)
	if !ok {
		return errors.New("args should is a string")
	}
	if path == "/" || path == "\\" {
		t := FileInfo{
			Name:  "/",
			IsDir: true,
		}
		*replay = t
		return nil
	}
	st, err := os.Stat(share.GetAbs(path))
	if err != nil {
		return err
	}
	t := FileInfo{
		Name:  st.Name(),
		Size:  st.Size(),
		Mode:  st.Mode(),
		MTime: st.ModTime(),
		IsDir: st.IsDir(),
	}
	*replay = t
	return nil
}

type ReadArgs struct {
	Path   string
	N      int
	Offset int64
}

func (*File) Read(args rpc.Args, replay *[]byte) error {
	readArgs, ok := args.Args.(ReadArgs)
	if !ok {
		return errors.New("args should is a ReadArgs")
	}

	f, err := os.Open(share.GetAbs(readArgs.Path))
	if err != nil {
		return err
	}
	defer f.Close()

	c := make([]byte, readArgs.N)
	n, err := f.ReadAt(c, readArgs.Offset)
	if err != nil {
		if err != io.EOF {
			return err
		}
		c = c[:n]
	}
	*replay = c
	return nil
}

type WriteArgs struct {
	Path   string
	Ctx    []byte
	Offset int64
}

func (*File) Write(args rpc.Args, replay *int) (err error) {
	writeArgs, ok := args.Args.(WriteArgs)
	if !ok {
		return errors.New("args should is a ReadArgs")
	}
	path := share.GetAbs(writeArgs.Path)
	defer synchronize(filepath.Dir(path))()
	defer func() {
		if err != nil {
			return
		}
		err = commit(writeArgs.Path, args.User, "Write")
	}()

	// 创建临时文件
	f, err := os.OpenFile(path+"."+args.Username, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	// 删除临时文件
	defer os.Remove(path + "." + args.Username)
	defer f.Close()

	// 修改文件内容
	err = f.Truncate(writeArgs.Offset + int64(len(writeArgs.Ctx)))
	if err != nil {
		return err
	}
	n, err := f.WriteAt(writeArgs.Ctx, writeArgs.Offset)
	if err != nil {
		if err != io.EOF {
			return err
		}
	}

	// 文件合并
	g, err := git.NewGit(share.GetDir(writeArgs.Path))
	if err != nil {
		return err
	}
	if err = g.MergeFile(writeArgs.Path, writeArgs.Path, writeArgs.Path+"."+args.Username); err != nil {
		return err
	}

	// 返回写入字节数
	*replay = n
	return nil
}

type TruncateArgs struct {
	Path string
	Size int64
}

func (*File) Truncate(args rpc.Args, replay *int) (err error) {
	truncateArgs, ok := args.Args.(TruncateArgs)
	if !ok {
		return errors.New("args should is a Truncate")
	}

	path := share.GetAbs(truncateArgs.Path)
	defer synchronize(filepath.Dir(path))()
	defer func() {
		if err != nil {
			return
		}
		commit(truncateArgs.Path, args.User, "Truncate")
	}()

	f, err := os.OpenFile(path, os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	return f.Truncate(truncateArgs.Size)
}

func (*File) MkDir(args rpc.Args, replay *int) (err error) {
	mmkdirArg, ok := args.Args.(string)
	if !ok {
		return errors.New("args should is a MkDir")
	}
	path := share.GetAbs(mmkdirArg)
	defer synchronize(filepath.Dir(path))()
	defer func() {
		if err != nil {
			return
		}
		commit(mmkdirArg, args.User, "MkDir")
	}()

	return os.MkdirAll(path, os.ModePerm)
}

type RenameArgs struct {
	OldPath, NewPath string
}

func (*File) Rename(args rpc.Args, replay *int) (err error) {
	renameArgs, ok := args.Args.(RenameArgs)
	if !ok {
		return errors.New("args should is a RenameArgs")
	}
	path := share.GetAbs(renameArgs.OldPath)
	defer synchronize(filepath.Dir(path))()
	defer func() {
		if err != nil {
			return
		}
		commit(renameArgs.OldPath, args.User, "Rename")
	}()

	return os.Rename(path, share.GetAbs(renameArgs.NewPath))
}

func (*File) Rm(args rpc.Args, replay *int) (err error) {
	rmArg, ok := args.Args.(string)
	if !ok {
		return errors.New("args should is a string")
	}
	path := share.GetAbs(rmArg)
	defer synchronize(filepath.Dir(path))()
	defer func() {
		if err != nil {
			return
		}
		commit(rmArg, args.User, "Rm")
	}()

	return os.Remove(path)
}

func (*File) MkNod(args rpc.Args, replay *int) (err error) {
	mknodArg, ok := args.Args.(string)
	if !ok {
		return errors.New("args should is a string")
	}
	path := share.GetAbs(mknodArg)
	defer synchronize(filepath.Dir(path))()
	defer func() {
		if err != nil {
			return
		}
		commit(mknodArg, args.User, "MkNod")
	}()

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return nil
}
