package tui

import (
	"log/slog"
	"math"
)

// Field sizes and presentation needs to get calculated *often*, because we want
// it to be responsive to content and to focus. At the same time, we don't want
// the sizing logic to be scattered all over the code. We also want this logic
// to be testable, which is easier if it's not too tightly coupled to the UI. To
// that end, the size/placement calculations are designed as a pure function...
// vanilla values come in, and vanilla values are returned.  The form can then
// use the return values to modify the models as needed.
//
// I *think* we can avoid needing to call this on every View() rendering, but as
// a pure function it would be safe to call if it turns out that's needed.

type Size struct {
	width  int
	height int
}

func (s Size) LogValue() slog.Value {
	return slog.GroupValue([]slog.Attr{
		{Key: "width", Value: slog.IntValue(s.width)},
		{Key: "height", Value: slog.IntValue(s.height)},
	}...)
}

// Limits can represent either content Limits or rendering Limits, depending on
// the context.  If a field should have a fixed size, min and max will be set to
// the same value. The caller knows enough about the sizing that it can assume
// and use whichever field it prefers.
type Limits struct {
	min int
	max int // a max of zero mean "no maximum" ... or do we require -1?
}

func (l Limits) LogValue() slog.Value {
	return slog.GroupValue([]slog.Attr{
		{Key: "min", Value: slog.IntValue(l.min)},
		{Key: "max", Value: slog.IntValue(l.max)},
	}...)
}

// Contents represents the actual content length, or the nominal (for
// placeholders).  I'm not 100% sure I need the nominal, though. The one case
// where it might make sense is in the header where a placeholder really
// shouldn't count as size... but that may be overkill.
type Contents struct {
	actual int
	// nominal int
}

func (c Contents) LogValue() slog.Value {
	return slog.GroupValue([]slog.Attr{
		{Key: "actual", Value: slog.IntValue(c.actual)},
		// {Key: "nominal", Value: slog.IntValue(c.nominal)},
	}...)
}

// FieldVals are the relevant limits/contents for each field.
type FieldVals[T slog.LogValuer] struct {
	typeWidth    T
	scopeWidth   T
	descWidth    T // TODO: add header constraints?
	bodyWidth    T
	bodyHeight   T
	footerWidth  T
	footerHeight T
}

func (l FieldVals[T]) LogValue() slog.Value {
	return slog.GroupValue([]slog.Attr{
		{Key: "typeWidth", Value: l.typeWidth.LogValue()},
		{Key: "scopeWidth", Value: l.scopeWidth.LogValue()},
		{Key: "descWidth", Value: l.descWidth.LogValue()},
		{Key: "bodyWidth", Value: l.bodyWidth.LogValue()},
		{Key: "bodyHeight", Value: l.bodyHeight.LogValue()},
		{Key: "footerWidth", Value: l.footerWidth.LogValue()},
		{Key: "footerHeight", Value: l.footerHeight.LogValue()},
	}...)
}

// calculateSizes takes the content limits and contents themselves, and returns
// the rendering limits for the fields.
func calculateSizes(bounds Size, limits FieldVals[Limits], focus field, contents FieldVals[Contents]) FieldVals[Limits] {
	slog.Debug("calculateSize", "width", bounds.width, "height", bounds.height)

	// Graphically, the UI looks like:
	//
	// (bounds is outer maximum...)
	// ┌────────────────────────────────────────────────────────────────────┐
	// │ (blank line for cleanliness)                                       │
	// │ <type>(<scope>): description                                       │
	// │╭──────╌───────╌╌╌─────────────────────────────────────────────────╮│
	// ││<body>                                                            ││
	// ││  width: min(configured line width, window width - 2)             ││
	// ││  "chromeHeight": 1 for cleanliness, 1 for header line,...        ││
	// ││                 2-? for footer, 8? for help, 3 for borders       ││
	// ││  minHeight: window height - otherHeight                          ││
	// ││  maxHight:  min(max(contentHeight, minHeight),                   ││
	// ││                 (window height - chromeHeight))                  ││
	// ││                                                                  ││
	// ││  if ideal height is < maxHeight, use dynamic height,             ││
	// ││  else use fixed maxHeight                                        ││
	// ││                                                                  ││
	// ││                                                                  ││
	// ││                                                                  ││
	// │├──────────────────────────────────────────────────────────────────┤│
	// ││<footer>                                                          ││
	// ││  width: min(configured line width, window width - 2)             ││
	// ││  height: do we *ever* let this grow?                             ││
	// │╰──────────────────────────────────────────────────────────────────╯│
	// │ <help> ...                                                         │
	// │   width: window width - 2 for alignment                            │
	// │   height: 8, always?                                               │
	// │                                                                    │
	// │                                                                    │
	// │                                                                    │
	// │ (blank line for cleanliness? maybe not)                            │
	// └────────────────────────────────────────────────────────────────────┘
	//
	// box chars: ─ │ ┌ ┐ └ ┘ ├ ┤ ┬ ┴ ┼ ╴ ╵ ╶ ╷ ┄ ┈ ╌ ╭ ╮ ╯ ╰

	const chromeHeight = 4 // initial blank line, 3 horizontal border lines
	const chromeWidth = 2  // 2 vertical border lines

	const helpHeight = 8 // do we *ever* allow shrinking or growing?

	// At a certain point, there's a minimum width/height we can reasonably lay
	// out inside of if we want to avoid zero- or negative-value fields. We can
	// calculate this from the minimum needed size for each field.
	const headerHeight = 1

	minBodyHeight := limits.bodyHeight.min
	if minBodyHeight <= 0 {
		minBodyHeight = 4 // subjective default
	}
	minFooterHeight := limits.footerHeight.min
	if minFooterHeight <= 0 {
		minFooterHeight = 2 // subjective default
	}
	minErrorHeight := 2

	// body and footer cay layout anywhere that the header can, width-wise
	const minHeaderWidth = 7 + 1 + 7 + 3 + 12 // type "(" scope "): " desc

	maxWidth := max(bounds.width, chromeWidth+minHeaderWidth)
	maxHeight := max(bounds.height, chromeHeight+headerHeight+minBodyHeight+minFooterHeight+minErrorHeight+helpHeight)

	// The sizing is *highly* dependent on which field we're calculating, and
	// some fields' size depends on what we have/calculate for other fields.
	var out FieldVals[Limits]

	// We actually lay out the body (and footer) first, as we *might* use the body
	// width when calculating the header/description to give it a default
	// maximum width (to prevent it from looking like it expands as you type).
	out.bodyWidth.min = minNonNegative(max(maxWidth-chromeWidth, limits.bodyWidth.min), limits.bodyWidth.max)
	out.bodyWidth.max = out.bodyWidth.min
	bodyHeight := limits.bodyHeight
	bodyHeight.min = max(bodyHeight.min, minBodyHeight)
	bodyHeight.max = minNonNegative(
		maxHeight-chromeHeight-headerHeight-minFooterHeight-minErrorHeight-helpHeight,
		maxNormalizer(bodyHeight.max),
	)
	// out.bodyHeight = shrinkableLimits(bodyHeight, contents.bodyHeight, false)
	out.bodyHeight = bodyHeight

	out.footerWidth.min = minNonNegative(
		max(maxWidth-chromeWidth, limits.footerWidth.min),
		maxNormalizer(limits.footerWidth.max),
	)
	out.footerWidth.max = out.footerWidth.min
	footerHeight := limits.footerHeight
	footerHeight.min = max(footerHeight.min, minFooterHeight)
	footerHeight.max = minNonNegative(
		maxHeight-chromeHeight-headerHeight-out.bodyHeight.max-minErrorHeight-helpHeight,
		maxNormalizer(footerHeight.max),
	)
	// out.footerHeight = shrinkableLimits(footerHeight, contents.footerHeight, false)
	out.footerHeight = footerHeight

	// Note that the *form* is responsible for accounting for the placeholders
	out.typeWidth = shrinkableLimits(limits.typeWidth, contents.typeWidth, focus == typeField)
	out.scopeWidth = shrinkableLimits(limits.scopeWidth, contents.scopeWidth, focus == scopeField)

	// calculate chrome in header line
	headerChromeWidth := chromeWidth + 2 // always start with initial space, trailing space, and ": "
	if out.scopeWidth.max > 0 {
		headerChromeWidth += 2 // add "()"
	}
	// the description isn't "shrinkable..."
	// slog.Info("desc", "outer", maxWidth, "remaining", maxWidth-headerChromeWidth-out.typeWidth-out.scopeWidth, "max", limits.descWidth.max+1)
	descWidthMax := limits.descWidth.max
	if descWidthMax <= 0 {
		descWidthMax = out.bodyWidth.max - out.typeWidth.max - out.scopeWidth.max - 1 // account extra 1 for focus (added below...)
		if out.scopeWidth.max > 0 {
			limits.descWidth.max -= 2 // account for "()"
		}
	}

	out.descWidth.min = minNonNegative(
		maxWidth-headerChromeWidth-out.typeWidth.max-out.scopeWidth.max,
		descWidthMax+1, // it doesn't shrink, so *always* allow 1 for focus
	)
	out.descWidth.max = out.descWidth.min

	return out
}

// shrinkableLimits calculates the measurement needed for a field that's
// allowed to shrink; often used to make a field "content-sized" when it doesn't
// have focus. The returned value is intended to
func shrinkableLimits(limits Limits, contents Contents, needFocusExtra bool) Limits {
	focusExtra := 0
	if needFocusExtra {
		focusExtra = 1
	}

	var out Limits

	out.min = limits.min
	out.max = minNonNegative(
		max(contents.actual, limits.min)+focusExtra,
		maxNormalizer(limits.max),
	)

	return out
}

// maxNormalizer helps with some logic (shrinkableLimits) where we want zero
// values for max to be treated as "unset" (because a max would never truly be
// zero, and because zero is the Zero value), and that logic is agnostic about
// the numbers... a *content* zero is a true zero, after all.  We don't want
// callers to have to know to treat "unset max" specially, so we do it here.
func maxNormalizer(max int) int {
	if max == 0 {
		return -1
	}
	return max
}

func minNonNegative(vals ...int) int {
	// foundMin := false
	ret := math.MaxInt

	for _, v := range vals {
		if v >= 0 {
			ret = min(ret, v)
			// foundMin = true
		}
	}

	// if !foundMin {
	// 	return 0
	// }

	return ret
}
