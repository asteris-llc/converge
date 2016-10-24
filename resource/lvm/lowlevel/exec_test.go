package lowlevel_test

import (
	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/asteris-llc/converge/resource/lvm/lowlevel"
	"github.com/stretchr/testify/assert"

	"testing"
)

// TestExecRun is test for osExec.Run
func TestExecRun(t *testing.T) {
	t.Parallel()

	t.Run("normal", func(t *testing.T) {
		defer logging.HideLogs(t)()
		e := lowlevel.MakeOsExec()
		err := e.Run("/bin/sh", []string{"-c", "true"})
		assert.NoError(t, err)
	})

	t.Run("fail", func(t *testing.T) {
		defer logging.HideLogs(t)()
		e := lowlevel.MakeOsExec()
		err := e.Run("/bin/sh", []string{"-c", "false"})
		assert.Error(t, err)
	})
}

// TestExecRead is test on osExec.Read
func TestExecRead(t *testing.T) {
	t.Parallel()

	t.Run("normal", func(t *testing.T) {
		defer logging.HideLogs(t)()
		e := lowlevel.MakeOsExec()
		out, err := e.Read("/bin/sh", []string{"-c", "echo foo"})
		assert.NoError(t, err)
		assert.Equal(t, "foo", out)
	})

	t.Run("failure", func(t *testing.T) {
		defer logging.HideLogs(t)()
		e := lowlevel.MakeOsExec()
		_, err := e.Read("/bin/sh", []string{"-c", "echo foo && false"})
		assert.Error(t, err)
		// FIXME: underlying exec.Command looks not return output on error
		//        would be nice to have all output in logs in case of error
		// assert.Equal(t, "foo", out)
	})

	t.Run("multiline", func(t *testing.T) {
		defer logging.HideLogs(t)()
		e := lowlevel.MakeOsExec()
		out, err := e.Read("/bin/sh", []string{"-c", "echo foo && echo bar"})
		assert.NoError(t, err)
		assert.Equal(t, "foo\nbar", out)
	})
}

// TestExecRunWithExitCode tests Exec.RunWithExitCode
func TestExecRunWithExitCode(t *testing.T) {
	t.Run("zero code", func(t *testing.T) {
		defer logging.HideLogs(t)()
		e := lowlevel.MakeOsExec()
		rc, err := e.RunWithExitCode("/bin/sh", []string{"-c", "true"})
		assert.NoError(t, err)
		assert.Equal(t, 0, rc)
	})

	t.Run("non-zero code", func(t *testing.T) {
		defer logging.HideLogs(t)()
		e := lowlevel.MakeOsExec()
		rc, err := e.RunWithExitCode("/bin/sh", []string{"-c", "exit 42"})
		assert.NoError(t, err)
		assert.Equal(t, 42, rc)
	})

	t.Run("failure", func(t *testing.T) {
		defer logging.HideLogs(t)()
		e := lowlevel.MakeOsExec()
		_, err := e.RunWithExitCode("/tricky/command/which/never-exists", []string{"-c", "true"})
		assert.Error(t, err)
	})
}

// TestExecReadWithExitCode tests Exec.ReadWithExitCode
func TestExecReadWithExitCode(t *testing.T) {
	t.Run("zero code", func(t *testing.T) {
		defer logging.HideLogs(t)()
		e := lowlevel.MakeOsExec()
		out, rc, err := e.ReadWithExitCode("/bin/sh", []string{"-c", "echo foo"})
		assert.NoError(t, err)
		assert.Equal(t, 0, rc)
		assert.Equal(t, "foo", out)
	})

	t.Run("non-zero code", func(t *testing.T) {
		defer logging.HideLogs(t)()
		e := lowlevel.MakeOsExec()
		out, rc, err := e.ReadWithExitCode("/bin/sh", []string{"-c", "echo foo && exit 42"})
		assert.NoError(t, err)
		assert.Equal(t, 42, rc)
		assert.Equal(t, "foo", out)
	})

	t.Run("failure", func(t *testing.T) {
		defer logging.HideLogs(t)()
		e := lowlevel.MakeOsExec()
		_, _, err := e.ReadWithExitCode("/tricky/command/which/never-exists", []string{"-c", "echo foo"})
		assert.Error(t, err)
	})
}

// TestExecLookup tests Exec.Lookup()
func TestExecLookup(t *testing.T) {
	t.Run("command `sh` which always exists", func(t *testing.T) {
		e := lowlevel.MakeOsExec()
		ok := e.Lookup("sh")
		assert.NoError(t, ok)
	})
	t.Run("some command which never exists", func(t *testing.T) {
		e := lowlevel.MakeOsExec()
		ok := e.Lookup("some-command-which-never-exists-in-normal-system")
		assert.Error(t, ok)
	})
}

// TestExecExists tests Exec.Exists()
func TestExecExists(t *testing.T) {
	t.Run("file /bin/sh exists", func(t *testing.T) {
		e := lowlevel.MakeOsExec()
		ok, err := e.Exists("/bin/sh")
		assert.NoError(t, err)
		assert.True(t, ok)
	})
	t.Run("some file which does not exists", func(t *testing.T) {
		e := lowlevel.MakeOsExec()
		ok, err := e.Exists("/tricky/file/which/never-exists")
		assert.NoError(t, err)
		assert.False(t, ok)
	})
}
