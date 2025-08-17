package types

import (
	"errors"
	"fmt"
)

// FirewallError represents firewall-specific errors
type FirewallError struct {
	Operation string
	Platform  Platform
	Cause     error
	Message   string
}

func (e *FirewallError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("firewall %s failed on %s: %s", e.Operation, e.Platform, e.Message)
	}
	if e.Cause != nil {
		return fmt.Sprintf("firewall %s failed on %s: %v", e.Operation, e.Platform, e.Cause)
	}
	return fmt.Sprintf("firewall %s failed on %s", e.Operation, e.Platform)
}

func (e *FirewallError) Unwrap() error {
	return e.Cause
}

// NewFirewallError creates a new firewall error
func NewFirewallError(operation string, platform Platform, cause error, message string) *FirewallError {
	return &FirewallError{
		Operation: operation,
		Platform:  platform,
		Cause:     cause,
		Message:   message,
	}
}

// PerformanceError represents performance monitoring errors
type PerformanceError struct {
	Component string
	Operation string
	Cause     error
	Message   string
}

func (e *PerformanceError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("performance %s/%s failed: %s", e.Component, e.Operation, e.Message)
	}
	if e.Cause != nil {
		return fmt.Sprintf("performance %s/%s failed: %v", e.Component, e.Operation, e.Cause)
	}
	return fmt.Sprintf("performance %s/%s failed", e.Component, e.Operation)
}

func (e *PerformanceError) Unwrap() error {
	return e.Cause
}

// NewPerformanceError creates a new performance error
func NewPerformanceError(component, operation string, cause error, message string) *PerformanceError {
	return &PerformanceError{
		Component: component,
		Operation: operation,
		Cause:     cause,
		Message:   message,
	}
}

// ServiceError represents service management errors
type ServiceError struct {
	ServiceName string
	Operation   string
	Cause       error
	Message     string
}

func (e *ServiceError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("service %s %s failed: %s", e.ServiceName, e.Operation, e.Message)
	}
	if e.Cause != nil {
		return fmt.Sprintf("service %s %s failed: %v", e.ServiceName, e.Operation, e.Cause)
	}
	return fmt.Sprintf("service %s %s failed", e.ServiceName, e.Operation)
}

func (e *ServiceError) Unwrap() error {
	return e.Cause
}

// NewServiceError creates a new service error
func NewServiceError(serviceName, operation string, cause error, message string) *ServiceError {
	return &ServiceError{
		ServiceName: serviceName,
		Operation:   operation,
		Cause:       cause,
		Message:     message,
	}
}

// RemoteError represents remote execution errors
type RemoteError struct {
	ServerID  string
	Host      string
	Operation string
	Cause     error
	Message   string
}

func (e *RemoteError) Error() string {
	server := e.ServerID
	if server == "" {
		server = e.Host
	}
	if e.Message != "" {
		return fmt.Sprintf("remote %s on %s failed: %s", e.Operation, server, e.Message)
	}
	if e.Cause != nil {
		return fmt.Sprintf("remote %s on %s failed: %v", e.Operation, server, e.Cause)
	}
	return fmt.Sprintf("remote %s on %s failed", e.Operation, server)
}

func (e *RemoteError) Unwrap() error {
	return e.Cause
}

// NewRemoteError creates a new remote error
func NewRemoteError(serverID, host, operation string, cause error, message string) *RemoteError {
	return &RemoteError{
		ServerID:  serverID,
		Host:      host,
		Operation: operation,
		Cause:     cause,
		Message:   message,
	}
}

// IsFirewallError checks if an error is a firewall error
func IsFirewallError(err error) bool {
	var firewallErr *FirewallError
	return errors.As(err, &firewallErr)
}

// IsPerformanceError checks if an error is a performance error
func IsPerformanceError(err error) bool {
	var perfErr *PerformanceError
	return errors.As(err, &perfErr)
}

// IsServiceError checks if an error is a service error
func IsServiceError(err error) bool {
	var serviceErr *ServiceError
	return errors.As(err, &serviceErr)
}

// IsRemoteError checks if an error is a remote error
func IsRemoteError(err error) bool {
	var remoteErr *RemoteError
	return errors.As(err, &remoteErr)
}
