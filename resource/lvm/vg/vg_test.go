package vg_test

import (
	"github.com/asteris-llc/converge/helpers/comparison"
	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource/lvm/testdata"
	"github.com/asteris-llc/converge/resource/lvm/testhelpers"
	"github.com/asteris-llc/converge/resource/lvm/vg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"testing"
)

// TestCreateVolume is a full-blown test, using fake engine to trace from high-level
// graph node vg.resourceVG, to commands passed to LVM tools. It cover only straighforward
// cases. Use mock-LVM for real tests of highlevel stuff.
func TestCreateVolume(t *testing.T) {
	t.Run("single device", func(t *testing.T) {
		lvm, me := testhelpers.MakeLvmWithMockExec()

		me.On("Getuid").Return(0)                  // assume, that we have root
		me.On("Lookup", mock.Anything).Return(nil) // assume, that all tools are here

		me.On("Read", "pvs", mock.Anything).Return("", nil)
		me.On("Read", "vgs", mock.Anything).Return("", nil)
		me.On("Run", "vgcreate", []string{"vg0", "/dev/sda1"}).Return(nil)

		fr := fakerenderer.New()

		r := vg.NewResourceVG(lvm, "vg0", []string{"/dev/sda1"})
		status, err := r.Check(fr)
		assert.NoError(t, err)
		assert.True(t, status.HasChanges())
		comparison.AssertDiff(t, status.Diffs(), "vg0", "<not exists>", "/dev/sda1")

		status, err = r.Apply()
		assert.NoError(t, err)
		me.AssertCalled(t, "Run", "vgcreate", []string{"vg0", "/dev/sda1"})
	})

	t.Run("multiple devices", func(t *testing.T) {
		lvm, me := testhelpers.MakeLvmWithMockExec()

		me.On("Getuid").Return(0)                  // assume, that we have root
		me.On("Lookup", mock.Anything).Return(nil) // assume, that all tools are here

		me.On("Read", "pvs", mock.Anything).Return(testdata.Pvs, nil)
		me.On("Read", "vgs", mock.Anything).Return(testdata.Vgs, nil)

		fr := fakerenderer.New()

		r := vg.NewResourceVG(lvm, "vg1", []string{"/dev/md127"})
		_, err := r.Check(fr)
		assert.Error(t, err)
	})

	t.Run("volume which already exists", func(t *testing.T) {
		lvm, me := testhelpers.MakeLvmWithMockExec()

		me.On("Getuid").Return(0)                  // assume, that we have root
		me.On("Lookup", mock.Anything).Return(nil) // assume, that all tools are here

		me.On("Read", "pvs", mock.Anything).Return(testdata.Pvs, nil)
		me.On("Read", "vgs", mock.Anything).Return(testdata.Vgs, nil)

		fr := fakerenderer.New()

		r := vg.NewResourceVG(lvm, "vg0", []string{"/dev/md127"})
		status, err := r.Check(fr)
		assert.NoError(t, err)
		assert.False(t, status.HasChanges())
	})
}
