package tui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/pkg/errors"

	"github.com/jaredreisinger/committed/pkg/commit"
	"github.com/jaredreisinger/committed/pkg/config"
)

var programTestHook *tea.Program

func RunCommitForm(cfg *config.Config, existingMsg *commit.Message) (*commit.Message, error) {
	m := newCommitForm(cfg, existingMsg)

	p := tea.NewProgram(m)
	programTestHook = p

	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	tuiModel, ok := finalModel.(commitForm)
	if !ok {
		return nil, errors.WithMessage(err, "unexpected model type")
	}

	resultMsg, err := tuiModel.Result()
	if err != nil {
		return nil, errors.WithMessage(err, "form not completed")
	}

	return resultMsg, nil
}

// TODO: set up super-simple editor
func RunSimpleEditor(existingMsg string) (string, error) {
	m := newSimpleEditor(existingMsg)

	p := tea.NewProgram(m)
	programTestHook = p

	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	tuiModel, ok := finalModel.(simpleEditor)
	if !ok {
		return "", errors.WithMessage(err, "unexpected model type")
	}

	resultMsg, err := tuiModel.Result()
	if err != nil {
		return "", errors.WithMessage(err, "form not completed")
	}

	return resultMsg, nil
}
