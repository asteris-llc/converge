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
		err := tasks[i].Apply()
		if errs[i] == "" {
			assert.NoError(t, err, fmt.Sprintf("Test Index: %d", i))
		} else {
			assert.EqualError(t, err, errs[i], fmt.Sprintf("Test Index: %d", i))
		}
	}
}
