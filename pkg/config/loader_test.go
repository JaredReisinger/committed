package config

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-openapi/testify/v2/assert"
)

func init() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))
}

// use testdata files...
func TestParseConfig_Simple(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join("testdata", "simple", "*"))
	assert.NoError(t, err)

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			t.Logf("testing %s", path)
			cfg, err := parseConfigFile(path)
			assert.NoError(t, err)

			assert.Equal(t, []string{"feat", "fix"}, cfg.Rules.TypeEnum.Value)
			assert.Equal(t, 50, cfg.Rules.SubjectMaxLength.Value)
		})
	}
}

func TestLoadConfig_Default(t *testing.T) {
	cfg, err := LoadConfig(t.TempDir())
	assert.NoError(t, err)

	types := cfg.Rules.TypeEnum.Value
	assert.NotNil(t, types)
	assert.NotEqual(t, []string{}, types)
}
