package tui

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
)

type keyMap struct {
	Cancel     key.Binding
	NextSingle key.Binding
	NextMulti  key.Binding
	Prev       key.Binding
	Submit     key.Binding
}

var defaultKeyMap = keyMap{
	Cancel: key.NewBinding(
		// key.WithKeys("ctrl+c", "esc"),
		// key.WithHelp("ctrl+c/esc", "cancel"),
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "cancel"),
	),

	NextSingle: key.NewBinding(
		key.WithKeys("tab", "enter"),
		key.WithHelp("tab/enter", "next field"),
	),

	NextMulti: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next field"),
		key.WithDisabled(),
	),

	Prev: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "previous field"),
	),

	Submit: key.NewBinding(
		key.WithKeys("ctrl+enter"),
		key.WithHelp("ctrl+enter", ""),
	),
}

// func (k *keyMap) forHelp() help.KeyMap {

// }

func (k *keyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.NextSingle,
		k.NextMulti,
		k.Prev,
		k.Cancel,
	}
}

func (k *keyMap) FullHelp() [][]key.Binding {
	return nil
}

func inTextareaKeyMap(taKeyMap textarea.KeyMap) help.KeyMap {
	// combine maps
	h := &helpMap{}
	h.short = []key.Binding{
		defaultKeyMap.NextSingle,
		defaultKeyMap.NextMulti,
		defaultKeyMap.Prev,
		defaultKeyMap.Cancel,
		taKeyMap.InsertNewline,

		// taKeyMap.CharacterBackward,
		// taKeyMap.CharacterForward,
		taKeyMap.DeleteAfterCursor,
		taKeyMap.DeleteBeforeCursor,
		// taKeyMap.DeleteCharacterBackward,
		// taKeyMap.DeleteCharacterForward,
		taKeyMap.DeleteWordBackward,
		taKeyMap.DeleteWordForward,
		// taKeyMap.LineEnd,
		// taKeyMap.LineNext,
		// taKeyMap.LinePrevious,
		// taKeyMap.LineStart,
		// taKeyMap.PageUp,
		// taKeyMap.PageDown,
		taKeyMap.Paste,
		taKeyMap.WordBackward,
		taKeyMap.WordForward,
		taKeyMap.InputBegin,
		taKeyMap.InputEnd,
		taKeyMap.UppercaseWordForward,
		taKeyMap.LowercaseWordForward,
		taKeyMap.CapitalizeWordForward,
		taKeyMap.TransposeCharacterBackward,
	}

	// for full help we need "columns of bindings"
	h.full = append(h.full,
		[]key.Binding{
			defaultKeyMap.NextSingle,
			defaultKeyMap.NextMulti,
			defaultKeyMap.Prev,
			defaultKeyMap.Cancel,
			taKeyMap.InsertNewline,
			// taKeyMap.CharacterBackward,
			// taKeyMap.CharacterForward,
			taKeyMap.DeleteAfterCursor,
			taKeyMap.DeleteBeforeCursor,
		},
		[]key.Binding{
			// taKeyMap.DeleteCharacterBackward,
			// taKeyMap.DeleteCharacterForward,
			taKeyMap.DeleteWordBackward,
			taKeyMap.DeleteWordForward,
			// taKeyMap.LineEnd,
			// taKeyMap.LineNext,
			// taKeyMap.LinePrevious,
			// taKeyMap.LineStart,
			// taKeyMap.PageUp,
			// taKeyMap.PageDown,
			taKeyMap.Paste,
			taKeyMap.WordBackward,
			taKeyMap.WordForward,
			taKeyMap.InputBegin,
			taKeyMap.InputEnd,
		},
		[]key.Binding{
			taKeyMap.UppercaseWordForward,
			taKeyMap.LowercaseWordForward,
			taKeyMap.CapitalizeWordForward,
			taKeyMap.TransposeCharacterBackward,
		},
	)

	return h
}

type helpMap struct {
	short []key.Binding
	full  [][]key.Binding
}

func (h *helpMap) ShortHelp() []key.Binding {
	return h.short
}

func (h *helpMap) FullHelp() [][]key.Binding {
	return h.full
}
