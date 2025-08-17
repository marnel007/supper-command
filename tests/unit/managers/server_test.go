package managers

import (
	"context"
	"testing"
	"time"

	"suppercommand/internal/managers/server"
	"suppercommand/internal/types"
)

func TestServerManager(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Run("create_manager", func(t *testing.T) {
		factory := server.NewFactory()
		manager, err := factory.CreateManager()
		if err != nil {
			t.Fatalf("Failed to create server manager: %v", err)
		}
		if manager == nil {
			t.Fatal("Manager should not be nil")
		}
	})

	t.Run("get_health_status", func(t *testing.T) {
		factory := server.NewFactory()
		manager, err := factory.CreateManager()
		if err != nil {
			t.Fatalf("Failed to create manager: %v", err)
		}

		health, err := manager.GetHealthStatus(ctx)
		if err != nil {
			t.Errorf("Failed to get health status: %v", err)
		}
		if health == nil {
			t.Error("Health status should not be nil")
		}

		// Validate health structure
		if health.Timestamp.IsZero() {
			t.Error("Health timestamp should be set")
		}
		if health.Status == "" {
			t.Error("Health status should not be empty")
		}
	})

	t.Run("list_services", func(t *testing.T) {
		factory := server.NewFactory()
		manager, err := factory.CreateManager()
		if err != nil {
			t.Fatalf("Failed to create manager: %v", err)
		}

		services, err := manager.ListServices(ctx)
		if err != nil {
			t.Errorf("Failed to list services: %v", err)
		}
		if services == nil {
			t.Error("Services should not be nil")
		}
	})

	t.Run("service_operations", func(t *testing.T) {
		factory := server.NewFactory()
		manager, err := factory.CreateManager()
		if err != nil {
			t.Fatalf("Failed to create manager: %v", err)
		}

		// Test service status (using a common service that should exist)
		status, err := manager.GetServiceStatus(ctx, "Spooler")
		if err != nil {
			t.Logf("Service status check failed (expected on some systems): %v", err)
		} else if status == nil {
			t.Error("Service status should not be nil when no error")
		}

		// Test service control (this will likely fail without admin rights, which is expected)
		err = manager.StartService(ctx, "test-service")
		if err == nil {
			t.Log("Service start succeeded unexpectedly")
		} else {
			t.Logf("Service start failed as expected: %v", err)
		}
	})

	t.Run("user_sessions", func(t *testing.T) {
		factory := server.NewFactory()
		manager, err := factory.CreateManager()
		if err != nil {
			t.Fatalf("Failed to create manager: %v", err)
		}

		sessions, err := manager.ListUserSessions(ctx)
		if err != nil {
			t.Errorf("Failed to list user sessions: %v", err)
		}
		if sessions == nil {
			t.Error("Sessions should not be nil")
		}
	})

	t.Run("backup_operations", func(t *testing.T) {
		factory := server.NewFactory()
		manager, err := factory.CreateManager()
		if err != nil {
			t.Fatalf("Failed to create manager: %v", err)
		}

		// List existing backups
		backups, err := manager.ListBackups(ctx)
		if err != nil {
			t.Errorf("Failed to list backups: %v", err)
		}
		if backups == nil {
			t.Error("Backups should not be nil")
		}

		// Create backup (might fail without proper setup)
		err = manager.CreateBackup(ctx, "test-backup")
		if err != nil {
			t.Logf("Backup creation failed as expected in test environment: %v", err)
		}
	})

	t.Run("alert_management", func(t *testing.T) {
		factory := server.NewFactory()
		manager, err := factory.CreateManager()
		if err != nil {
			t.Fatalf("Failed to create manager: %v", err)
		}

		// List alerts
		alerts, err := manager.ListAlerts(ctx)
		if err != nil {
			t.Errorf("Failed to list alerts: %v", err)
		}
		if alerts == nil {
			t.Error("Alerts should not be nil")
		}

		// Create test alert
		alert := &types.Alert{
			ID:           "test-alert",
			Type:         types.AlertTypeWarning,
			Message:      "Test alert message",
			Source:       "test",
			Timestamp:    time.Now(),
			Severity:     types.AlertSeverityMedium,
			Acknowledged: false,
		}

		err = manager.CreateAlert(ctx, alert)
		if err != nil {
			t.Errorf("Failed to create alert: %v", err)
		}

		// Acknowledge alert
		err = manager.AcknowledgeAlert(ctx, "test-alert")
		if err != nil {
			t.Errorf("Failed to acknowledge alert: %v", err)
		}

		// Clear alert
		err = manager.ClearAlert(ctx, "test-alert")
		if err != nil {
			t.Errorf("Failed to clear alert: %v", err)
		}
	})
}

func TestServerHealthChecks(t *testing.T) {
	ctx := context.Background()
	factory := server.NewFactory()
	manager, err := factory.CreateManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	t.Run("system_health", func(t *testing.T) {
		health, err := manager.CheckSystemHealth(ctx)
		if err != nil {
			t.Errorf("Failed to check system health: %v", err)
		}
		if health == nil {
			t.Error("System health should not be nil")
		}

		// Validate health components
		if health.CPU == nil {
			t.Error("CPU health should not be nil")
		}
		if health.Memory == nil {
			t.Error("Memory health should not be nil")
		}
		if health.Disk == nil {
			t.Error("Disk health should not be nil")
		}
	})

	t.Run("service_health", func(t *testing.T) {
		health, err := manager.CheckServiceHealth(ctx)
		if err != nil {
			t.Errorf("Failed to check service health: %v", err)
		}
		if health == nil {
			t.Error("Service health should not be nil")
		}
	})

	t.Run("network_health", func(t *testing.T) {
		health, err := manager.CheckNetworkHealth(ctx)
		if err != nil {
			t.Errorf("Failed to check network health: %v", err)
		}
		if health == nil {
			t.Error("Network health should not be nil")
		}
	})
}

func TestServerMetrics(t *testing.T) {
	ctx := context.Background()
	factory := server.NewFactory()
	manager, err := factory.CreateManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	t.Run("system_metrics", func(t *testing.T) {
		metrics, err := manager.GetSystemMetrics(ctx)
		if err != nil {
			t.Errorf("Failed to get system metrics: %v", err)
		}
		if metrics == nil {
			t.Error("System metrics should not be nil")
		}

		// Validate metrics
		if metrics.Uptime < 0 {
			t.Error("Uptime should not be negative")
		}
		if metrics.LoadAverage < 0 {
			t.Error("Load average should not be negative")
		}
	})

	t.Run("resource_usage", func(t *testing.T) {
		usage, err := manager.GetResourceUsage(ctx)
		if err != nil {
			t.Errorf("Failed to get resource usage: %v", err)
		}
		if usage == nil {
			t.Error("Resource usage should not be nil")
		}

		// Validate usage percentages
		if usage.CPUUsage < 0 || usage.CPUUsage > 100 {
			t.Errorf("CPU usage should be between 0-100, got %f", usage.CPUUsage)
		}
		if usage.MemoryUsage < 0 || usage.MemoryUsage > 100 {
			t.Errorf("Memory usage should be between 0-100, got %f", usage.MemoryUsage)
		}
	})
}

func TestServerConfiguration(t *testing.T) {
	ctx := context.Background()
	factory := server.NewFactory()
	manager, err := factory.CreateManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	t.Run("get_configuration", func(t *testing.T) {
		config, err := manager.GetConfiguration(ctx)
		if err != nil {
			t.Errorf("Failed to get configuration: %v", err)
		}
		if config == nil {
			t.Error("Configuration should not be nil")
		}
	})

	t.Run("validate_configuration", func(t *testing.T) {
		config := &types.ServerConfig{
			Name:     "test-server",
			Host:     "localhost",
			Port:     8080,
			Username: "testuser",
			Tags:     map[string]string{"env": "test"},
		}

		err := manager.ValidateConfiguration(ctx, config)
		if err != nil {
			t.Errorf("Failed to validate configuration: %v", err)
		}
	})
}
