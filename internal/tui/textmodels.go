package tui

import (
	"fmt"
	"image/color"
	"log/slog"

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

// ===== textModel ============================================================

// textModel attempts to simplify the life of the consumer at the cost of a more
// complicated model implementation.  A textModel is one and only one of
// textinput.Model or textarea.Model, which have very, very similar APIs (at
// least the parts we're using).
type textModel struct {
	isArea bool
	input  textinputModel
	area   textareaModel
	keymap textModelKeyMap
}

func newTextModel(isArea bool, placeholder string, initial string) textModel {
	// it is somewhat absurd to create both, but we'll be eating that memory
	// anyway and it makes the logic way easier here
	t := textModel{
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
var _ tea.Model = textModel{}

// "public" messages
type textModelValueChanged struct {
	Value string
}

func SendTextModelValueChanged(value string) tea.Cmd {
	return func() tea.Msg {
		return textModelValueChanged{value}
	}
}

func (t textModel) Init() tea.Cmd { return nil }

func (t textModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	t2 := textModel{t.isArea, input, area, t.keymap}

	after := t2.Value()

	if after != before {
		cmd = tea.Batch(cmd, SendTextModelValueChanged(after))
	}

	return t2, cmd
}

func (t textModel) View() tea.View {
	if !t.isArea {
		return t.input.View()
	}

	return t.area.View()
}

func (t textModel) clone() textModel {
	// slog.Debug("cloning textModel", "isArea", t.isArea)

	// can we just return t by value?
	return textModel{t.isArea, t.input, t.area, t.keymap}
}

// Other textinput/textarea APIs we need to shim...in the underlying models,
// some of these are just exported properties. In textModel, it *must* be
// getter/setter methods.

func (t textModel) Placeholder() string {
	if !t.isArea {
		return t.input.Placeholder
	}

	return t.area.Placeholder
}

func (t textModel) Value() string {
	if !t.isArea {
		return t.input.Value()
	}

	return t.area.Value()
}

func (t textModel) SetValue(value string) textModel {
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

func (t textModel) MinWidth() int {
	if !t.isArea {
		// for textinput CharLimit === MaxWidth
		return t.input.MinWidth()
	}

	return 0 // not needed for areas?

}

func (t textModel) SetMinWidth(width int) textModel {
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

func (t textModel) Focus() (tea.Model, tea.Cmd) {
	t2 := t.clone()
	var cmd tea.Cmd

	if !t.isArea {
		cmd = t2.input.Focus()
	} else {
		cmd = t2.area.Focus()
	}

	return t2, cmd
}

func (t textModel) Blur() (tea.Model, tea.Cmd) {
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
		// keys: ␤ ␈ ␡ ← ↑ → ↓ ↲ || NERDFONT:            
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

func (t textModel) KeyMap() textModelKeyMap {
	return t.keymap
}

func (t textModel) SetKeyMap(keys textModelKeyMap) textModel {
	input2 := t.input.SetKeyMap(keys.ForTextInput())
	area2 := t.area.SetKeyMap(keys.ForTextArea())
	return textModel{t.isArea, input2, area2, keys}
}

// We don't really care about the full map as much as we care about the
// key.Bindings for help, and we return it in the order we care about.  We also
// don't plumb this through to textinputModel and textareaModel, because then we
// have more places to update if we change the order.
func (t textModel) GetKeyBindings() []key.Binding {
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
