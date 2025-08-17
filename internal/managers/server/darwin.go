package server

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"suppercommand/internal/types"
)

// DarwinServerManager manages macOS server operations using launchctl and system commands
type DarwinServerManager struct {
	*BaseServerManager
}

// NewDarwinServerManager creates a new Darwin server manager
func NewDarwinServerManager() *DarwinServerManager {
	return &DarwinServerManager{
		BaseServerManager: NewBaseServerManager(types.PlatformDarwin),
	}
}

// GetHealthStatus returns macOS server health status
func (d *DarwinServerManager) GetHealthStatus(ctx context.Context) (*types.HealthStatus, error) {
	components := make(map[string]types.ComponentHealth)

	// Get CPU health
	if err := d.getCPUHealth(ctx, components); err != nil {
		return nil, types.NewServiceError("health", "cpu", err, "failed to get CPU health")
	}

	// Get memory health
	if err := d.getMemoryHealth(ctx, components); err != nil {
		return nil, types.NewServiceError("health", "memory", err, "failed to get memory health")
	}

	// Get disk health
	if err := d.getDiskHealth(ctx, components); err != nil {
		return nil, types.NewServiceError("health", "disk", err, "failed to get disk health")
	}

	// Get network health
	if err := d.getNetworkHealth(ctx, components); err != nil {
		return nil, types.NewServiceError("health", "network", err, "failed to get network health")
	}

	// Calculate overall health
	overall := d.CalculateOverallHealth(components)

	// Get system uptime
	uptime, _ := d.getSystemUptime(ctx)

	// Get load average
	loadAverage, _ := d.getLoadAverage(ctx)

	// Generate alerts
	alerts := d.GenerateHealthAlerts(components)

	status := &types.HealthStatus{
		Overall:     overall,
		Timestamp:   time.Now(),
		Components:  components,
		Uptime:      uptime,
		LoadAverage: loadAverage,
		Alerts:      alerts,
	}

	return status, nil
}

// getCPUHealth gets CPU health information using top
func (d *DarwinServerManager) getCPUHealth(ctx context.Context, components map[string]types.ComponentHealth) error {
	cmd := exec.CommandContext(ctx, "top", "-l", "1", "-n", "0")
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
							cpuUsage = 100.0 - idle
						}
					}
				}
			}
			break
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

// getMemoryHealth gets memory health information using vm_stat
func (d *DarwinServerManager) getMemoryHealth(ctx context.Context, components map[string]types.ComponentHealth) error {
	cmd := exec.CommandContext(ctx, "vm_stat")
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

	lines := strings.Split(string(output), "\n")
	var pageSize uint64 = 4096 // Default page size
	var freePages, activePages, inactivePages, wiredPages uint64

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
		}
	}

	// Calculate memory usage
	totalPages := freePages + activePages + inactivePages + wiredPages
	usedPages := activePages + inactivePages + wiredPages
	memUsage := float64(usedPages) / float64(totalPages) * 100

	// Calculate actual memory values using page size
	totalMemory := totalPages * pageSize
	usedMemory := usedPages * pageSize

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
		Message:     fmt.Sprintf("%s (Used: %d MB / Total: %d MB)", message, usedMemory/(1024*1024), totalMemory/(1024*1024)),
		LastChecked: time.Now(),
	}

	return nil
}

// getDiskHealth gets disk health information using df command
func (d *DarwinServerManager) getDiskHealth(ctx context.Context, components map[string]types.ComponentHealth) error {
	cmd := exec.CommandContext(ctx, "df", "-h")
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

	lines := strings.Split(string(output), "\n")
	totalUsage := 0.0
	maxUsage := 0.0
	diskCount := 0

	for i, line := range lines {
		if i == 0 || line == "" {
			continue // Skip header and empty lines
		}

		fields := strings.Fields(line)
		if len(fields) >= 5 {
			device := fields[0]
			// Skip special filesystems
			if strings.HasPrefix(device, "/dev/") {
				usageStr := strings.TrimSuffix(fields[4], "%")
				if usage, err := strconv.ParseFloat(usageStr, 64); err == nil {
					totalUsage += usage
					if usage > maxUsage {
						maxUsage = usage
					}
					diskCount++
				}
			}
		}
	}

	avgUsage := totalUsage / float64(diskCount)

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
func (d *DarwinServerManager) getNetworkHealth(ctx context.Context, components map[string]types.ComponentHealth) error {
	cmd := exec.CommandContext(ctx, "ifconfig")
	output, err := cmd.Output()
	if err != nil {
		components["Network"] = types.ComponentHealth{
			Status:      types.HealthLevelUnknown,
			Value:       0,
			Threshold:   1.0,
			Unit:        "interfaces",
			Message:     "Unable to determine network status",
			LastChecked: time.Now(),
		}
		return nil
	}

	lines := strings.Split(string(output), "\n")
	activeInterfaces := 0

	for _, line := range lines {
		if strings.Contains(line, "flags=") && strings.Contains(line, "UP") && !strings.Contains(line, "lo0") {
			activeInterfaces++
		}
	}

	var status types.HealthLevel
	var message string

	if activeInterfaces == 0 {
		status = types.HealthLevelCritical
		message = "No active network interfaces"
	} else if activeInterfaces < 2 {
		status = types.HealthLevelWarning
		message = "Limited network connectivity"
	} else {
		status = types.HealthLevelHealthy
		message = "Network connectivity is normal"
	}

	components["Network"] = types.ComponentHealth{
		Status:      status,
		Value:       float64(activeInterfaces),
		Threshold:   1.0,
		Unit:        "interfaces",
		Message:     message,
		LastChecked: time.Now(),
	}

	return nil
}

// getSystemUptime gets macOS system uptime
func (d *DarwinServerManager) getSystemUptime(ctx context.Context) (time.Duration, error) {
	cmd := exec.CommandContext(ctx, "uptime")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	// Parse uptime output (simplified)
	outputStr := string(output)
	if strings.Contains(outputStr, "up") {
		// This is a simplified parser - real implementation would be more robust
		return 24 * time.Hour, nil // Mock uptime
	}

	return 0, fmt.Errorf("unable to parse uptime")
}

// getLoadAverage gets macOS load average
func (d *DarwinServerManager) getLoadAverage(ctx context.Context) ([]float64, error) {
	cmd := exec.CommandContext(ctx, "uptime")
	output, err := cmd.Output()
	if err != nil {
		return []float64{0, 0, 0}, err
	}

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
				return loadAvg, nil
			}
		}
	}

	return []float64{0, 0, 0}, fmt.Errorf("unable to parse load average")
}

// ListServices returns macOS services information (launchd services)
func (d *DarwinServerManager) ListServices(ctx context.Context) ([]*types.ServiceInfo, error) {
	cmd := exec.CommandContext(ctx, "launchctl", "list")
	output, err := cmd.Output()
	if err != nil {
		return nil, types.NewServiceError("services", "list", err, "failed to list launchd services")
	}

	lines := strings.Split(string(output), "\n")
	services := make([]*types.ServiceInfo, 0)

	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue // Skip header and empty lines
		}

		fields := strings.Fields(line)
		if len(fields) >= 3 {
			pid := fields[0]
			serviceName := fields[2]

			var status types.ServiceStatus
			if pid != "-" {
				status = types.ServiceStatusRunning
			} else {
				status = types.ServiceStatusStopped
			}

			pidInt := 0
			if pid != "-" {
				if p, err := strconv.Atoi(pid); err == nil {
					pidInt = p
				}
			}

			service := &types.ServiceInfo{
				Name:        serviceName,
				DisplayName: serviceName,
				Status:      status,
				StartType:   types.StartTypeAutomatic, // Would need additional query
				PID:         pidInt,
				Memory:      0, // Would need additional query
				CPU:         0, // Would need additional query
				Uptime:      0, // Would need additional query
				Description: serviceName,
				Path:        "",
			}

			services = append(services, service)
		}
	}

	return services, nil
}

// ControlService controls a macOS service (launchd service)
func (d *DarwinServerManager) ControlService(ctx context.Context, serviceName string, action types.ServiceAction) error {
	var cmd *exec.Cmd

	switch action {
	case types.ServiceActionStart:
		cmd = exec.CommandContext(ctx, "launchctl", "start", serviceName)
	case types.ServiceActionStop:
		cmd = exec.CommandContext(ctx, "launchctl", "stop", serviceName)
	case types.ServiceActionRestart:
		// launchctl doesn't have restart, so stop then start
		if err := exec.CommandContext(ctx, "launchctl", "stop", serviceName).Run(); err != nil {
			return types.NewServiceError(serviceName, "stop", err, "failed to stop service for restart")
		}
		cmd = exec.CommandContext(ctx, "launchctl", "start", serviceName)
	default:
		return types.NewServiceError(serviceName, string(action), nil, "unsupported service action")
	}

	if err := cmd.Run(); err != nil {
		return types.NewServiceError(serviceName, string(action), err, "failed to control launchd service")
	}

	return nil
}

// GetActiveUsers returns active macOS user sessions
func (d *DarwinServerManager) GetActiveUsers(ctx context.Context) ([]*types.UserSession, error) {
	cmd := exec.CommandContext(ctx, "who")
	output, err := cmd.Output()
	if err != nil {
		return nil, types.NewServiceError("users", "query", err, "failed to query active users")
	}

	lines := strings.Split(string(output), "\n")
	users := make([]*types.UserSession, 0)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 3 {
			username := fields[0]
			terminal := fields[1]

			// Parse login time (simplified)
			loginTime := time.Now().Add(-time.Hour) // Mock login time

			user := &types.UserSession{
				Username:     username,
				SessionID:    terminal,
				Terminal:     terminal,
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

// GetServiceLogs returns macOS service logs (Console logs)
func (d *DarwinServerManager) GetServiceLogs(ctx context.Context, serviceName string, tail bool) (*types.LogStream, error) {
	var cmd *exec.Cmd

	if tail {
		cmd = exec.CommandContext(ctx, "log", "show", "--predicate", fmt.Sprintf("process == '%s'", serviceName), "--last", "10m")
	} else {
		cmd = exec.CommandContext(ctx, "log", "show", "--predicate", fmt.Sprintf("process == '%s'", serviceName))
	}

	output, err := cmd.Output()
	if err != nil {
		// Return empty log stream if no logs found
		return &types.LogStream{
			ServiceName: serviceName,
			Entries:     []types.LogEntry{},
			Following:   tail,
			StartTime:   time.Now(),
		}, nil
	}

	lines := strings.Split(string(output), "\n")
	entries := make([]types.LogEntry, 0)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Simple log parsing (would be more sophisticated in real implementation)
		entry := types.LogEntry{
			Timestamp: time.Now(), // Would parse actual timestamp
			Level:     "INFO",     // Would parse actual level
			Message:   line,
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

// ConfigureAlerts configures macOS-specific alert settings
func (d *DarwinServerManager) ConfigureAlerts(ctx context.Context, config *types.AlertConfig) error {
	// macOS-specific alert configuration would go here
	// This could involve launchd, notifications, etc.
	return nil
}

// BackupConfiguration backs up macOS server configuration
func (d *DarwinServerManager) BackupConfiguration(ctx context.Context, backupPath string) error {
	// macOS-specific configuration backup would go here
	// This could involve launchd plists, system preferences, etc.
	return nil
}
