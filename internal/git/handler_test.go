package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-openapi/testify/v2/assert"
)

func TestExtractArgs(t *testing.T) {
	const file = "/path/to/.git/COMMIT_EDITMSG"

	tests := []struct {
		name                string
		args                []string
		dryRun              bool
		expectedFile        string
		expectedMessageType MessageType
		hasError            bool
	}{
		{
			name:                "insufficient args",
			args:                []string{},
			hasError:            true,
			expectedMessageType: UnknownMessage,
		},
		{
			name:                "insufficient args allowed for dry run",
			args:                []string{},
			dryRun:              true,
			expectedMessageType: CommitMessage,
		},
		{
			name:                "minimum args",
			args:                []string{file},
			expectedFile:        file,
			expectedMessageType: CommitMessage,
		},
		{
			name:                "with source type",
			args:                []string{file, "message"},
			expectedFile:        file,
			expectedMessageType: CommitMessage,
		},
		{
			name:                "with source and object",
			args:                []string{file, "commit", "abc123"},
			expectedFile:        file,
			expectedMessageType: CommitMessage,
		},
		{
			name:                "extra args ignored",
			args:                []string{file, "merge", "abc123", "extra"},
			expectedFile:        file,
			expectedMessageType: CommitMessage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualFile, actualMessageType, err := ExtractArgs(tt.args, tt.dryRun)

			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedFile, actualFile)
			assert.Equal(t, tt.expectedMessageType, actualMessageType)
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
