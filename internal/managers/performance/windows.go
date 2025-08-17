package performance

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"suppercommand/internal/types"
)

// WindowsPerformanceAnalyzer manages performance analysis on Windows using WMI and PowerShell
type WindowsPerformanceAnalyzer struct {
	*BasePerformanceAnalyzer
}

// NewWindowsPerformanceAnalyzer creates a new Windows performance analyzer
func NewWindowsPerformanceAnalyzer() *WindowsPerformanceAnalyzer {
	return &WindowsPerformanceAnalyzer{
		BasePerformanceAnalyzer: NewBasePerformanceAnalyzer(types.PlatformWindows),
	}
}

// CollectMetrics collects performance metrics on Windows
func (w *WindowsPerformanceAnalyzer) CollectMetrics(ctx context.Context, duration time.Duration) (*types.PerformanceMetrics, error) {
	metrics := w.CollectBasicMetrics()
	metrics.Duration = duration

	// Collect CPU metrics
	if err := w.collectCPUMetrics(ctx, metrics); err != nil {
		return nil, types.NewPerformanceError("CPU", "collect", err, "failed to collect CPU metrics")
	}

	// Collect memory metrics
	if err := w.collectMemoryMetrics(ctx, metrics); err != nil {
		return nil, types.NewPerformanceError("Memory", "collect", err, "failed to collect memory metrics")
	}

	// Collect disk metrics
	if err := w.collectDiskMetrics(ctx, metrics); err != nil {
		return nil, types.NewPerformanceError("Disk", "collect", err, "failed to collect disk metrics")
	}

	// Collect network metrics
	if err := w.collectNetworkMetrics(ctx, metrics); err != nil {
		return nil, types.NewPerformanceError("Network", "collect", err, "failed to collect network metrics")
	}

	w.CalculateUsagePercentages(metrics)
	return metrics, nil
}

// collectCPUMetrics collects CPU performance metrics using PowerShell
func (w *WindowsPerformanceAnalyzer) collectCPUMetrics(ctx context.Context, metrics *types.PerformanceMetrics) error {
	// Get CPU usage using PowerShell
	cmd := exec.CommandContext(ctx, "powershell", "-Command",
		"Get-Counter '\\Processor(_Total)\\% Processor Time' | Select-Object -ExpandProperty CounterSamples | Select-Object -ExpandProperty CookedValue")

	output, err := cmd.Output()
	if err != nil {
		// Fallback to wmic
		return w.collectCPUMetricsWMIC(ctx, metrics)
	}

	cpuUsageStr := strings.TrimSpace(string(output))
	if cpuUsage, err := strconv.ParseFloat(cpuUsageStr, 64); err == nil {
		metrics.CPU.Usage = cpuUsage
	}

	// Get process count
	cmd = exec.CommandContext(ctx, "powershell", "-Command",
		"(Get-Process).Count")

	if output, err := cmd.Output(); err == nil {
		processCountStr := strings.TrimSpace(string(output))
		if processCount, err := strconv.Atoi(processCountStr); err == nil {
			metrics.CPU.Processes = processCount
		}
	}

	// Get load average (approximation using processor queue length)
	cmd = exec.CommandContext(ctx, "powershell", "-Command",
		"Get-Counter '\\System\\Processor Queue Length' | Select-Object -ExpandProperty CounterSamples | Select-Object -ExpandProperty CookedValue")

	if output, err := cmd.Output(); err == nil {
		queueLengthStr := strings.TrimSpace(string(output))
		if queueLength, err := strconv.ParseFloat(queueLengthStr, 64); err == nil {
			// Approximate load average
			metrics.CPU.LoadAverage = []float64{queueLength, queueLength, queueLength}
		}
	}

	return nil
}

// collectCPUMetricsWMIC collects CPU metrics using WMIC as fallback
func (w *WindowsPerformanceAnalyzer) collectCPUMetricsWMIC(ctx context.Context, metrics *types.PerformanceMetrics) error {
	cmd := exec.CommandContext(ctx, "wmic", "cpu", "get", "loadpercentage", "/value")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "LoadPercentage=") {
			valueStr := strings.TrimPrefix(line, "LoadPercentage=")
			valueStr = strings.TrimSpace(valueStr)
			if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
				metrics.CPU.Usage = value
				break
			}
		}
	}

	return nil
}

// collectMemoryMetrics collects memory performance metrics
func (w *WindowsPerformanceAnalyzer) collectMemoryMetrics(ctx context.Context, metrics *types.PerformanceMetrics) error {
	// Get memory information using PowerShell
	cmd := exec.CommandContext(ctx, "powershell", "-Command", `
		$mem = Get-CimInstance -ClassName Win32_OperatingSystem
		$totalMem = $mem.TotalVisibleMemorySize * 1024
		$freeMem = $mem.FreePhysicalMemory * 1024
		$usedMem = $totalMem - $freeMem
		
		$pageFile = Get-CimInstance -ClassName Win32_PageFileUsage
		$totalPageFile = ($pageFile | Measure-Object -Property AllocatedBaseSize -Sum).Sum * 1024 * 1024
		$usedPageFile = ($pageFile | Measure-Object -Property CurrentUsage -Sum).Sum * 1024 * 1024
		
		@{
			TotalMemory = $totalMem
			UsedMemory = $usedMem
			FreeMemory = $freeMem
			TotalPageFile = $totalPageFile
			UsedPageFile = $usedPageFile
		} | ConvertTo-Json
	`)

	output, err := cmd.Output()
	if err != nil {
		return w.collectMemoryMetricsWMIC(ctx, metrics)
	}

	var memInfo struct {
		TotalMemory   uint64 `json:"TotalMemory"`
		UsedMemory    uint64 `json:"UsedMemory"`
		FreeMemory    uint64 `json:"FreeMemory"`
		TotalPageFile uint64 `json:"TotalPageFile"`
		UsedPageFile  uint64 `json:"UsedPageFile"`
	}

	if err := json.Unmarshal(output, &memInfo); err == nil {
		metrics.Memory.Total = memInfo.TotalMemory
		metrics.Memory.Used = memInfo.UsedMemory
		metrics.Memory.Available = memInfo.FreeMemory
		metrics.Memory.SwapTotal = memInfo.TotalPageFile
		metrics.Memory.SwapUsed = memInfo.UsedPageFile
	}

	return nil
}

// collectMemoryMetricsWMIC collects memory metrics using WMIC as fallback
func (w *WindowsPerformanceAnalyzer) collectMemoryMetricsWMIC(ctx context.Context, metrics *types.PerformanceMetrics) error {
	// Get total physical memory
	cmd := exec.CommandContext(ctx, "wmic", "computersystem", "get", "TotalPhysicalMemory", "/value")
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "TotalPhysicalMemory=") {
				valueStr := strings.TrimPrefix(line, "TotalPhysicalMemory=")
				valueStr = strings.TrimSpace(valueStr)
				if value, err := strconv.ParseUint(valueStr, 10, 64); err == nil {
					metrics.Memory.Total = value
					break
				}
			}
		}
	}

	// Get available memory
	cmd = exec.CommandContext(ctx, "wmic", "OS", "get", "FreePhysicalMemory", "/value")
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "FreePhysicalMemory=") {
				valueStr := strings.TrimPrefix(line, "FreePhysicalMemory=")
				valueStr = strings.TrimSpace(valueStr)
				if value, err := strconv.ParseUint(valueStr, 10, 64); err == nil {
					metrics.Memory.Available = value * 1024 // Convert from KB to bytes
					metrics.Memory.Used = metrics.Memory.Total - metrics.Memory.Available
					break
				}
			}
		}
	}

	return nil
}

// collectDiskMetrics collects disk performance metrics
func (w *WindowsPerformanceAnalyzer) collectDiskMetrics(ctx context.Context, metrics *types.PerformanceMetrics) error {
	cmd := exec.CommandContext(ctx, "powershell", "-Command", `
		Get-CimInstance -ClassName Win32_LogicalDisk | Where-Object {$_.DriveType -eq 3} | ForEach-Object {
			@{
				Device = $_.DeviceID
				MountPoint = $_.DeviceID
				Total = $_.Size
				Used = $_.Size - $_.FreeSpace
				Available = $_.FreeSpace
			}
		} | ConvertTo-Json
	`)

	output, err := cmd.Output()
	if err != nil {
		return err
	}

	var diskInfos []struct {
		Device     string `json:"Device"`
		MountPoint string `json:"MountPoint"`
		Total      uint64 `json:"Total"`
		Used       uint64 `json:"Used"`
		Available  uint64 `json:"Available"`
	}

	// Handle both single object and array responses
	outputStr := strings.TrimSpace(string(output))
	if strings.HasPrefix(outputStr, "[") {
		if err := json.Unmarshal(output, &diskInfos); err != nil {
			return err
		}
	} else {
		var singleDisk struct {
			Device     string `json:"Device"`
			MountPoint string `json:"MountPoint"`
			Total      uint64 `json:"Total"`
			Used       uint64 `json:"Used"`
			Available  uint64 `json:"Available"`
		}
		if err := json.Unmarshal(output, &singleDisk); err != nil {
			return err
		}
		diskInfos = []struct {
			Device     string `json:"Device"`
			MountPoint string `json:"MountPoint"`
			Total      uint64 `json:"Total"`
			Used       uint64 `json:"Used"`
			Available  uint64 `json:"Available"`
		}{singleDisk}
	}

	for _, diskInfo := range diskInfos {
		metrics.Disk.Usage = append(metrics.Disk.Usage, types.DiskUsage{
			Device:     diskInfo.Device,
			MountPoint: diskInfo.MountPoint,
			Total:      diskInfo.Total,
			Used:       diskInfo.Used,
			Available:  diskInfo.Available,
		})
	}

	return nil
}

// collectNetworkMetrics collects network performance metrics
func (w *WindowsPerformanceAnalyzer) collectNetworkMetrics(ctx context.Context, metrics *types.PerformanceMetrics) error {
	cmd := exec.CommandContext(ctx, "powershell", "-Command", `
		Get-CimInstance -ClassName Win32_PerfRawData_Tcpip_NetworkInterface | Where-Object {$_.Name -notlike "*Loopback*" -and $_.Name -notlike "*Teredo*"} | ForEach-Object {
			@{
				Name = $_.Name
				BytesReceived = $_.BytesReceivedPerSec
				BytesSent = $_.BytesSentPerSec
				PacketsReceived = $_.PacketsReceivedPerSec
				PacketsSent = $_.PacketsSentPerSec
			}
		} | ConvertTo-Json
	`)

	output, err := cmd.Output()
	if err != nil {
		return err
	}

	var networkInfos []struct {
		Name            string `json:"Name"`
		BytesReceived   uint64 `json:"BytesReceived"`
		BytesSent       uint64 `json:"BytesSent"`
		PacketsReceived uint64 `json:"PacketsReceived"`
		PacketsSent     uint64 `json:"PacketsSent"`
	}

	// Handle both single object and array responses
	outputStr := strings.TrimSpace(string(output))
	if strings.HasPrefix(outputStr, "[") {
		if err := json.Unmarshal(output, &networkInfos); err != nil {
			return err
		}
	} else {
		var singleNetwork struct {
			Name            string `json:"Name"`
			BytesReceived   uint64 `json:"BytesReceived"`
			BytesSent       uint64 `json:"BytesSent"`
			PacketsReceived uint64 `json:"PacketsReceived"`
			PacketsSent     uint64 `json:"PacketsSent"`
		}
		if err := json.Unmarshal(output, &singleNetwork); err != nil {
			return err
		}
		networkInfos = []struct {
			Name            string `json:"Name"`
			BytesReceived   uint64 `json:"BytesReceived"`
			BytesSent       uint64 `json:"BytesSent"`
			PacketsReceived uint64 `json:"PacketsReceived"`
			PacketsSent     uint64 `json:"PacketsSent"`
		}{singleNetwork}
	}

	var totalBytesReceived, totalBytesSent, totalPacketsReceived, totalPacketsSent uint64

	for _, networkInfo := range networkInfos {
		netInterface := types.NetworkInterface{
			Name:            networkInfo.Name,
			BytesReceived:   networkInfo.BytesReceived,
			BytesSent:       networkInfo.BytesSent,
			PacketsReceived: networkInfo.PacketsReceived,
			PacketsSent:     networkInfo.PacketsSent,
			Errors:          0, // Would need additional WMI query
			Drops:           0, // Would need additional WMI query
		}

		metrics.Network.Interfaces = append(metrics.Network.Interfaces, netInterface)

		totalBytesReceived += networkInfo.BytesReceived
		totalBytesSent += networkInfo.BytesSent
		totalPacketsReceived += networkInfo.PacketsReceived
		totalPacketsSent += networkInfo.PacketsSent
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

// AnalyzePerformance analyzes Windows performance metrics
func (w *WindowsPerformanceAnalyzer) AnalyzePerformance(ctx context.Context, metrics *types.PerformanceMetrics) (*types.AnalysisReport, error) {
	// Use the mock analyzer's analysis logic for now
	mock := NewMockPerformanceAnalyzer()
	return mock.AnalyzePerformance(ctx, metrics)
}

// GenerateReport generates a Windows performance report
func (w *WindowsPerformanceAnalyzer) GenerateReport(ctx context.Context, detailed bool) (*types.PerformanceReport, error) {
	metrics, err := w.CollectMetrics(ctx, time.Second)
	if err != nil {
		return nil, err
	}

	analysis, err := w.AnalyzePerformance(ctx, metrics)
	if err != nil {
		return nil, err
	}

	report := &types.PerformanceReport{
		Timestamp: time.Now(),
		Summary:   w.generateSummary(analysis),
		Metrics:   metrics,
		Analysis:  analysis,
		Detailed:  detailed,
	}

	return report, nil
}

// SaveBaseline saves current metrics as a baseline
func (w *WindowsPerformanceAnalyzer) SaveBaseline(ctx context.Context, filepath string) error {
	// Use mock implementation for now
	mock := NewMockPerformanceAnalyzer()
	return mock.SaveBaseline(ctx, filepath)
}

// CompareBaseline compares current metrics against a saved baseline
func (w *WindowsPerformanceAnalyzer) CompareBaseline(ctx context.Context, baselinePath string) (*types.ComparisonReport, error) {
	// Use mock implementation for now
	mock := NewMockPerformanceAnalyzer()
	return mock.CompareBaseline(ctx, baselinePath)
}

// GetOptimizationSuggestions returns Windows-specific optimization suggestions
func (w *WindowsPerformanceAnalyzer) GetOptimizationSuggestions(ctx context.Context) ([]*types.OptimizationSuggestion, error) {
	metrics, err := w.CollectMetrics(ctx, time.Second)
	if err != nil {
		return nil, err
	}

	suggestions := make([]*types.OptimizationSuggestion, 0)

	// Windows-specific CPU suggestions
	if metrics.CPU.Usage > 80 {
		suggestions = append(suggestions, &types.OptimizationSuggestion{
			Category:    "CPU",
			Priority:    "High",
			Title:       "High CPU Usage on Windows",
			Description: "CPU usage is elevated. Check Task Manager for resource-intensive processes.",
			Impact:      "Improved system responsiveness",
			Commands:    []string{"tasklist /fo csv", "wmic process get name,processid,percentprocessortime"},
			Safe:        true,
		})
	}

	// Windows-specific memory suggestions
	if metrics.Memory.Usage > 85 {
		suggestions = append(suggestions, &types.OptimizationSuggestion{
			Category:    "Memory",
			Priority:    "High",
			Title:       "High Memory Usage on Windows",
			Description: "Memory usage is high. Consider closing unnecessary applications or increasing virtual memory.",
			Impact:      "Reduced memory pressure and improved performance",
			Commands:    []string{"tasklist /fo csv", "wmic process get name,processid,workingsetsize"},
			Safe:        true,
		})
	}

	// Windows-specific disk suggestions
	for _, disk := range metrics.Disk.Usage {
		if disk.Usage > 90 {
			suggestions = append(suggestions, &types.OptimizationSuggestion{
				Category:    "Disk",
				Priority:    "Critical",
				Title:       fmt.Sprintf("Disk %s Nearly Full", disk.Device),
				Description: "Disk space is critically low. Run Disk Cleanup or move files to free up space.",
				Impact:      "Prevent system instability and improve performance",
				Commands:    []string{"cleanmgr", "dism /online /cleanup-image /analyzecomponentstore"},
				Safe:        true,
			})
		}
	}

	return suggestions, nil
}

// generateSummary generates a summary string for the analysis
func (w *WindowsPerformanceAnalyzer) generateSummary(analysis *types.AnalysisReport) string {
	switch analysis.Overall {
	case types.HealthLevelHealthy:
		return "Windows system performance is healthy with no major issues detected."
	case types.HealthLevelWarning:
		return fmt.Sprintf("Windows system performance has some warnings: %s", strings.Join(analysis.Warnings, ", "))
	case types.HealthLevelCritical:
		return fmt.Sprintf("Windows system performance has critical issues: %s", strings.Join(analysis.Bottlenecks, ", "))
	default:
		return "Windows system performance status is unknown."
	}
}
