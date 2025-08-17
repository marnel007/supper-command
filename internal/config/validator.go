package config

import (
	"fmt"
	"strings"
	"time"
)

// ConfigValidator validates configuration values
type ConfigValidator struct{}

// NewConfigValidator creates a new configuration validator
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{}
}

// Validate validates the entire configuration
func (v *ConfigValidator) Validate(config *Config) error {
	if config == nil {
		return fmt.Errorf("configuration cannot be nil")
	}

	// Validate shell configuration
	if err := v.validateShellConfig(&config.Shell); err != nil {
		return fmt.Errorf("shell configuration error: %w", err)
	}

	// Validate intelligence configuration
	if err := v.validateIntelligenceConfig(&config.Intelligence); err != nil {
		return fmt.Errorf("intelligence configuration error: %w", err)
	}

	// Validate security configuration
	if err := v.validateSecurityConfig(&config.Security); err != nil {
		return fmt.Errorf("security configuration error: %w", err)
	}

	// Validate monitoring configuration
	if err := v.validateMonitoringConfig(&config.Monitoring); err != nil {
		return fmt.Errorf("monitoring configuration error: %w", err)
	}

	// Validate commands configuration
	if err := v.validateCommandsConfig(&config.Commands); err != nil {
		return fmt.Errorf("commands configuration error: %w", err)
	}

	return nil
}

// validateShellConfig validates shell configuration
func (v *ConfigValidator) validateShellConfig(config *ShellConfig) error {
	if config.Prompt == "" {
		return fmt.Errorf("prompt cannot be empty")
	}

	if config.HistorySize < 0 {
		return fmt.Errorf("history size cannot be negative")
	}

	if config.HistorySize > 100000 {
		return fmt.Errorf("history size too large (max: 100000)")
	}

	if config.Timeout < 0 {
		return fmt.Errorf("timeout cannot be negative")
	}

	if config.Timeout > 24*time.Hour {
		return fmt.Errorf("timeout too large (max: 24h)")
	}

	// Validate color scheme
	validSchemes := []string{"default", "dark", "light", "custom"}
	if !contains(validSchemes, config.Colors.Scheme) {
		return fmt.Errorf("invalid color scheme: %s (valid: %s)",
			config.Colors.Scheme, strings.Join(validSchemes, ", "))
	}

	return nil
}

// validateIntelligenceConfig validates intelligence configuration
func (v *ConfigValidator) validateIntelligenceConfig(config *IntelligenceConfig) error {
	if config.MaxSuggestions < 0 {
		return fmt.Errorf("max suggestions cannot be negative")
	}

	if config.MaxSuggestions > 100 {
		return fmt.Errorf("max suggestions too large (max: 100)")
	}

	if config.MinPatternOccurrences < 0 {
		return fmt.Errorf("min pattern occurrences cannot be negative")
	}

	if config.LearningThreshold < 0 {
		return fmt.Errorf("learning threshold cannot be negative")
	}

	if config.ResponseTimeout < 0 {
		return fmt.Errorf("response timeout cannot be negative")
	}

	if config.ResponseTimeout > 10*time.Second {
		return fmt.Errorf("response timeout too large (max: 10s)")
	}

	return nil
}

// validateSecurityConfig validates security configuration
func (v *ConfigValidator) validateSecurityConfig(config *SecurityConfig) error {
	if config.MaxInputLength < 1 {
		return fmt.Errorf("max input length must be positive")
	}

	if config.MaxInputLength > 1024*1024 {
		return fmt.Errorf("max input length too large (max: 1MB)")
	}

	// Validate command lists don't have conflicts
	for _, allowed := range config.AllowedCommands {
		if contains(config.BlockedCommands, allowed) {
			return fmt.Errorf("command '%s' cannot be both allowed and blocked", allowed)
		}
	}

	return nil
}

// validateMonitoringConfig validates monitoring configuration
func (v *ConfigValidator) validateMonitoringConfig(config *MonitoringConfig) error {
	// Validate log level
	validLevels := []string{"debug", "info", "warn", "error"}
	if !contains(validLevels, strings.ToLower(config.LogLevel)) {
		return fmt.Errorf("invalid log level: %s (valid: %s)",
			config.LogLevel, strings.Join(validLevels, ", "))
	}

	if config.MaxLogSize < 0 {
		return fmt.Errorf("max log size cannot be negative")
	}

	if config.MaxLogFiles < 0 {
		return fmt.Errorf("max log files cannot be negative")
	}

	if config.MetricsInterval < 0 {
		return fmt.Errorf("metrics interval cannot be negative")
	}

	if config.MetricsInterval > 24*time.Hour {
		return fmt.Errorf("metrics interval too large (max: 24h)")
	}

	return nil
}

// validateCommandsConfig validates commands configuration
func (v *ConfigValidator) validateCommandsConfig(config *CommandsConfig) error {
	if config.Timeout < 0 {
		return fmt.Errorf("timeout cannot be negative")
	}

	if config.MaxConcurrent < 1 {
		return fmt.Errorf("max concurrent must be positive")
	}

	if config.MaxConcurrent > 1000 {
		return fmt.Errorf("max concurrent too large (max: 1000)")
	}

	if config.RetryAttempts < 0 {
		return fmt.Errorf("retry attempts cannot be negative")
	}

	if config.RetryAttempts > 10 {
		return fmt.Errorf("retry attempts too large (max: 10)")
	}

	if config.RetryDelay < 0 {
		return fmt.Errorf("retry delay cannot be negative")
	}

	// Validate custom commands
	for name, cmdConfig := range config.CustomCommands {
		if name == "" {
			return fmt.Errorf("custom command name cannot be empty")
		}

		if cmdConfig.MaxArgs < 0 {
			return fmt.Errorf("max args for command '%s' cannot be negative", name)
		}

		if cmdConfig.Timeout < 0 {
			return fmt.Errorf("timeout for command '%s' cannot be negative", name)
		}
	}

	return nil
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
