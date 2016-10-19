package lowlevel_test

import (
	"github.com/asteris-llc/converge/resource/lvm/lowlevel"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExecRun(t *testing.T) {
	e := lowlevel.MakeOsExec()
	err := e.Run("/bin/sh", []string{"-c", "true"})
	assert.NoError(t, err)
}

func TestExecRunFail(t *testing.T) {
	e := lowlevel.MakeOsExec()
	err := e.Run("sh", []string{"-c", "false"})
	assert.Error(t, err)
}

func TestRead(t *testing.T) {
	e := lowlevel.MakeOsExec()
	out, err := e.Read("sh", []string{"-c", "echo foo"})
	assert.NoError(t, err)
	assert.Equal(t, "foo", out)
}

func TestQueryFail(t *testing.T) {
	e := lowlevel.MakeOsExec()
	_, err := e.Read("sh", []string{"-c", "echo foo && false"})
	assert.Error(t, err)
	// FIXME: underlying exec.Command looks not return output on error
	//        would be nice to have all output in logs in case of error
	// assert.Equal(t, "foo", out)
}

func TestReadMultiline(t *testing.T) {
	e := lowlevel.MakeOsExec()
	out, err := e.Read("sh", []string{"-c", "echo foo && echo bar"})
	assert.NoError(t, err)
	assert.Equal(t, "foo\nbar", out)
}
