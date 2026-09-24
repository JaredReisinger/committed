package tui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// TODO: we can choose our colors and use lighten/darken to create the blurred
// values.

// We build styles up from basic colors, and then aggregate into a structure
// that is particularly textinput/textarea-friendly. (We *could* decide to make
// our textmodel wrapper handle that, which might make user-customization
// easier.)

// do we need a light/dark axis, too?

type role int

const (
	foreground role = iota
	// background
	placeholder
	decoration
	cursor
)

type state int

const (
	blurred state = iota
	focused
)

type status int

const (
	defaultStatus status = iota
	invalid
	valid
)

// colors
var (
	placeholderColor = lipgloss.Color("238")

	defaultColors = map[status]map[state]map[role]color.Color{
		defaultStatus: {
			blurred: {
				foreground:  lipgloss.Color("244"), // #808080
				placeholder: placeholderColor,
				decoration:  lipgloss.Color("238"), // #444444
			},
			focused: {
				foreground:  lipgloss.Color("254"), // #e4e4e4
				placeholder: placeholderColor,
				decoration:  lipgloss.Color("27"), // #005fff
				cursor:      lipgloss.Color("39"), // #00afff
			},
		},
		invalid: {
			blurred: {
				foreground:  lipgloss.Color("167"), // #d75f5f
				placeholder: placeholderColor,
				decoration:  lipgloss.Color("88"), // #870000
			},
			focused: {
				foreground:  lipgloss.Color("217"), // #ffafaf
				placeholder: placeholderColor,
				decoration:  lipgloss.Color("196"), // #ff0000
				cursor:      lipgloss.Color("217"), // #ffafaf
			},
		},
		valid: {
			blurred: {
				foreground:  lipgloss.Color("77"), // #5fd75f
				placeholder: placeholderColor,
				decoration:  lipgloss.Color("28"), // #008700
			},
			focused: {
				foreground:  lipgloss.Color("157"), // #afffaf
				placeholder: placeholderColor,
				decoration:  lipgloss.Color("118"), // #87ff00
				cursor:      lipgloss.Color("157"), // #afffaf
			},
		},
	}
)

// decoration
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

	decorationFocusColor = defaultColors[defaultStatus][focused][decoration]
	decorationBlurColor  = defaultColors[defaultStatus][blurred][decoration]

	placeholderStyle = lipgloss.NewStyle().Foreground(defaultColors[defaultStatus][focused][placeholder]).Italic(true)

	defaultTextStyles = textStyles{
		Focused: textPartStyles{
			Text:        lipgloss.NewStyle().Foreground(defaultColors[defaultStatus][focused][foreground]),
			Placeholder: placeholderStyle,
		},
		Blurred: textPartStyles{
			Text:        lipgloss.NewStyle().Foreground(defaultColors[defaultStatus][blurred][foreground]),
			Placeholder: placeholderStyle,
		},
		Cursor: textCursorStyle{
			Color: defaultColors[defaultStatus][focused][cursor],
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
