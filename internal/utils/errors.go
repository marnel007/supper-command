package utils

import (
	"fmt"
	"suppercommand/internal/types"
	"suppercommand/pkg/errors"
)

// WrapFirewallError wraps an error as a firewall error
func WrapFirewallError(operation string, platform types.Platform, err error, message string) error {
	if err == nil {
		return types.NewFirewallError(operation, platform, nil, message)
	}
	return errors.Wrap(types.NewFirewallError(operation, platform, err, message), "firewall operation failed")
}

// WrapPerformanceError wraps an error as a performance error
func WrapPerformanceError(component, operation string, err error, message string) error {
	if err == nil {
		return types.NewPerformanceError(component, operation, nil, message)
	}
	return errors.Wrap(types.NewPerformanceError(component, operation, err, message), "performance operation failed")
}

// WrapServiceError wraps an error as a service error
func WrapServiceError(serviceName, operation string, err error, message string) error {
	if err == nil {
		return types.NewServiceError(serviceName, operation, nil, message)
	}
	return errors.Wrap(types.NewServiceError(serviceName, operation, err, message), "service operation failed")
}

// WrapRemoteError wraps an error as a remote error
func WrapRemoteError(serverID, host, operation string, err error, message string) error {
	if err == nil {
		return types.NewRemoteError(serverID, host, operation, nil, message)
	}
	return errors.Wrap(types.NewRemoteError(serverID, host, operation, err, message), "remote operation failed")
}

// FormatError formats an error with additional context
func FormatError(err error, context string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", context, err)
}

// IsRetryableError determines if an error is retryable
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for specific error types that are retryable
	if types.IsRemoteError(err) {
		// Network errors are typically retryable
		return true
	}

	// Add more retryable error conditions as needed
	return false
}

// GetErrorCategory returns the category of an error
func GetErrorCategory(err error) string {
	if err == nil {
		return "none"
	}

	switch {
	case types.IsFirewallError(err):
		return "firewall"
	case types.IsPerformanceError(err):
		return "performance"
	case types.IsServiceError(err):
		return "service"
	case types.IsRemoteError(err):
		return "remote"
	default:
		return "general"
	}
}
