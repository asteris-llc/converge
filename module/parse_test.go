package module_test

import (
	"testing"

	"github.com/asteris-llc/converge/module"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Parallel()

	m, err := module.New([]byte(basicModule))

	assert.NoError(t, err)

	// params
	assert.NotNil(t, m.Params["filename"])
	if assert.NotNil(t, m.Params["permissions"]) {
		assert.Equal(t, *m.Params["permissions"].Default, "0600")
	}

	// resources
	assert.Equal(t, len(m.Resources), 2)
}

func TestNewAnonymousParam(t *testing.T) {
	t.Parallel()

	_, err := module.New([]byte(`param {}`))
	if assert.Error(t, err) {
		assert.EqualError(t, err, "At 1:1: param has no name (expected `param \"name\"`)")
	}
}

func TestNewDuplicateParam(t *testing.T) {
	t.Parallel()

	_, err := module.New([]byte(duplicateParam))
	if assert.Error(t, err) {
		assert.EqualError(t, err, "At 3:1: duplicate param \"x\"")
	}
}

func TestNewAnonymousTask(t *testing.T) {
	t.Parallel()

	_, err := module.New([]byte(`task {}`))
	if assert.Error(t, err) {
		assert.EqualError(t, err, "At 1:1: task has no name (expected `task \"name\"`)")
	}
}

func TestNewAnonymousTemplate(t *testing.T) {
	t.Parallel()

	_, err := module.New([]byte(`template {}`))
	if assert.Error(t, err) {
		assert.EqualError(t, err, "At 1:1: template has no name (expected `template \"name\"`)")
	}
}

var (
	basicModule = `
param "filename" { }
param "permissions" { default = "0600" }

task "permission" {
  check = "stat -f \"%Op\" {{param \"filename\"}} tee /dev/stderr | grep -q {{param \"permission\"}}"
  apply = "chmod {{param \"permission\"}} {{param \"filename\"}}"
}

template "test" {
  content = ""
}
`

	duplicateParam = `
param "x" {}
param "x" {}`
)
