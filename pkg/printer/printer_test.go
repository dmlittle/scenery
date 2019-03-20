package printer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrint(t *testing.T) {
	t.Run("parses simple plans", func(tt *testing.T) {
		input := "this:\n  is:\n    an:\n      - example"

		expected := "this:\n" +
			"         is:\n" +
			"           an:\n" +
			"             - example"

		value := formatValue(string(input), 1)

		assert.Equal(tt, expected, value)
	})

}

func String(v string) *string {
	return &v
}
