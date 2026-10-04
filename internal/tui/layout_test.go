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
	baseConstraints := layoutConstraints{
		maxWidth:     40,
		maxHeight:    20,
		focus:        typeField,
		typeWidth:    dimension{0, 3, 10},
		scopeWidth:   dimension{0, 0, 10},
		descWidth:    dimension{0, 5, 20},
		bodyWidth:    dimension{0, 0, 40},
		bodyHeight:   dimension{0, 0, -1},
		footerWidth:  dimension{0, 0, 20},
		footerHeight: dimension{0, 0, -1},
	}

	baseExpected := layoutOutput{
		typeWidth:    4, // min+focus
		scopeWidth:   0,
		descWidth:    21,
		bodyWidth:    38,
		bodyHeight:   4,
		footerWidth:  20,
		footerHeight: 2,
	}

	actual := calculateSizes(baseConstraints)
	assert.Equal(t, baseExpected, actual)
}

func TestShrinkableDimension(t *testing.T) {
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
			actual := shrinkableDimension(dimension{c.inContent, c.inMin, c.inMax}, c.focusExtra)
			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestOptionalMin(t *testing.T) {
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
