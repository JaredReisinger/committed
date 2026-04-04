package tui

import (
	"testing"

	"github.com/go-openapi/testify/v2/assert"
	"github.com/jaredreisinger/committed/internal/config"
	"github.com/jaredreisinger/committed/pkg/commit"
)

func TestNewModel(t *testing.T) {
	cfg := config.DefaultConfig()
	m := newModel(cfg, nil)

	assert.NotNil(t, m.config)
	assert.Equal(t, typeField, m.focusedField)
	assert.NotEmpty(t, m.config.Types)
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

	assert.Equal(t, 0, m.typeIndex)
	assert.Equal(t, "add new feature", m.description.Value())
	assert.Equal(t, "detailed description", m.body.Value())
}

func TestValidateDescription(t *testing.T) {
	t.Skip("validation NYI")

	cfg := &config.Config{SubjectMaxLength: 50}
	m := newModel(cfg, nil)

	// Empty description should fail
	m.description.SetValue("")
	assert.Error(t, m.validateDescription())

	// Valid description should pass
	m.description.SetValue("add new feature")
	assert.NoError(t, m.validateDescription())

	// Too long description should fail
	longDescription := string(make([]byte, 51))
	m.description.SetValue(longDescription)
	assert.Error(t, m.validateDescription())
}
