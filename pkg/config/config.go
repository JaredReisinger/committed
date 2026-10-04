package config

import "github.com/jaredreisinger/committed/pkg/commitlint"

// REVIEW: Do we want a structured set of data that mirrors the commitlint
// validation configuration info?  Or do we just want something with a set of
// field validators?
//
// I'd set up validation to happen only on content change, but it turns out that
// bubbletea is pretty smart about only rendering (calling View) when something
// has changed.  Still, though, it seems better to control timing directly.

// Config contains the conventional commit conventions we use for TUI
// validation.
type Config struct {
	Extends string `json:"extends" yaml:"extends"`

	// Should this be a superset of commitlint rules?
	commitlint.Rules
}

// DefaultConfig is based on @commitlint/config-conventional values.
func DefaultConfig() *Config {
	return &Config{
		// Rules: commitlint.DefaultRules,
		Rules: commitlint.ConventionalConfigRules,
	}
}
