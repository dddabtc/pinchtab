package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// GetConfigPath returns the path to the config file
func GetConfigPath() string {
	return os.Getenv("BRIDGE_CONFIG") // No fallback; let the caller decide
}

// GetConfigPathOrDefault returns config path with sensible defaults
func GetConfigPathOrDefault() string {
	if configPath := os.Getenv("BRIDGE_CONFIG"); configPath != "" {
		return configPath
	}
	return filepath.Join(userConfigDir(), "config.json")
}

// LoadFileConfig loads the config file from disk
func LoadFileConfig() (*FileConfig, error) {
	configPath := GetConfigPathOrDefault()

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var fc FileConfig
	if err := json.Unmarshal(data, &fc); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &fc, nil
}

// SaveFileConfig writes the config file to disk
func SaveFileConfig(fc *FileConfig) error {
	configPath := GetConfigPathOrDefault()

	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal as JSON
	data, err := json.MarshalIndent(fc, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// SetConfigValue sets a single config value by key path
// Examples:
//
//	"server.port" = "9867"
//	"chrome.headless" = "true"
//	"chrome.maxTabs" = "50"
func SetConfigValue(key, value string) error {
	fc, err := LoadFileConfig()
	if err != nil {
		// If file doesn't exist, start with defaults
		fc = &FileConfig{}
	}

	parts := strings.Split(key, ".")
	if len(parts) < 2 {
		return fmt.Errorf("invalid key format: use 'section.key' (e.g., 'server.port')")
	}

	section := parts[0]
	fieldName := parts[1]

	switch section {
	case "server":
		if err := setServerField(fc, fieldName, value); err != nil {
			return err
		}
	case "chrome":
		if err := setChromeField(fc, fieldName, value); err != nil {
			return err
		}
	case "orchestrator":
		if err := setOrchestratorField(fc, fieldName, value); err != nil {
			return err
		}
	case "timeouts":
		if err := setTimeoutField(fc, fieldName, value); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown section: %s", section)
	}

	if err := SaveFileConfig(fc); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✅ Set %s.%s = %s\n", section, fieldName, value)
	return nil
}

func setServerField(fc *FileConfig, field, value string) error {
	switch field {
	case "port":
		fc.Port = value
	case "stateDir":
		fc.StateDir = value
	case "profileDir":
		fc.ProfileDir = value
	case "token":
		fc.Token = value
	case "cdpUrl":
		fc.CdpURL = value
	default:
		return fmt.Errorf("unknown server field: %s", field)
	}
	return nil
}

func setChromeField(fc *FileConfig, field, value string) error {
	switch field {
	case "headless":
		b, err := parseBool(value)
		if err != nil {
			return err
		}
		fc.Headless = &b
	case "maxTabs":
		n, err := parseInt(value)
		if err != nil {
			return err
		}
		fc.MaxTabs = &n
	case "noRestore":
		b, err := parseBool(value)
		if err != nil {
			return err
		}
		fc.NoRestore = b
	default:
		return fmt.Errorf("unknown chrome field: %s", field)
	}
	return nil
}

func setOrchestratorField(fc *FileConfig, field, value string) error {
	switch field {
	case "strategy":
		fc.Strategy = value
	case "allocationPolicy":
		fc.AllocationPolicy = value
	case "instancePortStart":
		n, err := parseInt(value)
		if err != nil {
			return err
		}
		fc.InstancePortStart = &n
	case "instancePortEnd":
		n, err := parseInt(value)
		if err != nil {
			return err
		}
		fc.InstancePortEnd = &n
	default:
		return fmt.Errorf("unknown orchestrator field: %s", field)
	}
	return nil
}

func setTimeoutField(fc *FileConfig, field, value string) error {
	switch field {
	case "actionSec", "timeoutSec":
		n, err := parseInt(value)
		if err != nil {
			return err
		}
		fc.TimeoutSec = n
	case "navigateSec":
		n, err := parseInt(value)
		if err != nil {
			return err
		}
		fc.NavigateSec = n
	default:
		return fmt.Errorf("unknown timeout field: %s", field)
	}
	return nil
}

// PatchConfigJSON merges a JSON object into the config
func PatchConfigJSON(jsonStr string) error {
	fc, err := LoadFileConfig()
	if err != nil {
		fc = &FileConfig{}
	}

	var patch map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &patch); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	// Apply patch by converting to JSON and back (simple merge)
	// This preserves all existing fields
	currentData, _ := json.Marshal(fc)
	var currentMap map[string]interface{}
	if err := json.Unmarshal(currentData, &currentMap); err != nil {
		return fmt.Errorf("failed to parse current config: %w", err)
	}

	// Merge patch into current
	for key, val := range patch {
		currentMap[key] = val
	}

	// Convert back to FileConfig
	mergedData, _ := json.Marshal(currentMap)
	if err := json.Unmarshal(mergedData, fc); err != nil {
		return fmt.Errorf("failed to apply patch: %w", err)
	}

	if err := SaveFileConfig(fc); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("✅ Config patched successfully")
	return nil
}

// PatchConfigYAML is an alias for PatchConfigJSON (parses YAML first)
func PatchConfigYAML(yamlStr string) error {
	var data interface{}
	if err := yaml.Unmarshal([]byte(yamlStr), &data); err != nil {
		return fmt.Errorf("invalid YAML: %w", err)
	}

	// Convert back to JSON for processing
	jsonBytes, _ := json.Marshal(data)
	return PatchConfigJSON(string(jsonBytes))
}

// ValidateConfig checks if the current config is valid
func ValidateConfig() (bool, []string) {
	fc, err := LoadFileConfig()
	if err != nil {
		return false, []string{fmt.Sprintf("Failed to load config: %v", err)}
	}

	var errs []string

	// Validate port
	if fc.Port == "" {
		errs = append(errs, "server.port is required")
	}

	// Validate port range
	if fc.InstancePortStart != nil && fc.InstancePortEnd != nil {
		if *fc.InstancePortStart >= *fc.InstancePortEnd {
			errs = append(errs, "instancePortStart must be less than instancePortEnd")
		}
	}

	// Validate timeouts
	if fc.TimeoutSec < 0 {
		errs = append(errs, "timeoutSec must be non-negative")
	}
	if fc.NavigateSec < 0 {
		errs = append(errs, "navigateSec must be non-negative")
	}

	// Validate strategy
	if fc.Strategy != "" && fc.Strategy != "simple" && fc.Strategy != "session" && fc.Strategy != "explicit" {
		errs = append(errs, fmt.Sprintf("invalid strategy: %s (must be simple, session, or explicit)", fc.Strategy))
	}

	// Validate allocation policy
	if fc.AllocationPolicy != "" && fc.AllocationPolicy != "fcfs" && fc.AllocationPolicy != "round_robin" && fc.AllocationPolicy != "random" {
		errs = append(errs, fmt.Sprintf("invalid allocationPolicy: %s (must be fcfs, round_robin, or random)", fc.AllocationPolicy))
	}

	return len(errs) == 0, errs
}

// DisplayConfig shows the current config in the specified format
func DisplayConfig(format string) error {
	configPath := GetConfigPathOrDefault()
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	// Parse JSON
	var config interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	switch format {
	case "json", "":
		// Pretty-print JSON
		pretty, _ := json.MarshalIndent(config, "", "  ")
		fmt.Println(string(pretty))
	case "yaml", "yml":
		// Convert to YAML
		yamlData, err := yaml.Marshal(config)
		if err != nil {
			return fmt.Errorf("failed to convert to YAML: %w", err)
		}
		fmt.Println(string(yamlData))
	default:
		return fmt.Errorf("unknown format: %s (use json or yaml)", format)
	}

	return nil
}

// Helper functions

func parseBool(s string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes", "on":
		return true, nil
	case "0", "false", "no", "off":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value: %s", s)
	}
}

func parseInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	if err != nil {
		return 0, fmt.Errorf("invalid integer value: %s", s)
	}
	return n, nil
}
