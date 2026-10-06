package tui

import (
	"log/slog"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/jaredreisinger/committed/pkg/commit"
	"github.com/jaredreisinger/committed/pkg/config"
	"github.com/jaredreisinger/committed/pkg/teautil"
)

// field is kind of a hybrid child model ID and tab-order value
type field int

const (
	// this is also the tab order
	typeField field = iota
	typeList
	scopeField
	descriptionField
	bodyField
	footerField
	// help?
	maxField
)

// mainForm represents the TUI state for conventional commit composition.
//
// It includes both local state and also the child models (which are, in a
// sense, also a kind of local state.)   The original AI-created class was an
// absolute mess, with no sense of immutability at all.  The addition of
// [teautil.Router] is helping a ton, however.
type mainForm struct {
	config        *config.Config
	existingMsg   *commit.Message // just for init, don't persist
	focusedField  field
	children      teautil.Router[field, tea.Model] // *focusable* children
	fieldErrs     map[field]error                  // field validation errors
	winWidth      int
	winHeight     int
	bodyMaxHeight int
	help          help.Model

	limits FieldVals[Limits]

	err error // should this be validation error(s) as well?
}

// newModel creates a new TUI model with the given configuration and optional
// existing message.
func newModel(cfg *config.Config, existingMsg *commit.Message) mainForm {
	// get existing message content...
	var initialType string
	var initialScope string
	var initialDesc string
	var initialBody string
	var initialFooter string

	// Do we really want to set initial values here, or is Init() better for
	// that?
	if existingMsg != nil {
		initialType = existingMsg.Type
		initialScope = existingMsg.Scope
		initialDesc = existingMsg.Description
		initialBody = existingMsg.Body
	}

	help := help.New()
	help.ShowAll = true
	help.SetWidth(80) // ??

	// get field limits from configuration
	limits := FieldVals[Limits]{
		typeWidth:    Limits{cfg.TypeMinLength.GetValue(0), cfg.TypeMaxLength.GetValue(0)},
		scopeWidth:   Limits{cfg.ScopeMinLength.GetValue(0), cfg.ScopeMaxLength.GetValue(0)},
		descWidth:    Limits{cfg.SubjectMinLength.GetValue(0), cfg.SubjectMaxLength.GetValue(0)},
		bodyWidth:    Limits{cfg.BodyMinLength.GetValue(0), cfg.BodyMaxLineLength.GetValue(0)},
		bodyHeight:   Limits{4, 0}, // ?? really?
		footerWidth:  Limits{cfg.FooterMinLength.GetValue(0), cfg.FooterMaxLineLength.GetValue(0)},
		footerHeight: Limits{1, 0}, // ?? really?
	}

	form := mainForm{
		config:       cfg,
		existingMsg:  existingMsg,
		focusedField: typeList,

		// There's an interesting question about whether the form's fields
		// should be *all* models or just focusable ones.  Perhaps `children`
		// should be *all* models (including help), and "focusableFields" should
		// be a separate list that we use to manage focus.  That would make
		// message passing more consistent and extensible.  (Just imagine if the
		// "help" started using/needing messages as opposed to being a
		// pure-display component.)

		children: teautil.NewRouter(map[field]tea.Model{
			typeField:        newTextModel(false, "type", initialType).SetCharLimit(limits.typeWidth.max),
			typeList:         newListModel(cfg.Rules.TypeEnum.Value),
			scopeField:       newTextModel(false, "scope", initialScope).SetCharLimit(limits.scopeWidth.max),
			descriptionField: newTextModel(false, "description", initialDesc).SetCharLimit(limits.descWidth.max),
			bodyField:        newTextModel(true, "message body", initialBody).SetDynamicHeight(true), //.SetMinHeight(6).SetDynamicHeight(true).SetMaxContentHeight(1000),
			footerField:      newTextModel(true, "footer", initialFooter).SetDynamicHeight(true),     //.SetMinHeight(2).SetDynamicHeight(true),
		}),

		fieldErrs: make(map[field]error, maxField),

		help:   help,
		limits: limits,
	}

	// perform initial validation?
	for f, child := range form.children.All() {
		t, ok := child.(textModel)
		if ok {
			form = form.validate(f, t.Value())
		}
	}

	return form
}

// func maxTypeLength(cfg *config.Config) int {
// 	if cfg.Rules.TypeMaxLength.Active() {
// 		return cfg.Rules.TypeMaxLength.Value
// 	}

// 	if cfg.Rules.TypeEnum.Set {
// 		maxTyp := 0
// 		for _, t := range cfg.Rules.TypeEnum.Value {
// 			w := len(t)
// 			if w > maxTyp {
// 				maxTyp = w
// 			}
// 		}
// 		return maxTyp
// 	}

// 	return 10 // ?
// }

// Init initializes the bubbletea program.
func (form mainForm) Init() tea.Cmd {
	// should other processing happen here?
	// return textinput.Blink
	// Set initial focus
	return setFocusCmd(typeList)
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
func (form mainForm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd // aggregated commands
	var cmd tea.Cmd
	handled := false // do we do this, or msg = nil?

	// // The message might be a wrapped message intended for a target model, *or*
	// // it might be a generic/unhandled message that go to the focused model. We
	// // de-duplicate that logic.
	// targetModel := m.focusedField

	switch msgT := msg.(type) {
	case tea.WindowSizeMsg:
		form = form.resize(msgT.Width, msgT.Height)
		handled = true

	case setFocusMsg:
		// If the focus *was* type or scope, we need to play some positioning
		// tricks to ensure the field isn't scrolled weirdly.

		// switch form.focusedField {
		// case typeField, scopeField:
		// 	t2 := form.children.MustGet(form.focusedField).(textModel).MoveToBegin().MoveToEnd()
		// 	form.children = form.children.Set(form.focusedField, t2)
		// }

		form, cmd = form.setFocus(msgT.field)
		cmds = append(cmds, cmd)
		// we also need to resize fields as focus changes...
		form = form.resize(form.winWidth, form.winHeight)
		handled = true

	case teautil.WrappedMsg[field]:
		switch innerMsg := msgT.Msg.(type) {
		case listSelectionChangedMsg:
			if msgT.Key == typeList {
				// The MoveToBegin/MoveToEnd ensures we don't have a weirdly
				// scrolled field afterwards
				t2 := form.children.MustGet(typeField).(textModel).SetValue(innerMsg.selectedItem).MoveToBegin()
				form.children = form.children.Set(typeField, t2)
				// changing the type *can* change the layout
				form = form.resize(form.winWidth, form.winHeight)
				t2 = form.children.MustGet(typeField).(textModel).MoveToEnd()
				form.children = form.children.Set(typeField, t2)
				handled = true

				// We also need to fake a "field changed" message!
				cmds = append(cmds, teautil.Wrap(SendTextModelValueChanged(t2.Value()), typeField))
			}

		case textModelValueChanged:
			slog.Debug("text model value changed", "field", msgT.Key, "value", innerMsg.Value)
			form = form.validate(msgT.Key, innerMsg.Value)
			handled = true // we don't let this percolate back down, do we?
		}

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

		case key.Matches(msgT, defaultKeyMap.NextSingle) ||
			key.Matches(msgT, defaultKeyMap.NextMulti):
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

		// if form.focusedField == typeField {
		// 	var typeList tea.Model
		// 	typeList, cmd = form.typeList.Update(msg)
		// 	form.typeList = typeList.(listModel)
		// } else {
		form.children, cmd = form.children.Update(msg, form.focusedField)
		// }

		// m.textareas, cmd = m.textareas.Update(msg, targetModel)
		cmds = append(cmds, cmd)

		// TODO: the resize logic should be in control of returning the "should
		// the body be dynamic or not" values.
		//
		// If the focused field is the body, check for dynamic->fixed size
		// changing
		if form.focusedField == bodyField {
			body := form.children.MustGet(bodyField).(textModel)
			if body.DynamicHeight() && body.Height() >= form.bodyMaxHeight {
				body = body.SetDynamicHeight(false)
				// body = body.SetMaxHeight(0)
				body = body.SetHeight(form.bodyMaxHeight)
			} else if body.Height() > form.bodyMaxHeight {
				body = body.SetHeight(form.bodyMaxHeight)
			} else {
				// bubbletea text doesn't expose its internal visual height value
				textHeight := strings.Count(commit.Wrap(body.Value(), body.Width()), "\n")
				body = body.SetDynamicHeight(textHeight < form.bodyMaxHeight)
				// body = body.SetHeight(max(6, textHeight))
				body = body.SetMaxHeight(form.bodyMaxHeight)
			}
			form.children = form.children.Set(bodyField, body)
		}
	}

	// // Clear any previous validation errors when user types
	// if m.err != nil && (m.focusedField == summaryField || m.focusedField == detailsField) {
	// 	m.err = nil
	// }

	return form, tea.Batch(cmds...)
}

// Field sizes and presentation needs to get calculated *often*, because we want
// it to be responsive to content and to focus. At the same time, we don't want
// the sizing logic to be scattered all over the code.

func (form mainForm) resize(winWidth int, winHeight int) mainForm {
	slog.Debug("window resize", "width", winWidth, "height", winHeight)

	form.winWidth = winWidth
	form.winHeight = winHeight

	typ := form.children.MustGet(typeField).(textModel)
	// TODO: size typeList field here as well!
	scope := form.children.MustGet(scopeField).(textModel)
	desc := form.children.MustGet(descriptionField).(textModel)
	body := form.children.MustGet(bodyField).(textModel)
	footer := form.children.MustGet(footerField).(textModel)

	// cfg := form.config

	// Set up the calculateSizes structure... could we cache a lot of this?
	// Might need to alter the constraints to not include current content size.
	// NOTE... no good way to get count of current body lines...
	contents := FieldVals[Contents]{
		typeWidth:    Contents{nominalLength(typ, true)},
		scopeWidth:   Contents{nominalLength(scope, form.focusedField == scopeField)},
		descWidth:    Contents{nominalLength(desc, true)},
		bodyWidth:    Contents{nominalLength(body, true)},
		bodyHeight:   Contents{0},
		footerWidth:  Contents{nominalLength(footer, true)},
		footerHeight: Contents{0},
	}

	// // shoehorn in max type size... this should happen differently...
	// if constraints.typeWidth.max == -1 {
	// 	constraints.typeWidth.max = maxTypeLength(cfg)
	// }

	bounds := Size{form.winWidth, form.winHeight}
	sizes := calculateSizes(bounds, form.limits, form.focusedField, contents)
	slog.Info("resize",
		"bounds", bounds.LogValue(),
		"limits", form.limits.LogValue(),
		"contents", contents.LogValue(),
		"sizes", sizes.LogValue(),
	)

	typ = typ.SetMinWidth(sizes.typeWidth.min).SetMaxWidth(sizes.typeWidth.max)
	scope = scope.SetMinWidth(sizes.scopeWidth.min).SetMaxWidth(sizes.scopeWidth.max)
	desc = desc.SetMinWidth(sizes.descWidth.min).SetWidth(sizes.descWidth.max)

	// body and footer widths are *not* dynamic
	body = body.SetWidth(sizes.bodyWidth.max).SetMinHeight(sizes.bodyHeight.min).SetMaxHeight(sizes.bodyHeight.max)
	footer = footer.SetWidth(sizes.footerWidth.max).SetMinHeight(sizes.footerHeight.min).SetMaxHeight(sizes.footerHeight.max)

	form.help.SetWidth(winWidth) // this *should* return an updated model!

	form.children = form.children.SetMap(map[field]tea.Model{
		typeField:        typ,
		scopeField:       scope,
		descriptionField: desc,
		bodyField:        body,
		footerField:      footer,
	})

	return form
}

// nominalLength handles ensuring there's space for the placeholder when there's
// no content
func nominalLength(t textModel, considerPlaceholder bool) int {
	val := t.Value()
	if val == "" && considerPlaceholder {
		val = t.Placeholder()
	}

	slog.Debug("nominalLength", "val", val, "length", len(val))
	return len(val)
}

// Validates *at least* the field/value. May validate other fields or
// combinations in some cases.
func (form mainForm) validate(f field, value string) mainForm {
	// I don't love that this is a ginormous function that validates individual
	// checks, but 🤷🏼‍♂️.  I still need to think about how to normalize the
	// commitlint rules into the validation config for the UI.
	//
	// Would it be helpful to take the union of all field-level settings and
	// have a validator that takes that union and the string value? That would
	// at least deduplicate things some.
	var err error
	switch f {
	case typeField:
		err = validate(value, &fieldRules{
			caseRule:           &form.config.TypeCase,
			multiCaseRule:      nil,
			delimiterStyleRule: nil,
			emptyRule:          &form.config.TypeEmpty,
			enumRule:           &form.config.TypeEnum,
			fullStopRule:       nil,
			maxLengthRule:      &form.config.TypeMaxLength,
			minLengthRule:      &form.config.TypeMinLength,
		})

	case scopeField:
		err = validate(value, &fieldRules{
			caseRule:           &form.config.ScopeCase,
			multiCaseRule:      nil,
			delimiterStyleRule: nil, // &form.config.ScopeDelimeterStyle, // []string instead of string!
			emptyRule:          &form.config.ScopeEmpty,
			enumRule:           &form.config.ScopeEnum,
			fullStopRule:       nil,
			maxLengthRule:      &form.config.ScopeMaxLength,
			minLengthRule:      &form.config.ScopeMinLength,
		})

	case descriptionField:
		err = validate(value, &fieldRules{
			caseRule:           nil,
			multiCaseRule:      &form.config.SubjectCase,
			delimiterStyleRule: nil,
			emptyRule:          &form.config.SubjectEmpty,
			enumRule:           nil,
			fullStopRule:       &form.config.SubjectFullStop,
			maxLengthRule:      &form.config.SubjectMaxLength,
			minLengthRule:      &form.config.SubjectMinLength,
		})

	case bodyField:
		err = validate(value, &fieldRules{
			caseRule:           &form.config.BodyCase,
			multiCaseRule:      nil,
			delimiterStyleRule: nil,
			emptyRule:          &form.config.BodyEmpty,
			enumRule:           nil,
			fullStopRule:       &form.config.BodyFullStop,
			maxLengthRule:      &form.config.BodyMaxLength,
			minLengthRule:      &form.config.BodyMinLength,
		})

	case footerField:
		err = validate(value, &fieldRules{
			caseRule:           nil,
			multiCaseRule:      nil,
			delimiterStyleRule: nil,
			emptyRule:          &form.config.FooterEmpty,
			enumRule:           nil,
			fullStopRule:       nil,
			maxLengthRule:      &form.config.FooterMaxLength,
			minLengthRule:      &form.config.FooterMinLength,
		})
	}

	form.fieldErrs[f] = err

	return form
}

type Focuser interface {
	Focus() (tea.Model, tea.Cmd)
	Blur() (tea.Model, tea.Cmd)
}

// updateFocus updates which field is currently focused.
func (form mainForm) setFocus(focusField field) (mainForm, tea.Cmd) {
	modelMap := map[field]tea.Model{}
	var cmds []tea.Cmd
	var cmd tea.Cmd

	// instead of a full loop, we could just blur the previous field!
	for f := range maxField {
		if child, ok := form.children.Get(f); ok {
			c2 := child
			// // This is okay, but I think I really want to update the value *as*
			// // the selection changes...
			// if f == typeField && form.focusedField == typeField {
			// 	c2 = c2.SetValue(form.config.Types[form.typeList.getSelectedIndex()]).MoveToEnd()
			// }

			focuser, ok := c2.(Focuser)
			if ok {
				if f == focusField {
					c2, cmd = focuser.Focus()
				} else {
					c2, cmd = focuser.Blur()
				}
			}

			modelMap[f] = c2
			cmds = append(cmds, teautil.Wrap(cmd, f))
		}
	}
	form.children = form.children.SetMap(modelMap)
	form.focusedField = focusField

	// NOTE: defaultKeyMap is a global, which isn't really the Elm Architecture
	// way...
	if form.focusedField != typeList {
		if form.children.MustGet(form.focusedField).(textModel).isArea {
			defaultKeyMap.NextSingle.SetEnabled(false)
			defaultKeyMap.NextMulti.SetEnabled(true)
		} else {
			defaultKeyMap.NextSingle.SetEnabled(true)
			defaultKeyMap.NextMulti.SetEnabled(false)
		}
	}

	return form, tea.Batch(cmds...)
}

// nextField moves to the next field.
func (form mainForm) nextField() tea.Cmd {
	return setFocusCmd(max((form.focusedField+1)%maxField, typeList))
}

// prevField moves to the previous field.
func (form mainForm) prevField() tea.Cmd {
	return setFocusCmd((form.focusedField + maxField - 1) % maxField)
}

// // validateDescription checks if the description field meets requirements.
// func (form mainForm) validateDescription() error {
// 	// description := strings.TrimSpace(desc.Value())
// 	// if description == "" {
// 	// 	return fmt.Errorf("description is required")
// 	// }
// 	// if len(description) > m.config.SubjectMaxLength {
// 	// 	return fmt.Errorf("description must be ≤ %d characters (currently %d)", m.config.SubjectMaxLength, len(description))
// 	// }
// 	return nil
// }

type KeyBinder interface {
	GetKeyBindings() []key.Binding
}

// View renders the TUI.
func (form mainForm) View() tea.View {
	// TODO: would it be better to define a type for the value/valid/decoration
	// info? I suspect that would be better than the parallel maps.
	type info struct {
		view    string
		valid   bool
		message string
		render  string
	}

	infos := make(map[field]info, form.children.Len())

	for f, child := range form.children.All() {
		var info info

		t, ok := child.(textModel)
		if ok {
			slog.Warn("getting child view content", "child", f, "width", t.Width(), "minWidth", t.MinWidth(), "content", t.Value())
		}

		info.view = child.View().Content
		// if t, ok := child.(textModel); ok && f <= descriptionField {
		// 	// textinput adds padding sometimes, which can cause it to render
		// 	// "width+1", and we're really particular about that.  We have to
		// 	// double-check and trim it if needed.
		// 	if lipgloss.Width(info.view) > t.Width() {
		// 		info.view = lipgloss.NewStyle().MaxWidth(t.Width()).Render(info.view)
		// 	}
		// }

		info.valid = form.fieldErrs[f] == nil

		// Figure out the decorator; depends on whether the field is valid, etc.
		// This also depends on where the focus is, as we play some interesting
		// games with borders to keep the rendering concise.
		var childState state = blurred
		var childStatus status = defaultStatus
		var childKind fieldKind = single
		if f == form.focusedField {
			childState = focused
		}
		// we *used* to have a "is the field even validate-able?" value, but
		// there's really not a ton of reason to support that
		if info.valid {
			childStatus = valid
		} else {
			childStatus = invalid
		}

		// TODO: make an interface for styling single/multi...
		if f == typeList || child.(textModel).isArea {
			childKind = multi
		}

		border := kindDecorations[childKind].BorderForeground(defaultColors[childStatus][childState][decoration])

		switch f {
		case typeField, scopeField, descriptionField:
			if form.focusedField >= bodyField {
				border = border.UnsetBorderBottom()
			}
		case bodyField:
			if form.focusedField < bodyField {
				border = border.UnsetBorderTop()
			}
			if form.focusedField == footerField {
				border = border.UnsetBorderBottom()
			}
		case footerField:
			if form.focusedField < footerField {
				border = border.UnsetBorderTop()
			}
		}

		info.render = border.Render(info.view)

		infos[f] = info
	}

	// If there's an error (validation!), make sure to show it. Do we do both
	// field and global validation?
	var errMsg string
	if form.fieldErrs[form.focusedField] != nil {
		errMsg = form.fieldErrs[form.focusedField].Error()
	} else if form.err != nil {
		errMsg = form.err.Error()
	}

	if errMsg != "" {
		slog.Debug("showing error", "err", errMsg)
		errMsg = lipgloss.NewStyle().Width(form.winWidth).Foreground(defaultColors[invalid][focused][decoration]).Render(errMsg)
	}

	// get the focused field key bindings...
	var keyBindings []key.Binding
	keyBinder, ok := form.children.MustGet(form.focusedField).(KeyBinder)
	if ok {
		keyBindings = keyBinder.GetKeyBindings()
	}

	// It seems ridiculous to recalculate the keymap on every single view, but
	// with an immutable pattern, we *can't* know if we have the same value call
	// after call. On the plus side, this should be a very fast calculation (and
	// maybe we can memoize the values and keep a pre-rendered view?)
	helpKeys := buildHelpKeys(keyBindings)

	// For the top line, we omit unused scope unless it has focus
	showScope := form.focusedField == scopeField || form.children.MustGet(scopeField).(textModel).Value() != ""
	header := lipgloss.JoinHorizontal(
		lipgloss.Top,
		" ", // one space to align body left border
		infos[typeField].render,
	)
	if showScope {
		header = lipgloss.JoinHorizontal(
			lipgloss.Top,
			header,
			"(",
			infos[scopeField].render,
			")",
		)
	}
	header = lipgloss.JoinHorizontal(
		lipgloss.Top,
		header,
		": ",
		infos[descriptionField].render,
	)

	mainView := lipgloss.JoinVertical(
		lipgloss.Left,
		"\n",
		header,
		infos[bodyField].render,
		infos[footerField].render,
	)

	if errMsg != "" {
		mainView = lipgloss.JoinVertical(
			lipgloss.Left,
			mainView,
			errMsg,
		)
	}

	mainView = lipgloss.JoinVertical(
		lipgloss.Left,
		mainView,
		form.help.View(helpKeys), // &defaultKeyMap),
	)

	layers := []*lipgloss.Layer{
		lipgloss.NewLayer(mainView),
	}

	if form.focusedField == typeField || form.focusedField == typeList {
		layers = append(layers, lipgloss.NewLayer(infos[typeList].render).X(10).Y(1))
	}

	comp := lipgloss.NewCompositor(layers...)

	return tea.NewView(comp.Render())
}

// Result returns the composed commit message and any error.
func (form mainForm) Result() (*commit.Message, error) {
	if form.err != nil {
		return nil, form.err
	}

	texts := make(map[field]string, form.children.Len())

	for f, t := range form.children.All() {
		if f == typeList {
			continue
		}
		texts[f] = strings.TrimSpace(t.(textModel).Value())
	}

	// if !m.done {
	// 	return nil, fmt.Errorf("form not completed")
	// }
	msg := &commit.Message{
		Type:        texts[typeField],
		Scope:       texts[scopeField],
		Description: texts[descriptionField],
		Body:        texts[bodyField],
		Footer:      texts[footerField],
	}

	return msg, nil
}
