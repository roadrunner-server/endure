package issues

import (
	"testing"

	"github.com/spiral/endure"
	"github.com/spiral/endure/tests/issues/issue33"
	"github.com/stretchr/testify/assert"
)

// Provided structure instead of function
func TestEndure_Issue33(t *testing.T) {
	c, err := endure.NewContainer(endure.DebugLevel, endure.RetryOnFail(true))
	assert.NoError(t, err)

	assert.Error(t, c.Register(&issue33.Plugin1{}))
	assert.NoError(t, c.Register(&issue33.Plugin2{}))
}
