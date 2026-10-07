package tui

import (
	"testing"

	"github.com/go-openapi/testify/v2/assert"

	"github.com/jaredreisinger/committed/pkg/commit"
	"github.com/jaredreisinger/committed/pkg/commitlint"
	"github.com/jaredreisinger/committed/pkg/config"
)

func TestNewModel(t *testing.T) {
	cfg := config.DefaultConfig()
	m := newModel(cfg, nil)

	assert.NotNil(t, m.config)
	assert.Equal(t, typeList, m.focusedField)
	assert.NotEmpty(t, m.config.Rules.TypeEnum.Value)
}

func TestNewModel_WithExistingMessage(t *testing.T) {
	cfg := &config.Config{
		Rules: commitlint.Rules{
			TypeEnum: commitlint.RuleValue[[]string]{
				Value: []string{"feat", "fix", "docs"},
				Set:   true,
			},
		},
	}
	existing := &commit.Message{
		Type:        "feat",
		Description: "add new feature",
		Body:        "detailed description",
	}

	m := newModel(cfg, existing)

	typ := m.children.MustGet(typeField).(textModel)
	desc := m.children.MustGet(descField).(textModel)
	body := m.children.MustGet(bodyField).(textModel)

	assert.Equal(t, "feat", typ.Value())
	assert.Equal(t, "add new feature", desc.Value())
	assert.Equal(t, "detailed description", body.Value())
}

// func TestValidateDescription(t *testing.T) {
// 	t.Skip("validation NYI")

// 	cfg := &config.Config{SubjectMaxLength: 50}
// 	m := newModel(cfg, nil)

// 	desc := m.children.MustGet(descriptionField).(textModel)

// 	// Empty description should fail
// 	desc.SetValue("")
// 	assert.Error(t, m.validateDescription())

// 	// Valid description should pass
// 	desc.SetValue("add new feature")
// 	assert.NoError(t, m.validateDescription())

// 	// Too long description should fail
// 	longDescription := string(make([]byte, 51))
// 	desc.SetValue(longDescription)
// 	assert.Error(t, m.validateDescription())
// }
