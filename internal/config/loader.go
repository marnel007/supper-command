package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// Loader interface defines configuration loading methods
type Loader interface {
	Load(path string) (*Config, error)
	LoadWithDefaults() *Config
	Validate(config *Config) error
	Watch(path string, callback func(*Config)) error
}

// BasicLoader implements the Loader interface
type BasicLoader struct {
	validator *ConfigValidator
}

// NewLoader creates a new configuration loader
func NewLoader() *BasicLoader {
	return &BasicLoader{
		validator: NewConfigValidator(),
	}
}

// Load loads configuration from a file
func (l *BasicLoader) Load(path string) (*Config, error) {
	if path == "" {
		return nil, fmt.Errorf("configuration path cannot be empty")
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := &Config{}
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse YAML config: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported config file format: %s", ext)
	}

	// Apply defaults for missing values
	l.applyDefaults(config)

	// Validate configuration
	if err := l.Validate(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// LoadWithDefaults returns a configuration with default values
func (l *BasicLoader) LoadWithDefaults() *Config {
	config := &Config{}
	l.applyDefaults(config)
	return config
}

// Validate validates a configuration
func (l *BasicLoader) Validate(config *Config) error {
	return l.validator.Validate(config)
}

// Watch watches a configuration file for changes (placeholder implementation)
func (l *BasicLoader) Watch(path string, callback func(*Config)) error {
	// This is a placeholder implementation
	// Real implementation would use file system watching
	return fmt.Errorf("configuration watching not implemented yet")
}

// applyDefaults applies default values to configuration
func (l *BasicLoader) applyDefaults(config *Config) {
	// Shell defaults
	if config.Shell.Prompt == "" {
		config.Shell.Prompt = "supershell> "
	}
	if config.Shell.HistorySize == 0 {
		config.Shell.HistorySize = 1000
	}
	if config.Shell.Timeout == 0 {
		config.Shell.Timeout = 30 * time.Second
	}
	if config.Shell.HistoryFile == "" {
		config.Shell.HistoryFile = "~/.supershell_history"
	}

	// Color defaults
	if config.Shell.Colors.CustomColors == nil {
		config.Shell.Colors.CustomColors = make(map[string]string)
	}
	if config.Shell.Colors.Scheme == "" {
		config.Shell.Colors.Scheme = "default"
	}
	config.Shell.Colors.Enabled = true
	config.Shell.AutoComplete = true
	config.Shell.SaveHistory = true

	// Intelligence defaults
	config.Intelligence.Enabled = true
	config.Intelligence.FuzzyMatchingEnabled = true
	config.Intelligence.SmartSuggestionsEnabled = true
	config.Intelligence.ExternalToolsEnabled = true
	config.Intelligence.LearningEnabled = true
	if config.Intelligence.MaxSuggestions == 0 {
		config.Intelligence.MaxSuggestions = 10
	}
	if config.Intelligence.MinPatternOccurrences == 0 {
		config.Intelligence.MinPatternOccurrences = 3
	}
	if config.Intelligence.LearningThreshold == 0 {
		config.Intelligence.LearningThreshold = 5
	}
	if config.Intelligence.ResponseTimeout == 0 {
		config.Intelligence.ResponseTimeout = 100 * time.Millisecond
	}
	if config.Intelligence.DataDirectory == "" {
		config.Intelligence.DataDirectory = "~/.supershell/intelligence"
	}

	// Security defaults
	config.Security.ValidationEnabled = true
	config.Security.SanitizationEnabled = true
	if config.Security.MaxInputLength == 0 {
		config.Security.MaxInputLength = 1024
	}
	config.Security.AllowElevation = true
	config.Security.LogSecurityEvents = true

	// Monitoring defaults
	config.Monitoring.Enabled = true
	if config.Monitoring.LogLevel == "" {
		config.Monitoring.LogLevel = "info"
	}
	if config.Monitoring.LogFile == "" {
		config.Monitoring.LogFile = "~/.supershell/logs/supershell.log"
	}
	config.Monitoring.LogRotation = true
	if config.Monitoring.MaxLogSize == 0 {
		config.Monitoring.MaxLogSize = 10 * 1024 * 1024 // 10MB
	}
	if config.Monitoring.MaxLogFiles == 0 {
		config.Monitoring.MaxLogFiles = 5
	}
	config.Monitoring.MetricsEnabled = true
	if config.Monitoring.MetricsInterval == 0 {
		config.Monitoring.MetricsInterval = 30 * time.Second
	}
	config.Monitoring.PerformanceTracking = true
	config.Monitoring.MemoryTracking = true

	// Commands defaults
	if config.Commands.Timeout == 0 {
		config.Commands.Timeout = 30 * time.Second
	}
	if config.Commands.MaxConcurrent == 0 {
		config.Commands.MaxConcurrent = 10
	}
	if config.Commands.RetryAttempts == 0 {
		config.Commands.RetryAttempts = 3
	}
	if config.Commands.RetryDelay == 0 {
		config.Commands.RetryDelay = 1 * time.Second
	}
	if config.Commands.CustomCommands == nil {
		config.Commands.CustomCommands = make(map[string]CommandConfig)
	}
}
