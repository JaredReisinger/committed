package config

// Config contains the conventional commit conventions we use for TUI
// validation.
type Config struct {
	Types             []string `json:"types" yaml:"types"`
	Scopes            []string `json:"scopes" yaml:"scopes"`
	SubjectMaxLength  int      `json:"subjectMaxLength" yaml:"subjectMaxLength"`
	BodyMaxLineLength int      `json:"bodyMaxLineLength" yaml:"bodyMaxLineLength"`
	HeaderMaxLength   int      `json:"headerMaxLength" yaml:"headerMaxLength"`
}

// DefaultConfig is based on @commitlint/config-conventional values.
func DefaultConfig() *Config {
	return &Config{
		Types: []string{ // ordered (not alphabetical!) based on common usage(?)
			"feat",
			"fix",
			"docs",
			"style",
			"refactor",
			"perf",
			"test",
			"chore",
			"revert",
			"build",
			"ci",
		},
		Scopes: []string{},
		// actual values...
		SubjectMaxLength:  100,
		BodyMaxLineLength: 100,
		HeaderMaxLength:   100,

		// // test values...
		// SubjectMaxLength:  50,
		// BodyMaxLineLength: 50,
		// HeaderMaxLength:   50,
	}
}
