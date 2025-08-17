package performance

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	"suppercommand/internal/types"
	"suppercommand/internal/utils"
)

// Factory creates performance analyzers based on the current platform
type Factory struct{}

// NewFactory creates a new performance analyzer factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateAnalyzer creates a performance analyzer for the current platform
func (f *Factory) CreateAnalyzer() (types.PerformanceAnalyzer, error) {
	platform := utils.GetCurrentPlatform()

	switch platform {
	case types.PlatformWindows:
		return NewWindowsPerformanceAnalyzer(), nil
	case types.PlatformLinux:
		return NewLinuxPerformanceAnalyzer(), nil
	case types.PlatformDarwin:
		return NewDarwinPerformanceAnalyzer(), nil
	default:
		return NewMockPerformanceAnalyzer(), nil
	}
}

// BasePerformanceAnalyzer provides common functionality for all performance analyzers
type BasePerformanceAnalyzer struct {
	platform types.Platform
}

// NewBasePerformanceAnalyzer creates a new base performance analyzer
func NewBasePerformanceAnalyzer(platform types.Platform) *BasePerformanceAnalyzer {
	return &BasePerformanceAnalyzer{
		platform: platform,
	}
}

// GetPlatform returns the platform this analyzer is for
func (b *BasePerformanceAnalyzer) GetPlatform() types.Platform {
	return b.platform
}

// CollectBasicMetrics collects basic Go runtime metrics
func (b *BasePerformanceAnalyzer) CollectBasicMetrics() *types.PerformanceMetrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return &types.PerformanceMetrics{
		Timestamp: time.Now(),
		CPU: types.CPUMetrics{
			Usage:     0, // Will be filled by platform-specific implementation
			CoreUsage: make([]float64, runtime.NumCPU()),
			Processes: 0, // Will be filled by platform-specific implementation
			Threads:   runtime.NumGoroutine(),
		},
		Memory: types.MemoryMetrics{
			Total:     0, // Will be filled by platform-specific implementation
			Used:      m.Alloc,
			Available: 0, // Will be filled by platform-specific implementation
			Usage:     0, // Will be calculated
			SwapTotal: 0, // Will be filled by platform-specific implementation
			SwapUsed:  0, // Will be filled by platform-specific implementation
			SwapUsage: 0, // Will be calculated
			Cached:    0, // Will be filled by platform-specific implementation
			Buffers:   0, // Will be filled by platform-specific implementation
		},
		Network: types.NetworkMetrics{
			Interfaces:      make([]types.NetworkInterface, 0),
			Connections:     0,
			BytesReceived:   0,
			BytesSent:       0,
			PacketsReceived: 0,
			PacketsSent:     0,
		},
		Disk: types.DiskMetrics{
			Usage:      make([]types.DiskUsage, 0),
			ReadSpeed:  0,
			WriteSpeed: 0,
		},
	}
}

// CalculateUsagePercentages calculates usage percentages for metrics
func (b *BasePerformanceAnalyzer) CalculateUsagePercentages(metrics *types.PerformanceMetrics) {
	// Calculate memory usage percentage
	if metrics.Memory.Total > 0 {
		metrics.Memory.Usage = float64(metrics.Memory.Used) / float64(metrics.Memory.Total) * 100
	}

	// Calculate swap usage percentage
	if metrics.Memory.SwapTotal > 0 {
		metrics.Memory.SwapUsage = float64(metrics.Memory.SwapUsed) / float64(metrics.Memory.SwapTotal) * 100
	}

	// Calculate disk usage percentages
	for i := range metrics.Disk.Usage {
		if metrics.Disk.Usage[i].Total > 0 {
			metrics.Disk.Usage[i].Usage = float64(metrics.Disk.Usage[i].Used) / float64(metrics.Disk.Usage[i].Total) * 100
		}
	}
}

// MockPerformanceAnalyzer provides a mock implementation for testing
type MockPerformanceAnalyzer struct {
	*BasePerformanceAnalyzer
	baselines map[string]*types.PerformanceMetrics
}

// NewMockPerformanceAnalyzer creates a new mock performance analyzer
func NewMockPerformanceAnalyzer() *MockPerformanceAnalyzer {
	return &MockPerformanceAnalyzer{
		BasePerformanceAnalyzer: NewBasePerformanceAnalyzer(utils.GetCurrentPlatform()),
		baselines:               make(map[string]*types.PerformanceMetrics),
	}
}

// CollectMetrics returns mock performance metrics
func (m *MockPerformanceAnalyzer) CollectMetrics(ctx context.Context, duration time.Duration) (*types.PerformanceMetrics, error) {
	// Simulate collection time
	time.Sleep(100 * time.Millisecond)

	metrics := m.CollectBasicMetrics()

	// Add mock data
	metrics.CPU.Usage = 25.5
	metrics.CPU.LoadAverage = []float64{1.2, 1.5, 1.8}
	metrics.CPU.Processes = 150

	metrics.Memory.Total = 16 * 1024 * 1024 * 1024    // 16GB
	metrics.Memory.Used = 8 * 1024 * 1024 * 1024      // 8GB
	metrics.Memory.Available = 8 * 1024 * 1024 * 1024 // 8GB
	metrics.Memory.SwapTotal = 4 * 1024 * 1024 * 1024 // 4GB
	metrics.Memory.SwapUsed = 1 * 1024 * 1024 * 1024  // 1GB

	metrics.Disk.Usage = []types.DiskUsage{
		{
			Device:     "/dev/sda1",
			MountPoint: "/",
			Total:      500 * 1024 * 1024 * 1024, // 500GB
			Used:       200 * 1024 * 1024 * 1024, // 200GB
			Available:  300 * 1024 * 1024 * 1024, // 300GB
		},
	}

	metrics.Network.Interfaces = []types.NetworkInterface{
		{
			Name:            "eth0",
			BytesReceived:   1024 * 1024 * 100, // 100MB
			BytesSent:       1024 * 1024 * 50,  // 50MB
			PacketsReceived: 10000,
			PacketsSent:     5000,
			Errors:          0,
			Drops:           0,
		},
	}

	metrics.Duration = duration

	m.CalculateUsagePercentages(metrics)

	return metrics, nil
}

// AnalyzePerformance analyzes performance metrics and returns a report
func (m *MockPerformanceAnalyzer) AnalyzePerformance(ctx context.Context, metrics *types.PerformanceMetrics) (*types.AnalysisReport, error) {
	report := &types.AnalysisReport{
		Timestamp:   time.Now(),
		Overall:     types.HealthLevelHealthy,
		Components:  make(map[string]types.HealthLevel),
		Bottlenecks: make([]string, 0),
		Warnings:    make([]string, 0),
		Suggestions: make([]*types.OptimizationSuggestion, 0),
	}

	// Analyze CPU
	if metrics.CPU.Usage > 80 {
		report.Components["CPU"] = types.HealthLevelCritical
		report.Bottlenecks = append(report.Bottlenecks, "High CPU usage")
		report.Overall = types.HealthLevelCritical
	} else if metrics.CPU.Usage > 60 {
		report.Components["CPU"] = types.HealthLevelWarning
		report.Warnings = append(report.Warnings, "Elevated CPU usage")
		if report.Overall == types.HealthLevelHealthy {
			report.Overall = types.HealthLevelWarning
		}
	} else {
		report.Components["CPU"] = types.HealthLevelHealthy
	}

	// Analyze Memory
	if metrics.Memory.Usage > 90 {
		report.Components["Memory"] = types.HealthLevelCritical
		report.Bottlenecks = append(report.Bottlenecks, "High memory usage")
		report.Overall = types.HealthLevelCritical
	} else if metrics.Memory.Usage > 75 {
		report.Components["Memory"] = types.HealthLevelWarning
		report.Warnings = append(report.Warnings, "Elevated memory usage")
		if report.Overall == types.HealthLevelHealthy {
			report.Overall = types.HealthLevelWarning
		}
	} else {
		report.Components["Memory"] = types.HealthLevelHealthy
	}

	// Analyze Disk
	diskHealth := types.HealthLevelHealthy
	for _, disk := range metrics.Disk.Usage {
		if disk.Usage > 95 {
			diskHealth = types.HealthLevelCritical
			report.Bottlenecks = append(report.Bottlenecks, fmt.Sprintf("Disk %s is nearly full", disk.Device))
			report.Overall = types.HealthLevelCritical
		} else if disk.Usage > 85 {
			if diskHealth == types.HealthLevelHealthy {
				diskHealth = types.HealthLevelWarning
			}
			report.Warnings = append(report.Warnings, fmt.Sprintf("Disk %s is getting full", disk.Device))
			if report.Overall == types.HealthLevelHealthy {
				report.Overall = types.HealthLevelWarning
			}
		}
	}
	report.Components["Disk"] = diskHealth

	// Generate suggestions
	m.generateSuggestions(report, metrics)

	return report, nil
}

// GenerateReport generates a comprehensive performance report
func (m *MockPerformanceAnalyzer) GenerateReport(ctx context.Context, detailed bool) (*types.PerformanceReport, error) {
	metrics, err := m.CollectMetrics(ctx, time.Second)
	if err != nil {
		return nil, err
	}

	analysis, err := m.AnalyzePerformance(ctx, metrics)
	if err != nil {
		return nil, err
	}

	report := &types.PerformanceReport{
		Timestamp: time.Now(),
		Summary:   m.generateSummary(analysis),
		Metrics:   metrics,
		Analysis:  analysis,
		Detailed:  detailed,
	}

	if detailed {
		// Add trend data for detailed reports
		report.Trends = []types.TrendData{
			{
				Timestamp: time.Now().Add(-time.Hour),
				Metric:    "CPU Usage",
				Value:     20.0,
				Unit:      "%",
			},
			{
				Timestamp: time.Now(),
				Metric:    "CPU Usage",
				Value:     metrics.CPU.Usage,
				Unit:      "%",
			},
		}
	}

	return report, nil
}

// SaveBaseline saves current metrics as a baseline
func (m *MockPerformanceAnalyzer) SaveBaseline(ctx context.Context, filepath string) error {
	metrics, err := m.CollectMetrics(ctx, time.Second)
	if err != nil {
		return types.NewPerformanceError("baseline", "save", err, "failed to collect metrics for baseline")
	}

	// In a real implementation, this would save to file
	m.baselines[filepath] = metrics
	return nil
}

// CompareBaseline compares current metrics against a saved baseline
func (m *MockPerformanceAnalyzer) CompareBaseline(ctx context.Context, baselinePath string) (*types.ComparisonReport, error) {
	baseline, exists := m.baselines[baselinePath]
	if !exists {
		return nil, types.NewPerformanceError("baseline", "compare", nil, "baseline not found")
	}

	current, err := m.CollectMetrics(ctx, time.Second)
	if err != nil {
		return nil, types.NewPerformanceError("baseline", "compare", err, "failed to collect current metrics")
	}

	report := &types.ComparisonReport{
		Timestamp:    time.Now(),
		BaselineDate: baseline.Timestamp,
		Current:      current,
		Baseline:     baseline,
		Improvements: make([]string, 0),
		Degradations: make([]string, 0),
	}

	// Compare CPU usage
	cpuDiff := current.CPU.Usage - baseline.CPU.Usage
	if cpuDiff > 10 {
		report.Degradations = append(report.Degradations, fmt.Sprintf("CPU usage increased by %.1f%%", cpuDiff))
	} else if cpuDiff < -10 {
		report.Improvements = append(report.Improvements, fmt.Sprintf("CPU usage decreased by %.1f%%", -cpuDiff))
	}

	// Compare memory usage
	memDiff := current.Memory.Usage - baseline.Memory.Usage
	if memDiff > 10 {
		report.Degradations = append(report.Degradations, fmt.Sprintf("Memory usage increased by %.1f%%", memDiff))
	} else if memDiff < -10 {
		report.Improvements = append(report.Improvements, fmt.Sprintf("Memory usage decreased by %.1f%%", -memDiff))
	}

	// Determine overall change
	if len(report.Improvements) > len(report.Degradations) {
		report.OverallChange = "Performance has improved"
	} else if len(report.Degradations) > len(report.Improvements) {
		report.OverallChange = "Performance has degraded"
	} else {
		report.OverallChange = "Performance is similar to baseline"
	}

	return report, nil
}

// GetOptimizationSuggestions returns optimization suggestions
func (m *MockPerformanceAnalyzer) GetOptimizationSuggestions(ctx context.Context) ([]*types.OptimizationSuggestion, error) {
	metrics, err := m.CollectMetrics(ctx, time.Second)
	if err != nil {
		return nil, err
	}

	suggestions := make([]*types.OptimizationSuggestion, 0)

	// CPU suggestions
	if metrics.CPU.Usage > 70 {
		suggestions = append(suggestions, &types.OptimizationSuggestion{
			Category:    "CPU",
			Priority:    "High",
			Title:       "High CPU Usage Detected",
			Description: "CPU usage is elevated. Consider identifying and optimizing CPU-intensive processes.",
			Impact:      "Improved system responsiveness",
			Commands:    []string{"perf analyze", "top", "htop"},
			Safe:        true,
		})
	}

	// Memory suggestions
	if metrics.Memory.Usage > 80 {
		suggestions = append(suggestions, &types.OptimizationSuggestion{
			Category:    "Memory",
			Priority:    "High",
			Title:       "High Memory Usage Detected",
			Description: "Memory usage is high. Consider closing unnecessary applications or adding more RAM.",
			Impact:      "Reduced memory pressure and improved performance",
			Commands:    []string{"free -h", "ps aux --sort=-%mem"},
			Safe:        true,
		})
	}

	return suggestions, nil
}

// generateSuggestions generates optimization suggestions based on analysis
func (m *MockPerformanceAnalyzer) generateSuggestions(report *types.AnalysisReport, metrics *types.PerformanceMetrics) {
	if metrics.CPU.Usage > 80 {
		report.Suggestions = append(report.Suggestions, &types.OptimizationSuggestion{
			Category:    "CPU",
			Priority:    "High",
			Title:       "Reduce CPU Load",
			Description: "High CPU usage detected. Consider stopping unnecessary processes.",
			Impact:      "Improved system responsiveness",
			Commands:    []string{"top", "htop", "ps aux"},
			Safe:        true,
		})
	}

	if metrics.Memory.Usage > 85 {
		report.Suggestions = append(report.Suggestions, &types.OptimizationSuggestion{
			Category:    "Memory",
			Priority:    "High",
			Title:       "Free Up Memory",
			Description: "High memory usage detected. Consider closing unused applications.",
			Impact:      "Reduced memory pressure",
			Commands:    []string{"free -h", "ps aux --sort=-%mem"},
			Safe:        true,
		})
	}
}

// generateSummary generates a summary string for the analysis
func (m *MockPerformanceAnalyzer) generateSummary(analysis *types.AnalysisReport) string {
	switch analysis.Overall {
	case types.HealthLevelHealthy:
		return "System performance is healthy with no major issues detected."
	case types.HealthLevelWarning:
		return fmt.Sprintf("System performance has some warnings: %s", strings.Join(analysis.Warnings, ", "))
	case types.HealthLevelCritical:
		return fmt.Sprintf("System performance has critical issues: %s", strings.Join(analysis.Bottlenecks, ", "))
	default:
		return "System performance status is unknown."
	}
}
