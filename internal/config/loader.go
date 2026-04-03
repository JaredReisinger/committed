package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var ErrNoConfigFound = errors.New("no commit configuration found")

// LoadConfig reads conventional commit config from .commitlintrc.json, .commitlintrc.yaml, or package.json.
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

func parseConfigFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	ext := filepath.Ext(path)
	cfg := DefaultConfig()

	if path == "package.json" || filepath.Base(path) == "package.json" {
		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			return nil, err
		}
		if commitizen, ok := raw["commitizen"]; ok {
			if czMap, ok := commitizen.(map[string]any); ok {
				if pathStr, ok := czMap["path"].(string); ok {
					cfg.Types = []string{pathStr}
				}
			}
		}
		return cfg, nil
	}

	if ext == ".json" {
		var raw struct {
			Rules map[string]any `json:"rules"`
		}
		if err := json.Unmarshal(data, &raw); err != nil {
			return nil, err
		}
		applyCommitlintRules(cfg, raw.Rules)
		return cfg, nil
	}

	if ext == ".yaml" || ext == ".yml" {
		var raw struct {
			Rules map[string]any `yaml:"rules"`
		}
		if err := yaml.Unmarshal(data, &raw); err != nil {
			return nil, err
		}
		applyCommitlintRules(cfg, raw.Rules)
		return cfg, nil
	}

	return nil, fmt.Errorf("unsupported config extension: %s", ext)
}

func applyCommitlintRules(cfg *Config, rules map[string]any) {
	if rules == nil {
		return
	}

	if t, ok := rules["type-enum"]; ok {
		if arr, ok := t.([]any); ok && len(arr) >= 3 {
			if list, ok := arr[2].([]any); ok {
				cfg.Types = make([]string, 0, len(list))
				for _, item := range list {
					if s, ok := item.(string); ok {
						cfg.Types = append(cfg.Types, s)
					}
				}
			}
		}
	}

	if s, ok := rules["subject-max-length"]; ok {
		if arr, ok := s.([]any); ok && len(arr) >= 3 {
			if n, ok := arr[2].(int); ok {
				cfg.SubjectMaxLength = n
			} else if n, ok := arr[2].(float64); ok {
				cfg.SubjectMaxLength = int(n)
			} else if n, ok := arr[2].(int64); ok {
				cfg.SubjectMaxLength = int(n)
			}
		}
	}

	if b, ok := rules["body-max-line-length"]; ok {
		if arr, ok := b.([]any); ok && len(arr) >= 3 {
			if n, ok := arr[2].(int); ok {
				cfg.BodyMaxLineLength = n
			} else if n, ok := arr[2].(float64); ok {
				cfg.BodyMaxLineLength = int(n)
			} else if n, ok := arr[2].(int64); ok {
				cfg.BodyMaxLineLength = int(n)
			}
		}
	}

	if h, ok := rules["header-max-length"]; ok {
		if arr, ok := h.([]any); ok && len(arr) >= 3 {
			if n, ok := arr[2].(int); ok {
				cfg.HeaderMaxLength = n
			} else if n, ok := arr[2].(float64); ok {
				cfg.HeaderMaxLength = int(n)
			} else if n, ok := arr[2].(int64); ok {
				cfg.HeaderMaxLength = int(n)
			}
		}
	}
}
