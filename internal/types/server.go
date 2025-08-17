package types

import (
	"context"
	"time"
)

// ServerManager defines the interface for server management operations
type ServerManager interface {
	GetHealthStatus(ctx context.Context) (*HealthStatus, error)
	ListServices(ctx context.Context) ([]*ServiceInfo, error)
	ControlService(ctx context.Context, serviceName string, action ServiceAction) error
	GetActiveUsers(ctx context.Context) ([]*UserSession, error)
	GetServiceLogs(ctx context.Context, serviceName string, tail bool) (*LogStream, error)
	ConfigureAlerts(ctx context.Context, config *AlertConfig) error
	BackupConfiguration(ctx context.Context, backupPath string) error
}

// HealthStatus represents the overall server health status
type HealthStatus struct {
	Overall     HealthLevel                `json:"overall"`
	Timestamp   time.Time                  `json:"timestamp"`
	Components  map[string]ComponentHealth `json:"components"`
	Uptime      time.Duration              `json:"uptime"`
	LoadAverage []float64                  `json:"load_average"`
	Alerts      []HealthAlert              `json:"alerts"`
}

// ComponentHealth represents the health of a system component
type ComponentHealth struct {
	Status      HealthLevel `json:"status"`
	Value       float64     `json:"value"`
	Threshold   float64     `json:"threshold"`
	Unit        string      `json:"unit"`
	Message     string      `json:"message"`
	LastChecked time.Time   `json:"last_checked"`
}

// ServiceInfo contains information about a system service
type ServiceInfo struct {
	Name        string        `json:"name"`
	DisplayName string        `json:"display_name"`
	Status      ServiceStatus `json:"status"`
	StartType   StartType     `json:"start_type"`
	PID         int           `json:"pid"`
	Memory      int64         `json:"memory"`      // Memory usage in bytes
	CPU         float64       `json:"cpu"`         // CPU usage percentage
	Uptime      time.Duration `json:"uptime"`      // Service uptime
	Description string        `json:"description"` // Service description
	Path        string        `json:"path"`        // Executable path
}

// UserSession contains information about an active user session
type UserSession struct {
	Username     string        `json:"username"`
	SessionID    string        `json:"session_id"`
	Terminal     string        `json:"terminal"`
	LoginTime    time.Time     `json:"login_time"`
	IdleTime     time.Duration `json:"idle_time"`
	RemoteHost   string        `json:"remote_host,omitempty"`
	ProcessCount int           `json:"process_count"`
}

// SessionEvent represents a session change event
type SessionEvent struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	Type      string                 `json:"type"`
	Username  string                 `json:"username"`
	SessionID string                 `json:"session_id"`
	Details   map[string]interface{} `json:"details"`
}

// LogStream represents a stream of log entries
type LogStream struct {
	ServiceName string     `json:"service_name"`
	Entries     []LogEntry `json:"entries"`
	Following   bool       `json:"following"`
	StartTime   time.Time  `json:"start_time"`
}

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Source    string    `json:"source"`
}

// AlertConfig contains alert configuration settings
type AlertConfig struct {
	Enabled       bool                      `json:"enabled"`
	Thresholds    map[string]AlertThreshold `json:"thresholds"`
	Notifications []NotificationConfig      `json:"notifications"`
	CheckInterval time.Duration             `json:"check_interval"`
}

// AlertThreshold defines a threshold for generating alerts
type AlertThreshold struct {
	Metric   string  `json:"metric"`   // CPU, Memory, Disk, etc.
	Warning  float64 `json:"warning"`  // Warning threshold
	Critical float64 `json:"critical"` // Critical threshold
	Unit     string  `json:"unit"`     // Percentage, bytes, etc.
	Enabled  bool    `json:"enabled"`
}

// NotificationConfig defines how alerts are delivered
type NotificationConfig struct {
	Type     string            `json:"type"`   // email, webhook, log
	Target   string            `json:"target"` // email address, webhook URL
	Enabled  bool              `json:"enabled"`
	Settings map[string]string `json:"settings"` // Additional settings
}

// HealthAlert represents a health-related alert
type HealthAlert struct {
	ID           string      `json:"id"`
	Level        HealthLevel `json:"level"`
	Component    string      `json:"component"`
	Message      string      `json:"message"`
	Timestamp    time.Time   `json:"timestamp"`
	Acknowledged bool        `json:"acknowledged"`
	Value        float64     `json:"value"`
	Threshold    float64     `json:"threshold"`
}

// BackupInfo contains information about a configuration backup
type BackupInfo struct {
	Path       string    `json:"path"`
	Timestamp  time.Time `json:"timestamp"`
	Size       int64     `json:"size"`
	Components []string  `json:"components"` // What was backed up
	Checksum   string    `json:"checksum"`
}

// ServiceControlResult contains the result of a service control operation
type ServiceControlResult struct {
	ServiceName    string        `json:"service_name"`
	Action         ServiceAction `json:"action"`
	Success        bool          `json:"success"`
	Message        string        `json:"message"`
	PreviousStatus ServiceStatus `json:"previous_status"`
	NewStatus      ServiceStatus `json:"new_status"`
	Duration       time.Duration `json:"duration"`
}

// GetHealthSummary returns a summary of the health status
func (h *HealthStatus) GetHealthSummary() string {
	switch h.Overall {
	case HealthLevelHealthy:
		return "System is running normally"
	case HealthLevelWarning:
		return "System has some warnings that need attention"
	case HealthLevelCritical:
		return "System has critical issues that require immediate attention"
	default:
		return "System health status is unknown"
	}
}

// IsRunning returns true if the service is running
func (s *ServiceInfo) IsRunning() bool {
	return s.Status == ServiceStatusRunning
}

// GetUptimeString returns a formatted uptime string
func (s *ServiceInfo) GetUptimeString() string {
	return FormatDuration(s.Uptime)
}

// GetMemoryString returns a formatted memory usage string
func (s *ServiceInfo) GetMemoryString() string {
	return FormatBytes(uint64(s.Memory))
}
