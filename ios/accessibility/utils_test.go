package accessibility

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertToStringList(t *testing.T) {
	t.Run("valid payload", func(t *testing.T) {
		payload := []interface{}{[]interface{}{"a", "b", "c"}}
		result, err := convertToStringList(payload)
		require.NoError(t, err)
		assert.Equal(t, []string{"a", "b", "c"}, result)
	})

	t.Run("payload element is not a list", func(t *testing.T) {
		assert.NotPanics(t, func() {
			payload := []interface{}{"not-a-list"}
			result, err := convertToStringList(payload)
			assert.Error(t, err)
			assert.Nil(t, result)
		})
	})

	t.Run("list element is not a string", func(t *testing.T) {
		assert.NotPanics(t, func() {
			payload := []interface{}{[]interface{}{"a", 42}}
			result, err := convertToStringList(payload)
			assert.Error(t, err)
			assert.Nil(t, result)
		})
	})

	t.Run("wrong payload length", func(t *testing.T) {
		payload := []interface{}{}
		result, err := convertToStringList(payload)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
