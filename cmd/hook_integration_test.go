package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jaredreisinger/committed/internal/config"
)

func TestHookIntegration(t *testing.T) {
	// Create a temporary directory for testing
	tmp := t.TempDir()

	// Create a sample commit message file
	msgFile := filepath.Join(tmp, "COMMIT_EDITMSG")
	initialContent := "feat: initial commit\n\nSome description here."
	err := os.WriteFile(msgFile, []byte(initialContent), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	// Change to temp directory and create config
	oldWd, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(oldWd)

	// Create a sample config file
	configContent := `{"rules":{"type-enum":[2,"always",["feat","fix","docs","style","refactor","perf","test","chore","revert"]]}}`
	err = os.WriteFile(".commitlintrc.json", []byte(configContent), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	// Test config loading
	cfg, err := config.LoadConfig(".")
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Types) != 9 {
		t.Errorf("expected 9 types, got %d", len(cfg.Types))
	}

	// Verify initial message file exists
	content, err := os.ReadFile(msgFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != initialContent {
		t.Errorf("expected %q, got %q", initialContent, string(content))
	}
}
