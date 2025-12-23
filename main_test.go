package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")

	configJSON := `{
		"label_actions": [
			{
				"label": "INBOX",
				"actions": [
					{
						"subject_filter": "Test",
						"download_attachment": true,
						"mark_as_read": false,
						"delete_email": false,
						"save_to": "/tmp/test"
					}
				]
			}
		]
	}`

	if err := os.WriteFile(configFile, []byte(configJSON), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Test loading config
	config, err := loadConfig(configFile)
	if err != nil {
		t.Fatalf("loadConfig() error = %v, want nil", err)
	}

	if config == nil {
		t.Fatal("loadConfig() config = nil, want non-nil")
	}

	if len(config.LabelActions) != 1 {
		t.Errorf("loadConfig() LabelActions length = %d, want 1", len(config.LabelActions))
	}

	if config.LabelActions[0].Label != "INBOX" {
		t.Errorf("loadConfig() Label = %v, want INBOX", config.LabelActions[0].Label)
	}

	if len(config.LabelActions[0].Actions) != 1 {
		t.Errorf("loadConfig() Actions length = %d, want 1", len(config.LabelActions[0].Actions))
	}

	action := config.LabelActions[0].Actions[0]
	if action.SubjectFilter != "Test" {
		t.Errorf("loadConfig() SubjectFilter = %v, want Test", action.SubjectFilter)
	}
	if !action.Download {
		t.Errorf("loadConfig() Download = %v, want true", action.Download)
	}
}

func TestLoadConfig_NotFound(t *testing.T) {
	// Test with non-existent file
	config, err := loadConfig("nonexistent.json")
	if err == nil {
		t.Error("loadConfig() error = nil, want error")
	}
	if config != nil {
		t.Errorf("loadConfig() config = %v, want nil", config)
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	// Create a temporary file with invalid JSON
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "invalid.json")

	if err := os.WriteFile(configFile, []byte("invalid json"), 0644); err != nil {
		t.Fatalf("Failed to write invalid config file: %v", err)
	}

	// Test loading invalid config
	config, err := loadConfig(configFile)
	if err == nil {
		t.Error("loadConfig() error = nil, want error for invalid JSON")
	}
	if config != nil {
		t.Errorf("loadConfig() config = %v, want nil", config)
	}
}

func TestLoadConfig_EmptyFile(t *testing.T) {
	// Create an empty file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "empty.json")

	if err := os.WriteFile(configFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to write empty config file: %v", err)
	}

	// Test loading empty config
	config, err := loadConfig(configFile)
	if err == nil {
		t.Error("loadConfig() error = nil, want error for empty file")
	}
	if config != nil {
		t.Errorf("loadConfig() config = %v, want nil", config)
	}
}

