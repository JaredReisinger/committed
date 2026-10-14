package commitlint

import (
	"testing"

	"github.com/go-openapi/testify/v2/assert"
)

func TestRulesMerge(t *testing.T) {
	rulesA := DefaultRules
	assert.False(t, rulesA.BodyCase.Active())
	assert.Equal(t, "DEFAULT", rulesA.BodyCase.GetValue("DEFAULT"))

	rulesA.Merge(&Rules{
		BodyCase: RuleValue[string]{Error, Always, "upper-case", true},
	})

	assert.True(t, rulesA.BodyCase.Active())
	assert.Equal(t, "upper-case", rulesA.BodyCase.GetValue("DEFAULT"))
}
