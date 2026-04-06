package tui

import (
	"image/color"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
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
//
// Can we define a common base? It'd be handy!

// ===== textinput ============================================================

// textinputModel shims and fixes [textinput.Model] to be a proper [tea.Model]
type textinputModel struct {
	textinput.Model
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

	return textinputModel{m}
}

// static compile-time check that the shim meets the interface
var _ tea.Model = textinputModel{}

func (t textinputModel) Init() tea.Cmd { return nil }

func (t textinputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	inner, cmd := t.Model.Update(msg)
	return textinputModel{inner}, cmd
}

func (t textinputModel) View() tea.View {
	return tea.NewView(t.Model.View())
}

// ===== textarea =============================================================

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
			Base:        defaultTextStyles.Focused.Text,
			Text:        defaultTextStyles.Focused.Text,
			Placeholder: defaultTextStyles.Focused.Placeholder,
		},
		Blurred: textarea.StyleState{
			Base:        defaultTextStyles.Blurred.Text,
			Text:        defaultTextStyles.Blurred.Text,
			Placeholder: defaultTextStyles.Blurred.Placeholder,
		},
		Cursor: textarea.CursorStyle{
			Color: defaultTextStyles.Cursor.Color,
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
	return tea.NewView(t.Model.View())
}

// ===== textModel ============================================================

// textModel attempts to simplify the life of the consumer at the cost of a more
// complicated model implementation.  A textModel is one and only one of
// textinput.Model or textarea.Model, which have very, very similar APIs (at
// least the parts we're using).
type textModel struct {
	isArea bool
	input  textinputModel
	area   textareaModel
}

func newTextModel(isArea bool, placeholder string, initial string) textModel {
	// it is somewhat absurd to create both, but we'll be eating that memory
	// anyway and it makes the logic way easier here
	t := textModel{
		isArea: isArea,
		input:  newTextInput(placeholder),
		area:   newTextArea(placeholder),
	}

	if initial != "" {
		t = t.SetValue(initial)
		t = t.MoveToBegin()
	}

	return t
}

// static compile-time check that the shim meets the interface
var _ tea.Model = textModel{}

func (t textModel) Init() tea.Cmd { return nil }

func (t textModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var input textinputModel
	var area textareaModel

	var m tea.Model
	var cmd tea.Cmd

	if !t.isArea {
		m, cmd = t.input.Update(msg)
		input = m.(textinputModel)
	} else {
		m, cmd = t.area.Update(msg)
		area = m.(textareaModel)
	}

	return textModel{t.isArea, input, area}, cmd
}

func (t textModel) View() tea.View {
	if !t.isArea {
		return t.input.View()
	}

	return t.area.View()
}

func (t textModel) clone() textModel {
	return textModel{
		isArea: t.isArea,
		input:  t.input,
		area:   t.area,
	}
}

// Other textinput/textarea APIs we need to shim...in the underlying models,
// some of these are just exported properties. In textModel, it *must* be
// getter/setter methods.

func (t textModel) Value() string {
	if !t.isArea {
		return t.input.Value()
	}

	return t.area.Value()
}

func (t textModel) SetValue(value string) textModel {
	t2 := t.clone()

	if !t.isArea {
		t2.input.SetValue(value)
	} else {
		t2.area.SetValue(value)
	}

	return t2
}

func (t textModel) MoveToBegin() textModel {
	t2 := t.clone()
	if !t.isArea {
		t2.input.CursorStart()
	} else {
		t2.area.MoveToBegin()
	}
	return t2
}

func (t textModel) MoveToEnd() textModel {
	t2 := t.clone()
	if !t.isArea {
		t2.input.CursorEnd()
	} else {
		t2.area.MoveToEnd()
	}
	return t2
}

func (t textModel) Height() int {
	if !t.isArea {
		return 1 // ? padding?
	}

	return t.area.Height()
}

func (t textModel) SetHeight(height int) textModel {
	if !t.isArea {
		// not allowed
		return t
	}

	t2 := t.clone()
	t2.area.SetHeight(height)
	return t2
}

func (t textModel) MinHeight() int {
	if !t.isArea {
		return 1 // ? padding?
	}

	return t.area.MinHeight
}

func (t textModel) SetMinHeight(height int) textModel {
	if !t.isArea {
		// not allowed
		return t
	}

	t2 := t.clone()
	t2.area.MinHeight = height
	return t2
}

func (t textModel) MaxHeight() int {
	if !t.isArea {
		return 1 // ? padding?
	}

	return t.area.MaxHeight
}

func (t textModel) SetMaxHeight(height int) textModel {
	if !t.isArea {
		// not allowed
		return t
	}

	t2 := t.clone()
	t2.area.MaxHeight = height
	return t2
}

func (t textModel) MaxContentHeight() int {
	if !t.isArea {
		return 1 // ? padding?
	}

	return t.area.MaxContentHeight
}

func (t textModel) SetMaxContentHeight(height int) textModel {
	if !t.isArea {
		// not allowed
		return t
	}

	t2 := t.clone()
	t2.area.MaxContentHeight = height
	return t2
}

func (t textModel) DynamicHeight() bool {
	if !t.isArea {
		return false
	}

	return t.area.DynamicHeight
}

func (t textModel) SetDynamicHeight(b bool) textModel {
	if !t.isArea {
		// not allowed!
		return t
	}

	t2 := t.clone()
	t2.area.DynamicHeight = b

	return t2
}

func (t textModel) Width() int {
	if !t.isArea {
		// for textinput CharLimit === MaxWidth
		return t.input.Width()
	}

	return t.area.Width()
}

func (t textModel) SetWidth(width int) textModel {
	t2 := t.clone()

	if !t.isArea {
		t2.input.SetWidth(width)
	} else {
		t2.area.SetWidth(width)
	}

	return t2
}

// weird that textinput has width/setwidth, and textarea has
// maxwidth/setmaxwidth... need to rationalize these!

func (t textModel) MaxWidth() int {
	if !t.isArea {
		// for textinput Width === MaxWidth
		return t.input.Width()
	}

	return t.area.MaxWidth
}

func (t textModel) SetMaxWidth(width int) textModel {
	t2 := t.clone()

	if !t.isArea {
		t2.input.SetWidth(width)
	} else {
		t2.area.MaxWidth = width
	}

	return t2
}

func (t textModel) CharLimit() int {
	if !t.isArea {
		// for textinput CharLimit === MaxWidth
		return t.input.CharLimit
	}

	return t.area.CharLimit
}

func (t textModel) SetCharLimit(width int) textModel {
	t2 := t.clone()

	if !t.isArea {
		t2.input.CharLimit = width
	} else {
		t2.area.CharLimit = width
	}

	return t2
}

func (t textModel) Focus() (textModel, tea.Cmd) {
	t2 := t.clone()
	var cmd tea.Cmd

	if !t.isArea {
		cmd = t2.input.Focus()
	} else {
		cmd = t2.area.Focus()
	}

	return t2, cmd
}

func (t textModel) Blur() (textModel, tea.Cmd) {
	t2 := t.clone()
	var cmd tea.Cmd

	// neither textinput nor text area returns a tea.Cmd, but we do for
	// consistency
	if !t.isArea {
		t2.input.Blur()
	} else {
		t2.area.Blur()
	}

	return t2, cmd
}

// We don't really care about the full map as much as we care about the
// key.Bindings for help, and we return it in the order we care about
func (t textModel) GetKeyBindings() []key.Binding {
	if !t.isArea {
		// textinput has no help! 🤦🏼‍♂️
		km := t.input.KeyMap
		bindings := []key.Binding{
			// km.CharacterBackward,
			// km.CharacterForward,
			km.DeleteAfterCursor,
			km.DeleteBeforeCursor,
			// km.DeleteCharacterBackward,
			// km.DeleteCharacterForward,
			km.DeleteWordBackward,
			km.DeleteWordForward,
			// km.LineEnd,
			// km.LineStart,
			km.Paste,
			km.WordBackward,
			km.WordForward,
			// km.AcceptSuggestion
			// km.NextSuggestion
			// km.PrevSuggestion
		}

		if bindings[0].Help().Desc == "" {
			// get the help from text area, hoping the keys are similar
			areaKeyMap := textarea.DefaultKeyMap()
			keyHelp := make(map[string]string, 24)
			for _, takb := range []key.Binding{
				areaKeyMap.CharacterBackward,
				areaKeyMap.CharacterForward,
				areaKeyMap.DeleteAfterCursor,
				areaKeyMap.DeleteBeforeCursor,
				areaKeyMap.DeleteCharacterBackward,
				areaKeyMap.DeleteCharacterForward,
				areaKeyMap.DeleteWordBackward,
				areaKeyMap.DeleteWordForward,
				areaKeyMap.InsertNewline,
				areaKeyMap.LineEnd,
				areaKeyMap.LineNext,
				areaKeyMap.LinePrevious,
				areaKeyMap.LineStart,
				areaKeyMap.PageUp,
				areaKeyMap.PageDown,
				areaKeyMap.Paste,
				areaKeyMap.WordBackward,
				areaKeyMap.WordForward,
				areaKeyMap.InputBegin,
				areaKeyMap.InputEnd,
				areaKeyMap.UppercaseWordForward,
				areaKeyMap.LowercaseWordForward,
				areaKeyMap.CapitalizeWordForward,
				areaKeyMap.TransposeCharacterBackward,
			} {
				help := takb.Help().Desc
				for _, key := range takb.Keys() {
					keyHelp[key] = help
				}
			}

			// now that we have the help text from textarea, apply it to any
			// matching keys for textinput
			newBindings := make([]key.Binding, 0, len(bindings))
			for _, b := range bindings {
				keyStr := b.Keys()[0]
				newBindings = append(newBindings,
					key.NewBinding(
						key.WithKeys(b.Keys()...),
						key.WithHelp(keyStr, keyHelp[keyStr]),
						// key.WithDisabled(),
					))
			}
			bindings = newBindings
		}

		return bindings
	}

	km := t.area.KeyMap
	return []key.Binding{
		km.InsertNewline,
		// km.CharacterBackward,
		// km.CharacterForward,
		km.DeleteAfterCursor,
		km.DeleteBeforeCursor,
		// km.DeleteCharacterBackward,
		// km.DeleteCharacterForward,
		km.DeleteWordBackward,
		km.DeleteWordForward,
		// km.LineEnd,
		// km.LineNext,
		// km.LinePrevious,
		// km.LineStart,
		// km.PageUp,
		// km.PageDown,
		km.Paste,
		km.WordBackward,
		km.WordForward,
		km.InputBegin,
		km.InputEnd,
		km.UppercaseWordForward,
		km.LowercaseWordForward,
		km.CapitalizeWordForward,
		km.TransposeCharacterBackward,
	}
}

// unified style support
type textStyles struct {
	Focused textPartStyles
	Blurred textPartStyles
	Cursor  textCursorStyle
}

type textPartStyles struct {
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

type textCursorStyle struct {
	Color color.Color
}

func (t textModel) Styles() textStyles {
	var s textStyles
	if !t.isArea {
		i := t.input.Styles()
		s = textStyles{
			Focused: textPartStyles{
				Text:        i.Focused.Text,
				Placeholder: i.Focused.Placeholder,
			},
			Blurred: textPartStyles{
				Text:        i.Blurred.Text,
				Placeholder: i.Blurred.Placeholder,
			},
		}
	} else {
		a := t.area.Styles()
		s = textStyles{
			Focused: textPartStyles{
				Text:        a.Focused.Text,
				Placeholder: a.Focused.Placeholder,
			},
			Blurred: textPartStyles{
				Text:        a.Blurred.Text,
				Placeholder: a.Blurred.Placeholder,
			},
		}

	}

	return s
}

func (t textModel) SetStyles(s textStyles) textModel {
	t2 := t.clone()

	if !t.isArea {
		t2.input.SetStyles(textinput.Styles{
			Focused: textinput.StyleState{
				Text:        s.Focused.Text,
				Placeholder: s.Focused.Placeholder,
			},
			Blurred: textinput.StyleState{
				Text:        s.Blurred.Text,
				Placeholder: s.Blurred.Placeholder,
			},
			Cursor: textinput.CursorStyle{
				Color: s.Cursor.Color,
			},
		})
	} else {
		t2.area.SetStyles(textarea.Styles{
			Focused: textarea.StyleState{
				Base:        s.Focused.Text,
				Text:        s.Focused.Text,
				Placeholder: s.Focused.Placeholder,
			},
			Blurred: textarea.StyleState{
				Base:        s.Blurred.Text,
				Text:        s.Blurred.Text,
				Placeholder: s.Blurred.Placeholder,
			},
			Cursor: textarea.CursorStyle{
				Color: s.Cursor.Color,
			},
		})
	}

	return t2
}
