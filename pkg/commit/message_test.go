package commit

import (
	"fmt"
	"testing"

	"github.com/go-openapi/testify/v2/assert"
)

func TestWrap(t *testing.T) {
	for i, tt := range []struct {
		input    string
		limit    int
		expected string
	}{
		{"lorem ipsum", 20, "lorem ipsum"},
		{"lorem ipsum", 6, "lorem\nipsum"},
		{"lorem\nipsum", 20, "lorem\nipsum"},
		{"lorem\n ipsum", 20, "lorem\n ipsum"},
		{"this\n  - item 1\n  - item2", 20, "this\n  - item 1\n  - item2"},
		{"lorem ipsum_but_way_too_long", 10, "lorem\nipsum_but_way_too_long"},
		// nice-ish handling of sentence separating whitespace
		{"end.  start", 10, "end. \nstart"},
		{"", 20, ""},
	} {
		t.Run(fmt.Sprintf("case %d", i+1), func(t *testing.T) {
			t.Parallel()
			actual := wrap(tt.input, tt.limit)
			assert.EqualT(t, tt.expected, actual)
		})
	}
}
