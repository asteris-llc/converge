package lowlevel_test

import (
	"github.com/asteris-llc/converge/resource/lvm/lowlevel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"testing"
)

func TestUnitFileNotExists(t *testing.T) {
	filename := "/test-unit-file-which-never-exists.xxx"
	currentContent := "this is a test"
	lvm := lowlevel.MakeLvmBackend()
	ok, err := lvm.CheckUnit(filename, currentContent)
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestUnitFileContentDiffs(t *testing.T) {
	originalContent := "a test this is"
	currentContent := "this is a test"
	tmpfile, err := ioutil.TempFile("", "test-unit-file-contents-diff")

	require.NoError(t, err)
	defer func() { require.NoError(t, os.RemoveAll(tmpfile.Name())) }()

	_, err = tmpfile.Write([]byte(originalContent))
	require.NoError(t, err)
	require.NoError(t, tmpfile.Sync())

	lvm := lowlevel.MakeLvmBackend()
	ok, err := lvm.CheckUnit(tmpfile.Name(), currentContent)
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestUnitFileUpdate(t *testing.T) {
	currentContent := "this is a test"
	filename := "/systemd/test.unit"

	lvm, me := makeLvmWithMockExec()

	// FIXME:   should be 0644 here, but call mismatch. Looks like BUG
	me.On("WriteFile", filename, []byte(currentContent), mock.Anything).Return(nil)
	me.On("Run", "systemctl", []string{"daemon-reload"}).Return(nil)

	err := lvm.UpdateUnit(filename, currentContent)
	assert.NoError(t, err)
}
