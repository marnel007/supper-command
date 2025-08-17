package server

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

// WindowsServerManager manages Windows server operations using PowerShell and WMI
type WindowsServerManager struct {
	*BaseServerManager
}

// NewWindowsServerManager creates a new Windows server manager
func NewWindowsServerManager() *WindowsServerManager {
	return &WindowsServerManager{
		BaseServerManager: NewBaseServerManager(types.PlatformWindows),
	}
}

// GetHealthStatus returns Windows server health status
func (w *WindowsServerManager) GetHealthStatus(ctx context.Context) (*types.HealthStatus, error) {
	components := make(map[string]types.ComponentHealth)

	// Get CPU health
	if err := w.getCPUHealth(ctx, components); err != nil {
		return nil, types.NewServiceError("health", "cpu", err, "failed to get CPU health")
	}

	// Get memory health
	if err := w.getMemoryHealth(ctx, components); err != nil {
		return nil, types.NewServiceError("health", "memory", err, "failed to get memory health")
	}

	// Get disk health
	if err := w.getDiskHealth(ctx, components); err != nil {
		return nil, types.NewServiceError("health", "disk", err, "failed to get disk health")
	}

	// Get network health
	if err := w.getNetworkHealth(ctx, components); err != nil {
		return nil, types.NewServiceError("health", "network", err, "failed to get network health")
	}

	// Calculate overall health
	overall := w.CalculateOverallHealth(components)

	// Get system uptime
	uptime, _ := w.getSystemUptime(ctx)

	// Generate alerts
	alerts := w.GenerateHealthAlerts(components)

	status := &types.HealthStatus{
		Overall:     overall,
		Timestamp:   time.Now(),
		Components:  components,
		Uptime:      uptime,
		LoadAverage: []float64{0, 0, 0}, // Windows doesn't have load average
		Alerts:      alerts,
	}

	return status, nil
}

// getCPUHealth gets CPU health information
func (w *WindowsServerManager) getCPUHealth(ctx context.Context, components map[string]types.ComponentHealth) error {
	cmd := exec.CommandContext(ctx, "powershell", "-Command",
		"Get-Counter '\\Processor(_Total)\\% Processor Time' | Select-Object -ExpandProperty CounterSamples | Select-Object -ExpandProperty CookedValue")

	output, err := cmd.Output()
	if err != nil {
		// Fallback to wmic
		return w.getCPUHealthWMIC(ctx, components)
	}

	cpuUsageStr := strings.TrimSpace(string(output))
	cpuUsage, err := strconv.ParseFloat(cpuUsageStr, 64)
	if err != nil {
		cpuUsage = 0
	}

	var status types.HealthLevel
	var message string

	if cpuUsage >= 90 {
		status = types.HealthLevelCritical
		message = "CPU usage is critically high"
	} else if cpuUsage >= 70 {
		status = types.HealthLevelWarning
		message = "CPU usage is elevated"
	} else {
		status = types.HealthLevelHealthy
		message = "CPU usage is normal"
	}

	components["CPU"] = types.ComponentHealth{
		Status:      status,
		Value:       cpuUsage,
		Threshold:   70.0,
		Unit:        "%",
		Message:     message,
		LastChecked: time.Now(),
	}

	return nil
}

// getCPUHealthWMIC gets CPU health using WMIC as fallback
func (w *WindowsServerManager) getCPUHealthWMIC(ctx context.Context, components map[string]types.ComponentHealth) error {
	cmd := exec.CommandContext(ctx, "wmic", "cpu", "get", "loadpercentage", "/value")
	output, err := cmd.Output()
	if err != nil {
		components["CPU"] = types.ComponentHealth{
			Status:      types.HealthLevelUnknown,
			Value:       0,
			Threshold:   70.0,
			Unit:        "%",
			Message:     "Unable to determine CPU usage",
			LastChecked: time.Now(),
		}
		return nil
	}

	lines := strings.Split(string(output), "\n")
	cpuUsage := 0.0

	for _, line := range lines {
		if strings.HasPrefix(line, "LoadPercentage=") {
			valueStr := strings.TrimPrefix(line, "LoadPercentage=")
			valueStr = strings.TrimSpace(valueStr)
			if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
				cpuUsage = value
				break
			}
		}
	}

	var status types.HealthLevel
	var message string

	if cpuUsage >= 90 {
		status = types.HealthLevelCritical
		message = "CPU usage is critically high"
	} else if cpuUsage >= 70 {
		status = types.HealthLevelWarning
		message = "CPU usage is elevated"
	} else {
		status = types.HealthLevelHealthy
		message = "CPU usage is normal"
	}

	components["CPU"] = types.ComponentHealth{
		Status:      status,
		Value:       cpuUsage,
		Threshold:   70.0,
		Unit:        "%",
		Message:     message,
		LastChecked: time.Now(),
	}

	return nil
}

// getMemoryHealth gets memory health information
func (w *WindowsServerManager) getMemoryHealth(ctx context.Context, components map[string]types.ComponentHealth) error {
	cmd := exec.CommandContext(ctx, "powershell", "-Command", `
		$mem = Get-CimInstance -ClassName Win32_OperatingSystem
		$totalMem = $mem.TotalVisibleMemorySize * 1024
		$freeMem = $mem.FreePhysicalMemory * 1024
		$usedMem = $totalMem - $freeMem
		$usage = ($usedMem / $totalMem) * 100
		$usage
	`)

	output, err := cmd.Output()
	if err != nil {
		components["Memory"] = types.ComponentHealth{
			Status:      types.HealthLevelUnknown,
			Value:       0,
			Threshold:   80.0,
			Unit:        "%",
			Message:     "Unable to determine memory usage",
			LastChecked: time.Now(),
		}
		return nil
	}

	memUsageStr := strings.TrimSpace(string(output))
	memUsage, err := strconv.ParseFloat(memUsageStr, 64)
	if err != nil {
		memUsage = 0
	}

	var status types.HealthLevel
	var message string

	if memUsage >= 95 {
		status = types.HealthLevelCritical
		message = "Memory usage is critically high"
	} else if memUsage >= 80 {
		status = types.HealthLevelWarning
		message = "Memory usage is elevated"
	} else {
		status = types.HealthLevelHealthy
		message = "Memory usage is normal"
	}

	components["Memory"] = types.ComponentHealth{
		Status:      status,
		Value:       memUsage,
		Threshold:   80.0,
		Unit:        "%",
		Message:     message,
		LastChecked: time.Now(),
	}

	return nil
}

// getDiskHealth gets disk health information
func (w *WindowsServerManager) getDiskHealth(ctx context.Context, components map[string]types.ComponentHealth) error {
	cmd := exec.CommandContext(ctx, "powershell", "-Command", `
		Get-CimInstance -ClassName Win32_LogicalDisk | Where-Object {$_.DriveType -eq 3} | ForEach-Object {
			$usage = (($_.Size - $_.FreeSpace) / $_.Size) * 100
			[PSCustomObject]@{
				Drive = $_.DeviceID
				Usage = $usage
			}
		} | ConvertTo-Json
	`)

	output, err := cmd.Output()
	if err != nil {
		components["Disk"] = types.ComponentHealth{
			Status:      types.HealthLevelUnknown,
			Value:       0,
			Threshold:   85.0,
			Unit:        "%",
			Message:     "Unable to determine disk usage",
			LastChecked: time.Now(),
		}
		return nil
	}

	var diskInfos []struct {
		Drive string  `json:"Drive"`
		Usage float64 `json:"Usage"`
	}

	// Handle both single object and array responses
	outputStr := strings.TrimSpace(string(output))
	if strings.HasPrefix(outputStr, "[") {
		if err := json.Unmarshal(output, &diskInfos); err != nil {
			return err
		}
	} else {
		var singleDisk struct {
			Drive string  `json:"Drive"`
			Usage float64 `json:"Usage"`
		}
		if err := json.Unmarshal(output, &singleDisk); err != nil {
			return err
		}
		diskInfos = []struct {
			Drive string  `json:"Drive"`
			Usage float64 `json:"Usage"`
		}{singleDisk}
	}

	// Calculate average disk usage
	totalUsage := 0.0
	maxUsage := 0.0
	for _, disk := range diskInfos {
		totalUsage += disk.Usage
		if disk.Usage > maxUsage {
			maxUsage = disk.Usage
		}
	}

	avgUsage := totalUsage / float64(len(diskInfos))

	var status types.HealthLevel
	var message string

	if maxUsage >= 95 {
		status = types.HealthLevelCritical
		message = "One or more disks are critically full"
	} else if maxUsage >= 85 {
		status = types.HealthLevelWarning
		message = "One or more disks are getting full"
	} else {
		status = types.HealthLevelHealthy
		message = "Disk usage is normal"
	}

	components["Disk"] = types.ComponentHealth{
		Status:      status,
		Value:       avgUsage,
		Threshold:   85.0,
		Unit:        "%",
		Message:     message,
		LastChecked: time.Now(),
	}

	return nil
}

// getNetworkHealth gets network health information
func (w *WindowsServerManager) getNetworkHealth(ctx context.Context, components map[string]types.ComponentHealth) error {
	// For Windows, we'll check network adapter status
	cmd := exec.CommandContext(ctx, "powershell", "-Command", `
		Get-NetAdapter | Where-Object {$_.Status -eq "Up"} | Measure-Object | Select-Object -ExpandProperty Count
	`)

	output, err := cmd.Output()
	if err != nil {
		components["Network"] = types.ComponentHealth{
			Status:      types.HealthLevelUnknown,
			Value:       0,
			Threshold:   1.0,
			Unit:        "adapters",
			Message:     "Unable to determine network status",
			LastChecked: time.Now(),
		}
		return nil
	}

	activeAdaptersStr := strings.TrimSpace(string(output))
	activeAdapters, err := strconv.ParseFloat(activeAdaptersStr, 64)
	if err != nil {
		activeAdapters = 0
	}

	var status types.HealthLevel
	var message string

	if activeAdapters == 0 {
		status = types.HealthLevelCritical
		message = "No active network adapters"
	} else if activeAdapters < 2 {
		status = types.HealthLevelWarning
		message = "Limited network connectivity"
	} else {
		status = types.HealthLevelHealthy
		message = "Network connectivity is normal"
	}

	components["Network"] = types.ComponentHealth{
		Status:      status,
		Value:       activeAdapters,
		Threshold:   1.0,
		Unit:        "adapters",
		Message:     message,
		LastChecked: time.Now(),
	}

	return nil
}

// getSystemUptime gets Windows system uptime
func (w *WindowsServerManager) getSystemUptime(ctx context.Context) (time.Duration, error) {
	cmd := exec.CommandContext(ctx, "powershell", "-Command", `
		(Get-CimInstance -ClassName Win32_OperatingSystem).LastBootUpTime
	`)

	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	bootTimeStr := strings.TrimSpace(string(output))
	bootTime, err := time.Parse("1/2/2006 3:04:05 PM", bootTimeStr)
	if err != nil {
		// Try alternative format
		bootTime, err = time.Parse("2006-01-02T15:04:05.000000-07:00", bootTimeStr)
		if err != nil {
			return 0, err
		}
	}

	uptime := time.Since(bootTime)
	return uptime, nil
}

// ListServices returns Windows services information
func (w *WindowsServerManager) ListServices(ctx context.Context) ([]*types.ServiceInfo, error) {
	cmd := exec.CommandContext(ctx, "powershell", "-Command", `
		Get-Service | ForEach-Object {
			$proc = Get-Process -Name $_.Name -ErrorAction SilentlyContinue
			[PSCustomObject]@{
				Name = $_.Name
				DisplayName = $_.DisplayName
				Status = $_.Status
				StartType = $_.StartType
				PID = if($proc) { $proc.Id } else { 0 }
				Memory = if($proc) { $proc.WorkingSet64 } else { 0 }
				CPU = if($proc) { $proc.CPU } else { 0 }
			}
		} | ConvertTo-Json
	`)

	output, err := cmd.Output()
	if err != nil {
		return nil, types.NewServiceError("services", "list", err, "failed to list Windows services")
	}

	var serviceInfos []struct {
		Name        string  `json:"Name"`
		DisplayName string  `json:"DisplayName"`
		Status      string  `json:"Status"`
		StartType   string  `json:"StartType"`
		PID         int     `json:"PID"`
		Memory      int64   `json:"Memory"`
		CPU         float64 `json:"CPU"`
	}

	if err := json.Unmarshal(output, &serviceInfos); err != nil {
		return nil, types.NewServiceError("services", "parse", err, "failed to parse service information")
	}

	services := make([]*types.ServiceInfo, 0, len(serviceInfos))
	for _, info := range serviceInfos {
		var status types.ServiceStatus
		switch strings.ToLower(info.Status) {
		case "running":
			status = types.ServiceStatusRunning
		case "stopped":
			status = types.ServiceStatusStopped
		default:
			status = types.ServiceStatusUnknown
		}

		var startType types.StartType
		switch strings.ToLower(info.StartType) {
		case "automatic":
			startType = types.StartTypeAutomatic
		case "manual":
			startType = types.StartTypeManual
		case "disabled":
			startType = types.StartTypeDisabled
		default:
			startType = types.StartTypeManual
		}

		service := &types.ServiceInfo{
			Name:        info.Name,
			DisplayName: info.DisplayName,
			Status:      status,
			StartType:   startType,
			PID:         info.PID,
			Memory:      info.Memory,
			CPU:         info.CPU,
			Uptime:      0, // Would need additional calculation
			Description: info.DisplayName,
			Path:        "", // Would need additional query
		}

		services = append(services, service)
	}

	return services, nil
}

// ControlService controls a Windows service
func (w *WindowsServerManager) ControlService(ctx context.Context, serviceName string, action types.ServiceAction) error {
	var cmd *exec.Cmd

	switch action {
	case types.ServiceActionStart:
		cmd = exec.CommandContext(ctx, "powershell", "-Command", fmt.Sprintf("Start-Service -Name '%s'", serviceName))
	case types.ServiceActionStop:
		cmd = exec.CommandContext(ctx, "powershell", "-Command", fmt.Sprintf("Stop-Service -Name '%s'", serviceName))
	case types.ServiceActionRestart:
		cmd = exec.CommandContext(ctx, "powershell", "-Command", fmt.Sprintf("Restart-Service -Name '%s'", serviceName))
	default:
		return types.NewServiceError(serviceName, string(action), nil, "unsupported service action")
	}

	if err := cmd.Run(); err != nil {
		return types.NewServiceError(serviceName, string(action), err, "failed to control Windows service")
	}

	return nil
}

// GetActiveUsers returns active Windows user sessions
func (w *WindowsServerManager) GetActiveUsers(ctx context.Context) ([]*types.UserSession, error) {
	cmd := exec.CommandContext(ctx, "query", "user")
	output, err := cmd.Output()
	if err != nil {
		return nil, types.NewServiceError("users", "query", err, "failed to query active users")
	}

	lines := strings.Split(string(output), "\n")
	users := make([]*types.UserSession, 0)

	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue // Skip header and empty lines
		}

		fields := strings.Fields(line)
		if len(fields) >= 4 {
			username := fields[0]
			sessionID := fields[1]

			// Parse login time (simplified)
			loginTime := time.Now().Add(-time.Hour) // Mock login time

			user := &types.UserSession{
				Username:     username,
				SessionID:    sessionID,
				Terminal:     "console",
				LoginTime:    loginTime,
				IdleTime:     0,
				RemoteHost:   "",
				ProcessCount: 0, // Would need additional query
			}

			users = append(users, user)
		}
	}

	return users, nil
}

// GetServiceLogs returns Windows service logs (Event Log)
func (w *WindowsServerManager) GetServiceLogs(ctx context.Context, serviceName string, tail bool) (*types.LogStream, error) {
	cmd := exec.CommandContext(ctx, "powershell", "-Command", fmt.Sprintf(`
		Get-WinEvent -FilterHashtable @{LogName='System'; ProviderName='Service Control Manager'} -MaxEvents 10 | 
		Where-Object {$_.Message -like '*%s*'} | 
		ForEach-Object {
			[PSCustomObject]@{
				TimeCreated = $_.TimeCreated
				LevelDisplayName = $_.LevelDisplayName
				Message = $_.Message
			}
		} | ConvertTo-Json
	`, serviceName))

	output, err := cmd.Output()
	if err != nil {
		// Return empty log stream if no events found
		return &types.LogStream{
			ServiceName: serviceName,
			Entries:     []types.LogEntry{},
			Following:   tail,
			StartTime:   time.Now(),
		}, nil
	}

	var logInfos []struct {
		TimeCreated      time.Time `json:"TimeCreated"`
		LevelDisplayName string    `json:"LevelDisplayName"`
		Message          string    `json:"Message"`
	}

	if err := json.Unmarshal(output, &logInfos); err != nil {
		return nil, types.NewServiceError(serviceName, "logs", err, "failed to parse service logs")
	}

	entries := make([]types.LogEntry, 0, len(logInfos))
	for _, info := range logInfos {
		entry := types.LogEntry{
			Timestamp: info.TimeCreated,
			Level:     info.LevelDisplayName,
			Message:   info.Message,
			Source:    serviceName,
		}
		entries = append(entries, entry)
	}

	logStream := &types.LogStream{
		ServiceName: serviceName,
		Entries:     entries,
		Following:   tail,
		StartTime:   time.Now(),
	}

	return logStream, nil
}

// ConfigureAlerts configures Windows-specific alert settings
func (w *WindowsServerManager) ConfigureAlerts(ctx context.Context, config *types.AlertConfig) error {
	// Windows-specific alert configuration would go here
	// This could involve Windows Event Log, Performance Counters, etc.
	return nil
}

// BackupConfiguration backs up Windows server configuration
func (w *WindowsServerManager) BackupConfiguration(ctx context.Context, backupPath string) error {
	// Windows-specific configuration backup would go here
	// This could involve registry, services, scheduled tasks, etc.
	return nil
}
