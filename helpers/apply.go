package helpers

import (
	"strconv"
	"testing"

	"github.com/asteris-llc/converge/resource"
	"github.com/stretchr/testify/assert"
)

func TaskApplyValidator(tasks []resource.Task, errs []string, t *testing.T) {
	assert.Equal(t, len(tasks), len(errs))
	for i := range tasks {
		err := tasks[i].Apply()
		if errs[i] == "" {
			assert.NoError(t, err, strconv.Itoa(i))
		} else {
			assert.EqualError(t, err, errs[i], strconv.Itoa(i))
		}
	}
}
