package tui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/pkg/errors"

	"github.com/jaredreisinger/committed/internal/config"
	"github.com/jaredreisinger/committed/pkg/commit"
)

func Run(cfg *config.Config, existingMsg *commit.Message) (*commit.Message, error) {
	model := NewModel(cfg, existingMsg)
	p := tea.NewProgram(model)

	finalModel, err := p.Run()
	if err != nil {
		return nil, errors.WithMessage(err, "error running TUI")
	}

	tuiModel, ok := finalModel.(Model)
	if !ok {
		return nil, errors.WithMessage(err, "unexpected model type")
	}

	resultMsg, err := tuiModel.Result()
	if err != nil {
		return nil, errors.WithMessage(err, "form not completed")
	}

	return resultMsg, nil
}
