package tui

import (
	"charm.land/bubbles/v2/help"
	tea "charm.land/bubbletea/v2"
)

// helpModel shims [help.Model] to make it actually meet the [tea.Model]
// interface.
type helpModel struct {
	help.Model
	keymap help.KeyMap
}

var _ tea.Model = helpModel{}

func newHelpModel() helpModel {
	h := help.New()
	h.SetWidth(80)
	h.ShowAll = true
	return helpModel{h, nil}
}

func (m helpModel) Init() tea.Cmd {
	return nil
}

// SetWidth sets the maximum width for the help view.
func (m helpModel) SetWidth(w int) helpModel {
	m2 := m.Model
	m2.SetWidth(w)
	return helpModel{m2, m.keymap}
}

// // Width returns the maximum width for the help view.
// func (m Model) Width() int {
// 	return m.width
// }

func (m helpModel) KeyMap() help.KeyMap {
	return m.keymap
}

func (m helpModel) SetKeyMap(keymap help.KeyMap) helpModel {
	return helpModel{m.Model, keymap}
}

func (m helpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	hm, cmd := m.Model.Update(msg)
	return helpModel{hm, m.keymap}, cmd
}

func (m helpModel) View() tea.View {
	var s string
	if m.keymap != nil {
		s = m.Model.View(m.keymap)
	}
	return tea.NewView(s)
}
