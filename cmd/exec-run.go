package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type execRun struct {
	cmd  *exec.Cmd
	name string
}

func newExecRun(cmd string, args ...string) *execRun {
	return &execRun{
		name: cmd,
		cmd:  exec.Command(cmd, args...),
	}
}

func (r *execRun) EnvParent() *execRun {
	r.cmd.Env = os.Environ()[:]
	return r
}

func (r *execRun) Env(name, value string) *execRun {
	r.cmd.Env = append(r.cmd.Env, fmt.Sprintf("%s=%s", name, value))
	return r
}

func (r *execRun) Std() *execRun {
	r.cmd.Stdin = os.Stdin
	r.cmd.Stdout = os.Stdout
	r.cmd.Stderr = os.Stderr
	return r
}

func (r *execRun) StdOut(wr io.Writer) *execRun {
	r.cmd.Stdout = wr
	r.cmd.Stderr = wr
	return r
}

func (r *execRun) Run() error {
	fmt.Println(r.name, strings.Join(r.cmd.Args, " "))
	return r.cmd.Run()
}

func (r *execRun) Start() error {
	return r.cmd.Start()
}

func (r *execRun) Wait() error {
	return r.cmd.Wait()
}
