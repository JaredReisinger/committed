package tui

import (
	"testing"

	"github.com/jaredreisinger/committed/internal/config"
	"github.com/jaredreisinger/committed/pkg/commit"
)

func TestNewModel(t *testing.T) {
	cfg := config.DefaultConfig()
	m := newModel(cfg, nil)

	if m.config == nil {
		t.Error("expected config to be set")
	}
	if m.focusedField != typeField {
		t.Error("expected initial focus on type field")
	}
	if len(m.config.Types) == 0 {
		t.Error("expected default types to be loaded")
	}
}

func TestNewModel_WithExistingMessage(t *testing.T) {
	cfg := &config.Config{
		Types: []string{"feat", "fix", "docs"},
	}
	existing := &commit.Message{
		Type:        "feat",
		Description: "add new feature",
		Body:        "detailed description",
	}

	m := newModel(cfg, existing)

	if m.typeIndex != 0 { // "feat" should be at index 0
		t.Errorf("expected typeIndex 0 for 'feat', got %d", m.typeIndex)
	}
	if m.summary.Value() != "add new feature" {
		t.Errorf("expected summary to be pre-populated, got %q", m.summary.Value())
	}
	if m.details.Value() != "detailed description" {
		t.Errorf("expected details to be pre-populated, got %q", m.details.Value())
	}
}

func TestValidateSummary(t *testing.T) {
	t.Skip("validation NYI")

	cfg := &config.Config{SubjectMaxLength: 50}
	m := newModel(cfg, nil)

	// Empty summary should fail
	m.summary.SetValue("")
	if err := m.validateSummary(); err == nil {
		t.Error("expected error for empty summary")
	}

	// Valid summary should pass
	m.summary.SetValue("add new feature")
	if err := m.validateSummary(); err != nil {
		t.Errorf("expected no error for valid summary, got %v", err)
	}

	// Too long summary should fail
	longSummary := string(make([]byte, 51))
	m.summary.SetValue(longSummary)
	if err := m.validateSummary(); err == nil {
		t.Error("expected error for too long summary")
	}
}
