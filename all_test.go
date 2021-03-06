package main_test

import (
	"io"
	"os"
	"os/exec"
	"strings"
)

type logcmd struct {
	f    *os.File
	cmds []string
}

func newlogger(fname string) (*logcmd, error) {
	f, err := os.Create(fname)
	if err != nil {
		return nil, err
	}
	return &logcmd{f: f, cmds: nil}, nil
}

func (cmd *logcmd) LastCmd() string {
	if len(cmd.cmds) <= 0 {
		return ""
	}
	return cmd.cmds[len(cmd.cmds)-1]
}
func (cmd *logcmd) Run(bin string, args ...string) error {
	cmd_line := ""
	{
		cargs := make([]string, 1, len(args)+1)
		cargs[0] = bin
		cargs = append(cargs, args...)
		cmd_line = strings.Join(cargs, " ")
	}
	cmd.cmds = append(cmd.cmds, cmd_line)
	c := exec.Command(bin, args...)
	c.Stdout = cmd.f
	c.Stderr = cmd.f

	_, err := cmd.f.WriteString("\n" + strings.Repeat("#", 80) + "\n")
	if err != nil {
		return err
	}

	_, err = cmd.f.WriteString("## " + cmd_line + "\n")
	if err != nil {
		return err
	}

	return c.Run()
}

func (cmd *logcmd) Close() error {
	return cmd.f.Close()
}

func (cmd *logcmd) Display() {
	cmd.f.Seek(0, 0)
	io.Copy(os.Stderr, cmd.f)
}

// EOF
