package tui

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"unicode"

	"github.com/jaredreisinger/committed/pkg/commitlint"
)

// fieldRules is the union of all field-specific rules. Some rules, like
// "leading blank", "max line length", and "trim" are handled by the message
// composition, not in the editor itself.
type fieldRules struct {
	caseRule           *commitlint.RuleValue[string]
	multiCaseRule      *commitlint.RuleValue[[]string]
	delimiterStyleRule *commitlint.RuleValue[string]
	emptyRule          *commitlint.RuleValue[commitlint.NoValue]
	enumRule           *commitlint.RuleValue[[]string]
	fullStopRule       *commitlint.RuleValue[string]
	maxLengthRule      *commitlint.RuleValue[int]
	minLengthRule      *commitlint.RuleValue[int]
}

type cases int

const (
	lowerCase    cases = 1 << iota // lower-case    -- lowercase
	upperCase                      // upper-case    -- UPPERCASE
	camelCase                      // camel-case    -- camelCase
	kebabCase                      // kebab-case    -- kebab-case
	pascalCase                     // pascal-case   -- PascalCase
	sentenceCase                   // sentence-case -- Sentence case
	snakeCase                      // snake-case    -- snake_case
	startCase                      // start-case    -- Start Case (a.k.a. Title Case)

	allCases = lowerCase | upperCase | camelCase | kebabCase | pascalCase | sentenceCase | snakeCase | startCase
)

var caseNames = []string{
	"lower-case",
	"upper-case",
	"camel-case",
	"kebab-case",
	"pascal-case",
	"sentence-case",
	"snake-case",
	"start-case",
}

func (c cases) String() string {
	set := make([]string, 0, len(caseNames))
	for i, name := range caseNames {
		if c&(1<<i) != 0 {
			set = append(set, name)
		}
	}
	return strings.Join(set, "|")
}

func (c *cases) UnmarshalString(s string) {
	*c = 0

	for i, name := range caseNames {
		if s == name {
			*c = cases(1 << i)
		}
	}
}

func validate(s string, rules *fieldRules) error {
	if rules == nil {
		return nil
	}

	slog.Debug("validating...", "string", s)

	var errs []error

	// although fieldRules defines the ruleset in alphabetical order, we use a
	// heuristic order here to put the "most likely to be important" rules
	// earlier in the result

	if rules.enumRule.Active() {
		inList := slices.Contains(rules.enumRule.Value, s)
		enumStr := strings.Join(rules.enumRule.Value, ", ")
		switch rules.enumRule.When {
		case commitlint.Always:
			if !inList {
				errs = append(errs, fmt.Errorf("must be one of: %s", enumStr))
			}
		case commitlint.Never:
			if inList {
				errs = append(errs, fmt.Errorf("must not be one of: %s", enumStr))
			}
		}
	}

	if rules.emptyRule.Active() {
		isEmpty := len(s) == 0
		shouldBeEmpty := rules.emptyRule.When == commitlint.Always

		if isEmpty != shouldBeEmpty {
			if shouldBeEmpty {
				errs = append(errs, errors.New("value is not empty, and should be"))
			} else {
				errs = append(errs, errors.New("value is empty, and should not be"))
			}
		}
	}

	if rules.minLengthRule.Active() {
		actual := len(s)
		expected := rules.minLengthRule.Value
		switch rules.minLengthRule.When {
		case commitlint.Always:
			if actual < expected {
				errs = append(errs, fmt.Errorf("must be at least length %d (currently %d)", expected, actual))
			}
		case commitlint.Never:
			// what does this even mean?

			// if actual >= expected {
			// 	errs = append(errs, fmt.Errorf("must be at least length %d (currently %d)", expected, actual))
			// }
		}
	}

	if rules.maxLengthRule.Active() {
		actual := len(s)
		expected := rules.maxLengthRule.Value
		switch rules.maxLengthRule.When {
		case commitlint.Always:
			if actual > expected {
				errs = append(errs, fmt.Errorf("must be no more than length %d (currently %d)", expected, actual))
			}
		case commitlint.Never:
			// what does this even mean?

			// if actual >= expected {
			// 	errs = append(errs, fmt.Errorf("must be at least length %d (currently %d)", expected, actual))
			// }
		}
	}

	if rules.caseRule.Active() {
		possibleCases := inferCases(s)

		var expectedCase cases
		expectedCase.UnmarshalString(rules.caseRule.Value)

		slog.Debug("casing inference", "string", s, "expected", expectedCase.String(), "possible", possibleCases.String())

		if possibleCases&expectedCase == 0 {
			errs = append(errs, fmt.Errorf("must be %s", expectedCase.String()))
		}
	}

	if rules.multiCaseRule.Active() {
		possibleCases := inferCases(s)

		var expectedCases cases
		for _, val := range rules.multiCaseRule.Value {
			var tmpCase cases
			tmpCase.UnmarshalString(val)
			expectedCases |= tmpCase
		}

		slog.Debug("casing inference", "string", s, "expected", expectedCases.String(), "possible", possibleCases.String())

		expectedStr := strings.Join(rules.multiCaseRule.Value, ", ")

		switch rules.multiCaseRule.When {
		case commitlint.Always:
			if possibleCases&expectedCases == 0 {
				errs = append(errs, fmt.Errorf("must be cased as one of: %s", expectedStr))
			}
		case commitlint.Never:
			if possibleCases&expectedCases != 0 {
				errs = append(errs, fmt.Errorf("must not be cased as one of: %s", expectedStr))
			}
		}
	}

	if rules.delimiterStyleRule.Active() {

	}

	// This one is obnoxious, as it will make the field border flicker or stay
	// red most of the time.  Maybe we should just auto-append?
	if rules.fullStopRule.Active() {
		hasFullStop := strings.HasSuffix(s, rules.fullStopRule.Value)
		switch rules.fullStopRule.When {
		case commitlint.Always:
			if !hasFullStop {
				errs = append(errs, fmt.Errorf("must end with %q", rules.fullStopRule.Value))
			}
		case commitlint.Never:
			if hasFullStop {
				errs = append(errs, fmt.Errorf("must not end with %q", rules.fullStopRule.Value))
			}
		}
	}

	return errors.Join(errs...)
}

func inferCases(s string) cases {
	possibleCases := allCases
	sawWhitespace := false
	for i, r := range s {
		if unicode.IsSpace(r) {
			sawWhitespace = true
			possibleCases = possibleCases &^ (camelCase | kebabCase | pascalCase | snakeCase)
		} else if unicode.IsLetter(r) {
			if unicode.IsUpper(r) {
				possibleCases = possibleCases &^ (lowerCase | kebabCase | snakeCase)
				if i == 0 || sawWhitespace {
					possibleCases = possibleCases &^ camelCase
				}

				if sawWhitespace {
					possibleCases = possibleCases &^ sentenceCase
				}
			} else if unicode.IsLower(r) {
				possibleCases = possibleCases &^ upperCase
				if i == 0 {
					possibleCases = possibleCases &^ (pascalCase | sentenceCase | startCase)
				} else if sawWhitespace {
					possibleCases = possibleCases &^ (camelCase | pascalCase | startCase)
				}
			}
			sawWhitespace = false
		} else {
			switch r {
			case '-':
				possibleCases = possibleCases &^ (camelCase | pascalCase | snakeCase)
			case '_':
				possibleCases = possibleCases &^ (camelCase | kebabCase | pascalCase)
			}
			// digit?
			sawWhitespace = false
		}
	}
	return possibleCases
}
