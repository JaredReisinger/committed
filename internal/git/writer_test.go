package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-openapi/testify/v2/assert"
)

func TestWriteMessageFile(t *testing.T) {
	tmp := t.TempDir()
	filepath := filepath.Join(tmp, "COMMIT_EDITMSG")

	content := "feat: add new feature\n\nThis is a detailed description."
	err := WriteMessageFile(filepath, content)
	assert.NoError(t, err)

	// Verify file was written correctly
	data, err := os.ReadFile(filepath)
	assert.NoError(t, err)

	expected := content + "\n" // Should add trailing newline
	assert.Equal(t, expected, string(data))

	// Verify permissions
	info, err := os.Stat(filepath)
	assert.NoError(t, err)

	mode := info.Mode()
	assert.Equal(t, 0o644, int(mode.Perm()))
}

func TestWriteMessageFile_AlreadyHasNewline(t *testing.T) {
	tmp := t.TempDir()
	filepath := filepath.Join(tmp, "COMMIT_EDITMSG")

	content := "feat: add feature\n"
	err := WriteMessageFile(filepath, content)
	assert.NoError(t, err)

	data, err := os.ReadFile(filepath)
	assert.NoError(t, err)

	assert.Equal(t, content, string(data))
}

func TestWriteMessageFile_WriteError(t *testing.T) {
	// Try to write to a directory (should fail)
	err := WriteMessageFile("/tmp", "content")
	assert.Error(t, err)
}
