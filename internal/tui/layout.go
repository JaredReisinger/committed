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

// Dimension represents the size constraints along a single dimension, either
// width or height.  We do this because some fields only size in one dimension,
// because sizing logic doesn't really care about direction as much as
// constraints, and to simplify the data we need to track.
type dimension struct {
	content int
	min     int
	max     int // a max of zero mean "no maximum"
}

func (d dimension) LogValue() slog.Value {
	return slog.GroupValue([]slog.Attr{
		{Key: "content", Value: slog.IntValue(d.content)},
		{Key: "min", Value: slog.IntValue(d.min)},
		{Key: "max", Value: slog.IntValue(d.max)},
	}...)
}

type layoutConstraints struct {
	maxWidth  int // the bounds, often the containing window
	maxHeight int
	focus     field

	// We *could* include the layout chrome here, to fully parameterize the
	// layout. That might be a future change, to allow users to change the
	// border styles. For now, though, it's hard-coded in calculateSizes.

	typeWidth    dimension
	scopeWidth   dimension
	descWidth    dimension // TODO: add header constraints?
	bodyWidth    dimension
	bodyHeight   dimension
	footerWidth  dimension
	footerHeight dimension
	// helpW/H -- TODO?
}

func (l layoutConstraints) LogValue() slog.Value {
	return slog.GroupValue([]slog.Attr{
		{Key: "maxWidth", Value: slog.IntValue(l.maxWidth)},
		{Key: "maxHeight", Value: slog.IntValue(l.maxHeight)},
		{Key: "focus", Value: slog.IntValue(int(l.focus))},
		{Key: "typeWidth", Value: l.typeWidth.LogValue()},
		{Key: "scopeWidth", Value: l.scopeWidth.LogValue()},
		{Key: "descWidth", Value: l.descWidth.LogValue()},
		{Key: "bodyWidth", Value: l.bodyWidth.LogValue()},
		{Key: "bodyHeight", Value: l.bodyHeight.LogValue()},
		{Key: "footerWidth", Value: l.footerWidth.LogValue()},
		{Key: "footerHeight", Value: l.footerHeight.LogValue()},
	}...)
}

// same fields as the constraints, boiled down to a single number
type layoutOutput struct {
	typeWidth    int
	scopeWidth   int
	descWidth    int
	bodyWidth    int
	bodyHeight   int
	footerWidth  int
	footerHeight int
}

func (l layoutOutput) LogValue() slog.Value {
	return slog.GroupValue([]slog.Attr{
		{Key: "typeWidth", Value: slog.IntValue(l.typeWidth)},
		{Key: "scopeWidth", Value: slog.IntValue(l.scopeWidth)},
		{Key: "descWidth", Value: slog.IntValue(l.descWidth)},
		{Key: "bodyWidth", Value: slog.IntValue(l.bodyWidth)},
		{Key: "bodyHeight", Value: slog.IntValue(l.bodyHeight)},
		{Key: "footerWidth", Value: slog.IntValue(l.footerWidth)},
		{Key: "footerHeight", Value: slog.IntValue(l.footerHeight)},
	}...)
}

func calculateSizes(in layoutConstraints) layoutOutput {
	slog.Debug("calculateSize", "maxWidth", in.maxWidth, "maxHeight", in.maxHeight)

	// The first thing we do is calculate how to distribute the height.
	// Graphically, the UI looks like:
	//
	// outer window...
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

	minBodyHeight := in.bodyHeight.min
	if minBodyHeight <= 0 {
		minBodyHeight = 4 // subjective default
	}
	minFooterHeight := in.footerHeight.min
	if minFooterHeight <= 0 {
		minFooterHeight = 2 // subjective default
	}

	// body and footer cay layout anywhere that the header can, width-wise
	const minHeaderWidth = 7 + 1 + 7 + 3 + 12 // type "(" scope "): " desc

	maxWidth := max(in.maxWidth, chromeWidth+minHeaderWidth)
	maxHeight := max(in.maxHeight, chromeHeight+headerHeight+minBodyHeight+minFooterHeight+helpHeight)

	// The sizing is *highly* dependent on which field we're calculating, and
	// some fields' size depends on what we have/calculate for other fields.
	var out layoutOutput

	// We actually lay out the body (and footer) first, as we might use the body
	// width when calculating the header/description to give it a default
	// maximum width (to prevent it from looking like it expands as you type).
	out.bodyWidth = minNonNegative(max(maxWidth-chromeWidth, in.bodyWidth.min), in.bodyWidth.max)
	bodyHeight := in.bodyHeight
	bodyHeight.min = max(bodyHeight.min, minBodyHeight)
	bodyHeight.max = minNonNegative(
		maxHeight-chromeHeight-headerHeight-minFooterHeight-helpHeight,
		bodyHeight.max,
	)
	out.bodyHeight = shrinkableDimension(bodyHeight, false)

	out.footerWidth = minNonNegative(max(maxWidth-chromeWidth, in.footerWidth.min), in.footerWidth.max)
	footerHeight := in.footerHeight
	footerHeight.min = max(footerHeight.min, minFooterHeight)
	footerHeight.max = minNonNegative(
		maxHeight-chromeHeight-headerHeight-out.bodyHeight-helpHeight,
		footerHeight.max,
	)
	out.footerHeight = shrinkableDimension(footerHeight, false)

	// Note that the *form* is responsible for accounting for the placeholders
	out.typeWidth = shrinkableDimension(in.typeWidth, in.focus == typeField)
	out.scopeWidth = shrinkableDimension(in.scopeWidth, in.focus == scopeField)

	// calculate chrome in header line
	headerChromeWidth := chromeWidth + 2 // always start with initial space, trailing space, and ": "
	if out.scopeWidth > 0 {
		headerChromeWidth += 2 // add "()"
	}
	// the description isn't "shrinkable..."
	// slog.Info("desc", "outer", maxWidth, "remaining", maxWidth-headerChromeWidth-out.typeWidth-out.scopeWidth, "max", in.descWidth.max+1)
	if in.descWidth.max < 0 {
		in.descWidth.max = out.bodyWidth - out.typeWidth - out.scopeWidth - 1 // account extra 1 for focus (added below...)
		if out.scopeWidth > 0 {
			in.descWidth.max -= 2 // account for "()"
		}
	}

	out.descWidth = minNonNegative(
		maxWidth-headerChromeWidth-out.typeWidth-out.scopeWidth,
		in.descWidth.max+1, // it doesn't shrink, so *always* allow 1 for focus
	)

	return out
}

// shrinkableDimension calculates the measurement needed for a field that's
// allowed to shrink; often used to make a field "content-sized" when it doesn't
// have focus.
func shrinkableDimension(in dimension, needFocusExtra bool) int {
	focusExtra := 0
	if needFocusExtra {
		focusExtra = 1
	}

	return minNonNegative(
		max(in.content, in.min)+focusExtra,
		in.max,
	)
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
