package core

import (
	"fmt"
	"log/slog"
	"os"
	"slices"

	"github.com/pkg/errors"

	"github.com/jaredreisinger/committed/internal/git"
	"github.com/jaredreisinger/committed/internal/tui"
	"github.com/jaredreisinger/committed/pkg/commit"
	"github.com/jaredreisinger/committed/pkg/config"
)

var testBypass = false

func Run(args []string, dryRun bool, logFile string) error {
	var logger *slog.Logger
	if logFile != "" {
		file, err := os.Create(logFile)
		if err != nil {
			return err
		}
		logger = slog.New(slog.NewJSONHandler(file, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	} else {
		logger = slog.New(slog.DiscardHandler)
	}
	slog.SetDefault(logger)

	slog.Info("commit initialized", "args", fmt.Sprintf("%s", args), "logFile", logFile, "dryRun", dryRun)

	file, messageType, err := git.ExtractArgs(args, dryRun)
	if err != nil {
		return errors.WithMessage(err, "error parsing hook arguments")
	}

	slog.Info("invoked with file", "filename", file, "messageType", messageType)

	useCommitForm := slices.Contains([]git.MessageType{git.CommitMessage, git.MergeMessage, git.SquashMessage}, messageType)

	if useCommitForm {
		cfg, err := config.LoadConfig(".")
		if err != nil {
			return errors.WithMessage(err, "error loading config")
		}

		var incoming string

		if file != "" {
			incoming, err = git.ReadMessageFile(file)
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
			msg, err = tui.RunCommitForm(cfg, msg)
			if err != nil {
				return err
			}
		}

		formattedMessage := msg.Format(cfg)
		if !dryRun {
			err = git.WriteMessageFile(file, formattedMessage)
			if err != nil {
				return errors.WithMessage(err, "error writing commit message")
			}
		} else {
			fmt.Printf("\nDRY-RUN: committed would have written:\n---\n%s---\n", formattedMessage)
		}
	} else {
		msg, err := tui.RunSimpleEditor("")
		if err != nil {
			return err
		}
		if !dryRun {
			fmt.Printf("\nNYI: committed would have written:\n---\n%s---\n", msg)
			err = nil
			// err = git.WriteMessageFile(file, formattedMessage)
			if err != nil {
				return errors.WithMessage(err, "error writing file")
			}
		} else {
			fmt.Printf("\nDRY-RUN: committed would have written:\n---\n%s---\n", msg)
		}
	}
	return nil
}
