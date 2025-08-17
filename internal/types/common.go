package types

import (
	"fmt"
	"time"
)

// Platform represents the operating system platform
type Platform string

const (
	PlatformWindows Platform = "windows"
	PlatformLinux   Platform = "linux"
	PlatformDarwin  Platform = "darwin"
)

// Direction represents firewall rule direction
type Direction string

const (
	DirectionInbound  Direction = "inbound"
	DirectionOutbound Direction = "outbound"
)

// Action represents firewall rule action
type Action string

const (
	ActionAllow Action = "allow"
	ActionBlock Action = "block"
)

// Protocol represents network protocol
type Protocol string

const (
	ProtocolTCP  Protocol = "tcp"
	ProtocolUDP  Protocol = "udp"
	ProtocolICMP Protocol = "icmp"
	ProtocolAny  Protocol = "any"
)

// ServiceStatus represents the status of a system service
type ServiceStatus string

const (
	ServiceStatusRunning ServiceStatus = "running"
	ServiceStatusStopped ServiceStatus = "stopped"
	ServiceStatusPending ServiceStatus = "pending"
	ServiceStatusUnknown ServiceStatus = "unknown"
)

// ServiceAction represents actions that can be performed on services
type ServiceAction string

const (
	ServiceActionStart   ServiceAction = "start"
	ServiceActionStop    ServiceAction = "stop"
	ServiceActionRestart ServiceAction = "restart"
	ServiceActionReload  ServiceAction = "reload"
)

// StartType represents how a service starts
type StartType string

const (
	StartTypeAutomatic StartType = "automatic"
	StartTypeManual    StartType = "manual"
	StartTypeDisabled  StartType = "disabled"
)

// AuthMethod represents authentication methods for remote connections
type AuthMethod string

const (
	AuthMethodKey      AuthMethod = "key"
	AuthMethodPassword AuthMethod = "password"
)

// HealthLevel represents the health status level
type HealthLevel string

const (
	HealthLevelHealthy  HealthLevel = "healthy"
	HealthLevelWarning  HealthLevel = "warning"
	HealthLevelCritical HealthLevel = "critical"
	HealthLevelUnknown  HealthLevel = "unknown"
)

// Timestamp represents a point in time with formatting utilities
type Timestamp struct {
	time.Time
}

// NewTimestamp creates a new timestamp with the current time
func NewTimestamp() Timestamp {
	return Timestamp{Time: time.Now()}
}

// FormatDuration formats a duration in a human-readable way
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return d.Round(time.Second).String()
	}
	if d < time.Hour {
		return d.Round(time.Minute).String()
	}
	if d < 24*time.Hour {
		return d.Round(time.Hour).String()
	}
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	return fmt.Sprintf("%dd %dh", days, hours)
}

// FormatBytes formats bytes in human readable format
func FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
