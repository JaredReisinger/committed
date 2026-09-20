package commitlint

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"math"
)

// TODO: flesh out Rules values?
type Config struct {
	// Rules map[string]any  `json:"rules" yaml:"rules"`
	Rules Rules `json:"rules" yaml:"rules"`
}

type Level int

const (
	Disabled Level = iota
	Warning
	Error
)

type Applicable int

const (
	Never Applicable = iota
	Always
)

func (a *Applicable) UnmarshalString(s string) error {
	switch s {
	case "never":
		*a = Never
	case "always":
		*a = Always
	default:
		return errors.New("unrecognized applicable string")
	}
	return nil
}

func (a *Applicable) UnmarshalJSONFrom(decoder *jsontext.Decoder) error {
	var s string

	err := json.UnmarshalDecode(decoder, &s)
	if err != nil {
		return err
	}
	return a.UnmarshalString(s)
}

// RuleValue[T] represents any kind of rule value; using a generic makes parsing
// "smart".
type RuleValue[T any] struct {
	Level Level
	When  Applicable
	Value T
	Set   bool
}

// should we use nesting with merged parsing? (Body.Case, Body.Empty, ...)
type Rules struct {
	BodyCase          RuleValue[string] `json:"body-case" yaml:"body-case"`
	BodyEmpty         RuleValue[bool]   `json:"body-empty" yaml:"body-empty"`
	BodyFullStop      RuleValue[string] `json:"body-full-stop" yaml:"body-full-stop"`
	BodyLeadingBlank  RuleValue[bool]   `json:"body-leading-blank" yaml:"body-leading-blank"`
	BodyMaxLength     RuleValue[int]    `json:"body-max-length" yaml:"body-max-length"`
	BodyMaxLineLength RuleValue[int]    `json:"body-max-line-length" yaml:"body-max-line-length"`
	BodyMinLength     RuleValue[int]    `json:"body-min-length" yaml:"body-min-length"`

	BreakingChangeExclamationMark RuleValue[bool] `json:"breaking-change-exclamation-mark" yaml:"breaking-change-exclamation-mark"`

	FooterEmpty         RuleValue[bool] `json:"footer-empty" yaml:"footer-empty"`
	FooterLeadingBlank  RuleValue[bool] `json:"footer-leading-blank" yaml:"footer-leading-blank"`
	FooterMaxLength     RuleValue[int]  `json:"footer-max-length" yaml:"footer-max-length"`
	FooterMaxLineLength RuleValue[int]  `json:"footer-max-line-length" yaml:"footer-max-line-length"`
	FooterMinLength     RuleValue[int]  `json:"footer-min-length" yaml:"footer-min-length"`

	HeaderCase      RuleValue[string] `json:"header-case" yaml:"header-case"`
	HeaderFullStop  RuleValue[string] `json:"header-full-stop" yaml:"header-full-stop"`
	HeaderMaxLength RuleValue[int]    `json:"header-max-length" yaml:"header-max-length"`
	HeaderMinLength RuleValue[int]    `json:"header-min-length" yaml:"header-min-length"`
	HeaderTrim      RuleValue[bool]   `json:"header-trim" header:"header-trim"`

	ReferencesEmpty RuleValue[bool] `json:"references-empty" yaml:"references-empty"`

	ScopeCase           RuleValue[string]   `json:"scope-case" yaml:"scope-case"`
	ScopeDelimiterStyle RuleValue[[]string] `json:"scope-delimiter-style" yaml:"scope-delimiter-style"`
	ScopeEmpty          RuleValue[bool]     `json:"scope-empty" yaml:"scope-empty"`
	ScopeEnum           RuleValue[[]string] `json:"scope-enum" yaml:"scope-enum"`
	ScopeMaxLength      RuleValue[int]      `json:"scope-max-length" yaml:"scope-max-length"`
	ScopeMinLength      RuleValue[int]      `json:"scope-min-length" yaml:"scope-min-length"`

	SignedOfBy RuleValue[string] `json:"signed-off-by" yaml:"signed-off-by"`

	SubjectCase            RuleValue[[]string] `json:"subject-case" yaml:"subject-case"`
	SubjectEmpty           RuleValue[bool]     `json:"subject-empty" yaml:"subject-empty"`
	SubjectExclamationMark RuleValue[bool]     `json:"subject-exclamation-mark" yaml:"subject-exclamation-mark"`
	SubjectFullStop        RuleValue[string]   `json:"subject-full-stop" yaml:"subject-full-stop"`
	SubjectMaxLength       RuleValue[int]      `json:"subject-max-length" yaml:"subject-max-length"`
	SubjectMinLength       RuleValue[int]      `json:"subject-min-length" yaml:"subject-min-length"`

	TrailerExists RuleValue[string] `json:"trailer-exists" yaml:"trailer-exists"`

	TypeCase      RuleValue[string]   `json:"type-case" yaml:"type-case"`
	TypeEmpty     RuleValue[bool]     `json:"type-empty" yaml:"type-empty"`
	TypeEnum      RuleValue[[]string] `json:"type-enum" yaml:"type-enum"`
	TypeMaxLength RuleValue[int]      `json:"type-max-length" yaml:"type-max-length"`
	TypeMinLength RuleValue[int]      `json:"type-min-length" yaml:"type-min-length"`
}

var DefaultRules = Rules{
	BodyCase:          RuleValue[string]{Disabled, Always, "lower-case", true},
	BodyEmpty:         RuleValue[bool]{Disabled, Never, false, true},
	BodyFullStop:      RuleValue[string]{Disabled, Never, ".", true},
	BodyLeadingBlank:  RuleValue[bool]{Disabled, Always, true, true},
	BodyMaxLength:     RuleValue[int]{Disabled, Always, math.MaxInt, true},
	BodyMaxLineLength: RuleValue[int]{Disabled, Always, math.MaxInt, true},
	BodyMinLength:     RuleValue[int]{Disabled, Always, 0, true},

	BreakingChangeExclamationMark: RuleValue[bool]{Disabled, Always, false, true},

	FooterEmpty:         RuleValue[bool]{Disabled, Never, false, true},
	FooterLeadingBlank:  RuleValue[bool]{Disabled, Never, false, true},
	FooterMaxLength:     RuleValue[int]{Disabled, Always, math.MaxInt, true},
	FooterMaxLineLength: RuleValue[int]{Disabled, Always, math.MaxInt, true},
	FooterMinLength:     RuleValue[int]{Disabled, Always, 0, true},

	HeaderCase:      RuleValue[string]{Disabled, Always, "lower-case", true},
	HeaderFullStop:  RuleValue[string]{Disabled, Never, ".", true},
	HeaderMaxLength: RuleValue[int]{Disabled, Always, math.MaxInt, true},
	HeaderMinLength: RuleValue[int]{Disabled, Always, 0, true},
	HeaderTrim:      RuleValue[bool]{Disabled, Always, true, true},

	ReferencesEmpty: RuleValue[bool]{Disabled, Never, false, true},

	ScopeCase:           RuleValue[string]{Disabled, Always, "lower-case", true},
	ScopeDelimiterStyle: RuleValue[[]string]{Disabled, Never, []string{`/`, `\`, `,`}, true},
	ScopeEmpty:          RuleValue[bool]{Disabled, Never, false, true},
	ScopeEnum:           RuleValue[[]string]{Disabled, Always, []string{}, true},
	ScopeMaxLength:      RuleValue[int]{Disabled, Always, math.MaxInt, true},
	ScopeMinLength:      RuleValue[int]{Disabled, Always, 0, true},

	SignedOfBy: RuleValue[string]{Disabled, Always, "Signed-off-by:", true},

	SubjectCase:            RuleValue[[]string]{Disabled, Never, []string{"sentence-case", "start-case", "pascal-case", "upper-case"}, true},
	SubjectEmpty:           RuleValue[bool]{Disabled, Never, false, true},
	SubjectExclamationMark: RuleValue[bool]{Disabled, Never, false, true},
	SubjectFullStop:        RuleValue[string]{Disabled, Never, ".", true},
	SubjectMaxLength:       RuleValue[int]{Disabled, Always, math.MaxInt, true},
	SubjectMinLength:       RuleValue[int]{Disabled, Always, 0, true},

	TrailerExists: RuleValue[string]{Disabled, Always, "Signed-off-by:", true},

	TypeCase:      RuleValue[string]{Disabled, Always, "lower-case", true},
	TypeEmpty:     RuleValue[bool]{Disabled, Never, false, true},
	TypeEnum:      RuleValue[[]string]{Disabled, Always, []string{"build", "chore", "ci", "docs", "feat", "fix", "perf", "refactor", "revert", "style", "test"}, true},
	TypeMaxLength: RuleValue[int]{Disabled, Always, math.MaxInt, true},
	TypeMinLength: RuleValue[int]{Disabled, Always, 0, true},
}
