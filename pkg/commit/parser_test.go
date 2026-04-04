package commit

import (
	"strings"
	"testing"
)

func TestParseMessage_ValidConventional(t *testing.T) {
	input := "feat(api): add user authentication\n\n- Add JWT token validation\n- Implement login endpoint\n\nCloses: #123"
	msg, err := ParseMessage(input)
	if err != nil {
		t.Fatal(err)
	}

	if msg.Type != "feat" {
		t.Errorf("expected type 'feat', got %q", msg.Type)
	}
	if msg.Scope != "api" {
		t.Errorf("expected scope 'api', got %q", msg.Scope)
	}
	if msg.Description != "add user authentication" {
		t.Errorf("expected description 'add user authentication', got %q", msg.Description)
	}
	if !strings.Contains(msg.Body, "Add JWT token validation") {
		t.Errorf("expected body to contain JWT validation, got %q", msg.Body)
	}
	if len(msg.Footers) != 1 || msg.Footers[0].Token != "Closes" || msg.Footers[0].Value != "#123" {
		t.Errorf("expected one footer 'Closes: #123', got %#v", msg.Footers)
	}
}

func TestParseMessage_NoScope(t *testing.T) {
	input := "fix: resolve null pointer exception"
	msg, err := ParseMessage(input)
	if err != nil {
		t.Fatal(err)
	}

	if msg.Type != "fix" {
		t.Errorf("expected type 'fix', got %q", msg.Type)
	}
	if msg.Scope != "" {
		t.Errorf("expected empty scope, got %q", msg.Scope)
	}
	if msg.Description != "resolve null pointer exception" {
		t.Errorf("expected description 'resolve null pointer exception', got %q", msg.Description)
	}
}

func TestParseMessage_WithBreakingChange(t *testing.T) {
	input := "feat!: add new breaking API\n\nBREAKING CHANGE: The old API is deprecated"
	msg, err := ParseMessage(input)
	if err != nil {
		t.Fatal(err)
	}

	if !msg.IsBreaking() {
		t.Error("expected message to be marked as breaking")
	}
}

func TestParseMessage_MalformedHeader(t *testing.T) {
	input := "this is not a conventional commit message"
	msg, err := ParseMessage(input)
	if err != nil {
		t.Fatal(err)
	}

	// Should treat entire message as description
	if msg.Description != input {
		t.Errorf("expected description to be entire message, got %q", msg.Description)
	}
	if msg.Type != "" {
		t.Errorf("expected empty type for malformed message, got %q", msg.Type)
	}
}

func TestParseMessage_EmptyMessage(t *testing.T) {
	msg, err := ParseMessage("")
	if err != nil {
		t.Fatal(err)
	}

	if msg.RawMessage != "" {
		t.Errorf("expected raw message to be empty, got %q", msg.RawMessage)
	}
}

func TestParseMessage_MultipleFooters(t *testing.T) {
	input := "feat: add feature\n\nSome body\n\nCloses: #1\nFixes: #2\nBREAKING CHANGE: This breaks stuff"
	msg, err := ParseMessage(input)
	if err != nil {
		t.Fatal(err)
	}

	if len(msg.Footers) != 3 {
		t.Errorf("expected 3 footers, got %d", len(msg.Footers))
	}

	expectedFooters := []Footer{
		{Token: "Closes", Value: "#1"},
		{Token: "Fixes", Value: "#2"},
		{Token: "BREAKING CHANGE", Value: "This breaks stuff"},
	}

	for i, expected := range expectedFooters {
		if i >= len(msg.Footers) {
			t.Errorf("missing footer %d: %v", i, expected)
			continue
		}
		actual := msg.Footers[i]
		if actual.Token != expected.Token || actual.Value != expected.Value {
			t.Errorf("footer %d: expected %v, got %v", i, expected, actual)
		}
	}
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

	if actual != expected {
		t.Errorf("expected:\n%q\ngot:\n%q", expected, actual)
	}
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
			if tt.message.IsBreaking() != tt.expected {
				t.Errorf("expected IsBreaking()=%v, got %v", tt.expected, tt.message.IsBreaking())
			}
		})
	}
}
