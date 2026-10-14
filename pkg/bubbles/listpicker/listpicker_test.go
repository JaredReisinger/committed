package listpicker

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/charmbracelet/x/exp/teatest/v2"
	"github.com/go-openapi/testify/v2/assert"
)

func TestListpicker(t *testing.T) {
	m := New([]string{"one", "two", "three"})

	m = m.SetSelectedIndex(1)
	assert.Equal(t, "two", m.SelectedItem())

	var ok bool
	m, ok = m.SetSelectedItem("three")
	assert.True(t, ok)
	assert.Equal(t, 2, m.SelectedIndex())

	m, ok = m.SetSelectedItem("MISSING")
	assert.False(t, ok)
	assert.Equal(t, 2, m.SelectedIndex())

	m = m.SetSelectedIndex(0)

	testModel := teatest.NewTestModel(t, m)

	// TODO: get the key bindings for this?
	testModel.Send(tea.KeyPressMsg{Code: tea.KeyDown}) // one->two
	testModel.Send(tea.KeyPressMsg{Code: tea.KeyDown}) // two->three
	testModel.Send(tea.KeyPressMsg{Code: tea.KeyUp})   // three->two

	testModel.Quit()
	testModel.WaitFinished(t, teatest.WithFinalTimeout(time.Second))

	finalModel := testModel.FinalModel(t)
	finalModelImpl, ok := finalModel.(Model)
	assert.True(t, ok)
	assert.Equal(t, 1, finalModelImpl.SelectedIndex())
	assert.Equal(t, "two", finalModelImpl.SelectedItem())
}
