package tui

import (
	"log/slog"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/jaredreisinger/committed/pkg/bubbles/text"
	"github.com/jaredreisinger/committed/pkg/teautil"
)

// simpleEditor represents the very, very simple "just a plain editor" TUI.
type simpleEditor struct {
	children teautil.Router[field, tea.Model] // *focusable* children

	tabstops     []field
	focusedField field

	winWidth  int
	winHeight int
}

func newSimpleEditor(existingMsg string) simpleEditor {
	return simpleEditor{
		children: teautil.NewRouter(map[field]tea.Model{
			bodyField: text.New(true, "message body", existingMsg).
				SetMaxContentHeight(100).
				SetDynamicHeight(true).
				SetMinHeight(10),
			helpField: newHelpModel(),
		}),
		tabstops:     []field{bodyField},
		focusedField: noField,
	}
}

// Init initializes the bubbletea program.
func (form simpleEditor) Init() tea.Cmd {
	// Set initial focus
	return setFocusCmd(bodyField) // kind of bogus...
}

// Update handles user input and updates the model.
func (form simpleEditor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd // aggregated commands
	var cmd tea.Cmd    // returned Cmd from delegated Update() calls
	handled := false

	switch msgT := msg.(type) {
	case tea.WindowSizeMsg:
		form = form.resize(msgT.Width, msgT.Height)
		handled = true

	case setFocusMsg:
		form, cmd = form.setFocus(msgT.field)
		cmds = append(cmds, cmd)
		handled = true

	case tea.KeyPressMsg:
		// REVIEW.... should we check the focus field for handling first?  I
		// have to think really hard about the order key messages should happen
		// in.  I think that perhaps it's (1) app-level focus/submit, (2)
		// field-level handling/typing.  This is kind of what's happening now,
		// so maybe it's fine.  But the keymap is conflated, and perhaps should
		// not be.
		handled = true // assume handled, set back to false in default case
		switch {
		case key.Matches(msgT, defaultKeyMap.Cancel):
			cmds = append(cmds, tea.Interrupt)

		case key.Matches(msgT, defaultKeyMap.NextFromSingle) ||
			key.Matches(msgT, defaultKeyMap.NextFromMulti):
			cmds = append(cmds, form.nextField())

		case key.Matches(msgT, defaultKeyMap.Prev):
			cmds = append(cmds, form.prevField())

		case key.Matches(msgT, defaultKeyMap.Submit):
			// TODO: validate?
			cmds = append(cmds, tea.Quit)

		default:
			handled = false
		}
	}

	// delegate other messages to the field/model with focus...
	if !handled {
		var cmd tea.Cmd
		form.children, cmd = form.children.Update(msg, form.focusedField)
		cmds = append(cmds, cmd)
	}

	return form, tea.Batch(cmds...)
}

// Field sizes and presentation needs to get calculated *often*, because we want
// it to be responsive to content and to focus. At the same time, we don't want
// the sizing logic to be scattered all over the code.

func (form simpleEditor) resize(winWidth int, winHeight int) simpleEditor {
	slog.Debug("window resize", "width", winWidth, "height", winHeight)

	form.winWidth = winWidth
	form.winHeight = winHeight

	body := form.children.MustGet(bodyField).(text.Model)
	help := form.children.MustGet(helpField).(helpModel)

	body = body.SetWidth(winWidth - 2)
	help = help.SetWidth(winWidth)

	form.children = form.children.SetMap(map[field]tea.Model{
		bodyField: body,
		helpField: help,
	})

	return form
}

// updateFocus updates which field is currently focused.
func (form simpleEditor) setFocus(focusField field) (simpleEditor, tea.Cmd) {
	var cmds []tea.Cmd

	if form.focusedField != focusField {
		modelMap := make(map[field]tea.Model, 3)

		// it's a "loop" just so we don't have to duplicate the "get from
		// router"/"set to router" logic
		for _, f := range []field{form.focusedField, focusField} {
			if child, ok := form.children.Get(f); ok {
				c2 := child
				var cmd tea.Cmd
				if focuser, ok := c2.(Focuser); ok {
					if f == focusField {
						c2, cmd = focuser.Focus()
					} else {
						c2, cmd = focuser.Blur()
					}

					modelMap[f] = c2
					cmds = append(cmds, teautil.Wrap(cmd, f))
				}
			}
		}

		// Also update the help keymap... we *could* do this inside the loop to
		// avoid calling form.children.Get() again, but since we're casting to a
		// different interface, this seems cleaner.
		var keyBindings []key.Binding
		if child, ok := form.children.Get(focusField); ok {
			if keyBinder, ok := child.(KeyBinder); ok {
				keyBindings = keyBinder.GetKeyBindings()
			}
		}

		help := form.children.MustGet(helpField).(helpModel)
		modelMap[helpField] = help.SetKeyMap(buildHelpKeys(simpleEditorKeyMap, keyBindings))

		form.children = form.children.SetMap(modelMap)
		form.focusedField = focusField
	}

	return form, tea.Batch(cmds...)
}

// nextField moves to the next field.
func (form simpleEditor) nextField() tea.Cmd {
	return setFocusCmd(circularDelta(form.tabstops, form.focusedField, 1))
}

// prevField moves to the previous field.
func (form simpleEditor) prevField() tea.Cmd {
	return setFocusCmd(circularDelta(form.tabstops, form.focusedField, -1))
}

// View renders the TUI.
func (form simpleEditor) View() tea.View {
	// TODO: would it be better to define a type for the value/valid/decoration
	// info? I suspect that would be better than the parallel maps.
	type info struct {
		view    string
		message string
		render  string
	}

	infos := make(map[field]info, form.children.Len())

	for f, child := range form.children.All() {
		var info info

		t, ok := child.(text.Model)
		if ok {
			slog.Debug("getting child view content", "child", f, "width", t.Width(), "minWidth", t.MinWidth(), "content", t.Value())
		}

		info.view = child.View().Content

		// Figure out the decorator; depends on whether the field is valid, etc.
		// This also depends on where the focus is, as we play some interesting
		// games with borders to keep the rendering concise.
		var childState state = blurred
		var childStatus status = defaultStatus
		var childKind fieldKind = single
		if f == form.focusedField {
			childState = focused
		}

		// TODO: make an interface for styling single/multi...
		textChild, ok := child.(text.Model)
		if f == typeList || (ok && textChild.IsArea()) {
			childKind = multi
		}

		border := kindDecorations[childKind].BorderForeground(defaultColors[childStatus][childState][decoration])

		// switch f {
		// case typeField, scopeField, descField:
		// 	if form.focusedField >= bodyField {
		// 		border = border.UnsetBorderBottom()
		// 	}
		// case bodyField:
		// 	if form.focusedField < bodyField {
		// 		border = border.UnsetBorderTop()
		// 	}
		// 	if form.focusedField == footerField {
		// 		border = border.UnsetBorderBottom()
		// 	}
		// case footerField:
		// 	if form.focusedField < footerField {
		// 		border = border.UnsetBorderTop()
		// 	}
		// case helpField:
		// 	border = lipgloss.NewStyle() // no border!
		// }

		info.render = border.Render(info.view)

		infos[f] = info
	}

	// // If there's an error (validation!), make sure to show it. Do we do both
	// // field and global validation?
	// var errMsg string
	// if form.fieldErrs[form.focusedField] != nil {
	// 	errMsg = form.fieldErrs[form.focusedField].Error()
	// } else if form.err != nil {
	// 	errMsg = form.err.Error()
	// }

	// if errMsg != "" {
	// 	slog.Debug("showing error", "err", errMsg)
	// 	errMsg = lipgloss.NewStyle().Width(form.winWidth).Foreground(defaultColors[invalid][focused][decoration]).Render(errMsg)
	// }

	// // For the top line, we omit unused scope unless it has focus
	// showScope := form.focusedField == scopeField || form.children.MustGet(scopeField).(text.Model).Value() != ""
	// header := lipgloss.JoinHorizontal(
	// 	lipgloss.Top,
	// 	" ", // one space to align body left border
	// 	infos[typeField].render,
	// )
	// if showScope {
	// 	header = lipgloss.JoinHorizontal(
	// 		lipgloss.Top,
	// 		header,
	// 		"(",
	// 		infos[scopeField].render,
	// 		")",
	// 	)
	// }
	// header = lipgloss.JoinHorizontal(
	// 	lipgloss.Top,
	// 	header,
	// 	": ",
	// 	infos[descField].render,
	// )

	mainView := lipgloss.JoinVertical(
		lipgloss.Left,
		"\n",
		// header,
		infos[bodyField].render,
		// infos[footerField].render,
	)

	// if errMsg != "" {
	// 	mainView = lipgloss.JoinVertical(
	// 		lipgloss.Left,
	// 		mainView,
	// 		errMsg,
	// 	)
	// }

	mainView = lipgloss.JoinVertical(
		lipgloss.Left,
		mainView,
		// form.help.View(helpKeys), // &defaultKeyMap),
		infos[helpField].render,
	)

	layers := []*lipgloss.Layer{
		lipgloss.NewLayer(mainView),
	}

	// if form.focusedField == typeField || form.focusedField == typeList {
	// 	layers = append(layers, lipgloss.NewLayer(infos[typeList].render).X(10).Y(1))
	// }

	comp := lipgloss.NewCompositor(layers...)

	return tea.NewView(comp.Render())
}

// Result returns the composed commit message and any error.
func (form simpleEditor) Result() (string, error) {
	texts := make(map[field]string, form.children.Len())

	for f, child := range form.children.All() {
		if textChild, ok := child.(text.Model); ok {
			texts[f] = strings.TrimSpace(textChild.Value())
		}
	}

	// if !m.done {
	// 	return nil, fmt.Errorf("form not completed")
	// }
	return texts[bodyField], nil
}
