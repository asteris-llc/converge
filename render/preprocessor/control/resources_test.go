package control_test

import (
	"testing"

	"github.com/asteris-llc/converge/render/preprocessor/control"
	"github.com/asteris-llc/converge/resource"
	"github.com/stretchr/testify/assert"
)

// TestInterfacesAreImplemented ensures that all types implement the correct
// interfaces
func TestInterfacesAreImplemented(t *testing.T) {
	t.Run("SwitchPreparer", func(t *testing.T) { assert.Implements(t, (*resource.Resource)(nil), new(control.SwitchPreparer)) })
	t.Run("SwitchResource", func(t *testing.T) { assert.Implements(t, (*resource.Task)(nil), new(control.SwitchResource)) })
	t.Run("CasePreparer", func(t *testing.T) { assert.Implements(t, (*resource.Resource)(nil), new(control.CasePreparer)) })
	t.Run("CaseResource", func(t *testing.T) { assert.Implements(t, (*resource.Task)(nil), new(control.CaseResource)) })
}
