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

	decorationFocusColor = lipgloss.Color("201")
	decorationBlurColor  = lipgloss.Color("238")

	textPlaceholderColor = lipgloss.Color("238")
	textBlurColor        = lipgloss.Color("244")
	textFocusColor       = lipgloss.Color("250")

	placeholderStyle = lipgloss.NewStyle().Foreground(textPlaceholderColor).Italic(true)

	defaultTextStyles = textStyles{
		Focused: textPartStyles{
			Text:        lipgloss.NewStyle().Foreground(textFocusColor),
			Placeholder: placeholderStyle,
		},
		Blurred: textPartStyles{
			Text:        lipgloss.NewStyle().Foreground(textBlurColor),
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

	focusSingle = singleDecoration.BorderForeground(decorationFocusColor)
	blurSingle  = singleDecoration.BorderForeground(decorationBlurColor)
	focusArea   = areaDecoration.BorderForeground(decorationFocusColor)
	blurArea    = areaDecoration.BorderForeground(decorationBlurColor)
)
