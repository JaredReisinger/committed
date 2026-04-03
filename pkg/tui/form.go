package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
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
)

// formModel represents the TUI state for conventional commit composition.
type formModel struct {
	config       *config.Config
	existingMsg  *commit.Message // just for init, don't persist
	focusedField field
	typ          textinput.Model
	summary      textinput.Model
	details      textinput.Model
	help         help.Model
	typeIndex    int // for navigating type enum
	err          error
}

// newModel creates a new TUI model with the given configuration and optional existing message.
func newModel(cfg *config.Config, existingMsg *commit.Message) formModel {
	// TODO: get actual width?
	typ := textinput.New()
	typ.Placeholder = "Select commit type"
	typ.CharLimit = 50
	typ.SetWidth(10)

	summary := textinput.New()
	summary.Placeholder = "Brief description of changes"
	summary.CharLimit = cfg.SubjectMaxLength
	summary.SetWidth(70)

	details := textinput.New()
	details.Placeholder = "Detailed description (optional)"
	details.SetWidth(80)

	help := help.New()
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

// updateFocus updates which field is currently focused.
func (m *formModel) updateFocus() {
	m.typ.Blur()
	m.summary.Blur()
	m.details.Blur()

	switch m.focusedField {
	case typeField:
		// Type field uses custom navigation, not text input
	case summaryField:
		m.summary.Focus() // returned tea.Cmd ignored!
	case detailsField:
		m.details.Focus() // returned tea.Cmd ignored!
	}
}

// nextField moves to the next field.
func (m *formModel) nextField() {
	switch m.focusedField {
	case typeField:
		m.focusedField = summaryField
	case summaryField:
		m.focusedField = detailsField
	case detailsField:
		// Stay on details field
	}
	m.updateFocus()
}

// prevField moves to the previous field.
func (m *formModel) prevField() {
	switch m.focusedField {
	case summaryField:
		m.focusedField = typeField
	case detailsField:
		m.focusedField = summaryField
	case typeField:
		// Stay on type field
	}
	m.updateFocus()
}

// validateSummary checks if the summary field meets requirements.
func (m *formModel) validateSummary() error {
	summary := strings.TrimSpace(m.summary.Value())
	if summary == "" {
		return fmt.Errorf("summary is required")
	}
	if len(summary) > m.config.SubjectMaxLength {
		return fmt.Errorf("summary must be ≤ %d characters (currently %d)", m.config.SubjectMaxLength, len(summary))
	}
	return nil
}

// Init initializes the bubbletea program.
func (m formModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles user input and updates the model.
func (m formModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	handled := false // do we do this, or msg = nil?

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		handled = true // assume handled, set back to false in default case
		switch {
		case key.Matches(msg, defaultKeyMap.Cancel):
			cmds = append(cmds, tea.Interrupt)

		case key.Matches(msg, defaultKeyMap.Next):
			m.nextField()

		case key.Matches(msg, defaultKeyMap.Prev):
			m.prevField()

		case key.Matches(msg, defaultKeyMap.Enter):
			// enter handling depends on focus...
			if m.focusedField == detailsField {
				// TODO: change to validate as user types?
				// Submit the form
				if err := m.validateSummary(); err != nil {
					m.err = err
				} else {
					m.err = nil // clear any error?
					cmds = append(cmds, tea.Quit)
				}
			} else {
				m.nextField()
			}

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
		// case typeField: // no model?
		case summaryField:
			m.summary, cmd = m.summary.Update(msg)
		case detailsField:
			m.details, cmd = m.details.Update(msg)
		}

		cmds = append(cmds, cmd)
	}

	// // Clear any previous validation errors when user types
	// if m.err != nil && (m.focusedField == summaryField || m.focusedField == detailsField) {
	// 	m.err = nil
	// }

	return m, tea.Batch(cmds...)
}

// View renders the TUI.
func (m formModel) View() tea.View {

	// get each sub-view...
	typ := fmt.Sprintf("[%s]", m.config.Types[m.typeIndex])

	view := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			typ,
			m.summary.View(),
		),
		"\n",
		m.details.View(),
		"\n",
		m.help.View(&defaultKeyMap),
		"\n",
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
		Type:        "",
		Description: strings.TrimSpace(m.summary.Value()),
		Body:        strings.TrimSpace(m.details.Value()),
	}

	if len(m.config.Types) > 0 && m.typeIndex >= 0 && m.typeIndex < len(m.config.Types) {
		msg.Type = m.config.Types[m.typeIndex]
	}

	return msg, nil
}
