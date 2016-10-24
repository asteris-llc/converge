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

// Exec is interface to `real system` also used for test injections
// FIXME: should be reviewed, and possible refactored as base for all
//        processes/filesystem operrations, to make the easy mockable
// FIXME: split to ExecInterface, FilesystemInterface (possible also SystemInterface for Getuid)
type Exec interface {
	Run(prog string, args []string) error
	RunWithExitCode(prog string, args []string) (int, error) // for mountpoint querying
	Read(prog string, args []string) (stdout string, err error)
	ReadWithExitCode(prog string, args []string) (stdout string, rc int, err error) // for blkid querying
	Lookup(prog string) error
	Getuid() int

	// unit read/write injection
	ReadFile(fn string) ([]byte, error)
	WriteFile(fn string, c []byte, p os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	Exists(path string) (bool, error)
}

type osExec struct {
}

// MakeOsExec create Exec backend
func MakeOsExec() Exec {
	return &osExec{}
}

func (*osExec) Run(prog string, args []string) error {
	log.WithField("module", "lvm").Infof("Executing %s: %v", prog, args)
	e := exec.Command(prog, args...).Run()
	if e == nil {
		log.WithField("module", "lvm").Debugf("%s: no error", prog)
	} else {
		log.WithField("module", "lvm").Debugf("%s: terminated with %s", prog, e.Error())
	}
	return e
}

func (e *osExec) RunWithExitCode(prog string, args []string) (int, error) {
	err := e.Run(prog, args)
	return exitStatus(err)
}

func (*osExec) ReadWithExitCode(prog string, args []string) (stdout string, rc int, err error) {
	rc = 0
	log.WithField("module", "lvm").Infof("Executing (read) %s: %v", prog, args)
	out, err := exec.Command(prog, args...).Output()
	strOut := strings.Trim(string(out), "\n ")
	if err != nil {
		rc, err = exitStatus(err)
		if err != nil {
			log.WithField("module", "lvm").Debugf("%s: terminated with %s", prog, err.Error())
			return "", 0, errors.Wrapf(err, "reading output of process %s: %s", prog, args)
		}
	}
	return strOut, rc, err
}

func (e *osExec) Read(prog string, args []string) (stdout string, err error) {
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

func (*osExec) Lookup(prog string) error {
	_, err := exec.LookPath(prog)
	return err
}

func (*osExec) ReadFile(fn string) ([]byte, error) {
	log.WithField("module", "lvm").Debugf("Reading %s...", fn)
	return ioutil.ReadFile(fn)
}

func (*osExec) WriteFile(fn string, content []byte, perm os.FileMode) error {
	log.WithField("module", "lvm").Debugf("Writing %s...", fn)
	return ioutil.WriteFile(fn, content, perm)
}

func (*osExec) MkdirAll(path string, perm os.FileMode) error {
	log.WithField("module", "lvm").Debugf("Make path %s...", path)
	return os.MkdirAll(path, perm)
}

func (*osExec) Exists(path string) (bool, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (*osExec) Getuid() int {
	return os.Getuid()
}
