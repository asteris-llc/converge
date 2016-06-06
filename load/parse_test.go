package load_test

import (
	"testing"

	"github.com/asteris-llc/converge/load"
	"github.com/asteris-llc/converge/resource"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	t.Parallel()

	m, err := load.Parse([]byte(basicModule))

	assert.NoError(t, err)

	// params
	params := m.Params()
	assert.NotNil(t, params["filename"])
	if assert.NotNil(t, params["permissions"]) {
		assert.Equal(t, params["permissions"].Default, resource.Value("0600"))
	}

	// resources
	assert.Equal(t, len(m.Resources), 4)
}

func getAllDeps(mod *resource.Module) []string {
	dependencies := make([]string, 0, 1)
	for _, r := range mod.Resources {
		dependencies = append(dependencies, r.Depends()...)
	}
	return dependencies
}

func TestParseDependcies(t *testing.T) {
	//t.Parallel()
	//Test simple Dependencies list
	mod, err := load.Parse([]byte(dependentModule))
	assert.NoError(t, err)
	assert.Equal(t, 4, len(mod.Resources))
	dependencies := getAllDeps(mod)
	assert.Equal(t, 2, len(dependencies))

	//Test empty Dependencies
	mod, err = load.Parse([]byte(emptyDependenciesModule))
	assert.NoError(t, err)
	dependencies = getAllDeps(mod)
	assert.Equal(t, 0, len(dependencies))

	//Test DependentCall
	mod, err = load.Parse([]byte(dependentCall))
	assert.NoError(t, err)
	assert.Equal(t, "module.y", getAllDeps(mod)[0])

	//Test task is a dependency of subsequent task.
	mod, err = load.Parse([]byte(taskDependenciesModule))
	assert.NoError(t, err)
	//First task
	r := mod.Resources[0]
	assert.Empty(t, r.Depends())
	//Second task
	r = mod.Resources[1]
	assert.Contains(t, r.Depends(), "task.a")
	//Third task
	r = mod.Resources[2]
	assert.Contains(t, r.Depends(), "task.b")

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
  check = "stat -f \"test%Op\" {{param \"filename\"}} tee /dev/stderr | grep -q {{param \"permission\"}}"
  apply = "chmod {{param \"permission\"}} {{param \"filename\"}}"
	depends = ["template.test", "param.permissions"]
}
`
	emptyDependenciesModule = `
param "filename" { }
param "permissions" { default = "0600" }


template "test" {
content = ""
}
task "permission" {
  check = "stat -f \"test%Op\" {{param \"filename\"}} tee /dev/stderr | grep -q {{param \"permission\"}}"
  apply = "chmod {{param \"permission\"}} {{param \"filename\"}}"
	depends = []
}
`

	taskDependenciesModule = `
	task "a" {
	check = ""
	apply = ""
}

	task "b" {
	check = ""
	apply = ""
}

	task "c" {
		check = ""
		apply = ""
	}

`

	duplicateParam = `
param "x" {}
param "x" {}`

	moduleCall = `
module "x" "y" {
	params = {
		arg1 = "z"
	}
}`

	dependentCall = `
	module "x" "y" {
	params = {
		arg1 = "z"
	}
}

	module "a" "b" {
	params = {
		arg1 = "z"
	}
	depends = ["module.y"]
}`

	duplicateTask = `
task "x" { }
task "x" { }
`
)
