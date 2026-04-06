package tui

import (
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

	// focusedTextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))

	placeholderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("232"))

	defaultTextStyles = textStyles{
		Focused: textPartStyles{
			Text:        lipgloss.NewStyle().Foreground(lipgloss.Color("250")),
			Placeholder: placeholderStyle,
		},
		Blurred: textPartStyles{
			Text:        lipgloss.NewStyle().Foreground(lipgloss.Color("232")),
			Placeholder: placeholderStyle,
		},
		Cursor: textCursorStyle{
			Color: lipgloss.Color("#900090"),
		},
	}
)
