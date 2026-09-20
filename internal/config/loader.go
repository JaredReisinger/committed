package config

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/jaredreisinger/committed/pkg/commitlint"
	"gopkg.in/yaml.v3"
)

// LoadConfig reads conventional commit config from .commitlintrc.json,
// .commitlintrc.yaml, or package.json.
func LoadConfig(workDir string) (*Config, error) {
	paths := []string{
		".commitlintrc.json",
		".commitlintrc.yaml",
		".commitlintrc.yml",
		"package.json",
	}

	for _, entry := range paths {
		candidate := filepath.Join(workDir, entry)
		slog.Debug("checking for config", "candidate", candidate)
		if _, err := os.Stat(candidate); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return nil, fmt.Errorf("stat config file %s: %w", candidate, err)
		}

		cfg, err := parseConfigFile(candidate)
		if err != nil {
			return nil, fmt.Errorf("parse config %s: %w", candidate, err)
		}
		return cfg, nil
	}

	return DefaultConfig(), nil
}

// We really need to struct drive this...

type packageJson struct {
	Commitlint *commitlint.Config `json:"commitlint"`
	Commitizen any                `json:"commitizen"`
}

func parseConfigFile(path string) (*Config, error) {
	slog.Debug("parsing config", "path", path)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	file := filepath.Base(path)
	ext := filepath.Ext(file)
	base := strings.TrimSuffix(file, ext)
	cfg := DefaultConfig()
	parsed := true

	switch base {
	case ".commitlintrc":
		raw := commitlint.Config{}
		switch ext {
		case ".json":
			err = json.Unmarshal(data, &raw)
		case ".yaml", ".yml":
			err = yaml.Unmarshal(data, &raw)
		}
		if err != nil {
			return nil, err
		}

		// slog.Debug("parsed .commitlintrc", "raw", raw)
		cfg.applyCommitlint(raw.Rules)

	case "package":
		if ext != ".json" {
			return nil, errors.New("bad extension")
		}
		raw := packageJson{}
		err = json.Unmarshal(data, &raw)
		if err != nil {
			return nil, err
		}

		// slog.Debug("parsed package.json", "raw", raw)
		if raw.Commitlint != nil {
			cfg.applyCommitlint(raw.Commitlint.Rules)
		}
	default:
		parsed = false
	}

	if parsed {
		slog.Debug("using config", "config", cfg)
		return cfg, nil
	}

	return nil, fmt.Errorf("unsupported config extension: %s", ext)
}

func (cfg *Config) applyCommitlint(rules commitlint.Rules) {
	slog.Debug("applying commitlint rules")

	if rules.TypeEnum.Set {
		cfg.Types = rules.TypeEnum.Value
	}

	if rules.SubjectMaxLength.Set {
		cfg.SubjectMaxLength = rules.SubjectMaxLength.Value
	}

	if rules.BodyMaxLineLength.Set {
		cfg.BodyMaxLineLength = rules.BodyMaxLineLength.Value
	}

	if rules.HeaderMaxLength.Set {
		cfg.HeaderMaxLength = rules.HeaderMaxLength.Value
	}
}
