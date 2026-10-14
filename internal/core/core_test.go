package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-openapi/testify/v2/assert"
)

const defaultPerms = 0o644

func TestRun(t *testing.T) {
	testBypass = true

	tmpFile := createTempMessageFile(t, "COMMIT_EDITMSG", `feat: simple feature`)

	err := Run([]string{tmpFile}, false, "")
	assert.NoError(t, err)
}

func TestRunWithLogFile(t *testing.T) {
	testBypass = true
	logFile := filepath.Join(t.TempDir(), "log.json")
	err := Run([]string{}, true, logFile)
	assert.NoError(t, err)

	b, err := os.ReadFile(logFile)
	assert.NoError(t, err)
	assert.Contains(t, string(b), "commit initialized")
}

func TestRunNoArgs(t *testing.T) {
	testBypass = true
	err := Run([]string{}, false, "")
	assert.Error(t, err)
}

func TestRunNoArgsDryRun(t *testing.T) {
	testBypass = true
	err := Run([]string{}, true, "")
	assert.NoError(t, err)
}

func TestRunSimpleEditor(t *testing.T) {
	testBypass = true

	tmpFile := createTempMessageFile(t, "TAG_EDITMSG", `tag message`)

	err := Run([]string{tmpFile}, false, "")
	assert.NoError(t, err)
}

func TestRunSimpleEditorDryRun(t *testing.T) {
	testBypass = true

	tmpFile := createTempMessageFile(t, "TAG_EDITMSG", `tag message`)

	err := Run([]string{tmpFile}, true, "")
	assert.NoError(t, err)
}

func createTempMessageFile(t *testing.T, filename string, content string) string {
	tmpFile := filepath.Join(t.TempDir(), filename)
	err := os.WriteFile(tmpFile, []byte(content), defaultPerms)
	assert.NoError(t, err)
	return tmpFile
}
