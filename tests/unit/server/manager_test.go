package server

import (
	"context"
	"testing"
	"time"

	"suppercommand/internal/managers/server"
	"suppercommand/internal/types"
)

func TestServerManagerFactory(t *testing.T) {
	factory := server.NewFactory()
	if factory == nil {
		t.Fatal("Factory should not be nil")
	}

	manager, err := factory.CreateManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	if manager == nil {
		t.Fatal("Manager should not be nil")
	}
}

func TestServerHealthStatus(t *testing.T) {
	factory := server.NewFactory()
	manager, err := factory.CreateManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	ctx := context.Background()
	health, err := manager.GetHealthStatus(ctx)
	if err != nil {
		t.Fatalf("Failed to get health status: %v", err)
	}

	if health == nil {
		t.Fatal("Health status should not be nil")
	}

	if health.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}

	if health.Components == nil {
		t.Error("Components should not be nil")
	}

	// Check for expected components
	expectedComponents := []string{"CPU", "Memory", "Disk", "Network"}
	for _, component := range expectedComponents {
		if _, exists := health.Components[component]; !exists {
			t.Errorf("Component %s should exist in health status", component)
		}
	}
}

func TestServerServiceManagement(t *testing.T) {
	factory := server.NewFactory()
	manager, err := factory.CreateManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	ctx := context.Background()

	// List services
	services, err := manager.ListServices(ctx)
	if err != nil {
		t.Fatalf("Failed to list services: %v", err)
	}

	if services == nil {
		t.Fatal("Services list should not be nil")
	}

	// Test with mock services
	if len(services) > 0 {
		service := services[0]

		// Test service control (using mock, so should succeed)
		err = manager.ControlService(ctx, service.Name, types.ServiceActionStart)
		if err != nil {
			t.Fatalf("Failed to start service: %v", err)
		}

		err = manager.ControlService(ctx, service.Name, types.ServiceActionStop)
		if err != nil {
			t.Fatalf("Failed to stop service: %v", err)
		}

		err = manager.ControlService(ctx, service.Name, types.ServiceActionRestart)
		if err != nil {
			t.Fatalf("Failed to restart service: %v", err)
		}
	}
}

func TestServerUserSessions(t *testing.T) {
	factory := server.NewFactory()
	manager, err := factory.CreateManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	ctx := context.Background()
	users, err := manager.GetActiveUsers(ctx)
	if err != nil {
		t.Fatalf("Failed to get active users: %v", err)
	}

	if users == nil {
		t.Fatal("Users list should not be nil")
	}

	// Validate user session structure
	for _, user := range users {
		if user.Username == "" {
			t.Error("Username should not be empty")
		}

		if user.LoginTime.IsZero() {
			t.Error("Login time should not be zero")
		}

		if user.SessionID == "" {
			t.Error("Session ID should not be empty")
		}
	}
}

func TestServerServiceLogs(t *testing.T) {
	factory := server.NewFactory()
	manager, err := factory.CreateManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	ctx := context.Background()

	// Test with a mock service name
	serviceName := "test-service"
	logStream, err := manager.GetServiceLogs(ctx, serviceName, false)
	if err != nil {
		t.Fatalf("Failed to get service logs: %v", err)
	}

	if logStream == nil {
		t.Fatal("Log stream should not be nil")
	}

	if logStream.ServiceName != serviceName {
		t.Errorf("Service name should be %s, got %s", serviceName, logStream.ServiceName)
	}

	if logStream.Entries == nil {
		t.Error("Log entries should not be nil")
	}
}

func TestServerAlertConfiguration(t *testing.T) {
	factory := server.NewFactory()
	manager, err := factory.CreateManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	ctx := context.Background()

	// Test alert configuration
	config := &types.AlertConfig{
		Enabled: true,
		Thresholds: map[string]types.AlertThreshold{
			"CPU": {
				Metric:   "CPU",
				Warning:  70.0,
				Critical: 90.0,
				Unit:     "%",
				Enabled:  true,
			},
		},
		CheckInterval: 30 * time.Second,
	}

	err = manager.ConfigureAlerts(ctx, config)
	if err != nil {
		t.Fatalf("Failed to configure alerts: %v", err)
	}
}

func TestServerConfigurationBackup(t *testing.T) {
	factory := server.NewFactory()
	manager, err := factory.CreateManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	ctx := context.Background()
	backupPath := "/tmp/server_config_backup_test"

	err = manager.BackupConfiguration(ctx, backupPath)
	if err != nil {
		t.Fatalf("Failed to backup configuration: %v", err)
	}
}

func TestServerHealthComponentValidation(t *testing.T) {
	tests := []struct {
		name      string
		component types.ComponentHealth
		wantErr   bool
	}{
		{
			name: "valid component",
			component: types.ComponentHealth{
				Status:      types.HealthLevelHealthy,
				Value:       50.0,
				Threshold:   80.0,
				Unit:        "%",
				Message:     "Component is healthy",
				LastChecked: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "invalid threshold",
			component: types.ComponentHealth{
				Status:    types.HealthLevelHealthy,
				Value:     50.0,
				Threshold: -10.0, // Invalid negative threshold
				Unit:      "%",
			},
			wantErr: true,
		},
		{
			name: "value exceeds percentage",
			component: types.ComponentHealth{
				Status:    types.HealthLevelHealthy,
				Value:     150.0, // Invalid: > 100% for percentage unit
				Threshold: 80.0,
				Unit:      "%",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateComponentHealth(&tt.component)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateComponentHealth() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestServerServiceValidation(t *testing.T) {
	tests := []struct {
		name    string
		service *types.ServiceInfo
		wantErr bool
	}{
		{
			name: "valid service",
			service: &types.ServiceInfo{
				Name:        "nginx",
				DisplayName: "Nginx HTTP Server",
				Status:      types.ServiceStatusRunning,
				StartType:   types.StartTypeAutomatic,
				PID:         1234,
				Memory:      50 * 1024 * 1024,
				CPU:         2.5,
			},
			wantErr: false,
		},
		{
			name: "missing name",
			service: &types.ServiceInfo{
				DisplayName: "Test Service",
				Status:      types.ServiceStatusRunning,
			},
			wantErr: true,
		},
		{
			name: "invalid PID",
			service: &types.ServiceInfo{
				Name:   "test-service",
				Status: types.ServiceStatusRunning,
				PID:    -1, // Invalid negative PID
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateServiceInfo(tt.service)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateServiceInfo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Helper validation functions for testing
func validateComponentHealth(component *types.ComponentHealth) error {
	if component.Threshold < 0 {
		return fmt.Errorf("threshold cannot be negative")
	}

	if component.Unit == "%" && (component.Value < 0 || component.Value > 100) {
		return fmt.Errorf("percentage value must be between 0 and 100")
	}

	return nil
}

func validateServiceInfo(service *types.ServiceInfo) error {
	if service.Name == "" {
		return fmt.Errorf("service name cannot be empty")
	}

	if service.PID < 0 {
		return fmt.Errorf("PID cannot be negative")
	}

	return nil
}

func BenchmarkServerHealthStatus(b *testing.B) {
	factory := server.NewFactory()
	manager, err := factory.CreateManager()
	if err != nil {
		b.Fatalf("Failed to create manager: %v", err)
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.GetHealthStatus(ctx)
		if err != nil {
			b.Fatalf("Failed to get health status: %v", err)
		}
	}
}

func BenchmarkServerListServices(b *testing.B) {
	factory := server.NewFactory()
	manager, err := factory.CreateManager()
	if err != nil {
		b.Fatalf("Failed to create manager: %v", err)
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.ListServices(ctx)
		if err != nil {
			b.Fatalf("Failed to list services: %v", err)
		}
	}
}

// Add missing import
import "fmt"