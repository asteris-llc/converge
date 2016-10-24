package lowlevel_test

import (
	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/asteris-llc/converge/resource/lvm/lowlevel"
	"github.com/stretchr/testify/assert"

	"testing"
)

// TestParseSize tests ParseSize()
func TestParseSize(t *testing.T) {
	t.Parallel()
	defer logging.HideLogs(t)()

	t.Run("normal absolute values", func(t *testing.T) {
		size, err := lowlevel.ParseSize("100G")
		assert.NoError(t, err)
		assert.Equal(t, int64(100), size.Size)
		assert.Equal(t, false, size.Relative)
		assert.Equal(t, "G", size.Unit)

		assert.Equal(t, "-L", size.Option())
		assert.Equal(t, "100G", size.String())
	})

	t.Run("normal relative values", func(t *testing.T) {
		size, err := lowlevel.ParseSize("100%FREE")
		assert.NoError(t, err)
		assert.Equal(t, int64(100), size.Size)
		assert.Equal(t, true, size.Relative)
		assert.Equal(t, "%FREE", size.Unit)

		assert.Equal(t, "-l", size.Option())
		assert.Equal(t, "100%FREE", size.String())
	})

	t.Run("percentage units", func(t *testing.T) {
		size, err := lowlevel.ParseSize("99%FREE")
		assert.NoError(t, err)
		assert.Equal(t, "%FREE", size.Unit)

		size, err = lowlevel.ParseSize("99%VG")
		assert.NoError(t, err)
		assert.Equal(t, "%VG", size.Unit)

		size, err = lowlevel.ParseSize("99%PVS")
		assert.NoError(t, err)
		assert.Equal(t, "%PVS", size.Unit)
	})

	t.Run("bad percentage unit", func(t *testing.T) {
		_, err := lowlevel.ParseSize("100%XYZ")
		assert.Error(t, err)
	})

	t.Run("bad percentage (overflow)", func(t *testing.T) {
		_, err := lowlevel.ParseSize("146%FREE")
		assert.Error(t, err)
	})

	t.Run("bad unit", func(t *testing.T) {
		_, err := lowlevel.ParseSize("146X")
		assert.Error(t, err)
	})
}
