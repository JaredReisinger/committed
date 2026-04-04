package hook

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-openapi/testify/v2/assert"
)

func TestExtractArgs(t *testing.T) {
	const file = "/path/to/.git/COMMIT_EDITMSG"

	tests := []struct {
		name           string
		args           []string
		dryRun         bool
		expectedFile   string
		expectedSource string
		expectedRef    string
		hasError       bool
	}{
		{
			name:     "insufficient args",
			args:     []string{},
			hasError: true,
		},
		{
			name:   "insufficient args allowed for dry run",
			args:   []string{},
			dryRun: true,
		},
		{
			name:         "minimum args",
			args:         []string{file},
			expectedFile: file,
		},
		{
			name:           "with source type",
			args:           []string{file, "message"},
			expectedFile:   file,
			expectedSource: "message",
		},
		{
			name:           "with source and object",
			args:           []string{file, "commit", "abc123"},
			expectedFile:   file,
			expectedSource: "commit",
			expectedRef:    "abc123",
		},
		{
			name:           "extra args ignored",
			args:           []string{file, "merge", "abc123", "extra"},
			expectedFile:   file,
			expectedSource: "merge",
			expectedRef:    "abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualFile, actualSource, actualRef, err := ExtractArgs(tt.args, tt.dryRun)

			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedFile, actualFile)
			assert.Equal(t, tt.expectedSource, actualSource)
			assert.Equal(t, tt.expectedRef, actualRef)
		})
	}
}

func TestReadMessageFile(t *testing.T) {
	tmp := t.TempDir()
	filepath := filepath.Join(tmp, "COMMIT_EDITMSG")

	content := "feat: add new feature\n\nThis is a detailed description.\n\nCloses: #123\n"
	err := os.WriteFile(filepath, []byte(content), 0o644)
	assert.NoError(t, err)

	result, err := ReadMessageFile(filepath)
	assert.NoError(t, err)

	expected := "feat: add new feature\n\nThis is a detailed description.\n\nCloses: #123"
	assert.Equal(t, expected, result)
}

func TestReadMessageFile_WithTrailingNewlines(t *testing.T) {
	tmp := t.TempDir()
	filepath := filepath.Join(tmp, "COMMIT_EDITMSG")

	content := "feat: add feature\n\n"
	err := os.WriteFile(filepath, []byte(content), 0o644)
	assert.NoError(t, err)

	result, err := ReadMessageFile(filepath)
	assert.NoError(t, err)

	expected := "feat: add feature"
	assert.Equal(t, expected, result)
}

func TestReadMessageFile_Nonexistent(t *testing.T) {
	_, err := ReadMessageFile("/nonexistent/file")
	assert.Error(t, err)
}

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
