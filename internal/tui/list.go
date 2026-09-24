package tui

import (
	"log/slog"
	"slices"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// list is a simple picker list, meant to pick the commit type

type listModel struct {
	items         []string
	selectedIndex int
}

func newListModel(items []string) listModel {
	slog.Debug("creating list", "items", items)
	l := listModel{
		items:         slices.Clone(items),
		selectedIndex: 0,
	}

	return l
}

// static compile-time check that the type meets the interface
var _ tea.Model = listModel{}

type listSelectionChangedMsg struct {
	selectedItemIndex int
	selectedItem      string
}

func listSelectionChangedCmd(index int, item string) tea.Cmd {
	return func() tea.Msg {
		return listSelectionChangedMsg{
			selectedItemIndex: index,
			selectedItem:      item,
		}
	}
}

func (l listModel) getSelectedIndex() int {
	return l.selectedIndex
}

func (l listModel) Init() tea.Cmd { return nil }

func (l listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// var cmds []tea.Cmd // aggregated commands
	var cmd tea.Cmd
	// handled := false // do we do this, or msg = nil?

	switch msgT := msg.(type) {
	case tea.KeyPressMsg:
		// handled = true // assume handled, set back to false in default case
		switch {
		case key.Matches(msgT, listKeyMap.Up):
			// we don't use commands because it's simple?
			l.selectedIndex = ((l.selectedIndex + len(l.items)) - 1) % len(l.items)
			cmd = listSelectionChangedCmd(l.selectedIndex, l.items[l.selectedIndex])

		case key.Matches(msgT, listKeyMap.Down):
			l.selectedIndex = (l.selectedIndex + 1) % len(l.items)
			cmd = listSelectionChangedCmd(l.selectedIndex, l.items[l.selectedIndex])

			// case key.Matches(msgT, listKeyMap.SelectAndNext):
			// 	// change parent focus?
			// case key.Matches(msgT, listKeyMap.SelectAndPrev):
			// 	// change parent focus?

			// default:
			// 	// handled = false
		}
	}

	return l, cmd
}

func (l listModel) View() tea.View {
	// slog.Debug("viewing list", "items", l.items)

	lines := make([]string, 0, len(l.items))
	for i, item := range l.items {
		style := defaultTextStyles.Blurred.Text
		if i == l.selectedIndex {
			style = defaultTextStyles.Focused.Text
		}
		lines = append(lines, style.Render(item))
	}
	view := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return tea.NewView(view)
}

func (l listModel) GetKeyBindings() []key.Binding {
	km := listKeyMap
	bindings := []key.Binding{
		km.Up,
		km.Down,
		km.SelectAndNext,
		km.SelectAndPrev,
	}
	return bindings
}

type listKeyMapX struct {
	Up            key.Binding
	Down          key.Binding
	SelectAndNext key.Binding
	SelectAndPrev key.Binding
}

var listKeyMap = listKeyMapX{
	Up: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "up"),
	),

	Down: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "down"),
	),

	SelectAndNext: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "select/next"),
	),

	SelectAndPrev: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "select/prev"),
	),
}
