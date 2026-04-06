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
	"github.com/jaredreisinger/committed/pkg/teautil"
)

type field int

const (
	typeField field = iota
	descriptionField
	bodyField
	// help?
	maxField
)

// formModel represents the TUI state for conventional commit composition.
type formModel struct {
	config       *config.Config
	existingMsg  *commit.Message // just for init, don't persist
	focusedField field
	models       teautil.Router[field, textareaShim]
	// typ          textarea.Model
	//scope
	// description   textarea.Model
	// body          textarea.Model
	bodyMaxHeight int
	help          help.Model
	typeIndex     int // for navigating type enum
	err           error
	log           string
}

func stdText(fullBorder bool) textarea.Model {
	t := textarea.New()
	t.SetStyles(textareaStyles(fullBorder))
	t.Prompt = ""
	t.ShowLineNumbers = false
	t.KeyMap.InsertNewline.SetEnabled(false)
	return t
}

// *sigh* ... textarea does not actually implement tea.Model correctly!
type textareaShim struct {
	textarea.Model
}

func (m textareaShim) Init() tea.Cmd { return nil }

func (m textareaShim) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	inner, cmd := m.Model.Update(msg)
	return textareaShim{inner}, cmd
}

func (m textareaShim) View() tea.View {
	return tea.NewView(m.Model.View())
}

// newModel creates a new TUI model with the given configuration and optional
// existing message.
func newModel(cfg *config.Config, existingMsg *commit.Message) formModel {
	typ := stdText(false)
	typ.Placeholder = "Select commit type"
	typ.SetHeight(1)
	typ.MaxHeight = 1

	description := stdText(false)
	description.Placeholder = "brief description of changes"
	description.SetHeight(1)
	description.MaxHeight = 1

	body := stdText(true)
	body.Placeholder = "Detailed description (optional)"
	body.KeyMap.InsertNewline.SetEnabled(true)
	// Border is *outside* the width, but prompt/line-numbers are inside, except
	// if there's *no* prompt or line numbers, the border is included in the
	// width. At least that what appears to happen. The code looks like it tries
	// to include margin+border+padding+content. 🤷🏼‍♂️
	body.DynamicHeight = true
	body.MinHeight = 10

	help := help.New()
	help.ShowAll = true
	help.SetWidth(80)

	m := formModel{
		config:       cfg,
		existingMsg:  existingMsg,
		focusedField: typeField,
		// typ:          typ,
		// description:  description,
		// body:         body,
		models: teautil.NewRouter(map[field]textareaShim{
			typeField:        {typ},
			descriptionField: {description},
			bodyField:        {body},
		}),
		help:      help,
		typeIndex: 0,
	}

	// Pre-populate fields if we have an existing message
	if existingMsg != nil {
		m = m.populateFromExisting()
	}

	// Set initial focus
	m = m.updateFocus()

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
func (m formModel) populateFromExisting() formModel {
	if m.existingMsg == nil {
		return m
	}

	// Set type
	if m.existingMsg.Type != "" {
		// Find the type in our config enum
		for i, t := range m.config.Types {
			if t == m.existingMsg.Type {
				m.typeIndex = i
				typ := m.models.MustGet(typeField)
				typ.SetValue(t)
				typ.MoveToBegin()
				m.models = m.models.Set(typeField, typ)
				break
			}
		}
	}

	if m.existingMsg.Description != "" {
		desc := m.models.MustGet(descriptionField)
		desc.SetValue(m.existingMsg.Description)
		desc.MoveToBegin()
		m.models = m.models.Set(descriptionField, desc)
	}

	if m.existingMsg.Body != "" {
		body := m.models.MustGet(bodyField)
		body.SetValue(m.existingMsg.Body)
		body.MoveToBegin()
		m.models = m.models.Set(bodyField, body)
	}

	return m
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

	// The message might be a wrapped message intended for a target model, *or*
	// it might be a generic/unhandled message that go to the focused model. We
	// de-duplicate that logic.
	targetModel := m.focusedField

	switch msgT := msg.(type) {
	case tea.WindowSizeMsg:
		m = m.resize(msgT.Width, msgT.Height)
		handled = true
	case tea.KeyPressMsg:
		handled = true // assume handled, set back to false in default case
		switch {
		case key.Matches(msgT, defaultKeyMap.Cancel):
			cmds = append(cmds, tea.Interrupt)

		case key.Matches(msgT, defaultKeyMap.NextSingle) ||
			key.Matches(msgT, defaultKeyMap.NextMulti):
			m = m.nextField()

		case key.Matches(msgT, defaultKeyMap.Prev):
			m = m.prevField()

		case key.Matches(msgT, defaultKeyMap.Submit):
			// TODO: validate?
			cmds = append(cmds, tea.Quit)
		default:
			handled = false
		}

	// case key.Matches(msg, defaultKeyMap.Enter):
	// 	// enter handling depends on focus...
	// 	if m.focusedField == bodyField {
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

	case teautil.WrappedMsg[int]:
		// forward the message to the appropriate model...
		targetModel = field(msgT.Key)
		msg = msgT.Msg
		// *don't* set handled!
	}

	// delegate other messages to the field/model with focus...
	if !handled {
		var cmd tea.Cmd
		m.models, cmd = m.models.Update(msg, targetModel)
		cmds = append(cmds, cmd)

		if targetModel == bodyField {
			body := m.models.MustGet(bodyField)
			if body.DynamicHeight && body.Height() >= m.bodyMaxHeight {
				body.DynamicHeight = false
				body.MaxHeight = 0
				body.SetHeight(m.bodyMaxHeight)
				m.models = m.models.Set(bodyField, body)
			}
		}
	}

	// // Clear any previous validation errors when user types
	// if m.err != nil && (m.focusedField == summaryField || m.focusedField == detailsField) {
	// 	m.err = nil
	// }

	return m, tea.Batch(cmds...)
}

func (m formModel) resize(width int, height int) formModel {
	m.log = fmt.Sprintf("%dx%d", width, height)

	// Border is *outside* the width, but prompt/line-numbers are inside, except
	// if there's *no* prompt or line numbers, the border is included in the
	// width. At least that what appears to happen. The code looks like it tries
	// to include margin+border+padding+content. 🤷🏼‍♂️

	typ := m.models.MustGet(typeField)
	desc := m.models.MustGet(descriptionField)
	body := m.models.MustGet(bodyField)

	xx, yy := typ.Styles().Focused.Base.GetFrameSize()
	cfg := m.config

	maxTyp := maxTypeLength(cfg)
	typ.CharLimit = maxTyp
	typ.SetWidth(maxTyp + xx + 1) // +1 for cursor

	subLimit := min(cfg.SubjectMaxLength, cfg.HeaderMaxLength-maxTyp)
	desc.CharLimit = subLimit
	desc.SetWidth(min(subLimit+xx+1, width-typ.Width())) // +1 for cursor

	body.SetWidth(min(cfg.BodyMaxLineLength+xx, width))
	helpHeight := 8
	logHeight := 1
	m.bodyMaxHeight = height - typ.Height() - yy - yy - helpHeight - logHeight - 1
	body.MaxHeight = m.bodyMaxHeight

	if body.DynamicHeight && body.Height() >= m.bodyMaxHeight {
		body.DynamicHeight = false
		body.MaxHeight = 0
		body.SetHeight(m.bodyMaxHeight)
	}

	// it would be nice to switch back to dynamic if the content shrinks, but
	// there's no easy way to get "how many visual lines are needed?"

	m.help.SetWidth(xx)

	m.models = m.models.Set(typeField, typ)
	m.models = m.models.Set(descriptionField, desc)
	m.models = m.models.Set(bodyField, body)

	return m
}

// updateFocus updates which field is currently focused.
func (m formModel) updateFocus() formModel {
	for key := range m.models.Keys() {
		f := m.models.MustGet(key)
		if key == m.focusedField {
			f.Focus()
		} else {
			f.Blur()
		}
		m.models = m.models.Set(key, f)
	}

	// NOTE: defaultKeyMap is a global, which isn't really the Elm Architecture
	// way...
	if m.focusedField == bodyField {
		defaultKeyMap.NextSingle.SetEnabled(false)
		defaultKeyMap.NextMulti.SetEnabled(true)
	} else {
		defaultKeyMap.NextSingle.SetEnabled(true)
		defaultKeyMap.NextMulti.SetEnabled(false)
	}

	return m
}

// nextField moves to the next field.
func (m formModel) nextField() formModel {
	m.focusedField = (m.focusedField + 1) % maxField
	return m.updateFocus()
}

// prevField moves to the previous field.
func (m formModel) prevField() formModel {
	m.focusedField = (m.focusedField + maxField - 1) % maxField
	return m.updateFocus()
}

// validateDescription checks if the description field meets requirements.
func (m formModel) validateDescription() error {
	// description := strings.TrimSpace(desc.Value())
	// if description == "" {
	// 	return fmt.Errorf("description is required")
	// }
	// if len(description) > m.config.SubjectMaxLength {
	// 	return fmt.Errorf("description must be ≤ %d characters (currently %d)", m.config.SubjectMaxLength, len(description))
	// }
	return nil
}

// View renders the TUI.
func (m formModel) View() tea.View {
	typ := m.models.MustGet(typeField)
	desc := m.models.MustGet(descriptionField)
	body := m.models.MustGet(bodyField)

	// TODO: use lipgloss.NewLayer() and compositor.Compose() to handle z-depth
	// rendering

	// Get the keymap of the focused model...
	var taKeyMap textarea.KeyMap

	switch m.focusedField {
	case typeField:
		taKeyMap = typ.KeyMap
	case descriptionField:
		taKeyMap = desc.KeyMap
	case bodyField:
		taKeyMap = body.KeyMap
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
			typ.View().Content,
			desc.View().Content,
		),
		body.View().Content,
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
	typ := m.models.MustGet(typeField)
	desc := m.models.MustGet(descriptionField)
	body := m.models.MustGet(bodyField)

	msg := &commit.Message{
		Type:        strings.TrimSpace(typ.Value()),
		Description: strings.TrimSpace(desc.Value()),
		Body:        strings.TrimSpace(body.Value()),
	}

	// if len(m.config.Types) > 0 && m.typeIndex >= 0 && m.typeIndex < len(m.config.Types) {
	// 	msg.Type = m.config.Types[m.typeIndex]
	// }

	return msg, nil
}
