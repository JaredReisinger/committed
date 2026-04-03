package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

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
	Commitlint *commitlintCfg `json:"commitlint"`
	Commitizen any            `json:"commitizen"`
}

func parseConfigFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	file := filepath.Base(path)
	ext := filepath.Ext(file)
	base := strings.TrimSuffix(file, ext)
	cfg := DefaultConfig()

	switch base {
	case ".commitlintrc":
		raw := commitlintCfg{}
		switch ext {
		case ".json":
			err = json.Unmarshal(data, &raw)
		case ".yaml", ".yml":
			err = yaml.Unmarshal(data, &raw)
		}
		if err != nil {
			return nil, err
		}

		// should take more than rules!
		applyCommitlintRules(cfg, raw.Rules)
		return cfg, nil

	case "package":
		if ext != ".json" {
			return nil, errors.New("bad extension")
		}
		raw := packageJson{}
		err = json.Unmarshal(data, &raw)
		if err != nil {
			return nil, err
		}

		if raw.Commitlint != nil {
			applyCommitlintRules(cfg, raw.Commitlint.Rules)
			return cfg, nil
		}
	}

	return nil, fmt.Errorf("unsupported config extension: %s", ext)
}
