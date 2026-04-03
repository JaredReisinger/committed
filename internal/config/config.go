package config

// Config contains the conventional commit conventions we use for TUI validation.
type Config struct {
	Types             []string `json:"types" yaml:"types"`
	Scopes            []string `json:"scopes" yaml:"scopes"`
	SubjectMaxLength  int      `json:"subjectMaxLength" yaml:"subjectMaxLength"`
	BodyMaxLineLength int      `json:"bodyMaxLineLength" yaml:"bodyMaxLineLength"`
	HeaderMaxLength   int      `json:"headerMaxLength" yaml:"headerMaxLength"`
}

func DefaultConfig() *Config {
	return &Config{
		Types:             []string{"feat", "fix", "docs", "style", "refactor", "perf", "test", "chore", "revert"},
		Scopes:            []string{},
		SubjectMaxLength:  72,
		BodyMaxLineLength: 100,
		HeaderMaxLength:   72,
	}
}
