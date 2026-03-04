package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSetConfigValue(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.json")
	_ = os.Setenv("BRIDGE_CONFIG", configPath)
	defer func() { _ = os.Unsetenv("BRIDGE_CONFIG") }()

	// Create default config first
	fc := DefaultFileConfig()
	if err := SaveFileConfig(&fc); err != nil {
		t.Fatalf("failed to save initial config: %v", err)
	}

	tests := []struct {
		key   string
		value string
		valid bool
	}{
		{"server.port", "9999", true},
		{"server.token", "secret-key", true},
		{"chrome.headless", "false", true},
		{"chrome.maxTabs", "100", true},
		{"orchestrator.strategy", "session", true},
		{"orchestrator.allocationPolicy", "round_robin", true},
		{"timeouts.actionSec", "30", true},
		{"invalid.key", "value", false},
		{"serverport", "9999", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			err := SetConfigValue(tt.key, tt.value)
			if tt.valid && err != nil {
				t.Errorf("expected no error for valid key, got %v", err)
			}
			if !tt.valid && err == nil {
				t.Errorf("expected error for invalid key, got nil")
			}
		})
	}
}

func TestPatchConfigJSON(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.json")
	_ = os.Setenv("BRIDGE_CONFIG", configPath)
	defer func() { _ = os.Unsetenv("BRIDGE_CONFIG") }()

	// Create default config
	fc := DefaultFileConfig()
	if err := SaveFileConfig(&fc); err != nil {
		t.Fatalf("failed to save initial config: %v", err)
	}

	// Apply patch
	patch := `{"port": "9999", "headless": false}`
	if err := PatchConfigJSON(patch); err != nil {
		t.Fatalf("failed to apply patch: %v", err)
	}

	// Load and verify
	updated, err := LoadFileConfig()
	if err != nil {
		t.Fatalf("failed to load updated config: %v", err)
	}

	if updated.Port != "9999" {
		t.Errorf("expected port 9999, got %s", updated.Port)
	}
	if updated.Headless == nil || *updated.Headless {
		t.Errorf("expected headless false, got %v", updated.Headless)
	}
}

func TestValidateConfig(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.json")
	_ = os.Setenv("BRIDGE_CONFIG", configPath)
	defer func() { _ = os.Unsetenv("BRIDGE_CONFIG") }()

	// Create valid config
	fc := DefaultFileConfig()
	if err := SaveFileConfig(&fc); err != nil {
		t.Fatalf("failed to save initial config: %v", err)
	}

	isValid, errs := ValidateConfig()
	if !isValid {
		t.Errorf("expected valid config, got errors: %v", errs)
	}

	// Test invalid strategy
	invalid := DefaultFileConfig()
	invalid.Strategy = "invalid"
	if err := SaveFileConfig(&invalid); err != nil {
		t.Fatalf("failed to save invalid config: %v", err)
	}

	isValid, errs = ValidateConfig()
	if isValid {
		t.Error("expected invalid config for bad strategy")
	}
	if len(errs) == 0 {
		t.Error("expected validation errors")
	}
}

func TestDisplayConfigJSON(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.json")
	_ = os.Setenv("BRIDGE_CONFIG", configPath)
	defer func() { _ = os.Unsetenv("BRIDGE_CONFIG") }()

	fc := DefaultFileConfig()
	if err := SaveFileConfig(&fc); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Should not error
	if err := DisplayConfig("json"); err != nil {
		t.Errorf("failed to display JSON config: %v", err)
	}
}

func TestDisplayConfigYAML(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.json")
	_ = os.Setenv("BRIDGE_CONFIG", configPath)
	defer func() { _ = os.Unsetenv("BRIDGE_CONFIG") }()

	fc := DefaultFileConfig()
	if err := SaveFileConfig(&fc); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Should not error
	if err := DisplayConfig("yaml"); err != nil {
		t.Errorf("failed to display YAML config: %v", err)
	}
}

func TestParseBool(t *testing.T) {
	tests := []struct {
		input string
		want  bool
		err   bool
	}{
		{"true", true, false},
		{"false", false, false},
		{"1", true, false},
		{"0", false, false},
		{"yes", true, false},
		{"no", false, false},
		{"invalid", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseBool(tt.input)
			if (err != nil) != tt.err {
				t.Errorf("parseBool error = %v, want %v", err, tt.err)
			}
			if err == nil && got != tt.want {
				t.Errorf("parseBool got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		input string
		want  int
		err   bool
	}{
		{"42", 42, false},
		{"0", 0, false},
		{"-10", -10, false},
		{"invalid", 0, true},
		{"abc123", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseInt(tt.input)
			if (err != nil) != tt.err {
				t.Errorf("parseInt error = %v, want %v", err, tt.err)
			}
			if err == nil && got != tt.want {
				t.Errorf("parseInt got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadSaveFileConfig(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.json")
	_ = os.Setenv("BRIDGE_CONFIG", configPath)
	defer func() { _ = os.Unsetenv("BRIDGE_CONFIG") }()

	// Save
	original := DefaultFileConfig()
	original.Port = "9999"
	if err := SaveFileConfig(&original); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Load
	loaded, err := LoadFileConfig()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if loaded.Port != "9999" {
		t.Errorf("expected port 9999, got %s", loaded.Port)
	}
}
