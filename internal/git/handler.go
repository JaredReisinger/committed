package git

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

type MessageType int

const (
	UnknownMessage MessageType = iota
	CommitMessage
	MergeMessage
	SquashMessage
	TagMessage
	RebasePicker
)

// ExtractArgs parses command line arguments, handling core.editor or the
// prepare-commit-msg or commit-msg hook patterns:
//
// core.editor:        <msg-file>
// prepare-commit-msg: <msg-file> <source> <sha1>
// commit-msg:         <msg-file>
//
// The <msg-file> is typically one of:
//
//	'.git/COMMIT_EDITMSG'   (commits, amends, cherry-pick, and rebase reword)
//	'.git/MERGE_MSG'        (merge)
//	'.git/SQUASH_MSG'       (squash)
//	'.git/TAG_EDITMSG'      (tag)
//	'.git/git-rebase-todo'  (rebase picker, should warn)
func ExtractArgs(args []string, dryRun bool) (string, MessageType, error) {
	c := len(args)

	if c <= 0 && !dryRun {
		err := errors.New("insufficient arguments: expected at least message file path")
		return "", UnknownMessage, err
	}

	var filename, source, ref string

	switch c {
	default:
		fallthrough
	case 3:
		ref = args[2]
		fallthrough
	case 2:
		source = args[1]
		fallthrough
	case 1:
		filename = args[0]
		fallthrough
	case 0:
		// no-op
	}

	if source != "" || ref != "" {
		slog.Warn("called as hook?", "msgSource", source, "ref", ref)
	}

	slog.Debug("extracted args", "count", c, "msgFile", filename, "msgSource", source, "ref", ref)

	var messageType MessageType
	var err error
	if filename != "" {
		messageType, err = InferMessageType(filename)
		if err != nil {
			return "", messageType, err
		}
	} else {
		messageType = CommitMessage
	}

	return filename, messageType, nil
}

func InferMessageType(filename string) (MessageType, error) {
	// In theory we should be able to trust the complete message path (as in
	// '.git/COMMIT_EDITMSG'), but (a) path separators will differ on Windows,
	// (b) a bare repo doesn't have '.git', and (c) we might as well be as
	// robust as possible.
	slog.Warn(strings.ToLower(filepath.Base(filename)))
	switch strings.ToLower(filepath.Base(filename)) {
	case "commit_editmsg":
		return CommitMessage, nil
	case "merge_msg":
		return MergeMessage, nil
	case "squash_msg":
		return SquashMessage, nil
	case "tag_editmsg":
		return TagMessage, nil
	case "git-rebase-todo":
		return RebasePicker, nil
	}

	return UnknownMessage, fmt.Errorf("unknown message file name %q", filename)
}

// ReadMessageFile reads the commit message from the specified file.
func ReadMessageFile(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("read commit message file: %w", err)
	}

	// Git commit message files may contain trailing newlines, trim them
	content := string(data)
	for len(content) > 0 && content[len(content)-1] == '\n' {
		content = content[:len(content)-1]
	}

	return content, nil
}
