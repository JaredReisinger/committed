package tui

import (
	"testing"
	"time"

	"github.com/charmbracelet/x/exp/teatest/v2"
	"github.com/go-openapi/testify/v2/assert"
)

func TestHelp(t *testing.T) {
	m := newHelpModel()
	testModel := teatest.NewTestModel(t, m)
	testModel.Quit()
	testModel.WaitFinished(t, teatest.WithFinalTimeout(time.Second))

	finalModel := testModel.FinalModel(t)
	actualFinalModel, ok := finalModel.(helpModel)
	assert.True(t, ok)
	assert.NotNil(t, actualFinalModel)
}
