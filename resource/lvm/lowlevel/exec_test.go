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
		err := e.Run("sh", []string{"-c", "false"})
		assert.Error(t, err)
	})
}

// TestExecRead is test on osExec.Read
func TestExecRead(t *testing.T) {
	t.Parallel()

	t.Run("normal", func(t *testing.T) {
		defer logging.HideLogs(t)()
		e := lowlevel.MakeOsExec()
		out, err := e.Read("sh", []string{"-c", "echo foo"})
		assert.NoError(t, err)
		assert.Equal(t, "foo", out)
	})

	t.Run("failure", func(t *testing.T) {
		defer logging.HideLogs(t)()
		e := lowlevel.MakeOsExec()
		_, err := e.Read("sh", []string{"-c", "echo foo && false"})
		assert.Error(t, err)
		// FIXME: underlying exec.Command looks not return output on error
		//        would be nice to have all output in logs in case of error
		// assert.Equal(t, "foo", out)
	})

	t.Run("multiline", func(t *testing.T) {
		defer logging.HideLogs(t)()
		e := lowlevel.MakeOsExec()
		out, err := e.Read("sh", []string{"-c", "echo foo && echo bar"})
		assert.NoError(t, err)
		assert.Equal(t, "foo\nbar", out)
	})
}
