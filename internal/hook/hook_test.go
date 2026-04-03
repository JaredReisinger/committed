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
	if err != nil {
		t.Fatal(err)
	}

	result, err := ReadMessageFile(filepath)
	if err != nil {
		t.Fatal(err)
	}

	expected := "feat: add new feature\n\nThis is a detailed description.\n\nCloses: #123"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestReadMessageFile_WithTrailingNewlines(t *testing.T) {
	tmp := t.TempDir()
	filepath := filepath.Join(tmp, "COMMIT_EDITMSG")

	content := "feat: add feature\n\n"
	err := os.WriteFile(filepath, []byte(content), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	result, err := ReadMessageFile(filepath)
	if err != nil {
		t.Fatal(err)
	}

	expected := "feat: add feature"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestReadMessageFile_Nonexistent(t *testing.T) {
	_, err := ReadMessageFile("/nonexistent/file")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestWriteMessageFile(t *testing.T) {
	tmp := t.TempDir()
	filepath := filepath.Join(tmp, "COMMIT_EDITMSG")

	content := "feat: add new feature\n\nThis is a detailed description."
	err := WriteMessageFile(filepath, content)
	if err != nil {
		t.Fatal(err)
	}

	// Verify file was written correctly
	data, err := os.ReadFile(filepath)
	if err != nil {
		t.Fatal(err)
	}

	expected := content + "\n" // Should add trailing newline
	if string(data) != expected {
		t.Errorf("expected %q, got %q", expected, string(data))
	}

	// Verify permissions
	info, err := os.Stat(filepath)
	if err != nil {
		t.Fatal(err)
	}
	mode := info.Mode()
	if mode.Perm() != 0o644 {
		t.Errorf("expected permissions 0644, got %o", mode.Perm())
	}
}

func TestWriteMessageFile_AlreadyHasNewline(t *testing.T) {
	tmp := t.TempDir()
	filepath := filepath.Join(tmp, "COMMIT_EDITMSG")

	content := "feat: add feature\n"
	err := WriteMessageFile(filepath, content)
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath)
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != content {
		t.Errorf("expected %q, got %q", content, string(data))
	}
}

func TestWriteMessageFile_WriteError(t *testing.T) {
	// Try to write to a directory (should fail)
	err := WriteMessageFile("/tmp", "content")
	if err == nil {
		t.Error("expected error when writing to directory")
	}
}
