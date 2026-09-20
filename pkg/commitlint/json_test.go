package commitlint

import (
	"encoding/json/v2"
	"testing"

	"github.com/go-openapi/testify/v2/assert"
)

func TestUnmarshalJSON(t *testing.T) {
	var actual RuleValue[int]
	err := json.Unmarshal([]byte(`[2, "always", 123]`), &actual)

	assert.NoError(t, err)
	assert.EqualT(t, 2, actual.Level)
	assert.EqualT(t, Always, actual.When)
	assert.EqualT(t, 123, actual.Value)
	assert.EqualT(t, true, actual.Set)

	err = json.Unmarshal([]byte(`[1, "never", 123]`), &actual)

	assert.NoError(t, err)
	assert.EqualT(t, 1, actual.Level)
	assert.EqualT(t, Never, actual.When)
	assert.EqualT(t, 123, actual.Value)
	assert.EqualT(t, true, actual.Set)

	err = json.Unmarshal([]byte(`"NOT ARRAY"`), &actual)
	assert.Error(t, err)

	err = json.Unmarshal([]byte(`[]`), &actual)
	assert.Error(t, err)

	err = json.Unmarshal([]byte(`["BAD", "always", 123]`), &actual)
	assert.Error(t, err)

	err = json.Unmarshal([]byte(`[2, "BAD", 123]`), &actual)
	assert.Error(t, err)

	err = json.Unmarshal([]byte(`[2, "always", "BAD"]`), &actual)
	assert.Error(t, err)
}
