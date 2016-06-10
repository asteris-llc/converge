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
	dependencies := make([]string, 0)
	for _, r := range mod.Resources {
		dependencies = append(dependencies, r.Depends()...)
	}
	return dependencies
}

func TestParseSimpleDependencies(t *testing.T) {
	t.Parallel()
	//Test simple Dependencies list
	mod, err := load.Parse([]byte(dependentModule))
	assert.NoError(t, err)
	assert.Equal(t, 5, len(mod.Resources))
	dependencies := getAllDeps(mod)
	assert.Equal(t, 3, len(dependencies))
}

func TestParseEmptyDependencies(t *testing.T) {
	t.Parallel()
	//Test empty Dependencies
	mod, err := load.Parse([]byte(emptyDependenciesModule))
	assert.NoError(t, err)
	dependencies := getAllDeps(mod)
	assert.Equal(t, 0, len(dependencies))
}

func TestDependentCall(t *testing.T) {
	t.Parallel()
	//Test DependentCall
	mod, err := load.Parse([]byte(dependentCall))
	assert.NoError(t, err)
	assert.Equal(t, "module.y", getAllDeps(mod)[0])
}

//TestAutoDependencies test that if the depends field is not set for a Task,
//the previously declared Task becomes that Task's dependency
func TestAutoDependencies(t *testing.T) {
	t.Parallel()
	//Test task is a dependency of subsequent task.
	mod, err := load.Parse([]byte(taskDependenciesModule))
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

func TestSameNameDependencies_Wrong(t *testing.T) {
	t.Parallel()
	_, err := load.Parse([]byte(sameNameDependenciesWrong))
	assert.Error(t, err)
}

func TestSameNameDependencies_Right(t *testing.T) {
	t.Parallel()
	_, err := load.Parse([]byte(sameNameDependenciesRight))
	assert.NoError(t, err)
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
		assert.EqualError(t, err, "At 3:1: duplicate param \"param.x\"")
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
		assert.EqualError(t, err, "At 3:1: duplicate task \"task.x\"")
	}
}

func TestBadName(t *testing.T) {
	t.Parallel()

	_, err := load.Parse([]byte(badName))
	if assert.Error(t, err) {
		assert.EqualError(t, err, `At 2:1: invalid name "task.a b"`)
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
	param "message" { default = "Hello, World!" }
	param "filename" { default = "test.txt" }

	task "morenothing" {
	check = ""
	apply = ""
}

	task "nothing" {
	check = ""
	apply = ""
	}

	task "render" {
	  check = "cat {{param 'filename'}} | tee /dev/stderr | grep -q '{{param 'message'}}'"
	  apply = "echo '{{param 'message'}}' > {{param 'filename'}} && cat {{param 'filename'}}"
	  depends = ["task.nothing", "task.morenothing"]
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

	sameNameDependenciesWrong = `
	template "a" {
		content = ""
	}

	task "a" {
		check = ""
		apply = ""
	}

	task "c" {
		check = ""
		apply = ""
		depends = ["a"]
	}
`

	sameNameDependenciesRight = `
	template "a" {
		content = ""
	}

	task "a" {
		check = ""
		apply = ""
	}

	task "c" {
		check = ""
		apply = ""
		depends = ["template.a"]
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

	badName = `
task "a b" { }
`
)
