package commit

import (
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"unicode"

	"github.com/jaredreisinger/committed/pkg/config"
)

// Message represents a parsed conventional commit message.
type Message struct {
	Type        string
	Scope       string
	Description string
	Body        string
	Footer      string
	Footers     []Footer // ???
	Breaking    bool
	RawMessage  string // not sure we need this
}

// Footer represents a footer in a conventional commit (e.g., "Closes #123").
type Footer struct {
	Token string
	Value string
}

// Format returns the formatted conventional commit message.
func (m *Message) Format(cfg *config.Config) string {
	var header strings.Builder
	var body strings.Builder
	var footer strings.Builder

	// Header: type(scope)[!]: description
	if !slices.Contains(cfg.Types, m.Type) {
		// Should we prevent exiting the editor in this case?
		slog.Warn("type is not permitted", "type", m.Type, "permitted", cfg.Types)
	}
	header.WriteString(m.Type)

	if m.Scope != "" {
		header.WriteString("(")
		header.WriteString(m.Scope)
		header.WriteString(")")
	}

	// TODO: check for breaking change to add "!"
	header.WriteString(": ")

	// TODO: should have a config for this?
	if len(m.Description) < 3 {
		slog.Warn("header description is too short", "min", 3, "actual", len(m.Description), "text", m.Description)
	}
	header.WriteString(m.Description)

	if header.Len() > cfg.HeaderMaxLength {
		slog.Warn("header line exceeds max length", "max", cfg.HeaderMaxLength, "actual", header.Len(), "text", header.String())
	}

	// Body -- the content coming out of the TUI is *not* wrapped; we need to do
	// so at the configured line-length
	if m.Body != "" {
		if cfg.BodyMaxLineLength > 0 {
			body.WriteString(wrap(m.Body, cfg.BodyMaxLineLength))
		}
	}

	// Footers
	if len(m.Footers) > 0 {
		// Do we need to handle wrapping here?
		for _, f := range m.Footers {
			if footer.Len() > 0 {
				footer.WriteRune('\n')
			}
			footer.WriteString(fmt.Sprintf("%s: %s", f.Token, f.Value))
		}
	} else if m.Footer != "" {
		footer.WriteString(m.Footer)
	}

	var message strings.Builder
	message.WriteString(header.String())

	if body.Len() > 0 {
		message.WriteString("\n\n")
		message.WriteString(body.String())
	}

	if footer.Len() > 0 {
		message.WriteString("\n\n")
		message.WriteString(footer.String())
	}

	// always end with newline (do we need to see if there already was one?)
	message.WriteRune('\n')

	return message.String()
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

// wrap performs word-wrapping. It's naive for now, but might add "nice"
// behavior like indenting list items (a la markdown) or hyphenation in the
// future.
func wrap(s string, limit int) string {
	// We run through the list, keeping track of the last whitespace we saw. We
	// do *not* compress consecutive whitespace into a single space, there are
	// legitimate cases where the author has either leading whitespace or is
	// aligning columns. Pathologically, one could have consecutive spaces that
	// cross a wrap boundary, but that seems very unlikely. In any case, we'll
	// wrap at the "expected" place, with one space getting eaten by the wrap
	// itself.  (It's an open issue if we want to eat all spaces after the
	// wrap... the case of "2 spaces after a sentence end" is one where we'd
	// rather *not* use the second space at the beginning of a line.)
	var word strings.Builder
	var line strings.Builder
	var out strings.Builder
	var lastSpace rune
	lastSpaceLen := 0

	for _, c := range s {
		if unicode.IsSpace(c) {
			// if the current word *does* fit on the line, include the last
			// whitespace and the word and keep accumulating (but look for hard
			// newlines!)
			fits := line.Len()+lastSpaceLen+word.Len() < limit
			if fits {
				if lastSpaceLen > 0 {
					line.WriteRune(lastSpace)
				}
				line.WriteString(word.String())
				word.Reset()
				lastSpace = c
				lastSpaceLen = 1
			}

			// If it didn't fit, *or* if this was a newline, go ahead and send
			// the line to the output builder.
			if !fits || c == '\n' {
				out.WriteString(line.String())
				out.WriteRune('\n')
				line.Reset()
				lastSpaceLen = 0
			}
		} else {
			word.WriteRune(c)
		}
	}

	// and a final check for any trailing word/line...
	if word.Len() > 0 || line.Len() > 0 {
		fits := line.Len()+lastSpaceLen+word.Len() < limit
		if fits {
			if lastSpaceLen > 0 {
				line.WriteRune(lastSpace)
			}
			line.WriteString(word.String())
			word.Reset()
		}

		if line.Len() > 0 {
			out.WriteString(line.String())
		}

		if word.Len() > 0 {
			out.WriteRune('\n')
			out.WriteString(word.String())
		}
	}

	return out.String()
}
