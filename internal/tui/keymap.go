package tui

import (
	"slices"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
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

func buildHelpKeys(textKeyBindings []key.Binding) help.KeyMap {
	// combine maps
	h := &helpMap{}

	h.short = []key.Binding{
		defaultKeyMap.NextSingle,
		defaultKeyMap.NextMulti,
		defaultKeyMap.Prev,
		defaultKeyMap.Cancel,
	}
	h.short = append(h.short, textKeyBindings...)

	// For the long list, it's 6 or 7 per group, and we can tell based on the
	// full length...(4+7) or (4+14), 11 or 18, so we pick 15 as the threshold
	chunkSize := 6
	if len(h.short) >= 15 {
		chunkSize = 7
	}

	for chunk := range slices.Chunk(h.short, chunkSize) {
		h.full = append(h.full, chunk)
	}

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
