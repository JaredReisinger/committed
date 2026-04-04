package tui

import (
	"charm.land/bubbles/v2/textarea"
	"charm.land/lipgloss/v2"
)

// TODO: we can choose our colors and use lighten/darken to create the blurred
// values.

var (
	underBorder = lipgloss.Border{
		Top:          " ",
		Bottom:       "─",
		Left:         " ",
		Right:        " ",
		TopLeft:      " ",
		TopRight:     " ",
		BottomLeft:   "╶",
		BottomRight:  "╴",
		MiddleLeft:   " ",
		MiddleRight:  " ",
		Middle:       " ",
		MiddleTop:    " ",
		MiddleBottom: " ",
	}

	placeholder = lipgloss.NewStyle().
			Foreground(lipgloss.Color("232"))
)

// The "fullBorder" style *always* shows the border, whereas the "under only"
// (fullBorder=false) disappears when blurred.
func textareaStyles(fullBorder bool) textarea.Styles {
	focusBorder := underBorder
	blurBorder := lipgloss.HiddenBorder()
	if fullBorder {
		focusBorder = lipgloss.NormalBorder()
		blurBorder = lipgloss.NormalBorder()
	}

	// There also appears to be a bug, no doubt related to the *frame* only
	// relying on Base and not having a dedicated style; if there is no content,
	// another border is drawn on the inside.  If I can figure out what style is
	// being used, setting it should fix the problem. Maybe end-of-buffer? Nope!
	return textarea.Styles{
		Focused: textarea.StyleState{
			Base: lipgloss.NewStyle().
				Border(focusBorder).
				BorderForeground(lipgloss.Color("201")),
			Text:        lipgloss.NewStyle().Foreground(lipgloss.Color("250")),
			CursorLine:  lipgloss.NewStyle().Foreground(lipgloss.Color("255")),
			Placeholder: placeholder,
		},
		Blurred: textarea.StyleState{
			Base: lipgloss.NewStyle().
				Border(blurBorder).
				BorderForeground(lipgloss.Color("250")),
			Text:        lipgloss.NewStyle().Foreground(lipgloss.Color("232")),
			Placeholder: placeholder,
		},
		Cursor: textarea.CursorStyle{
			Color: lipgloss.Color("#900090"),
			// Shape: tea.CursorUnderline, // shape does not seem to be passed
			// Blink: true,
		},
	}
}
