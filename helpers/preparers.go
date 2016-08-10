package helpers

import (
	"fmt"
	"testing"

	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource"
	"github.com/stretchr/testify/assert"
)

func PreparerValidator(t *testing.T, preparers []resource.Resource, errs []string) {
	assert.Equal(t, len(preparers), len(errs))
	fr := fakerenderer.FakeRenderer{}
	for i := range preparers {
		_, err := preparers[i].Prepare(&fr)
		if errs[i] == "" {
			assert.NoError(t, err, fmt.Sprintf("Test Index: %d", i))
		} else {
			assert.EqualError(t, err, errs[i], fmt.Sprintf("Test Index: %d", i))
		}
	}
}
