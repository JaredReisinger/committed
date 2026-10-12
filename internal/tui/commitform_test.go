package tui

import (
	"testing"

	"github.com/go-openapi/testify/v2/assert"

	"github.com/jaredreisinger/committed/pkg/bubbles/text"
	"github.com/jaredreisinger/committed/pkg/commit"
	"github.com/jaredreisinger/committed/pkg/commitlint"
	"github.com/jaredreisinger/committed/pkg/config"
)

func TestNewModel(t *testing.T) {
	cfg := config.DefaultConfig()
	m := newCommitForm(cfg, nil)

	assert.NotNil(t, m.config)
	assert.Equal(t, typeList, m.focusedField)
	assert.NotEmpty(t, m.config.Rules.TypeEnum.Value)
}

func TestNewModel_WithExistingMessage(t *testing.T) {
	cfg := &config.Config{
		TypeEnum: commitlint.RuleValue[[]string]{
			Value: []string{"feat", "fix", "docs"},
			Set:   true,
		},
	}
	existing := &commit.Message{
		Type:        "feat",
		Description: "add new feature",
		Body:        "detailed description",
	}

	m := newCommitForm(cfg, existing)

	typ := m.children.MustGet(typeField).(text.Model)
	desc := m.children.MustGet(descField).(text.Model)
	body := m.children.MustGet(bodyField).(text.Model)

	assert.Equal(t, "feat", typ.Value())
	assert.Equal(t, "add new feature", desc.Value())
	assert.Equal(t, "detailed description", body.Value())
}
