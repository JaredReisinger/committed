package tui

import (
	"testing"

	"github.com/go-openapi/testify/v2/assert"
	"github.com/jaredreisinger/committed/pkg/commitlint"
)

// We should be able to come up with the four tests: always/never and pass/fail
// (and maybe the levels as well?)

func TestValidationNoRules(t *testing.T) {
	// *commitlint.RuleValue[string]
	err := validate("ANYTHING", nil)
	assert.NoError(t, err)
}

func TestValidationCaseRule(t *testing.T) {
	rule := &commitlint.RuleValue[string]{
		Level: commitlint.Error,
		When:  commitlint.Always,
		Value: "lower-case",
		Set:   true,
	}

	err := validate("okay", &fieldRules{caseRule: rule})
	assert.NoError(t, err)

	err = validate("NOTGOOD", &fieldRules{caseRule: rule})
	assert.Error(t, err)
}

func TestValidationMultiCaseRule(t *testing.T) {
	rule := &commitlint.RuleValue[[]string]{
		Level: commitlint.Error,
		When:  commitlint.Always,
		Value: []string{"lower-case"},
		Set:   true,
	}

	err := validate("okay", &fieldRules{multiCaseRule: rule})
	assert.NoError(t, err)

	err = validate("NOTGOOD", &fieldRules{multiCaseRule: rule})
	assert.Error(t, err)
}

func TestValidationDelimiterStyleRule(t *testing.T) {
	t.Skip(t, "delimiter rule is not implemented")

	rule := &commitlint.RuleValue[string]{
		Level: commitlint.Error,
		When:  commitlint.Always,
		Value: "-",
		Set:   true,
	}

	err := validate("this-is-okay", &fieldRules{delimiterStyleRule: rule})
	assert.NoError(t, err)

	err = validate("NOT_GOOD", &fieldRules{delimiterStyleRule: rule})
	assert.Error(t, err)
}

func TestValidationEmptyRule(t *testing.T) {
	rule := &commitlint.RuleValue[commitlint.NoValue]{
		Level: commitlint.Error,
		When:  commitlint.Never,
		Value: commitlint.NoValue{},
		Set:   true,
	}

	err := validate("this-is-okay", &fieldRules{emptyRule: rule})
	assert.NoError(t, err)

	err = validate("", &fieldRules{emptyRule: rule})
	assert.Error(t, err)
}

func TestValidationEnumRule(t *testing.T) {
	rule := &commitlint.RuleValue[[]string]{
		Level: commitlint.Error,
		When:  commitlint.Always,
		Value: []string{"one", "two"},
		Set:   true,
	}

	err := validate("two", &fieldRules{enumRule: rule})
	assert.NoError(t, err)

	err = validate("NOTGOOD", &fieldRules{enumRule: rule})
	assert.Error(t, err)
}

func TestValidationFullStopRule(t *testing.T) {
	rule := &commitlint.RuleValue[string]{
		Level: commitlint.Error,
		When:  commitlint.Always,
		Value: ".",
		Set:   true,
	}

	err := validate("okay.", &fieldRules{fullStopRule: rule})
	assert.NoError(t, err)

	err = validate("BAD!", &fieldRules{fullStopRule: rule})
	assert.Error(t, err)
}

func TestValidationMaxLengthRule(t *testing.T) {
	rule := &commitlint.RuleValue[int]{
		Level: commitlint.Error,
		When:  commitlint.Always,
		Value: 5,
		Set:   true,
	}

	err := validate("good", &fieldRules{maxLengthRule: rule})
	assert.NoError(t, err)

	err = validate("THIS_IS_BAD", &fieldRules{maxLengthRule: rule})
	assert.Error(t, err)
}

func TestValidationMinLengthRule(t *testing.T) {
	rule := &commitlint.RuleValue[int]{
		Level: commitlint.Error,
		When:  commitlint.Always,
		Value: 5,
		Set:   true,
	}

	err := validate("this is good", &fieldRules{minLengthRule: rule})
	assert.NoError(t, err)

	err = validate("BAD", &fieldRules{minLengthRule: rule})
	assert.Error(t, err)
}
