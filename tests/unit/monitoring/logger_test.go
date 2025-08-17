package monitoring_test

import (
	"os"
	"testing"

	"suppercommand/internal/config"
	"suppercommand/internal/monitoring"
)

func TestLogger_Creation(t *testing.T) {
	logger := monitoring.NewLogger(config.MonitoringConfig{
		LogLevel: "info",
	})

	if logger == nil {
		t.Fatal("NewLogger() returned nil")
	}
}

func TestLogger_LogLevels(t *testing.T) {
	// Create logger
	logger := monitoring.NewLogger(config.MonitoringConfig{
		LogLevel: "debug",
	})

	// Test that logger methods can be called without panicking
	tests := []struct {
		name    string
		logFunc func()
	}{
		{
			name: "debug message",
			logFunc: func() {
				logger.Debug("debug message", monitoring.Field{Key: "test", Value: "value"})
			},
		},
		{
			name: "info message",
			logFunc: func() {
				logger.Info("info message", monitoring.Field{Key: "test", Value: "value"})
			},
		},
		{
			name: "warn message",
			logFunc: func() {
				logger.Warn("warn message", monitoring.Field{Key: "test", Value: "value"})
			},
		},
		{
			name: "error message",
			logFunc: func() {
				logger.Error("error message", os.ErrNotExist, monitoring.Field{Key: "test", Value: "value"})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just test that the function doesn't panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Log function panicked: %v", r)
				}
			}()

			tt.logFunc()
		})
	}
}

func TestLogger_Fields(t *testing.T) {
	logger := monitoring.NewLogger(config.MonitoringConfig{
		LogLevel: "info",
	})

	// Test that logging with fields doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Log function with fields panicked: %v", r)
		}
	}()

	// Log with fields
	logger.Info("test message",
		monitoring.Field{Key: "key1", Value: "value1"},
		monitoring.Field{Key: "key2", Value: 42},
	)
}
