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

type NoValue struct{}

func (r *RuleValue[T]) Active() bool {
	return r != nil && r.Set && r.Level != Disabled
}

func (r *RuleValue[T]) GetValue(defaultValue T) T {
	if r.Active() {
		return r.Value
	}
	return defaultValue
}

func (r *RuleValue[T]) Merge(other *RuleValue[T]) {
	if other != nil && other.Set {
		*r = *other
	}
}

// should we use nesting with merged parsing? (Body.Case, Body.Empty, ...)
type Rules struct {
	BodyCase          RuleValue[string]  `json:"body-case" yaml:"body-case"`
	BodyEmpty         RuleValue[NoValue] `json:"body-empty" yaml:"body-empty"`
	BodyFullStop      RuleValue[string]  `json:"body-full-stop" yaml:"body-full-stop"`
	BodyLeadingBlank  RuleValue[NoValue] `json:"body-leading-blank" yaml:"body-leading-blank"`
	BodyMaxLength     RuleValue[int]     `json:"body-max-length" yaml:"body-max-length"`
	BodyMaxLineLength RuleValue[int]     `json:"body-max-line-length" yaml:"body-max-line-length"`
	BodyMinLength     RuleValue[int]     `json:"body-min-length" yaml:"body-min-length"`

	BreakingChangeExclamationMark RuleValue[bool] `json:"breaking-change-exclamation-mark" yaml:"breaking-change-exclamation-mark"`

	FooterEmpty         RuleValue[NoValue] `json:"footer-empty" yaml:"footer-empty"`
	FooterLeadingBlank  RuleValue[NoValue] `json:"footer-leading-blank" yaml:"footer-leading-blank"`
	FooterMaxLength     RuleValue[int]     `json:"footer-max-length" yaml:"footer-max-length"`
	FooterMaxLineLength RuleValue[int]     `json:"footer-max-line-length" yaml:"footer-max-line-length"`
	FooterMinLength     RuleValue[int]     `json:"footer-min-length" yaml:"footer-min-length"`

	// header is the entire first line: `type(scope): description`
	HeaderCase      RuleValue[string]  `json:"header-case" yaml:"header-case"`
	HeaderFullStop  RuleValue[string]  `json:"header-full-stop" yaml:"header-full-stop"`
	HeaderMaxLength RuleValue[int]     `json:"header-max-length" yaml:"header-max-length"`
	HeaderMinLength RuleValue[int]     `json:"header-min-length" yaml:"header-min-length"`
	HeaderTrim      RuleValue[NoValue] `json:"header-trim" header:"header-trim"`

	ReferencesEmpty RuleValue[NoValue] `json:"references-empty" yaml:"references-empty"`

	ScopeCase           RuleValue[string]   `json:"scope-case" yaml:"scope-case"`
	ScopeDelimiterStyle RuleValue[[]string] `json:"scope-delimiter-style" yaml:"scope-delimiter-style"`
	ScopeEmpty          RuleValue[NoValue]  `json:"scope-empty" yaml:"scope-empty"`
	ScopeEnum           RuleValue[[]string] `json:"scope-enum" yaml:"scope-enum"`
	ScopeMaxLength      RuleValue[int]      `json:"scope-max-length" yaml:"scope-max-length"`
	ScopeMinLength      RuleValue[int]      `json:"scope-min-length" yaml:"scope-min-length"`

	SignedOfBy RuleValue[string] `json:"signed-off-by" yaml:"signed-off-by"`

	// "subject" is the "description"... no idea why they call it "subject" in
	// the rules
	SubjectCase            RuleValue[[]string] `json:"subject-case" yaml:"subject-case"`
	SubjectEmpty           RuleValue[NoValue]  `json:"subject-empty" yaml:"subject-empty"`
	SubjectExclamationMark RuleValue[NoValue]  `json:"subject-exclamation-mark" yaml:"subject-exclamation-mark"`
	SubjectFullStop        RuleValue[string]   `json:"subject-full-stop" yaml:"subject-full-stop"`
	SubjectMaxLength       RuleValue[int]      `json:"subject-max-length" yaml:"subject-max-length"`
	SubjectMinLength       RuleValue[int]      `json:"subject-min-length" yaml:"subject-min-length"`

	TrailerExists RuleValue[string] `json:"trailer-exists" yaml:"trailer-exists"`

	TypeCase      RuleValue[string]   `json:"type-case" yaml:"type-case"`
	TypeEmpty     RuleValue[NoValue]  `json:"type-empty" yaml:"type-empty"`
	TypeEnum      RuleValue[[]string] `json:"type-enum" yaml:"type-enum"`
	TypeMaxLength RuleValue[int]      `json:"type-max-length" yaml:"type-max-length"`
	TypeMinLength RuleValue[int]      `json:"type-min-length" yaml:"type-min-length"`
}

func (r *Rules) Merge(other *Rules) {
	r.BodyCase.Merge(&other.BodyCase)
	r.BodyEmpty.Merge(&other.BodyEmpty)
	r.BodyFullStop.Merge(&other.BodyFullStop)
	r.BodyLeadingBlank.Merge(&other.BodyLeadingBlank)
	r.BodyMaxLength.Merge(&other.BodyMaxLength)
	r.BodyMaxLineLength.Merge(&other.BodyMaxLineLength)
	r.BodyMinLength.Merge(&other.BodyMinLength)
	r.BreakingChangeExclamationMark.Merge(&other.BreakingChangeExclamationMark)
	r.FooterEmpty.Merge(&other.FooterEmpty)
	r.FooterLeadingBlank.Merge(&other.FooterLeadingBlank)
	r.FooterMaxLength.Merge(&other.FooterMaxLength)
	r.FooterMaxLineLength.Merge(&other.FooterMaxLineLength)
	r.FooterMinLength.Merge(&other.FooterMinLength)
	r.HeaderCase.Merge(&other.HeaderCase)
	r.HeaderFullStop.Merge(&other.HeaderFullStop)
	r.HeaderMaxLength.Merge(&other.HeaderMaxLength)
	r.HeaderMinLength.Merge(&other.HeaderMinLength)
	r.HeaderTrim.Merge(&other.HeaderTrim)
	r.ReferencesEmpty.Merge(&other.ReferencesEmpty)
	r.ScopeCase.Merge(&other.ScopeCase)
	r.ScopeDelimiterStyle.Merge(&other.ScopeDelimiterStyle)
	r.ScopeEmpty.Merge(&other.ScopeEmpty)
	r.ScopeEnum.Merge(&other.ScopeEnum)
	r.ScopeMaxLength.Merge(&other.ScopeMaxLength)
	r.ScopeMinLength.Merge(&other.ScopeMinLength)
	r.SignedOfBy.Merge(&other.SignedOfBy)
	r.SubjectCase.Merge(&other.SubjectCase)
	r.SubjectEmpty.Merge(&other.SubjectEmpty)
	r.SubjectExclamationMark.Merge(&other.SubjectExclamationMark)
	r.SubjectFullStop.Merge(&other.SubjectFullStop)
	r.SubjectMaxLength.Merge(&other.SubjectMaxLength)
	r.SubjectMinLength.Merge(&other.SubjectMinLength)
	r.TrailerExists.Merge(&other.TrailerExists)
	r.TypeCase.Merge(&other.TypeCase)
	r.TypeEmpty.Merge(&other.TypeEmpty)
	r.TypeEnum.Merge(&other.TypeEnum)
	r.TypeMaxLength.Merge(&other.TypeMaxLength)
	r.TypeMinLength.Merge(&other.TypeMinLength)
}

var DefaultRules = Rules{
	BodyCase:          RuleValue[string]{Disabled, Always, "lower-case", true},
	BodyEmpty:         RuleValue[NoValue]{Disabled, Never, NoValue{}, true},
	BodyFullStop:      RuleValue[string]{Disabled, Never, ".", true},
	BodyLeadingBlank:  RuleValue[NoValue]{Disabled, Always, NoValue{}, true},
	BodyMaxLength:     RuleValue[int]{Disabled, Always, math.MaxInt, true},
	BodyMaxLineLength: RuleValue[int]{Disabled, Always, math.MaxInt, true},
	BodyMinLength:     RuleValue[int]{Disabled, Always, 0, true},

	BreakingChangeExclamationMark: RuleValue[bool]{Disabled, Always, false, true},

	FooterEmpty:         RuleValue[NoValue]{Disabled, Never, NoValue{}, true},
	FooterLeadingBlank:  RuleValue[NoValue]{Disabled, Never, NoValue{}, true},
	FooterMaxLength:     RuleValue[int]{Disabled, Always, math.MaxInt, true},
	FooterMaxLineLength: RuleValue[int]{Disabled, Always, math.MaxInt, true},
	FooterMinLength:     RuleValue[int]{Disabled, Always, 0, true},

	HeaderCase:      RuleValue[string]{Disabled, Always, "lower-case", true},
	HeaderFullStop:  RuleValue[string]{Disabled, Never, ".", true},
	HeaderMaxLength: RuleValue[int]{Disabled, Always, math.MaxInt, true},
	HeaderMinLength: RuleValue[int]{Disabled, Always, 0, true},
	HeaderTrim:      RuleValue[NoValue]{Disabled, Always, NoValue{}, true},

	ReferencesEmpty: RuleValue[NoValue]{Disabled, Never, NoValue{}, true},

	ScopeCase:           RuleValue[string]{Disabled, Always, "lower-case", true},
	ScopeDelimiterStyle: RuleValue[[]string]{Disabled, Never, []string{`/`, `\`, `,`}, true},
	ScopeEmpty:          RuleValue[NoValue]{Disabled, Never, NoValue{}, true},
	ScopeEnum:           RuleValue[[]string]{Disabled, Always, []string{}, true},
	ScopeMaxLength:      RuleValue[int]{Disabled, Always, math.MaxInt, true},
	ScopeMinLength:      RuleValue[int]{Disabled, Always, 0, true},

	SignedOfBy: RuleValue[string]{Disabled, Always, "Signed-off-by:", true},

	SubjectCase:            RuleValue[[]string]{Disabled, Never, []string{"sentence-case", "start-case", "pascal-case", "upper-case"}, true},
	SubjectEmpty:           RuleValue[NoValue]{Disabled, Never, NoValue{}, true},
	SubjectExclamationMark: RuleValue[NoValue]{Disabled, Never, NoValue{}, true},
	SubjectFullStop:        RuleValue[string]{Disabled, Never, ".", true},
	SubjectMaxLength:       RuleValue[int]{Disabled, Always, math.MaxInt, true},
	SubjectMinLength:       RuleValue[int]{Disabled, Always, 0, true},

	TrailerExists: RuleValue[string]{Disabled, Always, "Signed-off-by:", true},

	TypeCase:      RuleValue[string]{Disabled, Always, "lower-case", true},
	TypeEmpty:     RuleValue[NoValue]{Disabled, Never, NoValue{}, true},
	TypeEnum:      RuleValue[[]string]{Disabled, Always, []string{"build", "chore", "ci", "docs", "feat", "fix", "perf", "refactor", "revert", "style", "test"}, true},
	TypeMaxLength: RuleValue[int]{Disabled, Always, math.MaxInt, true},
	TypeMinLength: RuleValue[int]{Disabled, Always, 0, true},
}

// @commitlint/config-conventional -- as of 2026-09-28 -- in the order defined
// in https://github.com/conventional-changelog/commitlint/blob/master/...
// %40commitlint/config-conventional/src/index.ts to make it easier to compare
var ConventionalConfigRules = Rules{
	BodyLeadingBlank:    RuleValue[NoValue]{Warning, Always, NoValue{}, true},
	BodyMaxLineLength:   RuleValue[int]{Error, Always, 100, true},
	FooterLeadingBlank:  RuleValue[NoValue]{Warning, Always, NoValue{}, true},
	FooterMaxLineLength: RuleValue[int]{Error, Always, 100, true},
	HeaderMaxLength:     RuleValue[int]{Error, Always, 100, true},
	HeaderTrim:          RuleValue[NoValue]{Error, Always, NoValue{}, true},
	SubjectCase:         RuleValue[[]string]{Error, Never, []string{"sentence-case", "start-case", "pascal-case", "upper-case"}, true},
	SubjectEmpty:        RuleValue[NoValue]{Error, Never, NoValue{}, true},
	SubjectFullStop:     RuleValue[string]{Error, Never, ".", true},
	TypeCase:            RuleValue[string]{Error, Always, "lower-case", true},
	TypeEmpty:           RuleValue[NoValue]{Error, Never, NoValue{}, true},
	TypeEnum:            RuleValue[[]string]{Error, Always, []string{"build", "chore", "ci", "docs", "feat", "fix", "perf", "refactor", "revert", "style", "test"}, true},
}

// @commitlint/config-conventional -- as of 2026-09-28 -- in the order defined
// in https://github.com/conventional-changelog/commitlint/blob/master/...
// %40commitlint/config-angular/src/index.ts to make it easier to compare
var AngularConfigRules = Rules{
	SubjectExclamationMark: RuleValue[NoValue]{Error, Never, NoValue{}, true},
	BodyLeadingBlank:       RuleValue[NoValue]{Warning, Always, NoValue{}, true},
	FooterLeadingBlank:     RuleValue[NoValue]{Warning, Always, NoValue{}, true},
	HeaderMaxLength:        RuleValue[int]{Error, Always, 72, true},
	ScopeCase:              RuleValue[string]{Error, Always, "lower-case", true},
	SubjectCase:            RuleValue[[]string]{Error, Never, []string{"sentence-case", "start-case", "pascal-case", "upper-case"}, true},
	SubjectEmpty:           RuleValue[NoValue]{Error, Never, NoValue{}, true},
	SubjectFullStop:        RuleValue[string]{Error, Never, ".", true},
	TypeCase:               RuleValue[string]{Error, Always, "lower-case", true},
	TypeEmpty:              RuleValue[NoValue]{Error, Never, NoValue{}, true},
	TypeEnum:               RuleValue[[]string]{Error, Always, []string{"build", "ci", "docs", "feat", "fix", "perf", "refactor", "revert", "style", "test"}, true},
}
