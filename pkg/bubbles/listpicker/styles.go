package listpicker

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
}

type PartStyles struct {
	Text lipgloss.Style
}

// do we need a light/dark axis, too?

type state int

const (
	blurred state = iota
	focused
)

// colors
var (
	defaultColors = map[state]color.Color{
		blurred: lipgloss.Color("244"), // #808080
		focused: lipgloss.Color("254"), // #e4e4e4
	}
)

var (
	defaultStyles = Styles{
		Focused: PartStyles{
			Text: lipgloss.NewStyle().Foreground(defaultColors[focused]),
		},
		Blurred: PartStyles{
			Text: lipgloss.NewStyle().Foreground(defaultColors[blurred]),
		},
	}
)
