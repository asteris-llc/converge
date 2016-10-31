// Copyright © 2016 Asteris, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lowlevel_test

import (
	"github.com/asteris-llc/converge/resource/lvm/lowlevel"
	"github.com/asteris-llc/converge/resource/lvm/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"io/ioutil"
	"os"
	"testing"
)

// TestLVMCheckUnit tests LVM.CheckUnit
func TestLVMCheckUnit(t *testing.T) {
	t.Run("unit file not exists", func(t *testing.T) {
		filename := "/test-unit-file-which-never-exists.xxx"
		currentContent := "this is a test"
		lvm := lowlevel.MakeLvmBackend()
		ok, err := lvm.CheckUnit(filename, currentContent)
		assert.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("unit file content diffs", func(t *testing.T) {
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
	})

}

// TestLVMUpdateUnit tests LVM.UpdateUnit()
func TestLVMUpdateUnit(t *testing.T) {
	t.Run("update unit file", func(t *testing.T) {
		currentContent := "this is a test"
		filename := "/systemd/test.unit"

		lvm, me := testhelpers.MakeLvmWithMockExec()

		me.On("WriteFile", filename, []byte(currentContent), (os.FileMode)(0644)).Return(nil)
		me.On("Run", "systemctl", []string{"daemon-reload"}).Return(nil)

		err := lvm.UpdateUnit(filename, currentContent)
		assert.NoError(t, err)
		me.AssertCalled(t, "WriteFile", filename, []byte(currentContent), (os.FileMode)(0644))
	})
}
