package tui

import (
	"fmt"
	"testing"

	"github.com/go-openapi/testify/v2/assert"
)

func TestCalculateSizes(t *testing.T) {
	// The input/output structures are bulky, so we pre-define some named base
	// cases, and we trust TestShrinkableDimension to handle a lot of the header
	// width nuance
	baseBounds := Size{40, 20}
	baseLimits := FieldVals[Limits]{
		typeWidth:    Limits{3, 10},
		scopeWidth:   Limits{0, 10},
		descWidth:    Limits{5, 20},
		bodyWidth:    Limits{0, 40},
		bodyHeight:   Limits{0, 0},
		footerWidth:  Limits{0, 20},
		footerHeight: Limits{0, 0},
	}

	baseContents := FieldVals[Contents]{
		typeWidth:    Contents{0},
		scopeWidth:   Contents{0},
		descWidth:    Contents{0},
		bodyWidth:    Contents{0},
		bodyHeight:   Contents{0},
		footerWidth:  Contents{0},
		footerHeight: Contents{0},
	}
	baseFocus := typeField

	baseExpected := FieldVals[Limits]{
		typeWidth:    Limits{3, 4}, // min+focus
		scopeWidth:   Limits{0, 0},
		descWidth:    Limits{21, 21},
		bodyWidth:    Limits{38, 38},
		bodyHeight:   Limits{4, 4},
		footerWidth:  Limits{20, 20},
		footerHeight: Limits{2, 2},
	}

	actual := calculateSizes(baseBounds, baseLimits, baseFocus, baseContents)
	assert.Equal(t, baseExpected, actual)
}

func TestShrinkableLimits(t *testing.T) {
	for i, c := range []struct {
		inContent  int
		inMin      int
		inMax      int
		focusExtra bool
		expected   int
	}{
		{0, 0, -1, false, 0},
		{0, 0, -1, true, 1},
		{0, 0, 10, false, 0},
		{0, 0, 10, true, 1},
		{5, 0, 10, false, 5},
		{5, 0, 10, true, 6},
		{10, 0, 5, false, 5},
		{10, 0, 5, false, 5},
		{0, 5, -1, false, 5},
		{0, 5, -1, true, 6},
		{0, 5, 10, false, 5},
		{0, 5, 10, true, 6},
		{0, 5, 5, true, 5},
	} {
		t.Run(fmt.Sprintf("case %d", i+1), func(t *testing.T) {
			t.Parallel()
			actual := shrinkableLimits(Limits{c.inMin, c.inMax}, Contents{c.inContent}, c.focusExtra)
			assert.Equal(t, Limits{c.inMin, c.expected}, actual)
		})
	}
}

func TestMinNonnegative(t *testing.T) {
	for i, c := range [][]int{
		// expected is first, remainder are args...
		{0, 0, 0},
		{1, 1, -1},
		{0, 1, 0},
		{1, 1, -5},
		{1, 1, 10},
		{4, -2, 10, -3, 9, 20, 4, 11, -3},
	} {
		expected := c[0]
		args := c[1:]
		t.Run(fmt.Sprintf("case %d", i+1), func(t *testing.T) {
			t.Parallel()
			actual := minNonNegative(args...)
			assert.Equal(t, expected, actual)
		})
	}

}
