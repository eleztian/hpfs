package git

import (
	"os"
	"regexp"
	"strings"
	"io/ioutil"
	"path/filepath"

	"bs-2018/hpfs-client/errors"
	"bs-2018/hpfs-client/p2p/share"
)

const GIT_NAME = ".git"

type Git struct {
	Ctx string
}

func init() {
	// 配置git, 显示中文
	cmd := CmdGit{
		Cmd: "config",
		Args: []string{
			"--global",
			"core.quotepath",
			"false",
		},
	}
	_, err := cmd.Exec(nil)
	if err != nil {
		panic(err)
	}
}

func NewGit(path string) (*Git, error) {
	if path == "" {
		return nil, errors.New("new git failed path should not empty")
	}
	g := &Git{path}

	_, err := os.Stat(path)
	if err != nil {
		if err == os.ErrNotExist {
			if err := g.Init(); err != nil {
				return nil, err
			}
			if err := g.Add("."); err != nil {
				return nil, err
			}
			if err := g.Commit("hpfs sys","", "init"); err != nil {
				return nil, err
			}
		}
	}

	return g, nil
}

type SHA1 string

func (s SHA1) Short() string {
	return string(s)[:7]
}

type Refs struct {
	Name   string
	Commit SHA1
}

// Init init a git work space.
func (g *Git) Init() error {
	cmd := CmdGit{
		Cmd:  "init",
		Dir:  g.Ctx,
		Args: []string{},
	}
	_, err := cmd.Exec(nil)
	return err
}

// NewRefs create a new refs.
func (g *Git) NewRefs(refs *Refs) error {
	path := filepath.Join(g.Ctx, GIT_NAME, "refs", "heads", refs.Name)
	return writeFile(path, []byte(refs.Commit))
}

// Head return current refs and a error.
func (g *Git) Head() (*Refs, error) {
	path := filepath.Join(g.Ctx, GIT_NAME, "HEAD")

	b, err := readFile(path)
	if err != nil {
		return nil, err
	}
	name := strings.TrimRightFunc(string(b[16:]), func(r rune) bool {
		if r == '\r' || r == '\n' {
			return true
		}
		return false
	})

	return g.Refs(name)
}

// SetHead change current refs to your give refs.
func (g *Git) SetHead(refs *Refs) error {
	if refs.Commit == "" {
		return errors.New("please give a current refs")
	}
	path := filepath.Join(g.Ctx, GIT_NAME, "HEAD")

	return writeFile(path, []byte("ref: refs/heads/"+refs.Name+"\n"))
}

// Refs return the refs info by your give name.
func (g *Git) Refs(refsName string) (*Refs, error) {
	path := filepath.Join(g.Ctx, GIT_NAME, "refs", "heads", refsName)

	b, err := readFile(path)
	if err != nil {
		return nil, err
	}
	return &Refs{Name: refsName, Commit: SHA1(b)}, nil
}

// RefsList return the all refs of this git workstation.
func (g *Git) RefsList() ([]*Refs, error) {
	path := filepath.Join(g.Ctx, GIT_NAME, "refs", "heads")
	files, err := readDir(path)
	if err != nil {
		return nil, err
	}
	result := make([]*Refs, 0)
	for _, fp := range files {
		r, err := g.Refs(fp)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}

// Checkout change zhe current refs to your give refs name.
// if it not exit create it and turn to.
func (g *Git) Checkout(refsName string) error {
	path := filepath.Join(g.Ctx, GIT_NAME, "refs", "heads", refsName)
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) { // the refs is not exit
			refs, err := g.Head() // get current refs
			if err != nil {
				return err
			}
			refs.Name = refsName
			err = g.NewRefs(refs) // create new refs
			if err != nil {
				return err
			}
			return g.SetHead(refs) // change current refs
		} else {
			return err
		}
	}
	refs, err := g.Refs(refsName)
	if err != nil {
		return err
	}
	return g.SetHead(refs)
}

// LogInfo git log information
type LogInfo struct {
	CommitSha1 SHA1
	TreeSha1   SHA1
	Author     string
	LongTime   string
	CommitCtx  string
	FileStat   []FileStat
}

// FileStat git file changed log information
type FileStat struct {
	Mode     string
	FileName string
}

// Log return the git's log information.
func (g *Git) Log() ([]*LogInfo, error) {
	cmd := CmdGit{
		Cmd: "log",
		Dir: g.Ctx,
		Args: []string{
			"--pretty=format:" + "||||-%hq-q%tq-q%anq-q%sq-q%cr-",
			"--name-status",

		},
	}
	oBuf, err := cmd.Exec(nil)
	if err != nil {
		return nil, err
	}
	ls := oBuf.String()
	logs := strings.Split(ls, "||||")
	result := make([]*LogInfo, 0)
	for _, v := range logs[1:] {
		v = strings.TrimSpace(v)
		reg := regexp.MustCompile("-([0-9a-fA-F]{7,})q-q([0-9a-fA-F]{7,})q-q(.*?)q-q(.*?)q-q([\\d].*?)-([\\s\\S]*)")
		l := reg.FindStringSubmatch(v)
		if len(l) == 7 {
			resultOne := LogInfo{
				CommitSha1: SHA1(l[1]),
				TreeSha1:   SHA1(l[2]),
				Author:     l[3],
				CommitCtx:  l[4],
				LongTime:   l[5],
			}
			if len(l[6]) > 0 && l[6][0] == '\n' {
				l[6] = l[6][1:]
			}
			stats := strings.Split(l[6], "\n")
			for index, stat := range stats {
				reg = regexp.MustCompile("([MAD])	([\\s\\S]*)")
				l = reg.FindStringSubmatch(stat)
				if len(l) == 3 {
					op := ""
					switch l[1] {
					case "M":
						op = "修改"
					case "A":
						op = "添加"
					case "D":
						op = "删除"
					}
					_, filename := filepath.Split(l[2])
					t := FileStat{Mode: op, FileName: filename}
					resultOne.FileStat = append(resultOne.FileStat, t)
					if index >= 5 {
						t := FileStat{Mode: "...", FileName: "..."}
						resultOne.FileStat = append(resultOne.FileStat, t)
						break
					}
				}
			}
			result = append(result, &resultOne)
		}
	}
	return result, nil
}

// Add update index to current refs.
func (g *Git) Add(fileNames ...string) error {
	cmd := CmdGit{
		Cmd:  "add",
		Dir:  g.Ctx,
		Args: fileNames,
	}
	_, err := cmd.Exec(nil)
	if err != nil {
		return err
	}
	return nil
}

// Commit commit cache info to current refs.
func (g *Git) Commit(username, email, note string) error {
	if email == "" {
		email = "unkown@email.com"
	}
	cmd := CmdGit{
		Cmd: "commit",
		Dir: g.Ctx,
		Args: []string{
			"-m", note,
			"--author=\"" + username + " <" + email + ">" + "\"",
		},
	}
	_, err := cmd.Exec(nil)
	if err != nil {
		return err
	}
	return nil
}

// Commit commit cache info to current refs.
func (g *Git) Merge(refs string, note string) error {
	cmd := CmdGit{
		Cmd: "merge",
		Dir: g.Ctx,
		Args: []string{
			refs,
			"-m", note,
		},
	}
	_, err := cmd.Exec(nil)
	if err != nil {
		return err
	}
	return nil
}

func getRelPath(p string) string {
	if p == "" {
		return ""
	}
	if p != "" && p[0] == '\\' {
		return p[1:]
	}
	return p
}

func (g *Git) MergeFile(p1, p2, p3 string) error {
	p1 = share.GetAbs(filepath.Clean(p1))
	p2 = share.GetAbs(filepath.Clean(p2))
	p3 = share.GetAbs(filepath.Clean(p3))
	cmd := CmdGit{
		Cmd: MergeFile,
		Dir: g.Ctx,
		Args: []string{
			p1, p2, p3,
			"--theirs",
		},
	}
	_, err := cmd.Exec(nil)
	if err != nil {
		return err
	}
	return nil
}

func (g *Git) Reset(sha1 SHA1) error {
	cmd := CmdGit{
		Cmd: "reset",
		Dir: g.Ctx,
		Args: []string{
			"--hard",
			string(sha1),
		},
	}
	_, err := cmd.Exec(nil)
	if err != nil {
		return err
	}
	return nil
}

func readFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ioutil.ReadAll(f)
}

func readDir(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	fis, err := f.Readdir(-1)
	if err != nil {
		return nil, err
	}

	result := make([]string, 0)
	for _, v := range fis {
		result = append(result, v.Name())
	}
	return result, nil
}

func writeFile(path string, ctx []byte) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	_, err = f.Write(ctx)
	return err
}
