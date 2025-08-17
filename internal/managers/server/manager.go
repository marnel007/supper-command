package server

import (
	"context"
	"fmt"
	"time"

	"suppercommand/internal/types"
	"suppercommand/internal/utils"
)

// Factory creates server managers based on the current platform
type Factory struct{}

// NewFactory creates a new server manager factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateManager creates a server manager for the current platform
func (f *Factory) CreateManager() (types.ServerManager, error) {
	platform := utils.GetCurrentPlatform()

	switch platform {
	case types.PlatformWindows:
		return NewWindowsServerManager(), nil
	case types.PlatformLinux:
		return NewLinuxServerManager(), nil
	case types.PlatformDarwin:
		return NewDarwinServerManager(), nil
	default:
		return NewMockServerManager(), nil
	}
}

// BaseServerManager provides common functionality for all server managers
type BaseServerManager struct {
	platform types.Platform
}

// NewBaseServerManager creates a new base server manager
func NewBaseServerManager(platform types.Platform) *BaseServerManager {
	return &BaseServerManager{
		platform: platform,
	}
}

// GetPlatform returns the platform this manager is for
func (b *BaseServerManager) GetPlatform() types.Platform {
	return b.platform
}

// CalculateOverallHealth calculates overall health from component health
func (b *BaseServerManager) CalculateOverallHealth(components map[string]types.ComponentHealth) types.HealthLevel {
	if len(components) == 0 {
		return types.HealthLevelUnknown
	}

	criticalCount := 0
	warningCount := 0
	healthyCount := 0

	for _, component := range components {
		switch component.Status {
		case types.HealthLevelCritical:
			criticalCount++
		case types.HealthLevelWarning:
			warningCount++
		case types.HealthLevelHealthy:
			healthyCount++
		}
	}

	// If any component is critical, overall is critical
	if criticalCount > 0 {
		return types.HealthLevelCritical
	}

	// If any component has warnings, overall is warning
	if warningCount > 0 {
		return types.HealthLevelWarning
	}

	// If all components are healthy, overall is healthy
	if healthyCount > 0 {
		return types.HealthLevelHealthy
	}

	return types.HealthLevelUnknown
}

// GenerateHealthAlerts generates health alerts based on component status
func (b *BaseServerManager) GenerateHealthAlerts(components map[string]types.ComponentHealth) []types.HealthAlert {
	alerts := make([]types.HealthAlert, 0)

	for componentName, component := range components {
		if component.Status == types.HealthLevelCritical || component.Status == types.HealthLevelWarning {
			alert := types.HealthAlert{
				ID:           fmt.Sprintf("%s_%d", componentName, time.Now().Unix()),
				Level:        component.Status,
				Component:    componentName,
				Message:      component.Message,
				Timestamp:    time.Now(),
				Acknowledged: false,
				Value:        component.Value,
				Threshold:    component.Threshold,
			}
			alerts = append(alerts, alert)
		}
	}

	return alerts
}

// MockServerManager provides a mock implementation for testing
type MockServerManager struct {
	*BaseServerManager
	services []*types.ServiceInfo
	users    []*types.UserSession
	alerts   []types.HealthAlert
}

// NewMockServerManager creates a new mock server manager
func NewMockServerManager() *MockServerManager {
	return &MockServerManager{
		BaseServerManager: NewBaseServerManager(utils.GetCurrentPlatform()),
		services:          make([]*types.ServiceInfo, 0),
		users:             make([]*types.UserSession, 0),
		alerts:            make([]types.HealthAlert, 0),
	}
}

// GetHealthStatus returns mock health status
func (m *MockServerManager) GetHealthStatus(ctx context.Context) (*types.HealthStatus, error) {
	// Create mock component health data
	components := map[string]types.ComponentHealth{
		"CPU": {
			Status:      types.HealthLevelHealthy,
			Value:       25.5,
			Threshold:   80.0,
			Unit:        "%",
			Message:     "CPU usage is normal",
			LastChecked: time.Now(),
		},
		"Memory": {
			Status:      types.HealthLevelWarning,
			Value:       75.2,
			Threshold:   80.0,
			Unit:        "%",
			Message:     "Memory usage is elevated",
			LastChecked: time.Now(),
		},
		"Disk": {
			Status:      types.HealthLevelHealthy,
			Value:       45.8,
			Threshold:   85.0,
			Unit:        "%",
			Message:     "Disk usage is normal",
			LastChecked: time.Now(),
		},
		"Network": {
			Status:      types.HealthLevelHealthy,
			Value:       12.3,
			Threshold:   90.0,
			Unit:        "%",
			Message:     "Network utilization is low",
			LastChecked: time.Now(),
		},
	}

	// Calculate overall health
	overall := m.CalculateOverallHealth(components)

	// Generate alerts
	alerts := m.GenerateHealthAlerts(components)

	// Mock system uptime
	uptime := 72 * time.Hour // 3 days

	status := &types.HealthStatus{
		Overall:     overall,
		Timestamp:   time.Now(),
		Components:  components,
		Uptime:      uptime,
		LoadAverage: []float64{1.2, 1.5, 1.8},
		Alerts:      alerts,
	}

	return status, nil
}

// ListServices returns mock service information
func (m *MockServerManager) ListServices(ctx context.Context) ([]*types.ServiceInfo, error) {
	// Return mock services if none configured
	if len(m.services) == 0 {
		m.services = []*types.ServiceInfo{
			&types.ServiceInfo{
				Name:        "nginx",
				DisplayName: "Nginx HTTP Server",
				Status:      types.ServiceStatusRunning,
				StartType:   types.StartTypeAutomatic,
				PID:         1234,
				Memory:      50 * 1024 * 1024, // 50MB
				CPU:         2.5,
				Uptime:      24 * time.Hour,
				Description: "High-performance HTTP server",
				Path:        "/usr/sbin/nginx",
			},
			&types.ServiceInfo{
				Name:        "mysql",
				DisplayName: "MySQL Database Server",
				Status:      types.ServiceStatusRunning,
				StartType:   types.StartTypeAutomatic,
				PID:         5678,
				Memory:      200 * 1024 * 1024, // 200MB
				CPU:         5.2,
				Uptime:      48 * time.Hour,
				Description: "MySQL database server",
				Path:        "/usr/sbin/mysqld",
			},
			&types.ServiceInfo{
				Name:        "redis",
				DisplayName: "Redis Cache Server",
				Status:      types.ServiceStatusStopped,
				StartType:   types.StartTypeManual,
				PID:         0,
				Memory:      0,
				CPU:         0,
				Uptime:      0,
				Description: "In-memory data structure store",
				Path:        "/usr/bin/redis-server",
			},
		}
	}

	return m.services, nil
}

// ControlService controls a mock service
func (m *MockServerManager) ControlService(ctx context.Context, serviceName string, action types.ServiceAction) error {
	// Find the service
	for i, service := range m.services {
		if service.Name == serviceName {
			switch action {
			case types.ServiceActionStart:
				if service.Status == types.ServiceStatusStopped {
					m.services[i].Status = types.ServiceStatusRunning
					m.services[i].PID = 9999 // Mock PID
					m.services[i].Uptime = 0 // Just started
				}
			case types.ServiceActionStop:
				if service.Status == types.ServiceStatusRunning {
					m.services[i].Status = types.ServiceStatusStopped
					m.services[i].PID = 0
					m.services[i].Uptime = 0
				}
			case types.ServiceActionRestart:
				m.services[i].Status = types.ServiceStatusRunning
				m.services[i].PID = 9999 // Mock PID
				m.services[i].Uptime = 0 // Just restarted
			}
			return nil
		}
	}

	return types.NewServiceError(serviceName, string(action), nil, "service not found")
}

// GetActiveUsers returns mock active user sessions
func (m *MockServerManager) GetActiveUsers(ctx context.Context) ([]*types.UserSession, error) {
	// Return mock users if none configured
	if len(m.users) == 0 {
		m.users = []*types.UserSession{
			&types.UserSession{
				Username:     "admin",
				SessionID:    "pts/0",
				Terminal:     "console",
				LoginTime:    time.Now().Add(-2 * time.Hour),
				IdleTime:     5 * time.Minute,
				RemoteHost:   "",
				ProcessCount: 15,
			},
			&types.UserSession{
				Username:     "developer",
				SessionID:    "pts/1",
				Terminal:     "ssh",
				LoginTime:    time.Now().Add(-30 * time.Minute),
				IdleTime:     2 * time.Minute,
				RemoteHost:   "192.168.1.100",
				ProcessCount: 8,
			},
		}
	}

	return m.users, nil
}

// GetServiceLogs returns mock service logs
func (m *MockServerManager) GetServiceLogs(ctx context.Context, serviceName string, tail bool) (*types.LogStream, error) {
	// Create mock log entries
	entries := []types.LogEntry{
		{
			Timestamp: time.Now().Add(-10 * time.Minute),
			Level:     "INFO",
			Message:   fmt.Sprintf("%s service started successfully", serviceName),
			Source:    serviceName,
		},
		{
			Timestamp: time.Now().Add(-5 * time.Minute),
			Level:     "INFO",
			Message:   "Processing requests normally",
			Source:    serviceName,
		},
		{
			Timestamp: time.Now().Add(-2 * time.Minute),
			Level:     "WARN",
			Message:   "High memory usage detected",
			Source:    serviceName,
		},
		{
			Timestamp: time.Now(),
			Level:     "INFO",
			Message:   "Service health check passed",
			Source:    serviceName,
		},
	}

	logStream := &types.LogStream{
		ServiceName: serviceName,
		Entries:     entries,
		Following:   tail,
		StartTime:   time.Now(),
	}

	return logStream, nil
}

// ConfigureAlerts configures mock alert settings
func (m *MockServerManager) ConfigureAlerts(ctx context.Context, config *types.AlertConfig) error {
	// Mock implementation - would store configuration in real implementation
	return nil
}

// BackupConfiguration creates a mock configuration backup
func (m *MockServerManager) BackupConfiguration(ctx context.Context, backupPath string) error {
	// Mock implementation - would backup actual configuration files
	return nil
}

// GetSystemUptime returns system uptime
func (m *MockServerManager) GetSystemUptime() time.Duration {
	// Mock uptime - in real implementation would get actual system uptime
	return 72 * time.Hour // 3 days
}

// GetLoadAverage returns system load average
func (m *MockServerManager) GetLoadAverage() []float64 {
	// Mock load average - in real implementation would get actual load
	return []float64{1.2, 1.5, 1.8}
}

// ValidateServiceName validates a service name
func (m *MockServerManager) ValidateServiceName(serviceName string) error {
	if serviceName == "" {
		return types.NewServiceError(serviceName, "validate", nil, "service name cannot be empty")
	}

	// Check if service exists
	services, err := m.ListServices(context.Background())
	if err != nil {
		return err
	}

	for _, service := range services {
		if service.Name == serviceName {
			return nil
		}
	}

	return types.NewServiceError(serviceName, "validate", nil, "service not found")
}
