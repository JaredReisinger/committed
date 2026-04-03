package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaredreisinger/committed/internal/config"
	"github.com/jaredreisinger/committed/pkg/commit"
)

type field int

const (
	typeField field = iota
	summaryField
	detailsField
)

// Model represents the TUI state for conventional commit composition.
type Model struct {
	config       *config.Config
	existingMsg  *commit.Message
	focusedField field
	typeInput    textinput.Model
	summaryInput textinput.Model
	detailsInput textinput.Model
	typeIndex    int // for navigating type enum
	err          error
	done         bool
}

// NewModel creates a new TUI model with the given configuration and optional existing message.
func NewModel(cfg *config.Config, existing *commit.Message) Model {
	ti := textinput.New()
	ti.Placeholder = "Select commit type"
	ti.CharLimit = 50
	ti.Width = 50

	si := textinput.New()
	si.Placeholder = "Brief description of changes"
	si.CharLimit = cfg.SubjectMaxLength
	si.Width = 72

	di := textinput.New()
	di.Placeholder = "Detailed description (optional)"
	di.Width = 72

	m := Model{
		config:       cfg,
		existingMsg:  existing,
		focusedField: typeField,
		typeInput:    ti,
		summaryInput: si,
		detailsInput: di,
		typeIndex:    0,
	}

	// Pre-populate fields if we have an existing message
	if existing != nil {
		m.populateFromExisting()
	}

	// Set initial focus
	m.updateFocus()

	return m
}

// populateFromExisting fills the form fields from an existing commit message.
func (m *Model) populateFromExisting() {
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
		m.summaryInput.SetValue(m.existingMsg.Description)
	}

	// Set details (body)
	if m.existingMsg.Body != "" {
		m.detailsInput.SetValue(m.existingMsg.Body)
	}
}

// updateFocus updates which field is currently focused.
func (m *Model) updateFocus() {
	m.typeInput.Blur()
	m.summaryInput.Blur()
	m.detailsInput.Blur()

	switch m.focusedField {
	case typeField:
		// Type field uses custom navigation, not text input
	case summaryField:
		m.summaryInput.Focus()
	case detailsField:
		m.detailsInput.Focus()
	}
}

// nextField moves to the next field.
func (m *Model) nextField() {
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
func (m *Model) prevField() {
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
func (m *Model) validateSummary() error {
	summary := strings.TrimSpace(m.summaryInput.Value())
	if summary == "" {
		return fmt.Errorf("summary is required")
	}
	if len(summary) > m.config.SubjectMaxLength {
		return fmt.Errorf("summary must be ≤ %d characters (currently %d)", m.config.SubjectMaxLength, len(summary))
	}
	return nil
}

// Init initializes the bubbletea program.
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles user input and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.err = fmt.Errorf("user cancelled")
			m.done = true
			return m, tea.Quit

		case "tab", "shift+tab":
			if msg.String() == "shift+tab" {
				m.prevField()
			} else {
				m.nextField()
			}
			return m, nil

		case "enter":
			switch m.focusedField {
			case typeField:
				m.nextField()
				return m, nil
			case summaryField:
				if err := m.validateSummary(); err != nil {
					m.err = err
					return m, nil
				}
				m.nextField()
				return m, nil
			case detailsField:
				// Submit the form
				if err := m.validateSummary(); err != nil {
					m.err = err
					return m, nil
				}
				m.done = true
				return m, tea.Quit
			}

		case "up", "down":
			if m.focusedField == typeField {
				if msg.String() == "up" {
					if m.typeIndex > 0 {
						m.typeIndex--
					}
				} else {
					if m.typeIndex < len(m.config.Types)-1 {
						m.typeIndex++
					}
				}
				return m, nil
			}
		}
	}

	// Handle text input updates
	var cmd tea.Cmd
	if m.focusedField == summaryField {
		m.summaryInput, cmd = m.summaryInput.Update(msg)
	} else if m.focusedField == detailsField {
		m.detailsInput, cmd = m.detailsInput.Update(msg)
	}

	// Clear any previous validation errors when user types
	if m.err != nil && (m.focusedField == summaryField || m.focusedField == detailsField) {
		m.err = nil
	}

	return m, cmd
}

// View renders the TUI.
func (m Model) View() string {
	var b strings.Builder

	b.WriteString("Compose Conventional Commit\n\n")

	// Type field
	b.WriteString("Type: ")
	if m.focusedField == typeField {
		b.WriteString("[")
		if len(m.config.Types) > 0 {
			b.WriteString(m.config.Types[m.typeIndex])
		}
		b.WriteString("]")
	} else {
		if len(m.config.Types) > 0 {
			b.WriteString(m.config.Types[m.typeIndex])
		}
	}
	b.WriteString(" (↑/↓ to navigate, Tab to next)\n\n")

	// Summary field
	b.WriteString("Summary: ")
	b.WriteString(m.summaryInput.View())
	charCount := len(m.summaryInput.Value())
	b.WriteString(fmt.Sprintf(" (%d/%d chars)", charCount, m.config.SubjectMaxLength))
	b.WriteString("\n\n")

	// Details field
	b.WriteString("Details: \n")
	b.WriteString(m.detailsInput.View())
	b.WriteString("\n\n")

	// Error display
	if m.err != nil {
		b.WriteString(fmt.Sprintf("Error: %s\n\n", m.err.Error()))
	}

	// Help text
	b.WriteString("Tab/Shift+Tab: navigate • Enter: next/submit • Esc/Ctrl+C: cancel")

	return b.String()
}

// Result returns the composed commit message and any error.
func (m Model) Result() (*commit.Message, error) {
	if m.err != nil {
		return nil, m.err
	}

	if !m.done {
		return nil, fmt.Errorf("form not completed")
	}

	msg := &commit.Message{
		Type:        "",
		Description: strings.TrimSpace(m.summaryInput.Value()),
		Body:        strings.TrimSpace(m.detailsInput.Value()),
	}

	if len(m.config.Types) > 0 && m.typeIndex >= 0 && m.typeIndex < len(m.config.Types) {
		msg.Type = m.config.Types[m.typeIndex]
	}

	return msg, nil
}
