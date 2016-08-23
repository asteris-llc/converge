package resource_test

import (
	"testing"

	"github.com/asteris-llc/converge/healthcheck"
	"github.com/asteris-llc/converge/resource"
	"github.com/stretchr/testify/assert"
)

func Test_Status_ImplementsCheck(t *testing.T) {
	assert.Implements(t, (*healthcheck.Check)(nil), new(resource.Status))
}
