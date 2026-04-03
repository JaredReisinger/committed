package hook

import (
	"fmt"
	"os"
)

// WriteMessageFile writes the formatted commit message to the specified file.
// It preserves file permissions and handles write errors gracefully.
func WriteMessageFile(filepath, content string) error {
	// Ensure content ends with a newline (git convention)
	if content != "" && content[len(content)-1] != '\n' {
		content += "\n"
	}

	err := os.WriteFile(filepath, []byte(content), 0o644)
	if err != nil {
		return fmt.Errorf("write commit message file: %w", err)
	}

	return nil
}
