package helpers

import (
	"fmt"
	"testing"

	"github.com/asteris-llc/converge/resource"
	"github.com/stretchr/testify/assert"
)

func SquashCheck(status1 string, willChange1 bool, err1 error, status2 string, willChange2 bool, err2 error) (string, bool, error) {
	s, c := fmt.Sprintf("%s\n%s", status1, status2), willChange1 || willChange2
	e := MultiErrorAppend(err1, err2)
	return s, c, e
}

type CheckValidator func(status string, willChange bool, err error, i int, t *testing.T)

func CheckValidatorCreator(status string, willChange bool, err string) CheckValidator {
	return func(s string, w bool, e error, i int, t *testing.T) {
		msg := fmt.Sprintf("Test index: %d", i)
		assert.Regexp(t, status, s, msg)
		assert.Equal(t, willChange, w, msg)
		if err == "" {
			assert.NoError(t, e, msg)
		} else {
			assert.EqualError(t, e, err, msg)
		}
	}
}

func TaskCheckValidator(tasks []resource.Task, checks []CheckValidator, t *testing.T) {
	assert.Equal(t, len(tasks), len(checks), fmt.Sprintf("Length missmatch. Given %d task but %d checks", len(tasks), len(checks)))
	for i := range tasks {
		assert.NotNil(t, checks[i])
		status, willChange, err := tasks[i].Check()
		checks[i](status, willChange, err, i, t)
	}
}
