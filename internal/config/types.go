package config

import "time"

// Config represents the main configuration structure
type Config struct {
	Shell        ShellConfig        `yaml:"shell" json:"shell"`
	Intelligence IntelligenceConfig `yaml:"intelligence" json:"intelligence"`
	Security     SecurityConfig     `yaml:"security" json:"security"`
	Monitoring   MonitoringConfig   `yaml:"monitoring" json:"monitoring"`
	Commands     CommandsConfig     `yaml:"commands" json:"commands"`
}

// ShellConfig contains shell-specific configuration
type ShellConfig struct {
	Prompt        string        `yaml:"prompt" json:"prompt"`
	HistorySize   int           `yaml:"history_size" json:"history_size"`
	Timeout       time.Duration `yaml:"timeout" json:"timeout"`
	Colors        ColorConfig   `yaml:"colors" json:"colors"`
	AutoComplete  bool          `yaml:"auto_complete" json:"auto_complete"`
	CaseSensitive bool          `yaml:"case_sensitive" json:"case_sensitive"`
	SaveHistory   bool          `yaml:"save_history" json:"save_history"`
	HistoryFile   string        `yaml:"history_file" json:"history_file"`
}

// ColorConfig contains color configuration
type ColorConfig struct {
	Enabled      bool              `yaml:"enabled" json:"enabled"`
	Scheme       string            `yaml:"scheme" json:"scheme"`
	CustomColors map[string]string `yaml:"custom_colors" json:"custom_colors"`
}

// IntelligenceConfig contains AI/ML feature configuration
type IntelligenceConfig struct {
	Enabled                 bool          `yaml:"enabled" json:"enabled"`
	FuzzyMatchingEnabled    bool          `yaml:"fuzzy_matching_enabled" json:"fuzzy_matching_enabled"`
	SmartSuggestionsEnabled bool          `yaml:"smart_suggestions_enabled" json:"smart_suggestions_enabled"`
	ExternalToolsEnabled    bool          `yaml:"external_tools_enabled" json:"external_tools_enabled"`
	LearningEnabled         bool          `yaml:"learning_enabled" json:"learning_enabled"`
	MaxSuggestions          int           `yaml:"max_suggestions" json:"max_suggestions"`
	MinPatternOccurrences   int           `yaml:"min_pattern_occurrences" json:"min_pattern_occurrences"`
	LearningThreshold       int           `yaml:"learning_threshold" json:"learning_threshold"`
	ResponseTimeout         time.Duration `yaml:"response_timeout" json:"response_timeout"`
	DataDirectory           string        `yaml:"data_directory" json:"data_directory"`
}

// SecurityConfig contains security-related configuration
type SecurityConfig struct {
	ValidationEnabled   bool     `yaml:"validation_enabled" json:"validation_enabled"`
	SanitizationEnabled bool     `yaml:"sanitization_enabled" json:"sanitization_enabled"`
	MaxInputLength      int      `yaml:"max_input_length" json:"max_input_length"`
	AllowedCommands     []string `yaml:"allowed_commands" json:"allowed_commands"`
	BlockedCommands     []string `yaml:"blocked_commands" json:"blocked_commands"`
	AllowElevation      bool     `yaml:"allow_elevation" json:"allow_elevation"`
	RequireConfirmation bool     `yaml:"require_confirmation" json:"require_confirmation"`
	LogSecurityEvents   bool     `yaml:"log_security_events" json:"log_security_events"`
	StrictMode          bool     `yaml:"strict_mode" json:"strict_mode"`
}

// MonitoringConfig contains monitoring and logging configuration
type MonitoringConfig struct {
	Enabled             bool          `yaml:"enabled" json:"enabled"`
	LogLevel            string        `yaml:"log_level" json:"log_level"`
	LogFile             string        `yaml:"log_file" json:"log_file"`
	LogRotation         bool          `yaml:"log_rotation" json:"log_rotation"`
	MaxLogSize          int64         `yaml:"max_log_size" json:"max_log_size"`
	MaxLogFiles         int           `yaml:"max_log_files" json:"max_log_files"`
	MetricsEnabled      bool          `yaml:"metrics_enabled" json:"metrics_enabled"`
	MetricsInterval     time.Duration `yaml:"metrics_interval" json:"metrics_interval"`
	PerformanceTracking bool          `yaml:"performance_tracking" json:"performance_tracking"`
	MemoryTracking      bool          `yaml:"memory_tracking" json:"memory_tracking"`
}

// CommandsConfig contains command-specific configuration
type CommandsConfig struct {
	Timeout          time.Duration            `yaml:"timeout" json:"timeout"`
	MaxConcurrent    int                      `yaml:"max_concurrent" json:"max_concurrent"`
	RetryAttempts    int                      `yaml:"retry_attempts" json:"retry_attempts"`
	RetryDelay       time.Duration            `yaml:"retry_delay" json:"retry_delay"`
	CustomCommands   map[string]CommandConfig `yaml:"custom_commands" json:"custom_commands"`
	DisabledCommands []string                 `yaml:"disabled_commands" json:"disabled_commands"`
}

// CommandConfig contains configuration for individual commands
type CommandConfig struct {
	Enabled          bool          `yaml:"enabled" json:"enabled"`
	Timeout          time.Duration `yaml:"timeout" json:"timeout"`
	RequireElevation bool          `yaml:"require_elevation" json:"require_elevation"`
	AllowedArgs      []string      `yaml:"allowed_args" json:"allowed_args"`
	BlockedArgs      []string      `yaml:"blocked_args" json:"blocked_args"`
	MaxArgs          int           `yaml:"max_args" json:"max_args"`
}
