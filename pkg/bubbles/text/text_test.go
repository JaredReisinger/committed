package text

import (
	"testing"

	"github.com/go-openapi/testify/v2/assert"
)

func TestTextModel(t *testing.T) {
	text := New(true, "PLACEHOLDER", "")
	assert.NotNil(t, text)

	assert.Nil(t, text.Init())

	text2, cmd := text.Update(nil)
	assert.NotEqual(t, text, text2)
	assert.Nil(t, cmd)

	// test a bunch of setters/getters
	assert.Equal(t, 5, text.SetCharLimit(5).CharLimit())
	assert.Equal(t, 6, text.SetWidth(6).Width())
	assert.Equal(t, 7, text.SetMaxWidth(7).MaxWidth())

	assert.Equal(t, 8, text.SetHeight(8).Height())
	assert.Equal(t, 9, text.SetMinHeight(9).MinHeight())
	assert.Equal(t, 10, text.SetMaxHeight(10).MaxHeight())
	assert.Equal(t, 11, text.SetMaxContentHeight(11).MaxContentHeight())

	assert.Equal(t, true, text.SetDynamicHeight(true).DynamicHeight())
	assert.Equal(t, false, text.SetDynamicHeight(false).DynamicHeight())
}
