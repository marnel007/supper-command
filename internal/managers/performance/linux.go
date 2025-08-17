package performance

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"suppercommand/internal/types"
)

// LinuxPerformanceAnalyzer manages performance analysis on Linux using /proc and system commands
type LinuxPerformanceAnalyzer struct {
	*BasePerformanceAnalyzer
}

// NewLinuxPerformanceAnalyzer creates a new Linux performance analyzer
func NewLinuxPerformanceAnalyzer() *LinuxPerformanceAnalyzer {
	return &LinuxPerformanceAnalyzer{
		BasePerformanceAnalyzer: NewBasePerformanceAnalyzer(types.PlatformLinux),
	}
}

// CollectMetrics collects performance metrics on Linux
func (l *LinuxPerformanceAnalyzer) CollectMetrics(ctx context.Context, duration time.Duration) (*types.PerformanceMetrics, error) {
	metrics := l.CollectBasicMetrics()
	metrics.Duration = duration

	// Collect CPU metrics
	if err := l.collectCPUMetrics(ctx, metrics); err != nil {
		return nil, types.NewPerformanceError("CPU", "collect", err, "failed to collect CPU metrics")
	}

	// Collect memory metrics
	if err := l.collectMemoryMetrics(ctx, metrics); err != nil {
		return nil, types.NewPerformanceError("Memory", "collect", err, "failed to collect memory metrics")
	}

	// Collect disk metrics
	if err := l.collectDiskMetrics(ctx, metrics); err != nil {
		return nil, types.NewPerformanceError("Disk", "collect", err, "failed to collect disk metrics")
	}

	// Collect network metrics
	if err := l.collectNetworkMetrics(ctx, metrics); err != nil {
		return nil, types.NewPerformanceError("Network", "collect", err, "failed to collect network metrics")
	}

	l.CalculateUsagePercentages(metrics)
	return metrics, nil
}

// collectCPUMetrics collects CPU performance metrics from /proc/stat and /proc/loadavg
func (l *LinuxPerformanceAnalyzer) collectCPUMetrics(ctx context.Context, metrics *types.PerformanceMetrics) error {
	// Get CPU usage from top command (simplified)
	cmd := exec.CommandContext(ctx, "top", "-bn1")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "%Cpu(s):") {
			// Parse line like: "%Cpu(s):  5.9 us,  2.9 sy,  0.0 ni, 91.2 id,  0.0 wa,  0.0 hi,  0.0 si,  0.0 st"
			parts := strings.Split(line, ",")
			if len(parts) >= 4 {
				idlePart := strings.TrimSpace(parts[3])
				if strings.Contains(idlePart, "id") {
					idleStr := strings.Fields(idlePart)[0]
					if idle, err := strconv.ParseFloat(idleStr, 64); err == nil {
						metrics.CPU.Usage = 100.0 - idle
					}
				}
			}
			break
		}
	}

	// Get load average
	cmd = exec.CommandContext(ctx, "cat", "/proc/loadavg")
	if output, err := cmd.Output(); err == nil {
		fields := strings.Fields(string(output))
		if len(fields) >= 3 {
			loadAvg := make([]float64, 3)
			for i := 0; i < 3; i++ {
				if val, err := strconv.ParseFloat(fields[i], 64); err == nil {
					loadAvg[i] = val
				}
			}
			metrics.CPU.LoadAverage = loadAvg
		}
	}

	// Get process count
	cmd = exec.CommandContext(ctx, "ps", "aux")
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		metrics.CPU.Processes = len(lines) - 2 // Subtract header and empty line
	}

	return nil
}

// collectMemoryMetrics collects memory performance metrics from /proc/meminfo
func (l *LinuxPerformanceAnalyzer) collectMemoryMetrics(ctx context.Context, metrics *types.PerformanceMetrics) error {
	cmd := exec.CommandContext(ctx, "cat", "/proc/meminfo")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	lines := strings.Split(string(output), "\n")
	memInfo := make(map[string]uint64)

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			key := strings.TrimSuffix(parts[0], ":")
			if value, err := strconv.ParseUint(parts[1], 10, 64); err == nil {
				// Convert from KB to bytes
				memInfo[key] = value * 1024
			}
		}
	}

	if total, ok := memInfo["MemTotal"]; ok {
		metrics.Memory.Total = total
	}

	if available, ok := memInfo["MemAvailable"]; ok {
		metrics.Memory.Available = available
		metrics.Memory.Used = metrics.Memory.Total - available
	} else if free, ok := memInfo["MemFree"]; ok {
		// Fallback calculation
		buffers := memInfo["Buffers"]
		cached := memInfo["Cached"]
		metrics.Memory.Available = free + buffers + cached
		metrics.Memory.Used = metrics.Memory.Total - metrics.Memory.Available
	}

	if swapTotal, ok := memInfo["SwapTotal"]; ok {
		metrics.Memory.SwapTotal = swapTotal
	}

	if swapFree, ok := memInfo["SwapFree"]; ok {
		metrics.Memory.SwapUsed = metrics.Memory.SwapTotal - swapFree
	}

	if cached, ok := memInfo["Cached"]; ok {
		metrics.Memory.Cached = cached
	}

	if buffers, ok := memInfo["Buffers"]; ok {
		metrics.Memory.Buffers = buffers
	}

	return nil
}

// collectDiskMetrics collects disk performance metrics using df command
func (l *LinuxPerformanceAnalyzer) collectDiskMetrics(ctx context.Context, metrics *types.PerformanceMetrics) error {
	cmd := exec.CommandContext(ctx, "df", "-B1") // Get sizes in bytes
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		if i == 0 || line == "" {
			continue // Skip header and empty lines
		}

		fields := strings.Fields(line)
		if len(fields) >= 6 {
			device := fields[0]
			mountPoint := fields[5]

			// Skip special filesystems
			if strings.HasPrefix(device, "/dev/") {
				total, _ := strconv.ParseUint(fields[1], 10, 64)
				used, _ := strconv.ParseUint(fields[2], 10, 64)
				available, _ := strconv.ParseUint(fields[3], 10, 64)

				diskUsage := types.DiskUsage{
					Device:     device,
					MountPoint: mountPoint,
					Total:      total,
					Used:       used,
					Available:  available,
				}

				metrics.Disk.Usage = append(metrics.Disk.Usage, diskUsage)
			}
		}
	}

	return nil
}

// collectNetworkMetrics collects network performance metrics from /proc/net/dev
func (l *LinuxPerformanceAnalyzer) collectNetworkMetrics(ctx context.Context, metrics *types.PerformanceMetrics) error {
	cmd := exec.CommandContext(ctx, "cat", "/proc/net/dev")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	lines := strings.Split(string(output), "\n")
	var totalBytesReceived, totalBytesSent, totalPacketsReceived, totalPacketsSent uint64

	for i, line := range lines {
		if i < 2 || line == "" {
			continue // Skip header lines and empty lines
		}

		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			continue
		}

		interfaceName := strings.TrimSpace(parts[0])
		if interfaceName == "lo" {
			continue // Skip loopback interface
		}

		fields := strings.Fields(parts[1])
		if len(fields) >= 16 {
			bytesReceived, _ := strconv.ParseUint(fields[0], 10, 64)
			packetsReceived, _ := strconv.ParseUint(fields[1], 10, 64)
			errorsReceived, _ := strconv.ParseUint(fields[2], 10, 64)
			dropsReceived, _ := strconv.ParseUint(fields[3], 10, 64)

			bytesSent, _ := strconv.ParseUint(fields[8], 10, 64)
			packetsSent, _ := strconv.ParseUint(fields[9], 10, 64)

			netInterface := types.NetworkInterface{
				Name:            interfaceName,
				BytesReceived:   bytesReceived,
				BytesSent:       bytesSent,
				PacketsReceived: packetsReceived,
				PacketsSent:     packetsSent,
				Errors:          errorsReceived,
				Drops:           dropsReceived,
			}

			metrics.Network.Interfaces = append(metrics.Network.Interfaces, netInterface)

			totalBytesReceived += bytesReceived
			totalBytesSent += bytesSent
			totalPacketsReceived += packetsReceived
			totalPacketsSent += packetsSent
		}
	}

	metrics.Network.BytesReceived = totalBytesReceived
	metrics.Network.BytesSent = totalBytesSent
	metrics.Network.PacketsReceived = totalPacketsReceived
	metrics.Network.PacketsSent = totalPacketsSent

	// Get connection count
	cmd = exec.CommandContext(ctx, "ss", "-tuln")
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		connectionCount := 0
		for _, line := range lines {
			if strings.Contains(line, "ESTAB") {
				connectionCount++
			}
		}
		metrics.Network.Connections = connectionCount
	}

	return nil
}

// AnalyzePerformance analyzes Linux performance metrics
func (l *LinuxPerformanceAnalyzer) AnalyzePerformance(ctx context.Context, metrics *types.PerformanceMetrics) (*types.AnalysisReport, error) {
	// Use the mock analyzer's analysis logic for now
	mock := NewMockPerformanceAnalyzer()
	return mock.AnalyzePerformance(ctx, metrics)
}

// GenerateReport generates a Linux performance report
func (l *LinuxPerformanceAnalyzer) GenerateReport(ctx context.Context, detailed bool) (*types.PerformanceReport, error) {
	metrics, err := l.CollectMetrics(ctx, time.Second)
	if err != nil {
		return nil, err
	}

	analysis, err := l.AnalyzePerformance(ctx, metrics)
	if err != nil {
		return nil, err
	}

	report := &types.PerformanceReport{
		Timestamp: time.Now(),
		Summary:   l.generateSummary(analysis),
		Metrics:   metrics,
		Analysis:  analysis,
		Detailed:  detailed,
	}

	return report, nil
}

// SaveBaseline saves current metrics as a baseline
func (l *LinuxPerformanceAnalyzer) SaveBaseline(ctx context.Context, filepath string) error {
	// Use mock implementation for now
	mock := NewMockPerformanceAnalyzer()
	return mock.SaveBaseline(ctx, filepath)
}

// CompareBaseline compares current metrics against a saved baseline
func (l *LinuxPerformanceAnalyzer) CompareBaseline(ctx context.Context, baselinePath string) (*types.ComparisonReport, error) {
	// Use mock implementation for now
	mock := NewMockPerformanceAnalyzer()
	return mock.CompareBaseline(ctx, baselinePath)
}

// GetOptimizationSuggestions returns Linux-specific optimization suggestions
func (l *LinuxPerformanceAnalyzer) GetOptimizationSuggestions(ctx context.Context) ([]*types.OptimizationSuggestion, error) {
	metrics, err := l.CollectMetrics(ctx, time.Second)
	if err != nil {
		return nil, err
	}

	suggestions := make([]*types.OptimizationSuggestion, 0)

	// Linux-specific CPU suggestions
	if metrics.CPU.Usage > 80 {
		suggestions = append(suggestions, &types.OptimizationSuggestion{
			Category:    "CPU",
			Priority:    "High",
			Title:       "High CPU Usage on Linux",
			Description: "CPU usage is elevated. Use top or htop to identify resource-intensive processes.",
			Impact:      "Improved system responsiveness",
			Commands:    []string{"top", "htop", "ps aux --sort=-%cpu"},
			Safe:        true,
		})
	}

	// Linux-specific memory suggestions
	if metrics.Memory.Usage > 85 {
		suggestions = append(suggestions, &types.OptimizationSuggestion{
			Category:    "Memory",
			Priority:    "High",
			Title:       "High Memory Usage on Linux",
			Description: "Memory usage is high. Consider freeing up memory or adding more RAM.",
			Impact:      "Reduced memory pressure and improved performance",
			Commands:    []string{"free -h", "ps aux --sort=-%mem", "sync && echo 3 > /proc/sys/vm/drop_caches"},
			Safe:        false, // drop_caches requires caution
		})
	}

	// Linux-specific disk suggestions
	for _, disk := range metrics.Disk.Usage {
		if disk.Usage > 90 {
			suggestions = append(suggestions, &types.OptimizationSuggestion{
				Category:    "Disk",
				Priority:    "Critical",
				Title:       fmt.Sprintf("Disk %s Nearly Full", disk.Device),
				Description: "Disk space is critically low. Clean up unnecessary files or expand storage.",
				Impact:      "Prevent system instability and improve performance",
				Commands:    []string{"du -sh /*", "find /tmp -type f -atime +7 -delete", "apt autoremove"},
				Safe:        false, // File deletion requires caution
			})
		}
	}

	// Load average suggestions
	if len(metrics.CPU.LoadAverage) > 0 && metrics.CPU.LoadAverage[0] > float64(metrics.CPU.Threads) {
		suggestions = append(suggestions, &types.OptimizationSuggestion{
			Category:    "CPU",
			Priority:    "Medium",
			Title:       "High Load Average",
			Description: "System load average is higher than the number of CPU cores, indicating potential bottleneck.",
			Impact:      "Better system responsiveness under load",
			Commands:    []string{"uptime", "iostat", "iotop"},
			Safe:        true,
		})
	}

	return suggestions, nil
}

// generateSummary generates a summary string for the analysis
func (l *LinuxPerformanceAnalyzer) generateSummary(analysis *types.AnalysisReport) string {
	switch analysis.Overall {
	case types.HealthLevelHealthy:
		return "Linux system performance is healthy with no major issues detected."
	case types.HealthLevelWarning:
		return fmt.Sprintf("Linux system performance has some warnings: %s", strings.Join(analysis.Warnings, ", "))
	case types.HealthLevelCritical:
		return fmt.Sprintf("Linux system performance has critical issues: %s", strings.Join(analysis.Bottlenecks, ", "))
	default:
		return "Linux system performance status is unknown."
	}
}
