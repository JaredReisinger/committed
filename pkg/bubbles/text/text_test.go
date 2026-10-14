package text

import (
	"fmt"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/exp/teatest/v2"
	"github.com/go-openapi/testify/v2/assert"
)

func TestTextModelNonArea(t *testing.T) {
	text := New(false, "PLACEHOLDER", "")
	assert.NotNil(t, text)

	assert.Nil(t, text.Init())

	text2, cmd := text.Update(nil)
	assert.NotEqual(t, text, text2)
	assert.Nil(t, cmd)

	// test a bunch of setters/getters
	assert.Equal(t, 5, text.SetCharLimit(5).CharLimit())
	assert.Equal(t, 6, text.SetWidth(6).Width())
	assert.Equal(t, 7, text.SetMinWidth(7).MinWidth())
	assert.Equal(t, 8, text.SetMaxWidth(8).MaxWidth())

	assert.Equal(t, 1, text.SetHeight(15).Height())
	assert.Equal(t, 1, text.SetMinHeight(16).MinHeight())
	assert.Equal(t, 1, text.SetMaxHeight(17).MaxHeight())
	assert.Equal(t, 1, text.SetMaxContentHeight(18).MaxContentHeight())

	assert.Equal(t, false, text.SetDynamicHeight(true).DynamicHeight())
	assert.Equal(t, false, text.SetDynamicHeight(false).DynamicHeight())

	assert.NotNil(t, text.SetStyles(defaultStyles).Styles())
}

func TestTextModelArea(t *testing.T) {
	text := New(true, "PLACEHOLDER", "")
	assert.NotNil(t, text)

	assert.Nil(t, text.Init())

	text2, cmd := text.Update(nil)
	assert.NotEqual(t, text, text2)
	assert.Nil(t, cmd)

	// test a bunch of setters/getters
	assert.Equal(t, 5, text.SetCharLimit(5).CharLimit())
	assert.Equal(t, 6, text.SetWidth(6).Width())
	assert.Equal(t, 0, text.SetMinWidth(7).MinWidth()) // not supported!
	assert.Equal(t, 8, text.SetMaxWidth(8).MaxWidth())

	assert.Equal(t, 15, text.SetHeight(15).Height())
	assert.Equal(t, 16, text.SetMinHeight(16).MinHeight())
	assert.Equal(t, 17, text.SetMaxHeight(17).MaxHeight())
	assert.Equal(t, 18, text.SetMaxContentHeight(18).MaxContentHeight())

	assert.Equal(t, true, text.SetDynamicHeight(true).DynamicHeight())
	assert.Equal(t, false, text.SetDynamicHeight(false).DynamicHeight())

	assert.NotNil(t, text.SetStyles(defaultStyles).Styles())
}

func TestTextModelUX(t *testing.T) {
	for _, isArea := range []bool{false, true} {
		t.Run(fmt.Sprintf("isArea: %t", isArea), func(t *testing.T) {
			t.Parallel()

			m := New(isArea, "placeholder", "INITIAL")
			m2, _ := m.Focus()
			m = m2.(Model)
			// use Blur, but *don't* overwrite the model... we should still be
			// able to type in the original one
			blurred, noCmd := m.Blur()
			assert.NotEqual(t, m, blurred)
			assert.Nil(t, noCmd)

			assert.Equal(t, isArea, m.IsArea())
			assert.Equal(t, "placeholder", m.Placeholder())

			testModel := teatest.NewTestModel(t, m)

			testModel.Send(tea.KeyPressMsg{Mod: tea.ModCtrl, Code: 'k'}) // kill line
			testModel.Type("here is some text")
			testModel.Send(tea.KeyPressMsg{Mod: tea.ModCtrl, Code: 'a'}) // start of line
			testModel.Type("first, ")

			testModel.Quit()
			testModel.WaitFinished(t, teatest.WithFinalTimeout(time.Second))

			finalModel := testModel.FinalModel(t)
			finalModelImpl, ok := finalModel.(Model)
			assert.True(t, ok)
			assert.Equal(t, "first, here is some text", finalModelImpl.Value())
		})
	}
}

func TestKeyMap(t *testing.T) {
	text := New(false, "", "")
	assert.NotNil(t, text)

	keymap := text.SetKeyMap(DefaultTextModelKeyMap()).KeyMap()

	actualSingleKeyMap := keymap.ForTextInput()

	assert.NotNil(t, actualSingleKeyMap)
	assert.Equal(t, keymap.AcceptSuggestion, actualSingleKeyMap.AcceptSuggestion)

	actualAreaKeyMap := keymap.ForTextArea()
	assert.NotNil(t, actualAreaKeyMap)
	assert.Equal(t, keymap.InsertNewline, actualAreaKeyMap.InsertNewline)
}

func TestKeyBindings(t *testing.T) {
	for _, isArea := range []bool{false, true} {
		t.Run(fmt.Sprintf("isArea: %t", isArea), func(t *testing.T) {
			t.Parallel()

			m := New(isArea, "placeholder", "INITIAL")

			bindings := m.GetKeyBindings()

			if isArea {
				assert.Equal(t, DefaultTextModelKeyMap().InsertNewline, bindings[0])
			} else {
				assert.Equal(t, DefaultTextModelKeyMap().DeleteAfterCursor, bindings[0])

			}
		})
	}
}
