package performance

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"suppercommand/internal/types"
	"suppercommand/internal/utils"
)

// OptimizationEngine provides automated performance optimization capabilities
type OptimizationEngine struct {
	analyzer *AdvancedAnalyzer
	platform types.Platform
	safeMode bool
	history  []OptimizationResult
}

// OptimizationResult represents the result of an optimization operation
type OptimizationResult struct {
	Timestamp   time.Time                     `json:"timestamp"`
	Suggestion  *types.OptimizationSuggestion `json:"suggestion"`
	Applied     bool                          `json:"applied"`
	Success     bool                          `json:"success"`
	Error       string                        `json:"error,omitempty"`
	BeforeScore float64                       `json:"before_score"`
	AfterScore  float64                       `json:"after_score,omitempty"`
	Commands    []string                      `json:"commands"`
	Output      string                        `json:"output,omitempty"`
}

// NewOptimizationEngine creates a new optimization engine
func NewOptimizationEngine(analyzer *AdvancedAnalyzer) *OptimizationEngine {
	return &OptimizationEngine{
		analyzer: analyzer,
		platform: utils.GetCurrentPlatform(),
		safeMode: true,
		history:  make([]OptimizationResult, 0),
	}
}

// SetSafeMode enables or disables safe mode (only safe optimizations)
func (o *OptimizationEngine) SetSafeMode(enabled bool) {
	o.safeMode = enabled
}

// IsSafeMode returns whether safe mode is enabled
func (o *OptimizationEngine) IsSafeMode() bool {
	return o.safeMode
}

// ApplyOptimizations applies optimization suggestions automatically
func (o *OptimizationEngine) ApplyOptimizations(ctx context.Context, suggestions []*types.OptimizationSuggestion) ([]OptimizationResult, error) {
	results := make([]OptimizationResult, 0)

	for _, suggestion := range suggestions {
		// Skip unsafe optimizations in safe mode
		if o.safeMode && !suggestion.Safe {
			result := OptimizationResult{
				Timestamp:  time.Now(),
				Suggestion: suggestion,
				Applied:    false,
				Success:    false,
				Error:      "Skipped unsafe optimization in safe mode",
			}
			results = append(results, result)
			continue
		}

		result := o.applySingleOptimization(ctx, suggestion)
		results = append(results, result)
		o.history = append(o.history, result)
	}

	// Keep only the last 50 optimization results
	if len(o.history) > 50 {
		o.history = o.history[len(o.history)-50:]
	}

	return results, nil
}

// applySingleOptimization applies a single optimization suggestion
func (o *OptimizationEngine) applySingleOptimization(ctx context.Context, suggestion *types.OptimizationSuggestion) OptimizationResult {
	result := OptimizationResult{
		Timestamp:  time.Now(),
		Suggestion: suggestion,
		Applied:    false,
		Success:    false,
		Commands:   suggestion.Commands,
	}

	// Get baseline performance score
	if metrics, err := o.collectCurrentMetrics(ctx); err == nil {
		result.BeforeScore = o.analyzer.CalculatePerformanceScore(metrics)
	}

	// Apply optimization based on category
	switch suggestion.Category {
	case "CPU":
		result = o.applyCPUOptimization(ctx, suggestion, result)
	case "Memory":
		result = o.applyMemoryOptimization(ctx, suggestion, result)
	case "Disk":
		result = o.applyDiskOptimization(ctx, suggestion, result)
	case "Network":
		result = o.applyNetworkOptimization(ctx, suggestion, result)
	default:
		result.Error = fmt.Sprintf("Unknown optimization category: %s", suggestion.Category)
		return result
	}

	// Get post-optimization performance score
	if result.Success {
		time.Sleep(2 * time.Second) // Wait for changes to take effect
		if metrics, err := o.collectCurrentMetrics(ctx); err == nil {
			result.AfterScore = o.analyzer.CalculatePerformanceScore(metrics)
		}
	}

	return result
}

// applyCPUOptimization applies CPU-related optimizations
func (o *OptimizationEngine) applyCPUOptimization(ctx context.Context, suggestion *types.OptimizationSuggestion, result OptimizationResult) OptimizationResult {
	switch o.platform {
	case types.PlatformWindows:
		return o.applyCPUOptimizationWindows(ctx, suggestion, result)
	case types.PlatformLinux:
		return o.applyCPUOptimizationLinux(ctx, suggestion, result)
	case types.PlatformDarwin:
		return o.applyCPUOptimizationDarwin(ctx, suggestion, result)
	default:
		result.Error = "CPU optimization not supported on this platform"
		return result
	}
}

// applyCPUOptimizationWindows applies Windows-specific CPU optimizations
func (o *OptimizationEngine) applyCPUOptimizationWindows(ctx context.Context, suggestion *types.OptimizationSuggestion, result OptimizationResult) OptimizationResult {
	// Example: Set power plan to high performance
	if strings.Contains(suggestion.Description, "High CPU usage") {
		cmd := exec.CommandContext(ctx, "powercfg", "/setactive", "8c5e7fda-e8bf-4a96-9a85-a6e23a8c635c")
		output, err := cmd.CombinedOutput()

		result.Applied = true
		result.Output = string(output)

		if err != nil {
			result.Error = fmt.Sprintf("Failed to set power plan: %v", err)
			return result
		}

		result.Success = true
		result.Output += "\nSet power plan to High Performance"
	}

	return result
}

// applyCPUOptimizationLinux applies Linux-specific CPU optimizations
func (o *OptimizationEngine) applyCPUOptimizationLinux(ctx context.Context, suggestion *types.OptimizationSuggestion, result OptimizationResult) OptimizationResult {
	// Example: Set CPU governor to performance
	if strings.Contains(suggestion.Description, "High CPU usage") {
		// Check if cpufreq-utils is available
		if _, err := exec.LookPath("cpufreq-set"); err == nil {
			cmd := exec.CommandContext(ctx, "cpufreq-set", "-g", "performance")
			output, err := cmd.CombinedOutput()

			result.Applied = true
			result.Output = string(output)

			if err != nil {
				result.Error = fmt.Sprintf("Failed to set CPU governor: %v", err)
				return result
			}

			result.Success = true
			result.Output += "\nSet CPU governor to performance"
		} else {
			result.Error = "cpufreq-utils not available for CPU optimization"
		}
	}

	return result
}

// applyCPUOptimizationDarwin applies macOS-specific CPU optimizations
func (o *OptimizationEngine) applyCPUOptimizationDarwin(ctx context.Context, suggestion *types.OptimizationSuggestion, result OptimizationResult) OptimizationResult {
	// macOS has limited user-controllable CPU optimizations
	// We can suggest checking Activity Monitor but can't automatically optimize
	result.Applied = true
	result.Success = true
	result.Output = "CPU optimization on macOS requires manual intervention via Activity Monitor"

	return result
}

// applyMemoryOptimization applies memory-related optimizations
func (o *OptimizationEngine) applyMemoryOptimization(ctx context.Context, suggestion *types.OptimizationSuggestion, result OptimizationResult) OptimizationResult {
	switch o.platform {
	case types.PlatformWindows:
		return o.applyMemoryOptimizationWindows(ctx, suggestion, result)
	case types.PlatformLinux:
		return o.applyMemoryOptimizationLinux(ctx, suggestion, result)
	case types.PlatformDarwin:
		return o.applyMemoryOptimizationDarwin(ctx, suggestion, result)
	default:
		result.Error = "Memory optimization not supported on this platform"
		return result
	}
}

// applyMemoryOptimizationWindows applies Windows-specific memory optimizations
func (o *OptimizationEngine) applyMemoryOptimizationWindows(ctx context.Context, suggestion *types.OptimizationSuggestion, result OptimizationResult) OptimizationResult {
	// Example: Clear standby memory (requires elevated privileges)
	if strings.Contains(suggestion.Description, "High memory usage") {
		// Use RAMMap or similar tool if available, or suggest manual cleanup
		result.Applied = true
		result.Success = true
		result.Output = "Memory optimization on Windows requires elevated privileges and specialized tools"
	}

	return result
}

// applyMemoryOptimizationLinux applies Linux-specific memory optimizations
func (o *OptimizationEngine) applyMemoryOptimizationLinux(ctx context.Context, suggestion *types.OptimizationSuggestion, result OptimizationResult) OptimizationResult {
	// Example: Drop caches (only if safe mode is disabled)
	if strings.Contains(suggestion.Description, "High memory usage") && !o.safeMode {
		// Sync first to ensure data integrity
		syncCmd := exec.CommandContext(ctx, "sync")
		if err := syncCmd.Run(); err != nil {
			result.Error = fmt.Sprintf("Failed to sync before dropping caches: %v", err)
			return result
		}

		// Drop page cache, dentries, and inodes
		cmd := exec.CommandContext(ctx, "sh", "-c", "echo 3 > /proc/sys/vm/drop_caches")
		output, err := cmd.CombinedOutput()

		result.Applied = true
		result.Output = string(output)

		if err != nil {
			result.Error = fmt.Sprintf("Failed to drop caches: %v", err)
			return result
		}

		result.Success = true
		result.Output += "\nDropped page cache, dentries, and inodes"
	} else {
		result.Applied = true
		result.Success = true
		result.Output = "Memory optimization requires unsafe operations or elevated privileges"
	}

	return result
}

// applyMemoryOptimizationDarwin applies macOS-specific memory optimizations
func (o *OptimizationEngine) applyMemoryOptimizationDarwin(ctx context.Context, suggestion *types.OptimizationSuggestion, result OptimizationResult) OptimizationResult {
	// macOS manages memory automatically, limited user control
	result.Applied = true
	result.Success = true
	result.Output = "Memory optimization on macOS is handled automatically by the system"

	return result
}

// applyDiskOptimization applies disk-related optimizations
func (o *OptimizationEngine) applyDiskOptimization(ctx context.Context, suggestion *types.OptimizationSuggestion, result OptimizationResult) OptimizationResult {
	// Disk optimizations are generally unsafe as they involve file operations
	if o.safeMode {
		result.Applied = false
		result.Error = "Disk optimizations are disabled in safe mode"
		return result
	}

	switch o.platform {
	case types.PlatformWindows:
		return o.applyDiskOptimizationWindows(ctx, suggestion, result)
	case types.PlatformLinux:
		return o.applyDiskOptimizationLinux(ctx, suggestion, result)
	case types.PlatformDarwin:
		return o.applyDiskOptimizationDarwin(ctx, suggestion, result)
	default:
		result.Error = "Disk optimization not supported on this platform"
		return result
	}
}

// applyDiskOptimizationWindows applies Windows-specific disk optimizations
func (o *OptimizationEngine) applyDiskOptimizationWindows(ctx context.Context, suggestion *types.OptimizationSuggestion, result OptimizationResult) OptimizationResult {
	// Example: Run disk cleanup
	if strings.Contains(suggestion.Description, "Disk space") {
		cmd := exec.CommandContext(ctx, "cleanmgr", "/sagerun:1")
		output, err := cmd.CombinedOutput()

		result.Applied = true
		result.Output = string(output)

		if err != nil {
			result.Error = fmt.Sprintf("Failed to run disk cleanup: %v", err)
			return result
		}

		result.Success = true
		result.Output += "\nRan disk cleanup utility"
	}

	return result
}

// applyDiskOptimizationLinux applies Linux-specific disk optimizations
func (o *OptimizationEngine) applyDiskOptimizationLinux(ctx context.Context, suggestion *types.OptimizationSuggestion, result OptimizationResult) OptimizationResult {
	// Example: Clean temporary files
	if strings.Contains(suggestion.Description, "Disk space") {
		// Clean /tmp directory of old files
		cmd := exec.CommandContext(ctx, "find", "/tmp", "-type", "f", "-atime", "+7", "-delete")
		output, err := cmd.CombinedOutput()

		result.Applied = true
		result.Output = string(output)

		if err != nil {
			result.Error = fmt.Sprintf("Failed to clean temporary files: %v", err)
			return result
		}

		result.Success = true
		result.Output += "\nCleaned old temporary files"
	}

	return result
}

// applyDiskOptimizationDarwin applies macOS-specific disk optimizations
func (o *OptimizationEngine) applyDiskOptimizationDarwin(ctx context.Context, suggestion *types.OptimizationSuggestion, result OptimizationResult) OptimizationResult {
	// macOS has built-in storage optimization
	result.Applied = true
	result.Success = true
	result.Output = "Disk optimization on macOS should be done through System Preferences > Storage"

	return result
}

// applyNetworkOptimization applies network-related optimizations
func (o *OptimizationEngine) applyNetworkOptimization(ctx context.Context, suggestion *types.OptimizationSuggestion, result OptimizationResult) OptimizationResult {
	// Network optimizations are generally safe
	result.Applied = true
	result.Success = true
	result.Output = "Network optimization suggestions provided for manual review"

	return result
}

// collectCurrentMetrics collects current performance metrics
func (o *OptimizationEngine) collectCurrentMetrics(ctx context.Context) (*types.PerformanceMetrics, error) {
	factory := NewFactory()
	analyzer, err := factory.CreateAnalyzer()
	if err != nil {
		return nil, err
	}

	return analyzer.CollectMetrics(ctx, time.Second)
}

// GetOptimizationHistory returns the optimization history
func (o *OptimizationEngine) GetOptimizationHistory() []OptimizationResult {
	return o.history
}

// SaveOptimizationHistory saves the optimization history to a file
func (o *OptimizationEngine) SaveOptimizationHistory(filepath string) error {
	data, err := json.MarshalIndent(o.history, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath, data, 0644)
}

// LoadOptimizationHistory loads the optimization history from a file
func (o *OptimizationEngine) LoadOptimizationHistory(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &o.history)
}

// ClearOptimizationHistory clears the optimization history
func (o *OptimizationEngine) ClearOptimizationHistory() {
	o.history = make([]OptimizationResult, 0)
}

// GetOptimizationStats returns statistics about optimization history
func (o *OptimizationEngine) GetOptimizationStats() map[string]interface{} {
	if len(o.history) == 0 {
		return map[string]interface{}{
			"total_optimizations":      0,
			"successful_optimizations": 0,
			"failed_optimizations":     0,
			"success_rate":             0.0,
		}
	}

	successful := 0
	failed := 0
	totalScoreImprovement := 0.0

	for _, result := range o.history {
		if result.Applied {
			if result.Success {
				successful++
				if result.AfterScore > 0 && result.BeforeScore > 0 {
					totalScoreImprovement += result.AfterScore - result.BeforeScore
				}
			} else {
				failed++
			}
		}
	}

	successRate := 0.0
	if successful+failed > 0 {
		successRate = float64(successful) / float64(successful+failed) * 100
	}

	avgScoreImprovement := 0.0
	if successful > 0 {
		avgScoreImprovement = totalScoreImprovement / float64(successful)
	}

	return map[string]interface{}{
		"total_optimizations":      len(o.history),
		"successful_optimizations": successful,
		"failed_optimizations":     failed,
		"success_rate":             successRate,
		"avg_score_improvement":    avgScoreImprovement,
	}
}
