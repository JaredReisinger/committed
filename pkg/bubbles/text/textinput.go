package text

import (
	"fmt"
	"log/slog"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Neither the textinput nor textarea bubbles properly implement tea.Model, so
// we shim them to fix them up.  It just so happens they are broken in the same
// way (no Init(), Update() returns the concrete model instead of the interface,
// and View() returns a string), but it's not clear we need a consolidated shim.
//
// We also define a New() function to make setting the initial values a little
// easier.

// textinputModel shims and fixes [textinput.Model] to be a proper [tea.Model].
// It also adds a "minimum width" feature which we need.
type textinputModel struct {
	textinput.Model
	minWidth int
}

func newTextInput(placeholder string) textinputModel {
	m := textinput.New()
	m.Prompt = ""
	m.Placeholder = placeholder

	m.SetStyles(textinput.Styles{
		Focused: textinput.StyleState{
			Text:        defaultTextStyles.Focused.Text,
			Placeholder: defaultTextStyles.Focused.Placeholder,
		},
		Blurred: textinput.StyleState{
			Text:        defaultTextStyles.Blurred.Text,
			Placeholder: defaultTextStyles.Blurred.Placeholder,
		},
		Cursor: textinput.CursorStyle{
			Color: defaultTextStyles.Cursor.Color,
		},
	})

	return textinputModel{m, len(placeholder)}
}

// static compile-time check that the shim meets the interface
var _ tea.Model = textinputModel{}

func (t textinputModel) Init() tea.Cmd { return nil }

func (t textinputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	inner, cmd := t.Model.Update(msg)
	return textinputModel{inner, t.minWidth}, cmd
}

func (t textinputModel) View() tea.View {
	origWidth := t.Model.Width()
	if origWidth < t.minWidth {
		slog.Debug("textinput temporarily bumping width",
			"from", origWidth,
			"to", t.minWidth,
			"content", fmt.Sprintf("[%s]", t.Value()),
			"placeholder", fmt.Sprintf("[%s]", t.Placeholder),
		)
		t.Model.SetWidth(t.minWidth)
	}
	view := t.Model.View()
	t.Model.SetWidth(origWidth) // t.Model isn't immutable, sadly

	// t.Model.View() *will* render an extra character for the cursor often, so
	// we check to see if it has focus or not, and potentially trim it.
	contentWidth := len(t.Model.Value())
	expectWidth := max(t.minWidth, contentWidth)
	if origWidth > 0 {
		expectWidth = min(expectWidth, origWidth)
	}
	viewWidth := lipgloss.Width(view)
	focused := t.Model.Focused()

	slog.Debug("checking textinput padding",
		"width", origWidth,
		"minWidth", t.minWidth,
		"contentWidth", contentWidth,
		"expectWidth", expectWidth,
		"view", fmt.Sprintf("[%s]", view),
		"viewWidth", viewWidth,
		"focused", focused,
	)
	if viewWidth > expectWidth && !t.Model.Focused() {
		// dead-reckon that we *always* trim one character...
		correctWidth := viewWidth - 1
		if expectWidth > 0 {
			correctWidth = expectWidth
		}
		view = lipgloss.NewStyle().Width(correctWidth).Render(view)
	} else if viewWidth < t.minWidth {
		view = lipgloss.NewStyle().PaddingRight(t.minWidth - viewWidth).Render(view)
	}

	return tea.NewView(view)
}

func (t textinputModel) MinWidth() int {
	return t.minWidth
}

func (t textinputModel) SetMinWidth(w int) textinputModel {
	// sort of like SetWidth, but does not clamp maximum
	return textinputModel{t.Model, w}
}

func (t textinputModel) KeyMap() textinput.KeyMap {
	return t.Model.KeyMap
}

func (t textinputModel) SetKeyMap(keys textinput.KeyMap) textinputModel {
	t2 := t.Model
	t2.KeyMap = keys
	return textinputModel{t2, t.minWidth}
}
