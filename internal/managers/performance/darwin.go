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

// DarwinPerformanceAnalyzer manages performance analysis on macOS using system commands
type DarwinPerformanceAnalyzer struct {
	*BasePerformanceAnalyzer
}

// NewDarwinPerformanceAnalyzer creates a new Darwin performance analyzer
func NewDarwinPerformanceAnalyzer() *DarwinPerformanceAnalyzer {
	return &DarwinPerformanceAnalyzer{
		BasePerformanceAnalyzer: NewBasePerformanceAnalyzer(types.PlatformDarwin),
	}
}

// CollectMetrics collects performance metrics on macOS
func (d *DarwinPerformanceAnalyzer) CollectMetrics(ctx context.Context, duration time.Duration) (*types.PerformanceMetrics, error) {
	metrics := d.CollectBasicMetrics()
	metrics.Duration = duration

	// Collect CPU metrics
	if err := d.collectCPUMetrics(ctx, metrics); err != nil {
		return nil, types.NewPerformanceError("CPU", "collect", err, "failed to collect CPU metrics")
	}

	// Collect memory metrics
	if err := d.collectMemoryMetrics(ctx, metrics); err != nil {
		return nil, types.NewPerformanceError("Memory", "collect", err, "failed to collect memory metrics")
	}

	// Collect disk metrics
	if err := d.collectDiskMetrics(ctx, metrics); err != nil {
		return nil, types.NewPerformanceError("Disk", "collect", err, "failed to collect disk metrics")
	}

	// Collect network metrics
	if err := d.collectNetworkMetrics(ctx, metrics); err != nil {
		return nil, types.NewPerformanceError("Network", "collect", err, "failed to collect network metrics")
	}

	d.CalculateUsagePercentages(metrics)
	return metrics, nil
}

// collectCPUMetrics collects CPU performance metrics using top and sysctl
func (d *DarwinPerformanceAnalyzer) collectCPUMetrics(ctx context.Context, metrics *types.PerformanceMetrics) error {
	// Get CPU usage from top command
	cmd := exec.CommandContext(ctx, "top", "-l", "1", "-n", "0")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "CPU usage:") {
			// Parse line like: "CPU usage: 5.26% user, 10.52% sys, 84.21% idle"
			parts := strings.Split(line, ",")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if strings.Contains(part, "idle") {
					fields := strings.Fields(part)
					if len(fields) >= 1 {
						idleStr := strings.TrimSuffix(fields[0], "%")
						if idle, err := strconv.ParseFloat(idleStr, 64); err == nil {
							metrics.CPU.Usage = 100.0 - idle
						}
					}
				}
			}
			break
		}
	}

	// Get load average using uptime
	cmd = exec.CommandContext(ctx, "uptime")
	if output, err := cmd.Output(); err == nil {
		outputStr := string(output)
		if strings.Contains(outputStr, "load averages:") {
			parts := strings.Split(outputStr, "load averages:")
			if len(parts) >= 2 {
				loadParts := strings.Fields(parts[1])
				if len(loadParts) >= 3 {
					loadAvg := make([]float64, 3)
					for i := 0; i < 3; i++ {
						loadStr := strings.TrimSpace(loadParts[i])
						if val, err := strconv.ParseFloat(loadStr, 64); err == nil {
							loadAvg[i] = val
						}
					}
					metrics.CPU.LoadAverage = loadAvg
				}
			}
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

// collectMemoryMetrics collects memory performance metrics using vm_stat
func (d *DarwinPerformanceAnalyzer) collectMemoryMetrics(ctx context.Context, metrics *types.PerformanceMetrics) error {
	// Get memory info using vm_stat
	cmd := exec.CommandContext(ctx, "vm_stat")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	lines := strings.Split(string(output), "\n")
	var pageSize uint64 = 4096 // Default page size
	var freePages, activePages, inactivePages, wiredPages, compressedPages uint64

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "page size of") {
			// Extract page size
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "of" && i+1 < len(parts) {
					if size, err := strconv.ParseUint(parts[i+1], 10, 64); err == nil {
						pageSize = size
					}
					break
				}
			}
		} else if strings.HasPrefix(line, "Pages free:") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				if pages, err := strconv.ParseUint(strings.TrimSuffix(parts[2], "."), 10, 64); err == nil {
					freePages = pages
				}
			}
		} else if strings.HasPrefix(line, "Pages active:") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				if pages, err := strconv.ParseUint(strings.TrimSuffix(parts[2], "."), 10, 64); err == nil {
					activePages = pages
				}
			}
		} else if strings.HasPrefix(line, "Pages inactive:") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				if pages, err := strconv.ParseUint(strings.TrimSuffix(parts[2], "."), 10, 64); err == nil {
					inactivePages = pages
				}
			}
		} else if strings.HasPrefix(line, "Pages wired down:") {
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				if pages, err := strconv.ParseUint(strings.TrimSuffix(parts[3], "."), 10, 64); err == nil {
					wiredPages = pages
				}
			}
		} else if strings.HasPrefix(line, "Pages occupied by compressor:") {
			parts := strings.Fields(line)
			if len(parts) >= 5 {
				if pages, err := strconv.ParseUint(strings.TrimSuffix(parts[4], "."), 10, 64); err == nil {
					compressedPages = pages
				}
			}
		}
	}

	// Calculate memory metrics
	totalPages := freePages + activePages + inactivePages + wiredPages + compressedPages
	metrics.Memory.Total = totalPages * pageSize
	metrics.Memory.Available = freePages * pageSize
	metrics.Memory.Used = (activePages + inactivePages + wiredPages) * pageSize

	// Get swap info using sysctl
	cmd = exec.CommandContext(ctx, "sysctl", "vm.swapusage")
	if output, err := cmd.Output(); err == nil {
		outputStr := string(output)
		// Parse output like: "vm.swapusage: total = 2048.00M  used = 617.25M  free = 1430.75M  (encrypted)"
		if strings.Contains(outputStr, "total =") {
			parts := strings.Fields(outputStr)
			for i, part := range parts {
				if part == "total" && i+2 < len(parts) {
					if totalStr := parts[i+2]; strings.HasSuffix(totalStr, "M") {
						if total, err := strconv.ParseFloat(strings.TrimSuffix(totalStr, "M"), 64); err == nil {
							metrics.Memory.SwapTotal = uint64(total * 1024 * 1024)
						}
					}
				} else if part == "used" && i+2 < len(parts) {
					if usedStr := parts[i+2]; strings.HasSuffix(usedStr, "M") {
						if used, err := strconv.ParseFloat(strings.TrimSuffix(usedStr, "M"), 64); err == nil {
							metrics.Memory.SwapUsed = uint64(used * 1024 * 1024)
						}
					}
				}
			}
		}
	}

	return nil
}

// collectDiskMetrics collects disk performance metrics using df command
func (d *DarwinPerformanceAnalyzer) collectDiskMetrics(ctx context.Context, metrics *types.PerformanceMetrics) error {
	cmd := exec.CommandContext(ctx, "df", "-b") // Get sizes in 512-byte blocks
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

				// Convert from 512-byte blocks to bytes
				diskUsage := types.DiskUsage{
					Device:     device,
					MountPoint: mountPoint,
					Total:      total * 512,
					Used:       used * 512,
					Available:  available * 512,
				}

				metrics.Disk.Usage = append(metrics.Disk.Usage, diskUsage)
			}
		}
	}

	return nil
}

// collectNetworkMetrics collects network performance metrics using netstat
func (d *DarwinPerformanceAnalyzer) collectNetworkMetrics(ctx context.Context, metrics *types.PerformanceMetrics) error {
	// Get network interface statistics
	cmd := exec.CommandContext(ctx, "netstat", "-ibn")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	lines := strings.Split(string(output), "\n")
	var totalBytesReceived, totalBytesSent, totalPacketsReceived, totalPacketsSent uint64

	for i, line := range lines {
		if i == 0 || line == "" {
			continue // Skip header and empty lines
		}

		fields := strings.Fields(line)
		if len(fields) >= 10 {
			interfaceName := fields[0]
			if interfaceName == "lo0" {
				continue // Skip loopback interface
			}

			packetsReceived, _ := strconv.ParseUint(fields[4], 10, 64)
			bytesReceived, _ := strconv.ParseUint(fields[6], 10, 64)
			packetsSent, _ := strconv.ParseUint(fields[7], 10, 64)
			bytesSent, _ := strconv.ParseUint(fields[9], 10, 64)

			netInterface := types.NetworkInterface{
				Name:            interfaceName,
				BytesReceived:   bytesReceived,
				BytesSent:       bytesSent,
				PacketsReceived: packetsReceived,
				PacketsSent:     packetsSent,
				Errors:          0, // Would need additional parsing
				Drops:           0, // Would need additional parsing
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
	cmd = exec.CommandContext(ctx, "netstat", "-an")
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		connectionCount := 0
		for _, line := range lines {
			if strings.Contains(line, "ESTABLISHED") {
				connectionCount++
			}
		}
		metrics.Network.Connections = connectionCount
	}

	return nil
}

// AnalyzePerformance analyzes macOS performance metrics
func (d *DarwinPerformanceAnalyzer) AnalyzePerformance(ctx context.Context, metrics *types.PerformanceMetrics) (*types.AnalysisReport, error) {
	// Use the mock analyzer's analysis logic for now
	mock := NewMockPerformanceAnalyzer()
	return mock.AnalyzePerformance(ctx, metrics)
}

// GenerateReport generates a macOS performance report
func (d *DarwinPerformanceAnalyzer) GenerateReport(ctx context.Context, detailed bool) (*types.PerformanceReport, error) {
	metrics, err := d.CollectMetrics(ctx, time.Second)
	if err != nil {
		return nil, err
	}

	analysis, err := d.AnalyzePerformance(ctx, metrics)
	if err != nil {
		return nil, err
	}

	report := &types.PerformanceReport{
		Timestamp: time.Now(),
		Summary:   d.generateSummary(analysis),
		Metrics:   metrics,
		Analysis:  analysis,
		Detailed:  detailed,
	}

	return report, nil
}

// SaveBaseline saves current metrics as a baseline
func (d *DarwinPerformanceAnalyzer) SaveBaseline(ctx context.Context, filepath string) error {
	// Use mock implementation for now
	mock := NewMockPerformanceAnalyzer()
	return mock.SaveBaseline(ctx, filepath)
}

// CompareBaseline compares current metrics against a saved baseline
func (d *DarwinPerformanceAnalyzer) CompareBaseline(ctx context.Context, baselinePath string) (*types.ComparisonReport, error) {
	// Use mock implementation for now
	mock := NewMockPerformanceAnalyzer()
	return mock.CompareBaseline(ctx, baselinePath)
}

// GetOptimizationSuggestions returns macOS-specific optimization suggestions
func (d *DarwinPerformanceAnalyzer) GetOptimizationSuggestions(ctx context.Context) ([]*types.OptimizationSuggestion, error) {
	metrics, err := d.CollectMetrics(ctx, time.Second)
	if err != nil {
		return nil, err
	}

	suggestions := make([]*types.OptimizationSuggestion, 0)

	// macOS-specific CPU suggestions
	if metrics.CPU.Usage > 80 {
		suggestions = append(suggestions, &types.OptimizationSuggestion{
			Category:    "CPU",
			Priority:    "High",
			Title:       "High CPU Usage on macOS",
			Description: "CPU usage is elevated. Use Activity Monitor to identify resource-intensive processes.",
			Impact:      "Improved system responsiveness",
			Commands:    []string{"top -o cpu", "ps aux | sort -k 3 -nr"},
			Safe:        true,
		})
	}

	// macOS-specific memory suggestions
	if metrics.Memory.Usage > 85 {
		suggestions = append(suggestions, &types.OptimizationSuggestion{
			Category:    "Memory",
			Priority:    "High",
			Title:       "High Memory Usage on macOS",
			Description: "Memory usage is high. Consider closing unnecessary applications or restarting memory-intensive apps.",
			Impact:      "Reduced memory pressure and improved performance",
			Commands:    []string{"vm_stat", "ps aux | sort -k 4 -nr"},
			Safe:        true,
		})
	}

	// macOS-specific disk suggestions
	for _, disk := range metrics.Disk.Usage {
		if disk.Usage > 90 {
			suggestions = append(suggestions, &types.OptimizationSuggestion{
				Category:    "Disk",
				Priority:    "Critical",
				Title:       fmt.Sprintf("Disk %s Nearly Full", disk.Device),
				Description: "Disk space is critically low. Use Storage Management or clean up unnecessary files.",
				Impact:      "Prevent system instability and improve performance",
				Commands:    []string{"du -sh /*", "find ~/Downloads -type f -atime +30"},
				Safe:        true,
			})
		}
	}

	return suggestions, nil
}

// generateSummary generates a summary string for the analysis
func (d *DarwinPerformanceAnalyzer) generateSummary(analysis *types.AnalysisReport) string {
	switch analysis.Overall {
	case types.HealthLevelHealthy:
		return "macOS system performance is healthy with no major issues detected."
	case types.HealthLevelWarning:
		return fmt.Sprintf("macOS system performance has some warnings: %s", strings.Join(analysis.Warnings, ", "))
	case types.HealthLevelCritical:
		return fmt.Sprintf("macOS system performance has critical issues: %s", strings.Join(analysis.Bottlenecks, ", "))
	default:
		return "macOS system performance status is unknown."
	}
}
