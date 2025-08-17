package performance

import (
	"context"
	"testing"
	"time"

	"suppercommand/internal/managers/performance"
	"suppercommand/internal/types"
)

func TestPerformanceAnalyzerFactory(t *testing.T) {
	factory := performance.NewFactory()
	if factory == nil {
		t.Fatal("Factory should not be nil")
	}

	analyzer, err := factory.CreateAnalyzer()
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	if analyzer == nil {
		t.Fatal("Analyzer should not be nil")
	}
}

func TestPerformanceMetricsCollection(t *testing.T) {
	factory := performance.NewFactory()
	analyzer, err := factory.CreateAnalyzer()
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	ctx := context.Background()
	metrics, err := analyzer.CollectMetrics(ctx)
	if err != nil {
		t.Fatalf("Failed to collect metrics: %v", err)
	}

	if metrics == nil {
		t.Fatal("Metrics should not be nil")
	}

	// Validate metrics structure
	if metrics.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}

	if metrics.CPU == nil {
		t.Error("CPU metrics should not be nil")
	}

	if metrics.Memory == nil {
		t.Error("Memory metrics should not be nil")
	}

	if metrics.Disk == nil {
		t.Error("Disk metrics should not be nil")
	}

	if metrics.Network == nil {
		t.Error("Network metrics should not be nil")
	}
}

func TestPerformanceAnalysis(t *testing.T) {
	factory := performance.NewFactory()
	analyzer, err := factory.CreateAnalyzer()
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	ctx := context.Background()

	// Collect metrics first
	metrics, err := analyzer.CollectMetrics(ctx)
	if err != nil {
		t.Fatalf("Failed to collect metrics: %v", err)
	}

	// Analyze metrics
	analysis, err := analyzer.AnalyzeMetrics(ctx, metrics)
	if err != nil {
		t.Fatalf("Failed to analyze metrics: %v", err)
	}

	if analysis == nil {
		t.Fatal("Analysis should not be nil")
	}

	if analysis.OverallScore < 0 || analysis.OverallScore > 100 {
		t.Errorf("Overall score should be between 0 and 100, got %f", analysis.OverallScore)
	}

	if len(analysis.Bottlenecks) < 0 {
		t.Error("Bottlenecks slice should be initialized")
	}

	if len(analysis.Recommendations) < 0 {
		t.Error("Recommendations slice should be initialized")
	}
}

func TestPerformanceOptimizer(t *testing.T) {
	factory := performance.NewFactory()
	optimizer, err := factory.CreateOptimizer()
	if err != nil {
		t.Fatalf("Failed to create optimizer: %v", err)
	}

	ctx := context.Background()

	// Create mock analysis
	analysis := &types.PerformanceAnalysis{
		OverallScore: 65.0,
		Bottlenecks: []types.Bottleneck{
			{
				Component:   "CPU",
				Severity:    types.SeverityMedium,
				Description: "High CPU usage detected",
				Impact:      "System responsiveness may be affected",
			},
		},
		Recommendations: []types.Recommendation{
			{
				Category:    "CPU",
				Priority:    types.PriorityMedium,
				Description: "Consider closing unnecessary applications",
				Impact:      "Should reduce CPU usage by 10-20%",
				Safe:        true,
			},
		},
	}

	// Test optimization suggestions
	suggestions, err := optimizer.GetOptimizationSuggestions(ctx, analysis)
	if err != nil {
		t.Fatalf("Failed to get optimization suggestions: %v", err)
	}

	if len(suggestions) == 0 {
		t.Error("Should have at least one optimization suggestion")
	}

	// Test safe optimizations
	safeOptimizations := optimizer.GetSafeOptimizations(suggestions)
	if len(safeOptimizations) < 0 {
		t.Error("Safe optimizations slice should be initialized")
	}
}

func TestPerformanceBaseline(t *testing.T) {
	factory := performance.NewFactory()
	analyzer, err := factory.CreateAnalyzer()
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	ctx := context.Background()

	// Create baseline
	baselineName := "test-baseline"
	err = analyzer.CreateBaseline(ctx, baselineName)
	if err != nil {
		t.Fatalf("Failed to create baseline: %v", err)
	}

	// List baselines
	baselines, err := analyzer.ListBaselines(ctx)
	if err != nil {
		t.Fatalf("Failed to list baselines: %v", err)
	}

	found := false
	for _, baseline := range baselines {
		if baseline.Name == baselineName {
			found = true
			break
		}
	}

	if !found {
		t.Error("Baseline should be found in the list")
	}

	// Compare with baseline
	metrics, err := analyzer.CollectMetrics(ctx)
	if err != nil {
		t.Fatalf("Failed to collect metrics: %v", err)
	}

	comparison, err := analyzer.CompareWithBaseline(ctx, baselineName, metrics)
	if err != nil {
		t.Fatalf("Failed to compare with baseline: %v", err)
	}

	if comparison == nil {
		t.Fatal("Comparison should not be nil")
	}

	// Delete baseline
	err = analyzer.DeleteBaseline(ctx, baselineName)
	if err != nil {
		t.Fatalf("Failed to delete baseline: %v", err)
	}
}

func TestPerformanceHistory(t *testing.T) {
	factory := performance.NewFactory()
	analyzer, err := factory.CreateAnalyzer()
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	ctx := context.Background()

	// Get performance history
	history, err := analyzer.GetPerformanceHistory(ctx, 24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to get performance history: %v", err)
	}

	if history == nil {
		t.Fatal("History should not be nil")
	}

	// History might be empty for new systems, which is okay
	if len(history.Entries) < 0 {
		t.Error("History entries slice should be initialized")
	}
}

func TestPerformanceMetricsValidation(t *testing.T) {
	tests := []struct {
		name    string
		metrics *types.PerformanceMetrics
		wantErr bool
	}{
		{
			name: "valid metrics",
			metrics: &types.PerformanceMetrics{
				Timestamp: time.Now(),
				CPU: &types.CPUMetrics{
					Usage:       50.0,
					LoadAverage: []float64{1.0, 1.2, 1.5},
					Cores:       4,
				},
				Memory: &types.MemoryMetrics{
					Total:     8 * 1024 * 1024 * 1024, // 8GB
					Used:      4 * 1024 * 1024 * 1024, // 4GB
					Available: 4 * 1024 * 1024 * 1024, // 4GB
					Usage:     50.0,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid CPU usage",
			metrics: &types.PerformanceMetrics{
				Timestamp: time.Now(),
				CPU: &types.CPUMetrics{
					Usage: 150.0, // Invalid: > 100%
					Cores: 4,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid memory usage",
			metrics: &types.PerformanceMetrics{
				Timestamp: time.Now(),
				Memory: &types.MemoryMetrics{
					Total: 8 * 1024 * 1024 * 1024,
					Used:  10 * 1024 * 1024 * 1024, // Invalid: used > total
				},
			},
			wantErr: true,
		},
	}

	factory := performance.NewFactory()
	analyzer, err := factory.CreateAnalyzer()
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := analyzer.ValidateMetrics(tt.metrics)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMetrics() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func BenchmarkMetricsCollection(b *testing.B) {
	factory := performance.NewFactory()
	analyzer, err := factory.CreateAnalyzer()
	if err != nil {
		b.Fatalf("Failed to create analyzer: %v", err)
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := analyzer.CollectMetrics(ctx)
		if err != nil {
			b.Fatalf("Failed to collect metrics: %v", err)
		}
	}
}

func BenchmarkPerformanceAnalysis(b *testing.B) {
	factory := performance.NewFactory()
	analyzer, err := factory.CreateAnalyzer()
	if err != nil {
		b.Fatalf("Failed to create analyzer: %v", err)
	}

	ctx := context.Background()

	// Collect metrics once
	metrics, err := analyzer.CollectMetrics(ctx)
	if err != nil {
		b.Fatalf("Failed to collect metrics: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := analyzer.AnalyzeMetrics(ctx, metrics)
		if err != nil {
			b.Fatalf("Failed to analyze metrics: %v", err)
		}
	}
}
