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

	tmpFile := createTempFile(t, `feat: simple feature`)

	err := Run([]string{tmpFile}, false)
	assert.NoError(t, err)
}

func createTempFile(t *testing.T, content string) string {
	tmpFile := filepath.Join(t.TempDir(), "MSG")
	err := os.WriteFile(tmpFile, []byte(content), defaultPerms)
	assert.NoError(t, err)
	return tmpFile
}
