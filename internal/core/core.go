package core

import (
	"github.com/pkg/errors"

	"github.com/jaredreisinger/committed/internal/config"
	"github.com/jaredreisinger/committed/internal/hook"
	"github.com/jaredreisinger/committed/pkg/commit"
	"github.com/jaredreisinger/committed/pkg/tui"
)

func Run(args []string) error {
	hookCtx, err := hook.ParseHookArgs(args)
	if err != nil {
		return errors.WithMessage(err, "error parsing hook arguments")
	}

	cfg, err := config.LoadConfig(".")
	if err != nil {
		return errors.WithMessage(err, "error loading config")
	}

	existingContent, err := hook.ReadMessageFile(hookCtx.MessageFilePath)
	if err != nil {
		return errors.WithMessage(err, "error reading commit message")
	}

	var existingMsg *commit.Message
	if existingContent != "" {
		parsed, err := commit.ParseMessage(existingContent)
		if err != nil {
			existingMsg = &commit.Message{Description: "", Body: existingContent}
		} else {
			existingMsg = parsed
		}
	}

	finalMsg, err := tui.Run(cfg, existingMsg)
	if err != nil {
		return errors.WithMessage(err, "error running TUI")
	}

	formatted := finalMsg.String()
	err = hook.WriteMessageFile(hookCtx.MessageFilePath, formatted)
	if err != nil {
		return errors.WithMessage(err, "error writing commit message")
	}

	return nil
}
