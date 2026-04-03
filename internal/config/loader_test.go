package config

import (
	"path/filepath"
	"testing"

	"github.com/go-openapi/testify/v2/assert"
)

// use testdata files...
func TestParseConfig_Simple(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join("testdata", "simple", "*"))
	assert.NoError(t, err)

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			cfg, err := parseConfigFile(path)
			assert.NoError(t, err)

			assert.Equal(t, []string{"feat", "fix"}, cfg.Types)
			assert.Equal(t, 50, cfg.SubjectMaxLength)
		})
	}
}

func TestLoadConfig_Default(t *testing.T) {
	cfg, err := LoadConfig(t.TempDir())
	assert.NoError(t, err)

	assert.NotNil(t, cfg.Types)
	assert.NotEqual(t, []string{}, cfg.Types)
}
