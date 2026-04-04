package hook

import (
	"errors"
	"fmt"
	"os"
)

// ExtractArgs parses command line arguments from commit-msg or prepare-commit-msg hook.
// The hook receives: <commit-msg-file> [<source>] [<SHA1>]
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
