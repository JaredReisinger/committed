package commit

import (
	"strings"
)

// Message represents a parsed conventional commit message.
type Message struct {
	Type        string
	Scope       string
	Description string
	Body        string
	Footers     []Footer
	Breaking    bool
	RawMessage  string // not sure we need this
}

// Footer represents a footer in a conventional commit (e.g., "Closes #123").
type Footer struct {
	Token string
	Value string
}

// String returns the formatted conventional commit message.
func (m *Message) String() string {
	var b strings.Builder

	// Header: type(scope): description
	header := m.Type
	if m.Scope != "" {
		header += "(" + m.Scope + ")"
	}
	header += ": " + m.Description
	b.WriteString(header)

	// Body
	if m.Body != "" {
		b.WriteString("\n\n")
		b.WriteString(m.Body)
	}

	// Footers
	if len(m.Footers) > 0 {
		if m.Body == "" {
			b.WriteString("\n\n")
		} else {
			b.WriteString("\n\n")
		}
		for i, footer := range m.Footers {
			if i > 0 {
				b.WriteString("\n")
			}
			b.WriteString(footer.Token + ": " + footer.Value)
		}
	}

	b.WriteString("\n")

	return b.String()
}

// IsBreaking returns true if this commit introduces a breaking change.
func (m *Message) IsBreaking() bool {
	if strings.Contains(m.Description, "!") {
		return true
	}
	for _, footer := range m.Footers {
		if footer.Token == "BREAKING CHANGE" {
			return true
		}
	}
	return false
}
