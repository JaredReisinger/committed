package tui

import (
	"charm.land/lipgloss/v2"
)

// TODO: we can choose our colors and use lighten/darken to create the blurred
// values.

var (
	underBorder = lipgloss.Border{
		// Top:          " ",
		Bottom: "─",
		// Left:         " ",
		// Right:        " ",
		// TopLeft:      " ",
		// TopRight:     " ",
		BottomLeft:  "╶",
		BottomRight: "╴",
		// MiddleLeft:   " ",
		// MiddleRight:  " ",
		// Middle:       " ",
		// MiddleTop:    " ",
		// MiddleBottom: " ",
	}

	// bodyBorder =

	// focusedTextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))

	placeholderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("238")).
				Italic(true)

	defaultTextStyles = textStyles{
		Focused: textPartStyles{
			Text:        lipgloss.NewStyle().Foreground(lipgloss.Color("250")),
			Placeholder: placeholderStyle,
		},
		Blurred: textPartStyles{
			Text:        lipgloss.NewStyle().Foreground(lipgloss.Color("244")),
			Placeholder: placeholderStyle,
		},
		Cursor: textCursorStyle{
			Color: lipgloss.Color("147"),
		},
	}

	singleDecoration = lipgloss.NewStyle().
				Border(underBorder, false, false, true, false)

	areaDecoration = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, true, true, true)

	focusColor = lipgloss.Color("201")
	blurColor  = lipgloss.Color("238")

	focusSingle = singleDecoration.BorderForeground(focusColor)
	blurSingle  = singleDecoration.BorderForeground(blurColor)
	focusArea   = areaDecoration.BorderForeground(focusColor)
	blurArea    = areaDecoration.BorderForeground(blurColor)
)
