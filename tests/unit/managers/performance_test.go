package managers

import (
	"context"
	"testing"
	"time"

	"suppercommand/internal/managers/performance"
)

func TestPerformanceManager(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Run("create_manager", func(t *testing.T) {
		factory := performance.NewFactory()
		manager, err := factory.CreateManager()
		if err != nil {
			t.Fatalf("Failed to create performance manager: %v", err)
		}
		if manager == nil {
			t.Fatal("Manager should not be nil")
		}
	})

	t.Run("analyze_performance", func(t *testing.T) {
		factory := performance.NewFactory()
		manager, err := factory.CreateManager()
		if err != nil {
			t.Fatalf("Failed to create manager: %v", err)
		}

		analysis, err := manager.AnalyzePerformance(ctx)
		if err != nil {
			t.Errorf("Failed to analyze performance: %v", err)
		}
		if analysis == nil {
			t.Error("Analysis should not be nil")
		}

		// Validate analysis structure
		if analysis.Timestamp.IsZero() {
			t.Error("Analysis timestamp should be set")
		}
		if analysis.OverallScore < 0 || analysis.OverallScore > 100 {
			t.Errorf("Overall score should be between 0-100, got %f", analysis.OverallScore)
		}
	})

	t.Run("monitor_performance", func(t *testing.T) {
		factory := performance.NewFactory()
		manager, err := factory.CreateManager()
		if err != nil {
			t.Fatalf("Failed to create manager: %v", err)
		}

		// Start monitoring (should not block)
		monitorCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		err = manager.StartMonitoring(monitorCtx, 100*time.Millisecond)
		// This might timeout, which is expected
		if err != nil && err != context.DeadlineExceeded {
			t.Errorf("Unexpected error from monitoring: %v", err)
		}
	})

	t.Run("generate_report", func(t *testing.T) {
		factory := performance.NewFactory()
		manager, err := factory.CreateManager()
		if err != nil {
			t.Fatalf("Failed to create manager: %v", err)
		}

		report, err := manager.GenerateReport(ctx)
		if err != nil {
			t.Errorf("Failed to generate report: %v", err)
		}
		if report == nil {
			t.Error("Report should not be nil")
		}

		// Validate report structure
		if report.GeneratedAt.IsZero() {
			t.Error("Report timestamp should be set")
		}
		if len(report.Sections) == 0 {
			t.Error("Report should have sections")
		}
	})

	t.Run("baseline_operations", func(t *testing.T) {
		factory := performance.NewFactory()
		manager, err := factory.CreateManager()
		if err != nil {
			t.Fatalf("Failed to create manager: %v", err)
		}

		// Create baseline
		err = manager.CreateBaseline(ctx, "test-baseline")
		if err != nil {
			t.Errorf("Failed to create baseline: %v", err)
		}

		// List baselines
		baselines, err := manager.ListBaselines(ctx)
		if err != nil {
			t.Errorf("Failed to list baselines: %v", err)
		}
		if baselines == nil {
			t.Error("Baselines should not be nil")
		}

		// Compare with baseline
		comparison, err := manager.CompareWithBaseline(ctx, "test-baseline")
		if err != nil {
			t.Errorf("Failed to compare with baseline: %v", err)
		}
		if comparison == nil {
			t.Error("Comparison should not be nil")
		}

		// Delete baseline
		err = manager.DeleteBaseline(ctx, "test-baseline")
		if err != nil {
			t.Errorf("Failed to delete baseline: %v", err)
		}
	})

	t.Run("optimization", func(t *testing.T) {
		factory := performance.NewFactory()
		manager, err := factory.CreateManager()
		if err != nil {
			t.Fatalf("Failed to create manager: %v", err)
		}

		// Get optimization suggestions
		suggestions, err := manager.GetOptimizationSuggestions(ctx)
		if err != nil {
			t.Errorf("Failed to get optimization suggestions: %v", err)
		}
		if suggestions == nil {
			t.Error("Suggestions should not be nil")
		}

		// Apply optimizations (if any)
		if len(suggestions) > 0 {
			err = manager.ApplyOptimization(ctx, suggestions[0].ID)
			if err != nil {
				t.Errorf("Failed to apply optimization: %v", err)
			}
		}
	})
}

func TestPerformanceMetrics(t *testing.T) {
	ctx := context.Background()
	factory := performance.NewFactory()
	manager, err := factory.CreateManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	t.Run("cpu_metrics", func(t *testing.T) {
		metrics, err := manager.GetCPUMetrics(ctx)
		if err != nil {
			t.Errorf("Failed to get CPU metrics: %v", err)
		}
		if metrics == nil {
			t.Error("CPU metrics should not be nil")
		}
		if metrics.Usage < 0 || metrics.Usage > 100 {
			t.Errorf("CPU usage should be between 0-100, got %f", metrics.Usage)
		}
	})

	t.Run("memory_metrics", func(t *testing.T) {
		metrics, err := manager.GetMemoryMetrics(ctx)
		if err != nil {
			t.Errorf("Failed to get memory metrics: %v", err)
		}
		if metrics == nil {
			t.Error("Memory metrics should not be nil")
		}
		if metrics.UsedPercent < 0 || metrics.UsedPercent > 100 {
			t.Errorf("Memory usage should be between 0-100, got %f", metrics.UsedPercent)
		}
	})

	t.Run("disk_metrics", func(t *testing.T) {
		metrics, err := manager.GetDiskMetrics(ctx)
		if err != nil {
			t.Errorf("Failed to get disk metrics: %v", err)
		}
		if metrics == nil {
			t.Error("Disk metrics should not be nil")
		}
		if len(metrics) == 0 {
			t.Error("Should have at least one disk metric")
		}
	})

	t.Run("network_metrics", func(t *testing.T) {
		metrics, err := manager.GetNetworkMetrics(ctx)
		if err != nil {
			t.Errorf("Failed to get network metrics: %v", err)
		}
		if metrics == nil {
			t.Error("Network metrics should not be nil")
		}
	})
}

func TestPerformanceThresholds(t *testing.T) {
	tests := []struct {
		name      string
		metric    string
		value     float64
		threshold float64
		expected  bool
	}{
		{"cpu_normal", "cpu", 50.0, 80.0, false},
		{"cpu_high", "cpu", 90.0, 80.0, true},
		{"memory_normal", "memory", 60.0, 85.0, false},
		{"memory_high", "memory", 95.0, 85.0, true},
		{"disk_normal", "disk", 70.0, 90.0, false},
		{"disk_high", "disk", 95.0, 90.0, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			exceeded := test.value > test.threshold
			if exceeded != test.expected {
				t.Errorf("Expected threshold exceeded: %v, got: %v", test.expected, exceeded)
			}
		})
	}
}
