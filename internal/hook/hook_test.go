package hook

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseHookArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected *HookContext
		hasError bool
	}{
		{
			name:     "insufficient args",
			args:     []string{"program"},
			expected: nil,
			hasError: true,
		},
		{
			name: "minimum args",
			args: []string{"program", "/path/to/.git/COMMIT_EDITMSG"},
			expected: &HookContext{
				MessageFilePath: "/path/to/.git/COMMIT_EDITMSG",
				SourceType:      "",
				SourceObject:    "",
			},
			hasError: false,
		},
		{
			name: "with source type",
			args: []string{"program", "/path/to/.git/COMMIT_EDITMSG", "message"},
			expected: &HookContext{
				MessageFilePath: "/path/to/.git/COMMIT_EDITMSG",
				SourceType:      "message",
				SourceObject:    "",
			},
			hasError: false,
		},
		{
			name: "with source and object",
			args: []string{"program", "/path/to/.git/COMMIT_EDITMSG", "commit", "abc123"},
			expected: &HookContext{
				MessageFilePath: "/path/to/.git/COMMIT_EDITMSG",
				SourceType:      "commit",
				SourceObject:    "abc123",
			},
			hasError: false,
		},
		{
			name: "extra args ignored",
			args: []string{"program", "/path/to/.git/COMMIT_EDITMSG", "merge", "abc123", "extra"},
			expected: &HookContext{
				MessageFilePath: "/path/to/.git/COMMIT_EDITMSG",
				SourceType:      "merge",
				SourceObject:    "abc123",
			},
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseHookArgs(tt.args)

			if tt.hasError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result.MessageFilePath != tt.expected.MessageFilePath {
				t.Errorf("expected MessageFilePath %q, got %q", tt.expected.MessageFilePath, result.MessageFilePath)
			}
			if result.SourceType != tt.expected.SourceType {
				t.Errorf("expected SourceType %q, got %q", tt.expected.SourceType, result.SourceType)
			}
			if result.SourceObject != tt.expected.SourceObject {
				t.Errorf("expected SourceObject %q, got %q", tt.expected.SourceObject, result.SourceObject)
			}
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
