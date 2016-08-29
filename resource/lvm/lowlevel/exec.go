package lowlevel

import (
	"os/exec"
)

type Exec interface {
	Run(prog string, args []string) error
	Read(prog string, args []string) (stdout string, err error)
}

type OsExec struct {
}

func (*OsExec) Run(prog string, args []string) error {
	return exec.Command(prog, args...).Wait()
}

func (*OsExec) Read(prog string, args []string) (stdout string, err error) {
	out, err := exec.Command(prog, args...).Output()
	if err != nil {
		return "", err
	}
	return string(out), err
}
