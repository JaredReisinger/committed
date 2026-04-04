package commit

import (
	"testing"

	"github.com/go-openapi/testify/v2/assert"
)

func TestParseMessage_ValidConventional(t *testing.T) {
	input := "feat(api): add user authentication\n\n- Add JWT token validation\n- Implement login endpoint\n\nCloses: #123"
	msg, err := ParseMessage(input)
	assert.NoError(t, err)

	assert.Equal(t, "feat", msg.Type)
	assert.Equal(t, "api", msg.Scope)
	assert.Equal(t, "add user authentication", msg.Description)
	assert.Contains(t, msg.Body, "Add JWT token validation")
	assert.Equal(t, []Footer{{Token: "Closes", Value: "#123"}}, msg.Footers)
	assert.Equal(t, false, msg.Breaking)
}

func TestParseMessage_NoScope(t *testing.T) {
	input := "fix: resolve null pointer exception"
	msg, err := ParseMessage(input)
	assert.NoError(t, err)

	assert.Equal(t, "fix", msg.Type)
	assert.Equal(t, "", msg.Scope)
	assert.Equal(t, "resolve null pointer exception", msg.Description)
	assert.Equal(t, "", msg.Body)
}

func TestParseMessage_WithBreakingChange(t *testing.T) {
	input := "feat!: add new breaking API\n\nBREAKING CHANGE: The old API is deprecated"
	msg, err := ParseMessage(input)
	assert.NoError(t, err)

	assert.True(t, msg.Breaking)
}

func TestParseMessage_MalformedHeader(t *testing.T) {
	input := "this is not a conventional commit message"
	msg, err := ParseMessage(input)
	assert.NoError(t, err)

	// Should treat entire message as description
	assert.Equal(t, input, msg.Body)
	assert.Equal(t, "", msg.Type)
	assert.Equal(t, "", msg.Description)
}

func TestParseMessage_EmptyMessage(t *testing.T) {
	msg, err := ParseMessage("")
	assert.NoError(t, err)
	assert.Equal(t, "", msg.RawMessage)
}

func TestParseMessage_MultipleFooters(t *testing.T) {
	input := "feat: add feature\n\nSome body\n\nCloses: #1\nFixes: #2\nBREAKING CHANGE: This breaks stuff"
	msg, err := ParseMessage(input)
	assert.NoError(t, err)

	expectedFooters := []Footer{
		{Token: "Closes", Value: "#1"},
		{Token: "Fixes", Value: "#2"},
		{Token: "BREAKING CHANGE", Value: "This breaks stuff"},
	}

	assert.Equal(t, expectedFooters, msg.Footers)
}

func TestMessage_String(t *testing.T) {
	msg := &Message{
		Type:        "feat",
		Scope:       "api",
		Description: "add user auth",
		Body:        "Some details here",
		Footers: []Footer{
			{Token: "Closes", Value: "#123"},
		},
	}

	expected := "feat(api): add user auth\n\nSome details here\n\nCloses: #123\n"
	actual := msg.String()

	assert.Equal(t, expected, actual)
}

func TestMessage_IsBreaking(t *testing.T) {
	tests := []struct {
		desc     string
		message  *Message
		expected bool
	}{
		{
			desc: "breaking change in description",
			message: &Message{
				Description: "add feature!",
			},
			expected: true,
		},
		{
			desc: "breaking change footer",
			message: &Message{
				Footers: []Footer{
					{Token: "BREAKING CHANGE", Value: "stuff"},
				},
			},
			expected: true,
		},
		{
			desc: "not breaking",
			message: &Message{
				Description: "add feature",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.message.IsBreaking())
		})
	}
}
