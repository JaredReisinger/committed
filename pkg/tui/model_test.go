package tui

import (
	"testing"

	"github.com/jaredreisinger/committed/internal/config"
	"github.com/jaredreisinger/committed/pkg/commit"
)

func TestNewModel(t *testing.T) {
	cfg := config.DefaultConfig()
	model := NewModel(cfg, nil)

	if model.config == nil {
		t.Error("expected config to be set")
	}
	if model.focusedField != typeField {
		t.Error("expected initial focus on type field")
	}
	if len(model.config.Types) == 0 {
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

	model := NewModel(cfg, existing)

	if model.typeIndex != 0 { // "feat" should be at index 0
		t.Errorf("expected typeIndex 0 for 'feat', got %d", model.typeIndex)
	}
	if model.summaryInput.Value() != "add new feature" {
		t.Errorf("expected summary to be pre-populated, got %q", model.summaryInput.Value())
	}
	if model.detailsInput.Value() != "detailed description" {
		t.Errorf("expected details to be pre-populated, got %q", model.detailsInput.Value())
	}
}

func TestValidateSummary(t *testing.T) {
	cfg := &config.Config{SubjectMaxLength: 50}
	model := NewModel(cfg, nil)

	// Empty summary should fail
	model.summaryInput.SetValue("")
	if err := model.validateSummary(); err == nil {
		t.Error("expected error for empty summary")
	}

	// Valid summary should pass
	model.summaryInput.SetValue("add new feature")
	if err := model.validateSummary(); err != nil {
		t.Errorf("expected no error for valid summary, got %v", err)
	}

	// Too long summary should fail
	longSummary := string(make([]byte, 51))
	model.summaryInput.SetValue(longSummary)
	if err := model.validateSummary(); err == nil {
		t.Error("expected error for too long summary")
	}
}
