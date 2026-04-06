package tui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/pkg/errors"

	"github.com/jaredreisinger/committed/internal/config"
	"github.com/jaredreisinger/committed/pkg/commit"
)

var testHookProgram *tea.Program

func Run(cfg *config.Config, existingMsg *commit.Message) (*commit.Message, error) {
	m := newModel(cfg, existingMsg)

	p := tea.NewProgram(m)
	testHookProgram = p

	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	tuiModel, ok := finalModel.(mainForm)
	if !ok {
		return nil, errors.WithMessage(err, "unexpected model type")
	}

	resultMsg, err := tuiModel.Result()
	if err != nil {
		return nil, errors.WithMessage(err, "form not completed")
	}

	return resultMsg, nil
}
