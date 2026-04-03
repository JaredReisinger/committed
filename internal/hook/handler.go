package hook

import (
	"errors"
	"fmt"
	"os"
)

// ParseHookArgs parses command line arguments from prepare-commit-msg hook.
// The hook receives: <commit-msg-file> [<source>] [<SHA1>]
func ParseHookArgs(args []string) (*HookContext, error) {
	if len(args) < 2 {
		return nil, errors.New("insufficient arguments: expected at least message file path")
	}

	ctx := &HookContext{
		MessageFilePath: args[1], // args[0] is program name, args[1] is message file
	}

	if len(args) >= 3 {
		ctx.SourceType = args[2]
	}

	if len(args) >= 4 {
		ctx.SourceObject = args[3]
	}

	return ctx, nil
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
