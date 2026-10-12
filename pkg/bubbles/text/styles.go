package text

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// TODO: we can choose our colors and use lighten/darken to create the blurred
// values.

// We build styles up from basic colors, and then aggregate into a structure
// that is particularly textinput/textarea-friendly. (We *could* decide to make
// our textModel wrapper handle that, which might make user-customization
// easier.)

// unified style support
type Styles struct {
	Focused PartStyles
	Blurred PartStyles
	Cursor  CursorStyle
}

type PartStyles struct {
	Text        lipgloss.Style
	Placeholder lipgloss.Style
	// Prompt      lipgloss.Style // we never use prompt

	// textinput ignored styles...

	// Suggestion lipgloss.Style

	// textarea ignored styles...

	// Base             lipgloss.Style // do we need this textarea style?
	// LineNumber       lipgloss.Style
	// CursorLineNumber lipgloss.Style
	// CursorLine       lipgloss.Style
	// EndOfBuffer      lipgloss.Style
}

type CursorStyle struct {
	Color color.Color
}

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

// colors
var (
	placeholderColor = lipgloss.Color("238")

	defaultColors = map[state]map[role]color.Color{
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
	}
)

var (
	placeholderStyle = lipgloss.NewStyle().Foreground(defaultColors[focused][placeholder]).Italic(true)

	defaultTextStyles = Styles{
		Focused: PartStyles{
			Text:        lipgloss.NewStyle().Foreground(defaultColors[focused][foreground]),
			Placeholder: placeholderStyle,
		},
		Blurred: PartStyles{
			Text:        lipgloss.NewStyle().Foreground(defaultColors[blurred][foreground]),
			Placeholder: placeholderStyle,
		},
		Cursor: CursorStyle{
			Color: defaultColors[focused][cursor],
		},
	}
)
