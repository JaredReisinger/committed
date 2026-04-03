package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_JSON(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, ".commitlintrc.json")
	data := []byte(`{"rules":{"type-enum":[2,"always",["feat","fix"]],"subject-max-length":[2,"always",50]}}`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Types) != 2 || cfg.Types[0] != "feat" || cfg.Types[1] != "fix" {
		t.Fatalf("unexpected types: %#v", cfg.Types)
	}
	if cfg.SubjectMaxLength != 50 {
		t.Fatalf("expected subject max length 50 got %d", cfg.SubjectMaxLength)
	}
}

func TestLoadConfig_YAML(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, ".commitlintrc.yaml")
	data := []byte("rules:\n  type-enum:\n    - 2\n    - always\n    - [feat, fix]\n  subject-max-length:\n    - 2\n    - always\n    - 60\n")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Types) != 2 || cfg.Types[1] != "fix" {
		t.Fatalf("unexpected types: %#v", cfg.Types)
	}
	if cfg.SubjectMaxLength != 60 {
		t.Fatalf("expected 60 got %d", cfg.SubjectMaxLength)
	}
}

func TestLoadConfig_PackageJSON(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "package.json")
	data := []byte(`{"commitizen":{"path":"@commitlint/config-conventional"}}`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Types) != 1 || cfg.Types[0] != "@commitlint/config-conventional" {
		t.Fatalf("unexpected types: %#v", cfg.Types)
	}
}

func TestLoadConfig_Default(t *testing.T) {
	cfg, err := LoadConfig(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Types) == 0 {
		t.Fatal("expected default types")
	}
}
