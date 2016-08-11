package helpers

import (
	"fmt"
	"testing"

	"github.com/asteris-llc/converge/resource"
	"github.com/stretchr/testify/assert"
)

func TaskApplyValidator(tasks []resource.Task, errs []string, t *testing.T) {
	assert.Equal(t, len(tasks), len(errs), fmt.Sprintf("Length missmatch. Given %d task but %d errors", len(tasks), len(errs)))
	for i := range tasks {
		msg := fmt.Sprintf("Test index: %d", i)
		err := tasks[i].Apply()
		if errs[i] == "" {
			assert.NoError(t, err, msg)
		} else {
			assert.EqualError(t, err, errs[i], msg)
		}
	}
}
