package tui

import (
	"charm.land/bubbles/v2/key"
)

type keyMap struct {
	Cancel key.Binding
	Next   key.Binding
	Prev   key.Binding
	Enter  key.Binding

	// Up     key.Binding
	// Down   key.Binding

	// cache help->keymap?
}

var defaultKeyMap = keyMap{
	Cancel: key.NewBinding(
		key.WithKeys("ctrl+c", "esc"),
		key.WithHelp("ctrl+c/esc", "cancel"),
	),

	Next: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next"),
	),

	Prev: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "previous"),
	),

	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "accept / next"),
	),

	// Up: key.NewBinding(
	// 	key.WithKeys("k", "up"),        // actual keybindings
	// 	key.WithHelp("↑/k", "move up"), // corresponding help text
	// ),

	// Down: key.NewBinding(
	// 	key.WithKeys("j", "down"),
	// 	key.WithHelp("↓/j", "move down"),
	// ),
}

// func (k *keyMap) forHelp() help.KeyMap {

// }

func (k *keyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Next, k.Prev, k.Enter, k.Cancel,
	}
}

func (k *keyMap) FullHelp() [][]key.Binding {
	return nil
}
