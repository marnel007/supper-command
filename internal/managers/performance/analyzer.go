package performance

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"time"

	"suppercommand/internal/types"
)

// AdvancedAnalyzer provides sophisticated performance analysis capabilities
type AdvancedAnalyzer struct {
	thresholds *PerformanceThresholds
	history    []types.PerformanceMetrics
}

// PerformanceThresholds defines thresholds for performance analysis
type PerformanceThresholds struct {
	CPU struct {
		Warning  float64 `json:"warning"`
		Critical float64 `json:"critical"`
	} `json:"cpu"`
	Memory struct {
		Warning  float64 `json:"warning"`
		Critical float64 `json:"critical"`
	} `json:"memory"`
	Disk struct {
		Warning  float64 `json:"warning"`
		Critical float64 `json:"critical"`
	} `json:"disk"`
	LoadAverage struct {
		Warning  float64 `json:"warning"`
		Critical float64 `json:"critical"`
	} `json:"load_average"`
}

// NewAdvancedAnalyzer creates a new advanced performance analyzer
func NewAdvancedAnalyzer() *AdvancedAnalyzer {
	return &AdvancedAnalyzer{
		thresholds: getDefaultThresholds(),
		history:    make([]types.PerformanceMetrics, 0),
	}
}

// getDefaultThresholds returns default performance thresholds
func getDefaultThresholds() *PerformanceThresholds {
	return &PerformanceThresholds{
		CPU: struct {
			Warning  float64 `json:"warning"`
			Critical float64 `json:"critical"`
		}{
			Warning:  70.0,
			Critical: 90.0,
		},
		Memory: struct {
			Warning  float64 `json:"warning"`
			Critical float64 `json:"critical"`
		}{
			Warning:  80.0,
			Critical: 95.0,
		},
		Disk: struct {
			Warning  float64 `json:"warning"`
			Critical float64 `json:"critical"`
		}{
			Warning:  85.0,
			Critical: 95.0,
		},
		LoadAverage: struct {
			Warning  float64 `json:"warning"`
			Critical float64 `json:"critical"`
		}{
			Warning:  2.0,
			Critical: 5.0,
		},
	}
}

// AnalyzeMetrics performs comprehensive analysis of performance metrics
func (a *AdvancedAnalyzer) AnalyzeMetrics(ctx context.Context, metrics *types.PerformanceMetrics) (*types.AnalysisReport, error) {
	report := &types.AnalysisReport{
		Timestamp:   time.Now(),
		Overall:     types.HealthLevelHealthy,
		Components:  make(map[string]types.HealthLevel),
		Bottlenecks: make([]string, 0),
		Warnings:    make([]string, 0),
		Suggestions: make([]*types.OptimizationSuggestion, 0),
	}

	// Analyze each component
	a.analyzeCPU(metrics, report)
	a.analyzeMemory(metrics, report)
	a.analyzeDisk(metrics, report)
	a.analyzeNetwork(metrics, report)

	// Determine overall health
	a.determineOverallHealth(report)

	// Generate optimization suggestions
	a.generateOptimizationSuggestions(metrics, report)

	// Add to history for trend analysis
	a.addToHistory(*metrics)

	return report, nil
}

// analyzeCPU analyzes CPU performance metrics
func (a *AdvancedAnalyzer) analyzeCPU(metrics *types.PerformanceMetrics, report *types.AnalysisReport) {
	cpuUsage := metrics.CPU.Usage

	if cpuUsage >= a.thresholds.CPU.Critical {
		report.Components["CPU"] = types.HealthLevelCritical
		report.Bottlenecks = append(report.Bottlenecks, fmt.Sprintf("Critical CPU usage: %.1f%%", cpuUsage))
	} else if cpuUsage >= a.thresholds.CPU.Warning {
		report.Components["CPU"] = types.HealthLevelWarning
		report.Warnings = append(report.Warnings, fmt.Sprintf("High CPU usage: %.1f%%", cpuUsage))
	} else {
		report.Components["CPU"] = types.HealthLevelHealthy
	}

	// Analyze load average if available
	if len(metrics.CPU.LoadAverage) > 0 {
		loadAvg := metrics.CPU.LoadAverage[0] // 1-minute load average
		if loadAvg >= a.thresholds.LoadAverage.Critical {
			report.Components["LoadAverage"] = types.HealthLevelCritical
			report.Bottlenecks = append(report.Bottlenecks, fmt.Sprintf("Critical load average: %.2f", loadAvg))
		} else if loadAvg >= a.thresholds.LoadAverage.Warning {
			report.Components["LoadAverage"] = types.HealthLevelWarning
			report.Warnings = append(report.Warnings, fmt.Sprintf("High load average: %.2f", loadAvg))
		} else {
			report.Components["LoadAverage"] = types.HealthLevelHealthy
		}
	}

	// Analyze process count
	if metrics.CPU.Processes > 500 {
		report.Warnings = append(report.Warnings, fmt.Sprintf("High process count: %d", metrics.CPU.Processes))
	}
}

// analyzeMemory analyzes memory performance metrics
func (a *AdvancedAnalyzer) analyzeMemory(metrics *types.PerformanceMetrics, report *types.AnalysisReport) {
	memoryUsage := metrics.Memory.Usage

	if memoryUsage >= a.thresholds.Memory.Critical {
		report.Components["Memory"] = types.HealthLevelCritical
		report.Bottlenecks = append(report.Bottlenecks, fmt.Sprintf("Critical memory usage: %.1f%%", memoryUsage))
	} else if memoryUsage >= a.thresholds.Memory.Warning {
		report.Components["Memory"] = types.HealthLevelWarning
		report.Warnings = append(report.Warnings, fmt.Sprintf("High memory usage: %.1f%%", memoryUsage))
	} else {
		report.Components["Memory"] = types.HealthLevelHealthy
	}

	// Analyze swap usage
	if metrics.Memory.SwapUsage > 50 {
		if metrics.Memory.SwapUsage > 80 {
			report.Components["Swap"] = types.HealthLevelCritical
			report.Bottlenecks = append(report.Bottlenecks, fmt.Sprintf("Critical swap usage: %.1f%%", metrics.Memory.SwapUsage))
		} else {
			report.Components["Swap"] = types.HealthLevelWarning
			report.Warnings = append(report.Warnings, fmt.Sprintf("High swap usage: %.1f%%", metrics.Memory.SwapUsage))
		}
	} else {
		report.Components["Swap"] = types.HealthLevelHealthy
	}
}

// analyzeDisk analyzes disk performance metrics
func (a *AdvancedAnalyzer) analyzeDisk(metrics *types.PerformanceMetrics, report *types.AnalysisReport) {
	overallDiskHealth := types.HealthLevelHealthy

	for _, disk := range metrics.Disk.Usage {
		diskHealth := types.HealthLevelHealthy

		if disk.Usage >= a.thresholds.Disk.Critical {
			diskHealth = types.HealthLevelCritical
			report.Bottlenecks = append(report.Bottlenecks,
				fmt.Sprintf("Critical disk usage on %s: %.1f%%", disk.Device, disk.Usage))
		} else if disk.Usage >= a.thresholds.Disk.Warning {
			diskHealth = types.HealthLevelWarning
			report.Warnings = append(report.Warnings,
				fmt.Sprintf("High disk usage on %s: %.1f%%", disk.Device, disk.Usage))
		}

		// Update overall disk health
		if diskHealth == types.HealthLevelCritical {
			overallDiskHealth = types.HealthLevelCritical
		} else if diskHealth == types.HealthLevelWarning && overallDiskHealth == types.HealthLevelHealthy {
			overallDiskHealth = types.HealthLevelWarning
		}
	}

	report.Components["Disk"] = overallDiskHealth
}

// analyzeNetwork analyzes network performance metrics
func (a *AdvancedAnalyzer) analyzeNetwork(metrics *types.PerformanceMetrics, report *types.AnalysisReport) {
	networkHealth := types.HealthLevelHealthy

	// Analyze network interfaces for errors and drops
	for _, iface := range metrics.Network.Interfaces {
		if iface.Errors > 100 {
			networkHealth = types.HealthLevelWarning
			report.Warnings = append(report.Warnings,
				fmt.Sprintf("Network errors detected on %s: %d", iface.Name, iface.Errors))
		}

		if iface.Drops > 50 {
			networkHealth = types.HealthLevelWarning
			report.Warnings = append(report.Warnings,
				fmt.Sprintf("Network packet drops on %s: %d", iface.Name, iface.Drops))
		}
	}

	// Analyze connection count
	if metrics.Network.Connections > 1000 {
		networkHealth = types.HealthLevelWarning
		report.Warnings = append(report.Warnings,
			fmt.Sprintf("High network connection count: %d", metrics.Network.Connections))
	}

	report.Components["Network"] = networkHealth
}

// determineOverallHealth determines the overall system health
func (a *AdvancedAnalyzer) determineOverallHealth(report *types.AnalysisReport) {
	criticalCount := 0
	warningCount := 0

	for _, health := range report.Components {
		switch health {
		case types.HealthLevelCritical:
			criticalCount++
		case types.HealthLevelWarning:
			warningCount++
		}
	}

	if criticalCount > 0 {
		report.Overall = types.HealthLevelCritical
	} else if warningCount > 0 {
		report.Overall = types.HealthLevelWarning
	} else {
		report.Overall = types.HealthLevelHealthy
	}
}

// generateOptimizationSuggestions generates optimization suggestions based on analysis
func (a *AdvancedAnalyzer) generateOptimizationSuggestions(metrics *types.PerformanceMetrics, report *types.AnalysisReport) {
	// CPU optimization suggestions
	if metrics.CPU.Usage > a.thresholds.CPU.Warning {
		priority := "Medium"
		if metrics.CPU.Usage > a.thresholds.CPU.Critical {
			priority = "High"
		}

		report.Suggestions = append(report.Suggestions, &types.OptimizationSuggestion{
			Category:    "CPU",
			Priority:    priority,
			Title:       "Optimize CPU Usage",
			Description: "High CPU usage detected. Consider identifying and optimizing CPU-intensive processes.",
			Impact:      "Improved system responsiveness and reduced power consumption",
			Commands:    []string{"top", "htop", "ps aux --sort=-%cpu"},
			Safe:        true,
		})
	}

	// Memory optimization suggestions
	if metrics.Memory.Usage > a.thresholds.Memory.Warning {
		priority := "Medium"
		if metrics.Memory.Usage > a.thresholds.Memory.Critical {
			priority = "High"
		}

		report.Suggestions = append(report.Suggestions, &types.OptimizationSuggestion{
			Category:    "Memory",
			Priority:    priority,
			Title:       "Optimize Memory Usage",
			Description: "High memory usage detected. Consider closing unnecessary applications or adding more RAM.",
			Impact:      "Reduced memory pressure and improved system stability",
			Commands:    []string{"free -h", "ps aux --sort=-%mem"},
			Safe:        true,
		})
	}

	// Swap optimization suggestions
	if metrics.Memory.SwapUsage > 50 {
		report.Suggestions = append(report.Suggestions, &types.OptimizationSuggestion{
			Category:    "Memory",
			Priority:    "Medium",
			Title:       "Reduce Swap Usage",
			Description: "High swap usage can slow down the system. Consider adding more RAM or optimizing memory usage.",
			Impact:      "Improved system performance and responsiveness",
			Commands:    []string{"swapon -s", "vmstat 1 5"},
			Safe:        true,
		})
	}

	// Disk optimization suggestions
	for _, disk := range metrics.Disk.Usage {
		if disk.Usage > a.thresholds.Disk.Warning {
			priority := "Medium"
			if disk.Usage > a.thresholds.Disk.Critical {
				priority = "High"
			}

			report.Suggestions = append(report.Suggestions, &types.OptimizationSuggestion{
				Category:    "Disk",
				Priority:    priority,
				Title:       fmt.Sprintf("Free Up Disk Space on %s", disk.Device),
				Description: "Disk space is running low. Clean up unnecessary files or expand storage.",
				Impact:      "Prevent system instability and improve performance",
				Commands:    []string{"du -sh /*", "find /tmp -type f -atime +7"},
				Safe:        false, // File operations require caution
			})
		}
	}

	// Load average suggestions
	if len(metrics.CPU.LoadAverage) > 0 && metrics.CPU.LoadAverage[0] > a.thresholds.LoadAverage.Warning {
		report.Suggestions = append(report.Suggestions, &types.OptimizationSuggestion{
			Category:    "CPU",
			Priority:    "Medium",
			Title:       "Reduce System Load",
			Description: "System load is higher than recommended. Consider reducing concurrent processes or upgrading hardware.",
			Impact:      "Better system responsiveness under load",
			Commands:    []string{"uptime", "iostat", "iotop"},
			Safe:        true,
		})
	}
}

// addToHistory adds metrics to the historical data
func (a *AdvancedAnalyzer) addToHistory(metrics types.PerformanceMetrics) {
	a.history = append(a.history, metrics)

	// Keep only the last 100 entries
	if len(a.history) > 100 {
		a.history = a.history[1:]
	}
}

// GenerateTrendAnalysis generates trend analysis from historical data
func (a *AdvancedAnalyzer) GenerateTrendAnalysis() []types.TrendData {
	if len(a.history) < 2 {
		return []types.TrendData{}
	}

	trends := make([]types.TrendData, 0)

	// Generate CPU usage trend
	for _, metrics := range a.history {
		trends = append(trends, types.TrendData{
			Timestamp: metrics.Timestamp,
			Metric:    "CPU Usage",
			Value:     metrics.CPU.Usage,
			Unit:      "%",
		})
	}

	// Generate memory usage trend
	for _, metrics := range a.history {
		trends = append(trends, types.TrendData{
			Timestamp: metrics.Timestamp,
			Metric:    "Memory Usage",
			Value:     metrics.Memory.Usage,
			Unit:      "%",
		})
	}

	// Sort by timestamp
	sort.Slice(trends, func(i, j int) bool {
		return trends[i].Timestamp.Before(trends[j].Timestamp)
	})

	return trends
}

// CalculatePerformanceScore calculates an overall performance score (0-100)
func (a *AdvancedAnalyzer) CalculatePerformanceScore(metrics *types.PerformanceMetrics) float64 {
	scores := make([]float64, 0)

	// CPU score (inverse of usage percentage)
	cpuScore := math.Max(0, 100-metrics.CPU.Usage)
	scores = append(scores, cpuScore)

	// Memory score (inverse of usage percentage)
	memoryScore := math.Max(0, 100-metrics.Memory.Usage)
	scores = append(scores, memoryScore)

	// Disk score (average of all disks, inverse of usage)
	if len(metrics.Disk.Usage) > 0 {
		diskTotal := 0.0
		for _, disk := range metrics.Disk.Usage {
			diskTotal += math.Max(0, 100-disk.Usage)
		}
		diskScore := diskTotal / float64(len(metrics.Disk.Usage))
		scores = append(scores, diskScore)
	}

	// Calculate weighted average
	if len(scores) == 0 {
		return 0
	}

	total := 0.0
	for _, score := range scores {
		total += score
	}

	return total / float64(len(scores))
}

// SaveThresholds saves custom thresholds to a file
func (a *AdvancedAnalyzer) SaveThresholds(filepath string) error {
	data, err := json.MarshalIndent(a.thresholds, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath, data, 0644)
}

// LoadThresholds loads custom thresholds from a file
func (a *AdvancedAnalyzer) LoadThresholds(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, a.thresholds)
}

// GetThresholds returns the current thresholds
func (a *AdvancedAnalyzer) GetThresholds() *PerformanceThresholds {
	return a.thresholds
}

// SetThresholds sets new thresholds
func (a *AdvancedAnalyzer) SetThresholds(thresholds *PerformanceThresholds) {
	a.thresholds = thresholds
}

// GetHistoryLength returns the number of historical entries
func (a *AdvancedAnalyzer) GetHistoryLength() int {
	return len(a.history)
}

// ClearHistory clears the historical data
func (a *AdvancedAnalyzer) ClearHistory() {
	a.history = make([]types.PerformanceMetrics, 0)
}
