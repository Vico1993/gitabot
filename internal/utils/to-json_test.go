package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToJsonOutput(t *testing.T) {
	content := map[string]interface{}{"foo": "bar"}

	res := ToJson(content)

	expected := `{
    "foo": "bar"
}`

	assert.Equal(t, expected, res)
}
