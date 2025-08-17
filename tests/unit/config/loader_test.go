package config_test

import (
	"testing"

	"suppercommand/internal/config"
)

func TestLoader_LoadWithDefaults(t *testing.T) {
	loader := config.NewLoader()
	cfg := loader.LoadWithDefaults()

	if cfg == nil {
		t.Fatal("LoadWithDefaults() returned nil")
	}

	// Test shell defaults
	if cfg.Shell.Prompt == "" {
		t.Error("Shell prompt should have default value")
	}
	if cfg.Shell.HistorySize <= 0 {
		t.Error("Shell history size should be positive")
	}

	// Test intelligence defaults
	if cfg.Intelligence.MaxSuggestions <= 0 {
		t.Error("Intelligence max suggestions should be positive")
	}

	// Test security defaults
	if cfg.Security.MaxInputLength <= 0 {
		t.Error("Security max input length should be positive")
	}

	// Test monitoring defaults
	if cfg.Monitoring.LogLevel == "" {
		t.Error("Monitoring log level should have default value")
	}

	// Test commands defaults
	if cfg.Commands.MaxConcurrent <= 0 {
		t.Error("Commands max concurrent should be positive")
	}
}

func TestConfigValidator_Validate(t *testing.T) {
	validator := config.NewConfigValidator()
	loader := config.NewLoader()
	cfg := loader.LoadWithDefaults()

	err := validator.Validate(cfg)
	if err != nil {
		t.Errorf("Validate() failed for default config: %v", err)
	}

	// Test invalid config
	invalidCfg := &config.Config{
		Shell: config.ShellConfig{
			Prompt:      "", // Invalid: empty prompt
			HistorySize: -1, // Invalid: negative history size
		},
	}

	err = validator.Validate(invalidCfg)
	if err == nil {
		t.Error("Validate() should fail for invalid config")
	}
}
