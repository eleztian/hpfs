package git

import (
	"bytes"
	"io"
	"os/exec"
)

func findGitCmd() (string, error) {
	return exec.LookPath("git")
}

type GitCmdT string

const (
	CatFile     GitCmdT = "cat-file"
	HashObject  GitCmdT = "hash-object"
	LsTree      GitCmdT = "ls-tree"
	CommitTree  GitCmdT = "commit-tree"
	UpdateIndex GitCmdT = "update-index"
	WriteTree   GitCmdT = "write-tree"
	ReadTree    GitCmdT = "read-tree"
	MergeFile   GitCmdT = "merge-file"
)

type CmdGit struct {
	Cmd  GitCmdT
	Dir  string
	Args []string
}

func (c *CmdGit) Exec(in io.Reader) (out *bytes.Buffer, err error) {
	out = bytes.NewBuffer(nil)
	cmd := exec.Command("git", string(c.Cmd))
	cmd.Args = append(cmd.Args, c.Args...)
	cmd.Dir = c.Dir
	cmd.Stdin = in
	cmd.Stdout = out
	err = cmd.Run()
	return
}
