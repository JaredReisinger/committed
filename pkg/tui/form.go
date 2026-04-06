package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/jaredreisinger/committed/internal/config"
	"github.com/jaredreisinger/committed/pkg/commit"
	"github.com/jaredreisinger/committed/pkg/teautil"
)

// field is kind of a hybrid child model ID and tab-order value
type field int

const (
	// this is also the tab order
	typeField field = iota
	scopeField
	descriptionField
	bodyField
	footerField
	// help?
	maxField
)

// form represents the TUI state for conventional commit composition.
//
// It includes both local state and also the child models (which are, in a
// sense, also a kind of local state.)   The original AI-created class was an
// absolute mess, with no sense of immutability at all.  The addition of
// [teautil.Router] is helping a ton, however.
type form struct {
	config       *config.Config
	existingMsg  *commit.Message // just for init, don't persist
	focusedField field
	texts        teautil.Router[field, textModel]
	// textinputs    teautil.Router[field, textinputModel]
	// textareas     teautil.Router[field, textareaModel]
	bodyMaxHeight int
	help          help.Model
	typeIndex     int // for navigating type enum
	err           error
	log           string
}

// newModel creates a new TUI model with the given configuration and optional
// existing message.
func newModel(cfg *config.Config, existingMsg *commit.Message) form {
	// get existing message content...
	var initialType string
	var initialScope string
	var initialDesc string
	var initialBody string
	var initialFooter string

	if existingMsg != nil {
		initialType = existingMsg.Type
		initialScope = existingMsg.Scope
		initialDesc = existingMsg.Description
		initialBody = existingMsg.Body
	}

	help := help.New()
	help.ShowAll = true
	help.SetWidth(80)

	m := form{
		config:       cfg,
		existingMsg:  existingMsg,
		focusedField: typeField,

		texts: teautil.NewRouter(map[field]textModel{
			typeField:        newTextModel(false, "type", initialType),
			scopeField:       newTextModel(false, "scope", initialScope),
			descriptionField: newTextModel(false, "description", initialDesc),
			bodyField:        newTextModel(true, "message body", initialBody),
			footerField:      newTextModel(true, "footer", initialFooter),
		}),

		help: help,

		typeIndex: 0, // ?
	}

	// // TODO: this should happen in Init!
	// //
	// // Pre-populate fields if we have an existing message
	// if existingMsg != nil {
	// 	m = m.populateFromExisting()
	// }

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

// // populateFromExisting fills the form fields from an existing commit message.
// func (m form) populateFromExisting() form {
// 	if m.existingMsg == nil {
// 		return m
// 	}

// 	// // Set type
// 	// if m.existingMsg.Type != "" {
// 	// 	// Find the type in our config enum
// 	// 	for i, t := range m.config.Types {
// 	// 		if t == m.existingMsg.Type {
// 	// 			m.typeIndex = i
// 	// 			typ := m.textareas.MustGet(typeField)
// 	// 			typ.SetValue(t)
// 	// 			typ.MoveToBegin()
// 	// 			m.textareas = m.textareas.Set(typeField, typ)
// 	// 			break
// 	// 		}
// 	// 	}
// 	// }

// 	// if m.existingMsg.Description != "" {
// 	// 	desc := m.textareas.MustGet(descriptionField)
// 	// 	desc.SetValue(m.existingMsg.Description)
// 	// 	desc.MoveToBegin()
// 	// 	m.textareas = m.textareas.Set(descriptionField, desc)
// 	// }

// 	// if m.existingMsg.Body != "" {
// 	// 	body := m.textareas.MustGet(bodyField)
// 	// 	body.SetValue(m.existingMsg.Body)
// 	// 	body.MoveToBegin()
// 	// 	m.textareas = m.textareas.Set(bodyField, body)
// 	// }

// 	return m
// }

// Init initializes the bubbletea program.
func (m form) Init() tea.Cmd {
	// should other processing happen here?
	// return textinput.Blink
	// Set initial focus
	return setFocusCmd(typeField)
}

type setFocusMsg struct {
	field
}

func setFocusCmd(f field) tea.Cmd {
	return func() tea.Msg {
		return setFocusMsg{field: f}
	}
}

// Update handles user input and updates the model.
func (m form) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd // aggregated commands
	var cmd tea.Cmd
	handled := false // do we do this, or msg = nil?

	// // The message might be a wrapped message intended for a target model, *or*
	// // it might be a generic/unhandled message that go to the focused model. We
	// // de-duplicate that logic.
	// targetModel := m.focusedField

	switch msgT := msg.(type) {
	case tea.WindowSizeMsg:
		m = m.resize(msgT.Width, msgT.Height)
		handled = true
	case setFocusMsg:
		m, cmd = m.setFocus(msgT.field)
		cmds = append(cmds, cmd)
	case tea.KeyPressMsg:
		handled = true // assume handled, set back to false in default case
		switch {
		case key.Matches(msgT, defaultKeyMap.Cancel):
			cmds = append(cmds, tea.Interrupt)

		case key.Matches(msgT, defaultKeyMap.NextSingle) ||
			key.Matches(msgT, defaultKeyMap.NextMulti):
			cmds = append(cmds, m.nextField())

		case key.Matches(msgT, defaultKeyMap.Prev):
			cmds = append(cmds, m.prevField())

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

		// case teautil.WrappedMsg[int]:
		// 	// forward the message to the appropriate model...
		// 	targetModel = field(msgT.Key)
		// 	msg = msgT.Msg
		// 	// *don't* set handled!
	}

	// delegate other messages to the field/model with focus...
	if !handled {
		var cmd tea.Cmd
		m.texts, cmd = m.texts.Update(msg, m.focusedField)
		// m.textareas, cmd = m.textareas.Update(msg, targetModel)
		cmds = append(cmds, cmd)

		// If the focused field is the body, check for dynamic->fixed size
		// changing
		if m.focusedField == bodyField {
			body := m.texts.MustGet(bodyField)
			if body.DynamicHeight() && body.Height() >= m.bodyMaxHeight {
				body = body.SetDynamicHeight(false)
				body = body.SetMaxHeight(0)
				body = body.SetHeight(m.bodyMaxHeight)
				m.texts = m.texts.Set(bodyField, body)
			}
		}
	}

	// // Clear any previous validation errors when user types
	// if m.err != nil && (m.focusedField == summaryField || m.focusedField == detailsField) {
	// 	m.err = nil
	// }

	return m, tea.Batch(cmds...)
}

func (m form) resize(width int, height int) form {
	m.log = fmt.Sprintf("%dx%d", width, height)

	// We're going to manage the field borders separate from the text controls
	// because textarea has some rendering challenges with doubling borders with
	// an empty field.  For now, we're calc'ing 2x and 2y for borders...
	// bw := 2
	bh := 2

	typ := m.texts.MustGet(typeField)
	scope := m.texts.MustGet(scopeField)
	desc := m.texts.MustGet(descriptionField)
	body := m.texts.MustGet(bodyField)
	footer := m.texts.MustGet(footerField)

	cfg := m.config

	maxTyp := maxTypeLength(cfg)
	typ = typ.SetCharLimit(maxTyp)
	typ = typ.SetWidth(maxTyp + 1) // +1 for cursor

	scope = scope.SetWidth(10)

	descLimit := min(cfg.SubjectMaxLength, cfg.HeaderMaxLength-maxTyp)
	desc = desc.SetCharLimit(descLimit)
	desc = desc.SetWidth(min(descLimit+1, width-typ.Width())) // +1 for cursor

	body = body.SetWidth(min(cfg.BodyMaxLineLength, width))
	helpHeight := 8
	logHeight := 1
	m.bodyMaxHeight = height - typ.Height() - bh - bh - helpHeight - logHeight - 1
	body = body.SetMaxHeight(m.bodyMaxHeight)

	if body.DynamicHeight() && body.Height() >= m.bodyMaxHeight {
		body = body.SetDynamicHeight(false)
		body = body.SetMaxHeight(0)
		body = body.SetHeight(m.bodyMaxHeight)
	}
	// it would be nice to switch back to dynamic if the content shrinks, but
	// there's no easy way to get "how many visual lines are needed?"

	// not sizing footer yet!

	m.help.SetWidth(width)

	m.texts = m.texts.SetMap(map[field]textModel{
		typeField:        typ,
		scopeField:       scope,
		descriptionField: desc,
		bodyField:        body,
		footerField:      footer,
	})

	return m
}

// updateFocus updates which field is currently focused.
func (m form) setFocus(focusField field) (form, tea.Cmd) {
	modelMap := map[field]textModel{}
	var cmds []tea.Cmd
	var t2 textModel
	var cmd tea.Cmd

	// instead of a full loop, we could just blur the previous field!
	for f := range maxField {
		if t, ok := m.texts.Get(f); ok {
			if f == focusField {
				t2, cmd = t.Focus()
			} else {
				t2, cmd = t.Blur()
			}
			modelMap[f] = t2
			cmds = append(cmds, teautil.Wrap(cmd, f))
		}
	}
	m.texts = m.texts.SetMap(modelMap)
	m.focusedField = focusField

	// NOTE: defaultKeyMap is a global, which isn't really the Elm Architecture
	// way...
	if m.texts.MustGet(m.focusedField).isArea {
		defaultKeyMap.NextSingle.SetEnabled(false)
		defaultKeyMap.NextMulti.SetEnabled(true)
	} else {
		defaultKeyMap.NextSingle.SetEnabled(true)
		defaultKeyMap.NextMulti.SetEnabled(false)
	}

	return m, tea.Batch(cmds...)
}

// nextField moves to the next field.
func (m form) nextField() tea.Cmd {
	return setFocusCmd((m.focusedField + 1) % maxField)
}

// prevField moves to the previous field.
func (m form) prevField() tea.Cmd {
	return setFocusCmd((m.focusedField + maxField - 1) % maxField)
}

// validateDescription checks if the description field meets requirements.
func (m form) validateDescription() error {
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
func (m form) View() tea.View {

	// get all the child model views (Router helper?)
	views := make(map[field]tea.View, m.texts.Len())

	for f, t := range m.texts.All() {
		views[f] = t.View()
	}

	// get the focused field key bindings...
	textKeyBindings := m.texts.MustGet(m.focusedField).GetKeyBindings()

	// It seems ridiculous to recalculate the keymap on every single view, but
	// with an immutable pattern, we *can't* know if we have the same value call
	// after call. On the plus side, this should be a very fast calculation (and
	// maybe we can memoize the values and keep a pre-rendered view?)
	helpKeys := buildHelpKeys(textKeyBindings)

	// TODO: use lipgloss.NewLayer() and compositor.Compose() to handle z-depth
	// rendering

	view := lipgloss.JoinVertical(
		lipgloss.Left,
		// "\n",
		// ruler...
		// "          1         2         3         4         5         6         7         8         9         0\n",
		// " 1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890",
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			views[typeField].Content,
			"(",
			views[scopeField].Content,
			"):",
			views[descriptionField].Content,
		),
		views[bodyField].Content,
		views[footerField].Content,
		m.help.View(helpKeys), // &defaultKeyMap),

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
func (m form) Result() (*commit.Message, error) {
	if m.err != nil {
		return nil, m.err
	}

	texts := make(map[field]string, m.texts.Len())

	for f, t := range m.texts.All() {
		texts[f] = strings.TrimSpace(t.Value())
	}

	// if !m.done {
	// 	return nil, fmt.Errorf("form not completed")
	// }
	msg := &commit.Message{
		Type:        texts[typeField],
		Scope:       texts[scopeField],
		Description: texts[descriptionField],
		Body:        texts[bodyField],
		// Footer:      texts[footerField], // a slice?
	}

	// if len(m.config.Types) > 0 && m.typeIndex >= 0 && m.typeIndex < len(m.config.Types) {
	// 	msg.Type = m.config.Types[m.typeIndex]
	// }

	return msg, nil
}
