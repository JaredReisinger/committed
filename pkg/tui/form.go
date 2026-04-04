package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/jaredreisinger/committed/internal/config"
	"github.com/jaredreisinger/committed/pkg/commit"
)

type field int

const (
	typeField field = iota
	summaryField
	detailsField
	// help?
	maxField
)

// formModel represents the TUI state for conventional commit composition.
type formModel struct {
	config           *config.Config
	existingMsg      *commit.Message // just for init, don't persist
	focusedField     field
	typ              textarea.Model
	summary          textarea.Model
	details          textarea.Model
	detailsMaxHeight int
	help             help.Model
	typeIndex        int // for navigating type enum
	err              error
	log              string
}

func stdText(fullBorder bool) textarea.Model {
	t := textarea.New()
	t.SetStyles(textareaStyles(fullBorder))
	t.Prompt = ""
	t.ShowLineNumbers = false
	t.KeyMap.InsertNewline.SetEnabled(false)
	return t
}

// newModel creates a new TUI model with the given configuration and optional
// existing message.
func newModel(cfg *config.Config, existingMsg *commit.Message) formModel {
	typ := stdText(false)
	typ.Placeholder = "Select commit type"
	// typ.CharLimit = 50
	// typ.SetWidth(maxTyp)
	typ.SetHeight(1)
	typ.MaxHeight = 1

	summary := stdText(false)
	summary.Placeholder = "brief description of changes"
	// subLimit := cfg.SubjectMaxLength
	// if cfg.HeaderMaxLength-maxTyp < subLimit {
	// 	subLimit = cfg.HeaderMaxLength - maxTyp
	// }
	// summary.CharLimit = subLimit
	// summary.SetWidth(subLimit)
	summary.SetHeight(1)
	summary.MaxHeight = 1

	details := stdText(true)
	details.Placeholder = "Detailed description (optional)"
	details.KeyMap.InsertNewline.SetEnabled(false)
	// Border is *outside* the width, but prompt/line-numbers are inside, except
	// if there's *no* prompt or line numbers, the border is included in the
	// width. At least that what appears to happen. The code looks like it tries
	// to include margin+border+padding+content. 🤷🏼‍♂️
	// details.SetWidth(cfg.BodyMaxLineLength)
	details.DynamicHeight = true
	details.MinHeight = 10
	// details.SetHeight(10)
	// details.MaxHeight = 20

	help := help.New()
	help.ShowAll = true
	help.SetWidth(80)

	m := formModel{
		config:       cfg,
		existingMsg:  existingMsg,
		focusedField: typeField,
		typ:          typ,
		summary:      summary,
		details:      details,
		help:         help,
		typeIndex:    0,
	}

	// Pre-populate fields if we have an existing message
	if existingMsg != nil {
		m.populateFromExisting()
	}

	// Set initial focus
	m.updateFocus()

	return m
}

func maxTypeLength(cfg *config.Config) int {
	maxTyp := 0
	for _, t := range cfg.Types {
		w := len(t)
		if w > maxTyp {
			maxTyp = w
		}
	}
	return maxTyp
}

// populateFromExisting fills the form fields from an existing commit message.
func (m *formModel) populateFromExisting() {
	if m.existingMsg == nil {
		return
	}

	// Set type
	if m.existingMsg.Type != "" {
		// Find the type in our config enum
		for i, t := range m.config.Types {
			if t == m.existingMsg.Type {
				m.typeIndex = i
				m.typ.SetValue(t)
				break
			}
		}
	}

	// Set summary (description)
	if m.existingMsg.Description != "" {
		m.summary.SetValue(m.existingMsg.Description)
	}

	// Set details (body)
	if m.existingMsg.Body != "" {
		m.details.SetValue(m.existingMsg.Body)
	}
}

// Init initializes the bubbletea program.
func (m formModel) Init() tea.Cmd {
	// should other processing happen here?
	return textinput.Blink
}

// Update handles user input and updates the model.
func (m formModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	handled := false // do we do this, or msg = nil?

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.resize(msg.Width, msg.Height)
		handled = true
	case tea.KeyPressMsg:
		handled = true // assume handled, set back to false in default case
		switch {
		case key.Matches(msg, defaultKeyMap.Cancel):
			cmds = append(cmds, tea.Interrupt)

		case key.Matches(msg, defaultKeyMap.NextSingle) ||
			key.Matches(msg, defaultKeyMap.NextMulti):
			m.nextField()

		case key.Matches(msg, defaultKeyMap.Prev):
			m.prevField()

		case key.Matches(msg, defaultKeyMap.Submit):
			// TODO: validate?
			cmds = append(cmds, tea.Quit)

		// case key.Matches(msg, defaultKeyMap.Enter):
		// 	// enter handling depends on focus...
		// 	if m.focusedField == detailsField {
		// 		// // TODO: change to validate as user types?
		// 		// // Submit the form
		// 		// if err := m.validateSummary(); err != nil {
		// 		// 	m.err = err
		// 		// } else {
		// 		// 	m.err = nil // clear any error?
		// 		// 	cmds = append(cmds, tea.Quit)
		// 		// }

		// 		// delegate enter to details!
		// 		handled = false
		// 	} else {
		// 		m.nextField()
		// 	}

		default:
			handled = false
		}

		// case tea.KeyMsg:
		// 	switch msg.String() {

		// 	case "up", "down":
		// 		if m.focusedField == typeField {
		// 			if msg.String() == "up" {
		// 				if m.typeIndex > 0 {
		// 					m.typeIndex--
		// 				}
		// 			} else {
		// 				if m.typeIndex < len(m.config.Types)-1 {
		// 					m.typeIndex++
		// 				}
		// 			}
		// 			return m, nil
		// 		}
		// 	}
	}

	// delegate other messages to the field/model with focus...
	if !handled {
		var cmd tea.Cmd
		switch m.focusedField {
		case typeField:
			m.typ, cmd = m.typ.Update(msg)
		case summaryField:
			m.summary, cmd = m.summary.Update(msg)
		case detailsField:
			m.details, cmd = m.details.Update(msg)
			if m.details.DynamicHeight && m.details.Height() >= m.detailsMaxHeight {
				m.details.DynamicHeight = false
				m.details.MaxHeight = 0
				m.details.SetHeight(m.detailsMaxHeight)
			}
		}

		cmds = append(cmds, cmd)
	}

	// // Clear any previous validation errors when user types
	// if m.err != nil && (m.focusedField == summaryField || m.focusedField == detailsField) {
	// 	m.err = nil
	// }

	return m, tea.Batch(cmds...)
}

func (m *formModel) resize(width int, height int) {
	m.log = fmt.Sprintf("%dx%d", width, height)

	// Border is *outside* the width, but prompt/line-numbers are inside, except
	// if there's *no* prompt or line numbers, the border is included in the
	// width. At least that what appears to happen. The code looks like it tries
	// to include margin+border+padding+content. 🤷🏼‍♂️

	xx, yy := m.typ.Styles().Focused.Base.GetFrameSize()
	cfg := m.config

	maxTyp := maxTypeLength(cfg)
	m.typ.CharLimit = maxTyp
	m.typ.SetWidth(maxTyp + xx + 1) // +1 for cursor

	subLimit := min(cfg.SubjectMaxLength, cfg.HeaderMaxLength-maxTyp)
	m.summary.CharLimit = subLimit
	m.summary.SetWidth(min(subLimit+xx+1, width-m.typ.Width())) // +1 for cursor

	m.details.SetWidth(min(cfg.BodyMaxLineLength+xx, width))
	helpHeight := 8
	logHeight := 1
	m.detailsMaxHeight = height - m.typ.Height() - yy - yy - helpHeight - logHeight - 1
	m.details.MaxHeight = m.detailsMaxHeight

	if m.details.DynamicHeight && m.details.Height() >= m.detailsMaxHeight {
		m.details.DynamicHeight = false
		m.details.MaxHeight = 0
		m.details.SetHeight(m.detailsMaxHeight)
	}

	// it would be nice to switch back to dynamic if the content shrinks, but
	// there's no easy way to get "how many visual lines are needed?"

	m.help.SetWidth(xx)
}

// updateFocus updates which field is currently focused.
func (m *formModel) updateFocus() {
	m.typ.Blur()
	m.summary.Blur()
	m.details.Blur()

	// NOTE: we *should not* have to update the insert-newline keybinding. As a
	// value struct, each textarea should have its own copy. And yet... they
	// seem to be conflated, for some reason.

	switch m.focusedField {
	case typeField:
		m.typ.Focus()
		m.typ.KeyMap.InsertNewline.SetEnabled(false)
		defaultKeyMap.NextSingle.SetEnabled(true)
		defaultKeyMap.NextMulti.SetEnabled(false)
	case summaryField:
		m.summary.Focus() // returned tea.Cmd ignored!
		m.summary.KeyMap.InsertNewline.SetEnabled(false)
		defaultKeyMap.NextSingle.SetEnabled(true)
		defaultKeyMap.NextMulti.SetEnabled(false)
	case detailsField:
		m.details.Focus() // returned tea.Cmd ignored!
		m.details.KeyMap.InsertNewline.SetEnabled(true)
		defaultKeyMap.NextSingle.SetEnabled(false)
		defaultKeyMap.NextMulti.SetEnabled(true)
	}
}

// nextField moves to the next field.
func (m *formModel) nextField() {
	m.focusedField = (m.focusedField + 1) % maxField
	m.updateFocus()
}

// prevField moves to the previous field.
func (m *formModel) prevField() {
	m.focusedField = (m.focusedField + maxField - 1) % maxField
	m.updateFocus()
}

// validateSummary checks if the summary field meets requirements.
func (m *formModel) validateSummary() error {
	// summary := strings.TrimSpace(m.summary.Value())
	// if summary == "" {
	// 	return fmt.Errorf("summary is required")
	// }
	// if len(summary) > m.config.SubjectMaxLength {
	// 	return fmt.Errorf("summary must be ≤ %d characters (currently %d)", m.config.SubjectMaxLength, len(summary))
	// }
	return nil
}

// View renders the TUI.
func (m formModel) View() tea.View {

	// TODO: use lipgloss.NewLayer() and compositor.Compose() to handle z-depth
	// rendering

	// Get the keymap of the focused model...
	var taKeyMap textarea.KeyMap

	switch m.focusedField {
	case typeField:
		taKeyMap = m.typ.KeyMap
	case summaryField:
		taKeyMap = m.summary.KeyMap
	case detailsField:
		taKeyMap = m.details.KeyMap
	}

	// ... we shouldn't recalc the keymap every view...
	keyMap := inTextareaKeyMap(taKeyMap)

	view := lipgloss.JoinVertical(
		lipgloss.Left,
		// "\n",
		// ruler...
		// "          1         2         3         4         5         6         7         8         9         0\n",
		// " 1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890",
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.typ.View(),
			m.summary.View(),
		),
		m.details.View(),
		m.help.View(keyMap), // &defaultKeyMap),

		// // debug config info
		// fmt.Sprintf(
		// 	"header=%d, subject=%d, body=%d, log=%s",
		// 	m.config.HeaderMaxLength,
		// 	m.config.SubjectMaxLength,
		// 	m.config.BodyMaxLineLength,
		// 	m.log,
		// ),
	)

	return tea.NewView(view)

}

// Result returns the composed commit message and any error.
func (m formModel) Result() (*commit.Message, error) {
	if m.err != nil {
		return nil, m.err
	}

	// if !m.done {
	// 	return nil, fmt.Errorf("form not completed")
	// }

	msg := &commit.Message{
		Type:        strings.TrimSpace(m.typ.Value()),
		Description: strings.TrimSpace(m.summary.Value()),
		Body:        strings.TrimSpace(m.details.Value()),
	}

	// if len(m.config.Types) > 0 && m.typeIndex >= 0 && m.typeIndex < len(m.config.Types) {
	// 	msg.Type = m.config.Types[m.typeIndex]
	// }

	return msg, nil
}
