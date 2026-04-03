package config

// TODO: flesh out Rules values?
type commitlintCfg struct {
	Rules map[string]any `json:"rules" yaml:"rules"`
}

// applying to default config may not be quite right... it depends on what
// defaults commitlint defines

// We'll need more comprehensive testing of rule types, I suspect.
func applyCommitlintRules(cfg *Config, rules map[string]any) {
	if rules == nil {
		return
	}

	// so gross!
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
