package hook

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
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
//	.git/COMMIT_EDITMSG   (commits, amends, cherry-pick, and rebase reword)
//	.git/MERGE_MSG        (merge)
//	.git/SQUASH_MSG       (squash)
//	.git/TAG_EDITMSG      (tag)
//	.git/git-rebase-todo  (rebase picker, should warn)
func ExtractArgs(args []string, dryRun bool) (file string, source string, ref string, err error) {
	c := len(args)

	if c <= 0 && !dryRun {
		err = errors.New("insufficient arguments: expected at least message file path")
		return
	}

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
		file = args[0]
		fallthrough
	case 0:
		// no-op
	}

	slog.Debug("extracted args", "count", c, "msgFile", file, "msgSource", source, "ref", ref)

	return
}

// ReadMessageFile reads the commit message from the specified file.
func ReadMessageFile(filepath string) (string, error) {
	data, err := os.ReadFile(filepath)
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
