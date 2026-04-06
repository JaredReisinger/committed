package core

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/jaredreisinger/committed/internal/config"
	"github.com/jaredreisinger/committed/internal/hook"
	"github.com/jaredreisinger/committed/internal/tui"
	"github.com/jaredreisinger/committed/pkg/commit"
)

var testBypass = false

func Run(args []string, dryRun bool) error {
	file, _, _, err := hook.ExtractArgs(args, dryRun)
	if err != nil {
		return errors.WithMessage(err, "error parsing hook arguments")
	}

	cfg, err := config.LoadConfig(".")
	if err != nil {
		return errors.WithMessage(err, "error loading config")
	}

	var incoming string

	if file != "" {
		incoming, err = hook.ReadMessageFile(file)
		if err != nil {
			return errors.WithMessage(err, "error reading commit message")
		}
	}

	// ParseMessage should *really* take the config, shouldn't it?
	msg, err := commit.ParseMessage(incoming)
	if err != nil {
		msg = &commit.Message{Body: incoming}
	}

	// Haven't figured out a good way to test the UI yet!
	if !testBypass {
		msg, err = tui.Run(cfg, msg)
		if err != nil {
			return err
		}
	}

	if !dryRun {
		err = hook.WriteMessageFile(file, msg.String())
		if err != nil {
			return errors.WithMessage(err, "error writing commit message")
		}
	} else {
		fmt.Printf("DRY-RUN: committed would have written:\n---\n%s---\n", msg.String())
	}

	return nil
}
