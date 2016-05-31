package load_test

import (
	"testing"

	"github.com/asteris-llc/converge/load"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	t.Parallel()

	m, err := load.Parse([]byte(basicModule))

	assert.NoError(t, err)

	// params
	assert.NotNil(t, m.Params["filename"])
	if assert.NotNil(t, m.Params["permissions"]) {
		assert.Equal(t, *m.Params["permissions"].Default, "0600")
	}

	// resources
	assert.Equal(t, len(m.Resources), 2)
}

func TestParseDependcies(t *testing.T) {
	t.Parallel()

	m, err := load.Parse([]byte(dependentModule))

	assert.NoError(t, err)

	// params
	assert.NotNil(t, m.Params["filename"])
	if assert.NotNil(t, m.Params["permissions"]) {
		assert.Equal(t, *m.Params["permissions"].Default, "0600")
	}
	// resources
	assert.Equal(t, len(m.Resources), 2)
	dependencies := []string{}
	for _, r := range m.Resources {
		dependencies = append(dependencies, r.Depends()...)
	}
	assert.Equal(t, len(dependencies), 1)
}

func TestParseAnonymousParam(t *testing.T) {
	t.Parallel()

	_, err := load.Parse([]byte(`param {}`))
	if assert.Error(t, err) {
		assert.EqualError(t, err, "At 1:1: param has no name (expected `param \"name\"`)")
	}
}

func TestParseDuplicateParam(t *testing.T) {
	t.Parallel()

	_, err := load.Parse([]byte(duplicateParam))
	if assert.Error(t, err) {
		assert.EqualError(t, err, "At 3:1: duplicate param \"x\"")
	}
}

func TestParseAnonymousTask(t *testing.T) {
	t.Parallel()

	_, err := load.Parse([]byte(`task {}`))
	if assert.Error(t, err) {
		assert.EqualError(t, err, "At 1:1: task has no name (expected `task \"name\"`)")
	}
}

func TestParseAnonymousTemplate(t *testing.T) {
	t.Parallel()

	_, err := load.Parse([]byte(`template {}`))
	if assert.Error(t, err) {
		assert.EqualError(t, err, "At 1:1: template has no name (expected `template \"name\"`)")
	}
}

func TestParseModuleCall(t *testing.T) {
	t.Parallel()

	mod, err := load.Parse([]byte(moduleCall))
	assert.NoError(t, err)

	assert.Equal(t, len(mod.Resources), 1)
}

func TestParseDependentCall(t *testing.T) {
	t.Parallel()

	mod, err := load.Parse([]byte(dependentCall))
	assert.NoError(t, err)
	assert.Equal(t, len(mod.Resources), 2)
	dependencies := make([]string, 0, 1)
	for _, r := range mod.Resources {
		dependencies = append(dependencies, r.Depends()...)
	}
	assert.Equal(t, len(dependencies), 1)
}

func TestParseDuplicateTask(t *testing.T) {
	t.Parallel()

	_, err := load.Parse([]byte(duplicateTask))
	if assert.Error(t, err) {
		assert.EqualError(t, err, "At 3:1: duplicate task \"x\"")
	}
}

var (
	basicModule = `
param "filename" { }
param "permissions" { default = "0600" }


template "test" {
content = ""
}
task "permission" {
  check = "stat -f \"%Op\" {{param \"filename\"}} tee /dev/stderr | grep -q {{param \"permission\"}}"
  apply = "chmod {{param \"permission\"}} {{param \"filename\"}}"
}
`

	dependentModule = `
param "filename" { }
param "permissions" { default = "0600" }

template "test" {
content = ""
}

task "permission" {
check = "stat -f \"%Op\" {{param \"filename\"}} tee /dev/stderr | grep -q {{param \"permission\"}}"
apply = "chmod {{param \"permission\"}} {{param \"filename\"}}"
depends = ["test"]
}
`

	duplicateParam = `
param "x" {}
param "x" {}`

	moduleCall = `
module "x" "y" {
  arg1 = "z"
}`

	dependentCall = `
	module "x" "y" {
	arg1 = "z"
}
module "a" "b" {
	arg1 = "c"
	depends = ["y"]
}
`

	duplicateTask = `
task "x" { }
task "x" { }
`
)
