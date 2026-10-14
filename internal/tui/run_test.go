package tui

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/charmbracelet/x/exp/teatest/v2"
	"github.com/go-openapi/testify/v2/assert"

	"github.com/jaredreisinger/committed/pkg/commit"
	"github.com/jaredreisinger/committed/pkg/config"
)

func TestRunCommitForm(t *testing.T) {
	cfg := config.DefaultConfig()
	m := newCommitForm(cfg, &commit.Message{})
	testModel := teatest.NewTestModel(t, m)

	testModel.Send(tea.KeyPressMsg{Code: tea.KeyDown}) // build->chore
	testModel.Send(tea.KeyPressMsg{Code: tea.KeyDown}) // chore->ci
	testModel.Send(tea.KeyPressMsg{Code: tea.KeyDown}) // ci->docs
	testModel.Send(tea.KeyPressMsg{Code: tea.KeyDown}) // docs->feat
	testModel.Send(tea.KeyPressMsg{Code: tea.KeyDown}) // feat->fix
	testModel.Send(tea.KeyPressMsg{Code: tea.KeyTab})  // accept "fix" -> scope

	testModel.Type("foo")

	testModel.Send(tea.KeyPressMsg{Code: tea.KeyTab}) // scope -> desc

	testModel.Type("testing 1 2 3")

	testModel.Send(tea.KeyPressMsg{Code: tea.KeyTab}) // desc -> body

	testModel.Type("this is the body")

	testModel.Quit()
	testModel.WaitFinished(t, teatest.WithFinalTimeout(time.Second))

	finalModel := testModel.FinalModel(t)
	commitForm2, ok := finalModel.(commitForm)
	assert.True(t, ok)

	msg, err := commitForm2.Result()
	assert.NoError(t, err)
	assert.Equal(t, &commit.Message{
		Type:        "fix",
		Scope:       "foo",
		Description: "testing 1 2 3",
		Body:        "this is the body",
	}, msg)
}

func TestRunSimpleEditor(t *testing.T) {
	m := newSimpleEditor("")
	testModel := teatest.NewTestModel(t, m)

	testModel.Send(tea.KeyPressMsg{Code: tea.KeyDown}) // dummy...

	testModel.Type("foo")

	testModel.Quit()
	testModel.WaitFinished(t, teatest.WithFinalTimeout(time.Second))

	finalModel := testModel.FinalModel(t)
	simpleEditor2, ok := finalModel.(simpleEditor)
	assert.True(t, ok)

	msg, err := simpleEditor2.Result()
	assert.NoError(t, err)
	assert.Equal(t, "foo", msg)
}
