package security

import (
	"regexp"
	"strings"

	"suppercommand/pkg/errors"
)

// Validator interface defines input validation methods
type Validator interface {
	ValidateCommand(cmd string, args []string) error
	SanitizeInput(input string) (string, error)
	CheckPrivileges(cmd Command) (*PrivilegeInfo, error)
}

// Command interface for privilege checking
type Command interface {
	Name() string
	RequiresElevation() bool
	SupportedPlatforms() []string
}

// PrivilegeInfo contains privilege information
type PrivilegeInfo struct {
	IsElevated    bool
	CanElevate    bool
	RequiredLevel PrivilegeLevel
	Platform      string
	Capabilities  []string
}

// PrivilegeLevel represents different privilege levels
type PrivilegeLevel int

const (
	PrivilegeLevelUser PrivilegeLevel = iota
	PrivilegeLevelAdmin
	PrivilegeLevelRoot
)

// ValidationResult contains validation results
type ValidationResult struct {
	IsValid   bool
	Errors    []ValidationError
	Warnings  []string
	Sanitized string
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
	Code    string
}

// BasicValidator implements the Validator interface
type BasicValidator struct {
	dangerousPatterns []*regexp.Regexp
	allowedChars      *regexp.Regexp
	maxInputLength    int
}

// NewValidator creates a new validator instance
func NewValidator() *BasicValidator {
	// Define dangerous patterns that could indicate injection attempts
	dangerousPatterns := []*regexp.Regexp{
		regexp.MustCompile(`[;&|><\$` + "`" + `]`),   // Shell metacharacters
		regexp.MustCompile(`\.\./`),                  // Path traversal
		regexp.MustCompile(`^-`),                     // Commands starting with dash
		regexp.MustCompile(`\x00`),                   // Null bytes
		regexp.MustCompile(`[\r\n]`),                 // Line breaks
		regexp.MustCompile(`(rm\s+-rf|del\s+/[sq])`), // Dangerous delete commands
	}

	// Allow alphanumeric, spaces, and common safe punctuation
	allowedChars := regexp.MustCompile(`^[a-zA-Z0-9\s\-_./:@]+$`)

	return &BasicValidator{
		dangerousPatterns: dangerousPatterns,
		allowedChars:      allowedChars,
		maxInputLength:    1024, // Maximum input length
	}
}

// ValidateCommand validates a command and its arguments
func (v *BasicValidator) ValidateCommand(cmd string, args []string) error {
	// Check command name
	if err := v.validateString(cmd, "command"); err != nil {
		return err
	}

	// Check each argument with more lenient validation
	for i, arg := range args {
		if err := v.validateArgument(arg, "argument"); err != nil {
			return errors.Wrap(err, "argument %d validation failed", i)
		}
	}

	return nil
}

// SanitizeInput sanitizes user input
func (v *BasicValidator) SanitizeInput(input string) (string, error) {
	if len(input) > v.maxInputLength {
		return "", errors.NewValidationError("input too long")
	}

	// Remove null bytes
	sanitized := strings.ReplaceAll(input, "\x00", "")

	// Remove carriage returns and line feeds
	sanitized = strings.ReplaceAll(sanitized, "\r", "")
	sanitized = strings.ReplaceAll(sanitized, "\n", "")

	// Trim whitespace
	sanitized = strings.TrimSpace(sanitized)

	return sanitized, nil
}

// CheckPrivileges checks if a command can be executed with current privileges
func (v *BasicValidator) CheckPrivileges(cmd Command) (*PrivilegeInfo, error) {
	// This is a placeholder implementation
	// Real implementation would check actual system privileges
	return &PrivilegeInfo{
		IsElevated:    false,
		CanElevate:    true,
		RequiredLevel: PrivilegeLevelUser,
		Platform:      "unknown",
		Capabilities:  []string{},
	}, nil
}

// validateString validates a single string input
func (v *BasicValidator) validateString(input, fieldName string) error {
	if len(input) == 0 {
		return errors.NewValidationError("%s cannot be empty", fieldName)
	}

	if len(input) > v.maxInputLength {
		return errors.NewValidationError("%s too long", fieldName)
	}

	// Check for dangerous patterns
	for _, pattern := range v.dangerousPatterns {
		if pattern.MatchString(input) {
			return errors.NewSecurityError("potentially dangerous input detected in %s", fieldName)
		}
	}

	return nil
}

// validateArgument validates command arguments with more lenient rules
func (v *BasicValidator) validateArgument(input, fieldName string) error {
	if len(input) > v.maxInputLength {
		return errors.NewValidationError("%s too long", fieldName)
	}

	// Check for dangerous patterns but skip the "starts with dash" pattern for arguments
	dangerousPatterns := []*regexp.Regexp{
		regexp.MustCompile(`[;&|><\$` + "`" + `]`),   // Shell metacharacters
		regexp.MustCompile(`\.\./`),                  // Path traversal
		regexp.MustCompile(`\x00`),                   // Null bytes
		regexp.MustCompile(`[\r\n]`),                 // Line breaks
		regexp.MustCompile(`(rm\s+-rf|del\s+/[sq])`), // Dangerous delete commands
	}

	for _, pattern := range dangerousPatterns {
		if pattern.MatchString(input) {
			return errors.NewSecurityError("potentially dangerous input detected in %s", fieldName)
		}
	}

	return nil
}
