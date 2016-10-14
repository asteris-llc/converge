package lowlevel

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

type Exec interface {
	Run(prog string, args []string) error
	RunExitCode(prog string, args []string) (int, error) // for mountpoint querying
	Read(prog string, args []string) (stdout string, err error)
	ReadWithExitCode(prog string, args []string) (stdout string, rc int, err error) // for blkid querying

	// unit read/write injection
	ReadFile(fn string) ([]byte, error)
	WriteFile(fn string, c []byte, p os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	Exists(path string) (bool, error)
}

type OsExec struct {
}

func (*OsExec) Run(prog string, args []string) error {
	log.WithField("module", "lvm").Infof("Executing %s: %v", prog, args)
	e := exec.Command(prog, args...).Run()
	if e == nil {
		log.WithField("module", "lvm").Debugf("%s: no error", prog)
	} else {
		log.WithField("module", "lvm").Debugf("%s: terminated with %s", prog, e.Error())
	}
	return e
}

func (e *OsExec) RunExitCode(prog string, args []string) (int, error) {
	err := e.Run(prog, args)
	return exitStatus(err)
}

func (*OsExec) ReadWithExitCode(prog string, args []string) (stdout string, rc int, err error) {
	log.WithField("module", "lvm").Infof("Executing (read) %s: %v", prog, args)
	out, err := exec.Command(prog, args...).Output()
	if err != nil {
		if rc, err := exitStatus(err); err != nil {
			log.WithField("module", "lvm").Debugf("%s: terminated with %s", prog, err.Error())
			return "", 0, errors.Wrapf(err, "reading output of process %s: %s", prog, args)
		} else {
			return strings.Trim(string(out), "\n "), rc, nil
		}
	}
	return strings.Trim(string(out), "\n "), 0, err
}

func (e *OsExec) Read(prog string, args []string) (stdout string, err error) {
	out, rc, err := e.ReadWithExitCode(prog, args)
	if err != nil {
		return "", err
	}
	if rc != 0 {
		return "", fmt.Errorf("process %s: %s terminated with status code %d", prog, args, rc)
	}
	return out, nil
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

func (*OsExec) ReadFile(fn string) ([]byte, error) {
	log.WithField("module", "lvm").Debugf("Reading %s...", fn)
	return ioutil.ReadFile(fn)
}

func (*OsExec) WriteFile(fn string, content []byte, perm os.FileMode) error {
	log.WithField("module", "lvm").Debugf("Writing %s...", fn)
	return ioutil.WriteFile(fn, content, perm)
}

func (*OsExec) MkdirAll(path string, perm os.FileMode) error {
	log.WithField("module", "lvm").Debugf("Make path %s...", path)
	return os.MkdirAll(path, perm)
}

func (*OsExec) Exists(path string) (bool, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
