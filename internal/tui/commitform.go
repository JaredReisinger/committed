package tui

import (
	"log/slog"
	"slices"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/jaredreisinger/committed/pkg/bubbles/listpicker"
	"github.com/jaredreisinger/committed/pkg/bubbles/text"
	"github.com/jaredreisinger/committed/pkg/commit"
	"github.com/jaredreisinger/committed/pkg/config"
	"github.com/jaredreisinger/committed/pkg/teautil"
)

// field is the key for child models
type field int

const (
	// this is *not* the tab order
	typeField field = iota
	typeList
	scopeField
	descField
	bodyField

	// I'm considering have 2 footer fields per footer, as per Conventional
	// Commits spec items 8, 9, and 10:
	//
	// 8. One or more footers MAY be provided one blank line after the body.
	// Each footer MUST consist of a word token, followed by either a :<space>
	// or <space># separator, followed by a string value (this is inspired by
	// the git trailer convention).
	//
	// 9. A footer’s token MUST use - in place of whitespace characters, e.g.,
	// Acked-by (this helps differentiate the footer section from a
	// multi-paragraph body). An exception is made for BREAKING CHANGE, which
	// MAY also be used as a token.
	//
	// 10. A footer’s value MAY contain spaces and newlines, and parsing MUST
	// terminate when the next valid footer token/separator pair is observed.
	//
	// The challenge is in figuring out "who" is responsible for the 2-value
	// management; I think getting the layout right makes it need to be the
	// form, but it would be nice to encapsulate the tabbing... but, meh?
	//
	// And... it may be overkill to do that much for footers vs. everything
	// else.
	footerField

	// footerToken1Field
	// footerValue1Field
	// ...

	// helpField ?  I'm not sure it follows a true tea.Model pattern
	helpField
	maxField
	noField field = -1
)

// commitForm represents the TUI state for conventional commit composition.
//
// It includes both local state and also the child models (which are, in a
// sense, also a kind of local state.)   The original AI-created class was an
// absolute mess, with no sense of immutability at all.  The addition of
// [teautil.Router] is helping a ton, however.
type commitForm struct {
	config      *config.Config
	existingMsg *commit.Message // just for init, don't persist

	children  teautil.Router[field, tea.Model] // *focusable* children
	fieldErrs map[field]error                  // field validation errors

	tabstops     []field // see focus handlers and newForm
	focusedField field

	winWidth  int
	winHeight int
	// help      help.Model // move to children?

	limits FieldVals[Limits]

	err error // should this be validation error(s) as well?
}

// newCommitForm creates a new TUI model with the given configuration and
// optional existing message.
func newCommitForm(cfg *config.Config, existingMsg *commit.Message) commitForm {
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

	// help := help.New()
	// help.ShowAll = true
	// help.SetWidth(80) // ??

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

	list := listpicker.New(cfg.Rules.TypeEnum.Value)
	if initialType != "" {
		var ok bool
		list, ok = list.SetSelectedItem(initialType)
		if !ok {
			slog.Warn("incoming type is not valid")
		}
	} else {
		initialType = list.SelectedItem()
	}

	// We want the user to be able to opt out of the type list, or ever setting
	// focus in the type field and *only* using the type list, so we keep a list
	// of tabstops that's distinct from the children.  We *could* keep a
	// separate tabIndex, and/or use tabstops[tabIndex] instead of
	// focusedField... but the list of tabstops is small enough that a linear
	// search is good enough.
	tabstops := []field{
		// by default, we don't edit the type field directly
		typeList,
		scopeField,
		descField,
		bodyField,
		footerField,
	}

	form := commitForm{
		config:      cfg,
		existingMsg: existingMsg,

		// There's an interesting question about whether the form's fields
		// should be *all* models or just focusable ones.  Perhaps `children`
		// should be *all* models (including help), and "focusableFields" should
		// be a separate list that we use to manage focus.  That would make
		// message passing more consistent and extensible.  (Just imagine if the
		// "help" started using/needing messages as opposed to being a
		// pure-display component.)

		children: teautil.NewRouter(map[field]tea.Model{
			typeField:   text.New(false, "type", initialType).SetCharLimit(limits.typeWidth.max),
			typeList:    list,
			scopeField:  text.New(false, "scope", initialScope).SetCharLimit(limits.scopeWidth.max),
			descField:   text.New(false, "description", initialDesc).SetCharLimit(limits.descWidth.max),
			bodyField:   text.New(true, "message body", initialBody).SetMaxContentHeight(100).SetDynamicHeight(true).SetMinHeight(4),
			footerField: text.New(true, "footer", initialFooter).SetMaxContentHeight(50).SetMinHeight(1),
			helpField:   newHelpModel(),
		}),

		fieldErrs: make(map[field]error, maxField),

		tabstops:     tabstops,
		focusedField: noField,

		// help:   help,
		limits: limits,
	}

	// perform initial validation?
	for f, child := range form.children.All() {
		t, ok := child.(text.Model)
		if ok {
			form = form.validate(f, t.Value())
		}
	}

	return form
}

// Init initializes the bubbletea program.
func (form commitForm) Init() tea.Cmd {
	// should other processing happen here?
	// return textinput.Blink
	// Set initial focus
	return setFocusCmd(form.tabstops[0])
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
func (form commitForm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd // aggregated commands
	var cmd tea.Cmd    // returned Cmd from delegated Update() calls
	handled := false

	switch msgT := msg.(type) {
	case tea.WindowSizeMsg:
		form = form.resize(msgT.Width, msgT.Height)
		handled = true

	case setFocusMsg:
		// If the focus *was* type or scope, we need to play some positioning
		// tricks to ensure the field isn't scrolled weirdly.
		switch form.focusedField {
		case typeField, scopeField:
			if child, ok := form.children.Get(form.focusedField); ok {
				if t, ok := child.(text.Model); ok {
					t2 := t.MoveToBegin().MoveToEnd()
					form.children = form.children.Set(form.focusedField, t2)
				}
			}
		}

		form, cmd = form.setFocus(msgT.field)
		cmds = append(cmds, cmd)
		// we also need to resize fields as focus changes...
		form = form.resize(form.winWidth, form.winHeight)
		handled = true

	case teautil.WrappedMsg[field]:
		// Our teautil.Router doesn't have a mechanism for child models to
		// "broadcast up" messages for the parent, but we *can* inspect any
		// wrapped messages and look for the ones we care about. This has the
		// added benefit of telling us which model sent the message (via the
		// Key), which can be important to know.
		switch innerMsg := msgT.Msg.(type) {
		case listpicker.SelectionChangedMsg:
			if msgT.Key == typeList {
				// The MoveToBegin/MoveToEnd ensures we don't have a weirdly
				// scrolled field afterwards
				t2 := form.children.MustGet(typeField).(text.Model).SetValue(innerMsg.SelectedItem).MoveToBegin()
				form.children = form.children.Set(typeField, t2)
				// changing the type *can* change the layout
				form = form.resize(form.winWidth, form.winHeight)
				t2 = form.children.MustGet(typeField).(text.Model).MoveToEnd()
				form.children = form.children.Set(typeField, t2)
				handled = true

				// We also need to fake a "field changed" message!
				cmds = append(cmds, teautil.Wrap(text.SendValueChanged(t2.Value()), typeField))
			}

		case text.ValueChanged:
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

func (form commitForm) resize(winWidth int, winHeight int) commitForm {
	slog.Debug("window resize", "width", winWidth, "height", winHeight)

	form.winWidth = winWidth
	form.winHeight = winHeight

	typ := form.children.MustGet(typeField).(text.Model)
	// TODO: size typeList field here as well!
	scope := form.children.MustGet(scopeField).(text.Model)
	desc := form.children.MustGet(descField).(text.Model)
	body := form.children.MustGet(bodyField).(text.Model)
	footer := form.children.MustGet(footerField).(text.Model)
	help := form.children.MustGet(helpField).(helpModel)

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

	// body and footer widths are *not* dynamic, but the heights are
	body = body.SetWidth(sizes.bodyWidth.max).SetMinHeight(sizes.bodyHeight.min).SetMaxHeight(sizes.bodyHeight.max)
	footer = footer.SetWidth(sizes.footerWidth.max).SetMinHeight(sizes.footerHeight.min).SetMaxHeight(sizes.footerHeight.max)

	help = help.SetWidth(winWidth)

	form.children = form.children.SetMap(map[field]tea.Model{
		typeField:   typ,
		scopeField:  scope,
		descField:   desc,
		bodyField:   body,
		footerField: footer,
		helpField:   help,
	})

	return form
}

// nominalLength handles ensuring there's space for the placeholder when there's
// no content
func nominalLength(t text.Model, considerPlaceholder bool) int {
	val := t.Value()
	if val == "" && considerPlaceholder {
		val = t.Placeholder()
	}

	slog.Debug("nominalLength", "val", val, "length", len(val))
	return len(val)
}

// Validates *at least* the field/value. May validate other fields or
// combinations in some cases.
func (form commitForm) validate(f field, value string) commitForm {
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

	case descField:
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
func (form commitForm) setFocus(focusField field) (commitForm, tea.Cmd) {
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
		modelMap[helpField] = help.SetKeyMap(buildHelpKeys(defaultKeyMap, keyBindings))

		form.children = form.children.SetMap(modelMap)
		form.focusedField = focusField
	}

	return form, tea.Batch(cmds...)
}

// Looks up v in s, then calculates the index delta, wrapping around if needed.
func circularDelta[S ~[]E, E comparable](s S, v E, delta int) E {
	curIndex := slices.Index(s, v)
	if curIndex < 0 {
		return s[0]
	}
	// we *could* loop to ensure the new index is in [0, len), but we can play
	// with some math to make it easier. "delta % len" ensures a value in the
	// range (-len, len), adding another len ensures it's not negative.
	l := len(s)
	newIndex := (curIndex + (delta % l) + l) % l
	return s[newIndex]
}

// nextField moves to the next field.
func (form commitForm) nextField() tea.Cmd {
	return setFocusCmd(circularDelta(form.tabstops, form.focusedField, 1))
}

// prevField moves to the previous field.
func (form commitForm) prevField() tea.Cmd {
	return setFocusCmd(circularDelta(form.tabstops, form.focusedField, -1))
}

type KeyBinder interface {
	GetKeyBindings() []key.Binding
}

// View renders the TUI.
func (form commitForm) View() tea.View {
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

		t, ok := child.(text.Model)
		if ok {
			slog.Debug("getting child view content", "child", f, "width", t.Width(), "minWidth", t.MinWidth(), "content", t.Value())
		}

		info.view = child.View().Content
		info.valid = form.fieldErrs[f] == nil

		// REVIEW: should the validation/border be restricted to the text
		// fields, with typeList and help handled *compeltely* separately?

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
		textChild, ok := child.(text.Model)
		if f == typeList || (ok && textChild.IsArea()) {
			childKind = multi
		}

		border := kindDecorations[childKind].BorderForeground(defaultColors[childStatus][childState][decoration])

		switch f {
		case typeField, scopeField, descField:
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
		case helpField:
			border = lipgloss.NewStyle() // no border!
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

	// For the top line, we omit unused scope unless it has focus
	showScope := form.focusedField == scopeField || form.children.MustGet(scopeField).(text.Model).Value() != ""
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
		infos[descField].render,
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
		// form.help.View(helpKeys), // &defaultKeyMap),
		infos[helpField].render,
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
func (form commitForm) Result() (*commit.Message, error) {
	if form.err != nil {
		return nil, form.err
	}

	texts := make(map[field]string, form.children.Len())

	for f, child := range form.children.All() {
		if textChild, ok := child.(text.Model); ok {
			texts[f] = strings.TrimSpace(textChild.Value())
		}
	}

	// if !m.done {
	// 	return nil, fmt.Errorf("form not completed")
	// }
	msg := &commit.Message{
		Type:        texts[typeField],
		Scope:       texts[scopeField],
		Description: texts[descField],
		Body:        texts[bodyField],
		Footer:      texts[footerField],
	}

	return msg, nil
}
