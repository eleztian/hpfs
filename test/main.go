package main

import (
	"os"
	"io"
	"time"
	"io/ioutil"
)

func main() {
}


func WriteFile(path string, n int64) (t time.Duration, err error) {
	start := time.Now()
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return
	}
	defer f.Close()
	tmp := []byte("hello")
	all := 0
	for i:=int64(0); i < n; i++{
		n, err := f.Write(tmp)
		if err != nil {
			return 0, err
		}
		all += n
	}
	return time.Since(start), nil
}

func ReadDir(path string) (t time.Duration, err error) {
	start := time.Now()
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	_, err = f.Readdir(-1)
	if err != nil {
		return
	}
	return time.Since(start), nil
}

func CopyFile(dst, src string) (t time.Duration, err error) {
	f2, err := os.OpenFile(src, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer f2.Close()
	start := time.Now()
	f, err := os.Open(dst)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	_, err = io.Copy(f2, f)
	if err != nil {
		return 0, err
	}
	t = time.Since(start)

	return
}

func ReadFile(path string) (t time.Duration, err error) {
	start := time.Now()
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	_, err = ioutil.ReadAll(f)
	if err != nil {
		return
	}
	t = time.Since(start)
	return
}