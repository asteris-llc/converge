package lowlevel_test

import (
	"github.com/asteris-llc/converge/resource/lvm/lowlevel"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseSizeSimple(t *testing.T) {
	size, option, unit, err := lowlevel.ParseSize("100G")
	assert.NoError(t, err)
	assert.Equal(t, int64(100), size)
	assert.Equal(t, "L", option)
	assert.Equal(t, "G", unit)
}

func TestParseSizeSimplePercents(t *testing.T) {
	size, option, unit, err := lowlevel.ParseSize("100%FREE")
	assert.NoError(t, err)
	assert.Equal(t, int64(100), size)
	assert.Equal(t, "l", option)
	assert.Equal(t, "%FREE", unit)
}

func TestParsePercents(t *testing.T) {
	_, _, unit, err := lowlevel.ParseSize("99%FREE")
	assert.NoError(t, err)
	assert.Equal(t, "%FREE", unit)

	_, _, unit, err = lowlevel.ParseSize("99%VG")
	assert.NoError(t, err)
	assert.Equal(t, "%VG", unit)

	_, _, unit, err = lowlevel.ParseSize("99%PVS")
	assert.NoError(t, err)
	assert.Equal(t, "%PVS", unit)
}

func TestBadPercentageUnit(t *testing.T) {
	_, _, _, err := lowlevel.ParseSize("100%XXX")
	assert.Error(t, err)
}

func TestBadPercentage(t *testing.T) {
	_, _, _, err := lowlevel.ParseSize("146%FREE")
	assert.Error(t, err)
}

func TestBadUnit(t *testing.T) {
	_, _, _, err := lowlevel.ParseSize("146X")
	assert.Error(t, err)
}
