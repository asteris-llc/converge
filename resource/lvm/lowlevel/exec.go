package lowlevel

import (
	"os/exec"
	"strings"
	"syscall"
)

type Exec interface {
	Run(prog string, args []string) error
	RunExitCode(prog string, args []string) (int, error) // for mountpoint querying
	Read(prog string, args []string) (stdout string, err error)

	// unit read/write injection
	//    ReadFile(fn string) (string, error)
	//    WriteFile(fn string, c string) error
}

type OsExec struct {
}

func (*OsExec) Run(prog string, args []string) error {
	return exec.Command(prog, args...).Run()
}

func (e *OsExec) RunExitCode(prog string, args []string) (int, error) {
	err := e.Run(prog, args)
	return exitStatus(err)
}

func (*OsExec) Read(prog string, args []string) (stdout string, err error) {
	out, err := exec.Command(prog, args...).Output()
	if err != nil {
		return "", err
	}
	return strings.Trim(string(out), "\n "), err
}

func exitStatus(err error) (int, error) {
	if exiterr, ok := err.(*exec.ExitError); ok {
		// The program has exited with an exit code != 0
		// This works on both Unix and Windows. Although package
		// syscall is generally platform dependent, WaitStatus is
		// defined for both Unix and Windows and in both cases has
		// an ExitStatus() method with the same signature.
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus(), nil
		}
	}
	return 0, err
}
