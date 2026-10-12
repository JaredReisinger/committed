// Package text normalizes the interface between charmbracelet/bubbles's
// [textinput.Model] and [textarea.Model], also fixing those up in places where
// they break the strict ELM/model paradigm.  One place where text.Model differs
// greatly is that it *dies not* render its own border/decoration. That is left
// entirely to the container.
package text

import (
	"log/slog"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

// text.Model attempts to simplify the life of the consumer at the cost of a
// more complicated model implementation.  A Model is one and only one of
// textinput.Model or textarea.Model, which have very, very similar APIs (at
// least the parts we're using).
type Model struct {
	isArea bool
	input  textinputModel
	area   textareaModel
	keymap textModelKeyMap
}

func New(isArea bool, placeholder string, initial string) Model {
	// it is somewhat absurd to create both, but we'll be eating that memory
	// anyway and it makes the logic way easier here
	t := Model{
		isArea: isArea,
		input:  newTextInput(placeholder),
		area:   newTextArea(placeholder),
		keymap: DefaultTextModelKeyMap(),
	}

	if initial != "" {
		t = t.SetValue(initial)
		t = t.MoveToBegin()
	}

	return t
}

// static compile-time check that the shim meets the interface
var _ tea.Model = Model{}

// "public" messages
type ValueChanged struct {
	Value string
}

func SendValueChanged(value string) tea.Cmd {
	return func() tea.Msg {
		return ValueChanged{value}
	}
}

func (t Model) IsArea() bool {
	return t.isArea
}

func (t Model) Init() tea.Cmd { return nil }

func (t Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var input textinputModel
	var area textareaModel

	var m tea.Model
	var cmd tea.Cmd

	before := t.Value()

	if !t.isArea {
		m, cmd = t.input.Update(msg)
		input = m.(textinputModel)
	} else {
		m, cmd = t.area.Update(msg)
		area = m.(textareaModel)
	}

	t2 := Model{t.isArea, input, area, t.keymap}

	after := t2.Value()

	if after != before {
		cmd = tea.Batch(cmd, SendValueChanged(after))
	}

	return t2, cmd
}

func (t Model) View() tea.View {
	if !t.isArea {
		return t.input.View()
	}

	return t.area.View()
}

func (t Model) clone() Model {
	// slog.Debug("cloning textModel", "isArea", t.isArea)

	// can we just return t by value?
	return Model{t.isArea, t.input, t.area, t.keymap}
}

// Other textinput/textarea APIs we need to shim...in the underlying models,
// some of these are just exported properties. In textModel, it *must* be
// getter/setter methods.

func (t Model) Placeholder() string {
	if !t.isArea {
		return t.input.Placeholder
	}

	return t.area.Placeholder
}

func (t Model) Value() string {
	if !t.isArea {
		return t.input.Value()
	}

	return t.area.Value()
}

func (t Model) SetValue(value string) Model {
	t2 := t.clone()

	// before := t2.Value()

	if !t.isArea {
		t2.input.SetValue(value)
	} else {
		t2.area.SetValue(value)
	}

	// after := t2.Value()

	// if after != before {
	// 	// Sadly, this is *outside* the Update() loop, which means we can't
	// 	// trigger a change event.  The caller will have to manage it.

	// 	// slog.Debug("content changed", "before", before, "after", after, "validation", t2.validationMessage)
	// }

	return t2
}

func (t Model) MoveToBegin() Model {
	t2 := t.clone()
	if !t.isArea {
		t2.input.CursorStart()
	} else {
		t2.area.MoveToBegin()
	}
	return t2
}

func (t Model) MoveToEnd() Model {
	t2 := t.clone()
	if !t.isArea {
		t2.input.CursorEnd()
	} else {
		t2.area.MoveToEnd()
	}
	return t2
}

func (t Model) Height() int {
	if !t.isArea {
		return 1 // ? padding?
	}

	return t.area.Height()
}

func (t Model) SetHeight(height int) Model {
	if !t.isArea {
		// not allowed
		return t
	}

	t2 := t.clone()
	t2.area.SetHeight(height)
	return t2
}

func (t Model) MinHeight() int {
	if !t.isArea {
		return 1 // ? padding?
	}

	return t.area.MinHeight
}

func (t Model) SetMinHeight(height int) Model {
	if !t.isArea {
		// not allowed
		return t
	}

	t2 := t.clone()
	t2.area.MinHeight = height
	return t2
}

func (t Model) MaxHeight() int {
	if !t.isArea {
		return 1 // ? padding?
	}

	return t.area.MaxHeight
}

func (t Model) SetMaxHeight(height int) Model {
	if !t.isArea {
		// not allowed
		return t
	}

	t2 := t.clone()
	t2.area.MaxHeight = height
	return t2
}

func (t Model) MaxContentHeight() int {
	if !t.isArea {
		return 1 // ? padding?
	}

	return t.area.MaxContentHeight
}

func (t Model) SetMaxContentHeight(height int) Model {
	if !t.isArea {
		// not allowed
		return t
	}

	t2 := t.clone()
	t2.area.MaxContentHeight = height
	return t2
}

func (t Model) DynamicHeight() bool {
	if !t.isArea {
		return false
	}

	return t.area.DynamicHeight
}

func (t Model) SetDynamicHeight(b bool) Model {
	if !t.isArea {
		// not allowed!
		return t
	}

	t2 := t.clone()
	t2.area.DynamicHeight = b

	return t2
}

func (t Model) Width() int {
	if !t.isArea {
		// for textinput CharLimit === MaxWidth
		return t.input.Width()
	}

	return t.area.Width()
}

func (t Model) SetWidth(width int) Model {
	t2 := t.clone()

	if !t.isArea {
		t2.input.SetWidth(width)
	} else {
		t2.area.SetWidth(width)
	}

	return t2
}

func (t Model) MinWidth() int {
	if !t.isArea {
		// for textinput CharLimit === MaxWidth
		return t.input.MinWidth()
	}

	return 0 // not needed for areas?

}

func (t Model) SetMinWidth(width int) Model {
	t2 := t.clone()

	if !t.isArea {
		// We implement SetMinWidth immutably, so we have to save back the model
		t2.input = t2.input.SetMinWidth(width)
	} else {
		// nop!
		slog.Warn("textModel area cannot SetMinWidth")
	}

	return t2
}

// weird that textinput has Width/SetWidth, and textarea has
// MaxWidth/SetMaxWidth... need to rationalize these!

func (t Model) MaxWidth() int {
	if !t.isArea {
		// for textinput Width === MaxWidth
		return t.input.Width()
	}

	return t.area.MaxWidth
}

func (t Model) SetMaxWidth(width int) Model {
	t2 := t.clone()

	if !t.isArea {
		t2.input.SetWidth(width)
	} else {
		t2.area.MaxWidth = width
	}

	return t2
}

func (t Model) CharLimit() int {
	if !t.isArea {
		// for textinput CharLimit === MaxWidth
		return t.input.CharLimit
	}

	return t.area.CharLimit
}

func (t Model) SetCharLimit(width int) Model {
	t2 := t.clone()

	if !t.isArea {
		t2.input.CharLimit = width
	} else {
		t2.area.CharLimit = width
	}

	return t2
}

func (t Model) Focus() (tea.Model, tea.Cmd) {
	t2 := t.clone()
	var cmd tea.Cmd

	if !t.isArea {
		cmd = t2.input.Focus()
	} else {
		cmd = t2.area.Focus()
	}

	return t2, cmd
}

func (t Model) Blur() (tea.Model, tea.Cmd) {
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

// Since we're merging textinput and textarea, we need a keymap that's the union
// of those, and then pass the appropriate subset as needed.

// textarea                      textinput
// --------------------------------------------------------------------------
// CapitalizeWordForward
// CharacterBackward             CharacterBackward
// CharacterForward              CharacterForward
// CopySelection
// DeleteAfterCursor             DeleteAfterCursor
// DeleteBeforeCursor            DeleteBeforeCursor
// DeleteCharacterBackward       DeleteCharacterBackward
// DeleteCharacterForward        DeleteCharacterForward
// DeleteWordBackward            DeleteWordBackward
// DeleteWordForward             DeleteWordForward
// InputBegin
// InputEnd
// InsertNewline
// LineEnd                       LineEnd
// LineNext
// LinePrevious
// LineStart                     LineStart
// --                            NextSuggestion
// LowercaseWordForward
// PageDown
// PageUp
// Paste                         Paste
// --                            PrevSuggestion
// SelectAll
// SelectCharacterBackward
// SelectCharacterForward
// SelectLineDown
// SelectLineUp
// SelectWordBackward
// SelectWordForward
// TransposeCharacterBackward
// UppercaseWordForward
// WordBackward                  WordBackward
// WordForward                   WordForward
// --------------------------------------------------------------------------

// textModelKeyMap is the union of [textinput.KeyMap] and [textarea.KeyMap],
// defined alphabetically (to make sure we don't' miss anything)
type textModelKeyMap struct {
	AcceptSuggestion           key.Binding // *textinput* only
	CapitalizeWordForward      key.Binding // textarea only
	CharacterBackward          key.Binding // both
	CharacterForward           key.Binding // both
	CopySelection              key.Binding // textarea only
	DeleteAfterCursor          key.Binding // both
	DeleteBeforeCursor         key.Binding // both
	DeleteCharacterBackward    key.Binding // both
	DeleteCharacterForward     key.Binding // both
	DeleteWordBackward         key.Binding // both
	DeleteWordForward          key.Binding // both
	InputBegin                 key.Binding // textarea only
	InputEnd                   key.Binding // textarea only
	InsertNewline              key.Binding // textarea only
	LineEnd                    key.Binding // both
	LineNext                   key.Binding // textarea only
	LinePrevious               key.Binding // textarea only
	LineStart                  key.Binding // both
	NextSuggestion             key.Binding // *textinput* only
	LowercaseWordForward       key.Binding // textarea only
	PageDown                   key.Binding // textarea only
	PageUp                     key.Binding // textarea only
	Paste                      key.Binding // both
	PrevSuggestion             key.Binding // *textinput* only
	SelectAll                  key.Binding // textarea only
	SelectCharacterBackward    key.Binding // textarea only
	SelectCharacterForward     key.Binding // textarea only
	SelectLineDown             key.Binding // textarea only
	SelectLineUp               key.Binding // textarea only
	SelectWordBackward         key.Binding // textarea only
	SelectWordForward          key.Binding // textarea only
	TransposeCharacterBackward key.Binding // textarea only
	UppercaseWordForward       key.Binding // textarea only
	WordBackward               key.Binding // both
	WordForward                key.Binding // both
}

// key.Binding conflates key bindings and help text, which I don't love. I might
// come up with a "can act like key.Binding" replacement if I want to allow for
// multiple key bindings (vim binding, emacs binding, pico/nano bindings...)
func DefaultTextModelKeyMap() textModelKeyMap {
	// *Should* we include keys in the map that aren't described? The concern is
	// "magic" sequences that we might want for use later.
	return textModelKeyMap{
		// keys: ␤ ␈ ␡ ← ↑ → ↓ ↲ || NerdFont:            
		AcceptSuggestion:           key.NewBinding(key.WithDisabled()), // unbound would be better?
		CapitalizeWordForward:      newBinding("alt+c", "capitalize word forward", "alt+c"),
		CharacterBackward:          newBinding("←/ctrl+b", "char backward", "left", "ctrl+b"),
		CharacterForward:           newBinding("→/ctrl+f", "char forward", "right", "ctrl+f"),
		CopySelection:              newBinding("ctrl+shift+c", "copy selection", "ctrl+shift+c"),
		DeleteAfterCursor:          newBinding("ctrl+k", "delete after cursor", "ctrl+k"),
		DeleteBeforeCursor:         newBinding("ctrl+u", "delete before cursor", "ctrl+u"),
		DeleteCharacterBackward:    newBinding("backspace", "delete char backward", "backspace", "ctrl+h"),
		DeleteCharacterForward:     newBinding("delete", "delete char forward", "delete", "ctrl+u"),
		DeleteWordBackward:         newBinding("alt+backspace", "delete word backward", "alt+backspace", "ctrl+w"),
		DeleteWordForward:          newBinding("alt+delete", "delete word forward", "alt+delete", "alt+d", "ctrl+delete"),
		InputBegin:                 newBinding("alt+<", "beginning of text", "alt+<", "ctrl+home"),
		InputEnd:                   newBinding("alt+>", "end of text", "alt+>", "ctrl+end"),
		InsertNewline:              newBinding("↲/enter", "insert newline", "enter", "ctrl+m"),
		LineEnd:                    newBinding("end/ctrl+e", "line end", "end", "ctrl+e"),
		LineNext:                   newBinding("↓/down", "next line", "down", "ctrl+n"),
		LinePrevious:               newBinding("↑/up", "previous line", "up", "ctrl+p"),
		LineStart:                  newBinding("home/ctrl+a", "line start", "home", "ctrl+a"),
		NextSuggestion:             key.NewBinding(key.WithDisabled()), // unbound would be better?
		LowercaseWordForward:       newBinding("alt+l", "lowercase word forward", "alt+l"),
		PageDown:                   newBinding("pgdown", "page down", "pgdown"),
		PageUp:                     newBinding("pgup", "page up", "pgup"),
		Paste:                      newBinding("ctrl+v", "paste", "ctrl+v"),
		PrevSuggestion:             key.NewBinding(key.WithDisabled()), // unbound would be better?
		SelectAll:                  newBinding("", "", ""),
		SelectCharacterBackward:    newBinding("shift+←", "select character backward", "shift+left"),
		SelectCharacterForward:     newBinding("shift+→", "select character forward", "shift+right"),
		SelectLineDown:             newBinding("shift+↓", "select line down", "shift+down"),
		SelectLineUp:               newBinding("shift+↑", "select line up", "shift+up"),
		SelectWordBackward:         newBinding("ctrl+shift+←", "select word backward", "ctrl+shift+left", "alt+shift+left", "alt+shift+b"),
		SelectWordForward:          newBinding("ctrl+shift+→", "select word forward", "ctrl+shift+right", "alt+shift+right", "alt+shift+f"),
		TransposeCharacterBackward: newBinding("ctrl+t", "transpose character backward", "ctrl+t"),
		UppercaseWordForward:       newBinding("alt+u", "uppercase word forward", "alt+u"),
		WordBackward:               newBinding("ctrl+←", "previous word", "ctrl+left", "alt+left", "alt+b"),
		WordForward:                newBinding("ctrl+→", "next word", "ctrl+right", "alt+right", "alt+f"),
	}
}

// a more-compact keys.NewBinding...
func newBinding(display string, help string, keys ...string) key.Binding {
	return key.NewBinding(key.WithKeys(keys...), key.WithHelp(display, help))
}

func (keys textModelKeyMap) ForTextInput() textinput.KeyMap {
	return textinput.KeyMap{
		CharacterForward:        keys.CharacterForward,
		CharacterBackward:       keys.CharacterBackward,
		WordForward:             keys.WordForward,
		WordBackward:            keys.WordBackward,
		DeleteWordBackward:      keys.DeleteWordBackward,
		DeleteWordForward:       keys.DeleteWordForward,
		DeleteAfterCursor:       keys.DeleteAfterCursor,
		DeleteBeforeCursor:      keys.DeleteBeforeCursor,
		DeleteCharacterBackward: keys.DeleteCharacterBackward,
		DeleteCharacterForward:  keys.DeleteCharacterForward,
		LineStart:               keys.LineStart,
		LineEnd:                 keys.LineEnd,
		Paste:                   keys.Paste,
		AcceptSuggestion:        keys.AcceptSuggestion,
		NextSuggestion:          keys.NextSuggestion,
		PrevSuggestion:          keys.PrevSuggestion,
	}
}

func (keys textModelKeyMap) ForTextArea() textarea.KeyMap {
	return textarea.KeyMap{
		CharacterBackward:          keys.CharacterBackward,
		CharacterForward:           keys.CharacterForward,
		DeleteAfterCursor:          keys.DeleteAfterCursor,
		DeleteBeforeCursor:         keys.DeleteBeforeCursor,
		DeleteCharacterBackward:    keys.DeleteCharacterBackward,
		DeleteCharacterForward:     keys.DeleteCharacterForward,
		DeleteWordBackward:         keys.DeleteWordBackward,
		DeleteWordForward:          keys.DeleteWordForward,
		InsertNewline:              keys.InsertNewline,
		LineEnd:                    keys.LineEnd,
		LineNext:                   keys.LineNext,
		LinePrevious:               keys.LinePrevious,
		LineStart:                  keys.LineStart,
		PageUp:                     keys.PageUp,
		PageDown:                   keys.PageDown,
		Paste:                      keys.Paste,
		WordBackward:               keys.WordBackward,
		WordForward:                keys.WordForward,
		InputBegin:                 keys.InputBegin,
		InputEnd:                   keys.InputEnd,
		UppercaseWordForward:       keys.UppercaseWordForward,
		LowercaseWordForward:       keys.LowercaseWordForward,
		CapitalizeWordForward:      keys.CapitalizeWordForward,
		TransposeCharacterBackward: keys.TransposeCharacterBackward,
		SelectCharacterForward:     keys.SelectCharacterForward,
		SelectCharacterBackward:    keys.SelectCharacterBackward,
		SelectWordForward:          keys.SelectWordForward,
		SelectWordBackward:         keys.SelectWordBackward,
		SelectLineUp:               keys.SelectLineUp,
		SelectLineDown:             keys.SelectLineDown,
		SelectAll:                  keys.SelectAll,
		CopySelection:              keys.CopySelection,
	}
}

func (t Model) KeyMap() textModelKeyMap {
	return t.keymap
}

func (t Model) SetKeyMap(keys textModelKeyMap) Model {
	input2 := t.input.SetKeyMap(keys.ForTextInput())
	area2 := t.area.SetKeyMap(keys.ForTextArea())
	return Model{t.isArea, input2, area2, keys}
}

// We don't really care about the full map as much as we care about the
// key.Bindings for help, and we return it in the order we care about.  We also
// don't plumb this through to textinputModel and textareaModel, because then we
// have more places to update if we change the order.
func (t Model) GetKeyBindings() []key.Binding {
	var bindings []key.Binding
	commonBindings := []key.Binding{
		// t.keymap.CharacterBackward,
		// t.keymap.CharacterForward,
		t.keymap.DeleteAfterCursor,
		t.keymap.DeleteBeforeCursor,
		// t.keymap.DeleteCharacterBackward,
		// t.keymap.DeleteCharacterForward,
		t.keymap.DeleteWordBackward,
		t.keymap.DeleteWordForward,
		t.keymap.Paste,
		t.keymap.WordBackward,
		t.keymap.WordForward,
	}

	if !t.isArea {
		bindings = commonBindings
	} else {
		bindings = []key.Binding{
			t.keymap.InsertNewline,
		}
		bindings = append(bindings, commonBindings...)
		bindings = append(bindings, []key.Binding{
			t.keymap.InputBegin,
			t.keymap.InputEnd,
			t.keymap.UppercaseWordForward,
			t.keymap.LowercaseWordForward,
			t.keymap.CapitalizeWordForward,
			t.keymap.TransposeCharacterBackward,
		}...)
	}

	// we used to have a helper that set up key help, copying from
	// textarea.KeyMap's binding.Help
	return bindings
}

func (t Model) Styles() Styles {
	var s Styles
	if !t.isArea {
		i := t.input.Styles()
		s = Styles{
			Focused: PartStyles{
				Text:        i.Focused.Text,
				Placeholder: i.Focused.Placeholder,
			},
			Blurred: PartStyles{
				Text:        i.Blurred.Text,
				Placeholder: i.Blurred.Placeholder,
			},
		}
	} else {
		a := t.area.Styles()
		s = Styles{
			Focused: PartStyles{
				Text:        a.Focused.Text,
				Placeholder: a.Focused.Placeholder,
			},
			Blurred: PartStyles{
				Text:        a.Blurred.Text,
				Placeholder: a.Blurred.Placeholder,
			},
		}

	}

	return s
}

func (t Model) SetStyles(s Styles) Model {
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
