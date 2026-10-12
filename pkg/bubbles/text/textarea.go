package text

import (
	"log/slog"

	"charm.land/bubbles/v2/textarea"
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

// textareaModel shims and fixes [textarea.Model] to be a proper [tea.Model]
type textareaModel struct {
	textarea.Model
}

func newTextArea(placeholder string) textareaModel {
	m := textarea.New()
	m.Prompt = ""
	m.ShowLineNumbers = false
	m.Placeholder = placeholder
	m.DynamicHeight = true
	m.MinHeight = 10 //?

	m.SetStyles(textarea.Styles{
		Focused: textarea.StyleState{
			Base:        defaultStyles.Focused.Text,
			Text:        defaultStyles.Focused.Text,
			Placeholder: defaultStyles.Focused.Placeholder,
		},
		Blurred: textarea.StyleState{
			Base:        defaultStyles.Blurred.Text,
			Text:        defaultStyles.Blurred.Text,
			Placeholder: defaultStyles.Blurred.Placeholder,
		},
		Cursor: textarea.CursorStyle{
			Color: defaultStyles.Cursor.Color,
		},
	})

	return textareaModel{m}
}

// static compile-time check that the shim meets the interface
var _ tea.Model = textareaModel{}

func (t textareaModel) Init() tea.Cmd { return nil }

func (t textareaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	inner, cmd := t.Model.Update(msg)
	return textareaModel{inner}, cmd
}

func (t textareaModel) View() tea.View {
	// textarea doesn't always render up to its minimum height?
	view := t.Model.View()
	_, h := lipgloss.Size(view)
	if h < t.Model.MinHeight {
		slog.Warn("textarea rendering too short", "minHeight", t.Model.MinHeight, "renderedHeight", h, "view", view)
		view = lipgloss.NewStyle().PaddingBottom(t.Model.MinHeight - h).Render(view)
	} else {
		slog.Debug("textarea rendering OK", "minHeight", t.Model.MinHeight, "renderedHeight", h, "view", view)
	}
	return tea.NewView(view)
}

func (t textareaModel) KeyMap() textarea.KeyMap {
	return t.Model.KeyMap
}

func (t textareaModel) SetKeyMap(keys textarea.KeyMap) textareaModel {
	t2 := t.Model
	t2.KeyMap = keys
	return textareaModel{t2}
}
