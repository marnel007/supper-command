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

// LinuxServerManager manages Linux server operations using systemctl and system commands
type LinuxServerManager struct {
	*BaseServerManager
	useSystemd bool
}

// NewLinuxServerManager creates a new Linux server manager
func NewLinuxServerManager() *LinuxServerManager {
	manager := &LinuxServerManager{
		BaseServerManager: NewBaseServerManager(types.PlatformLinux),
	}

	// Check if systemd is available
	manager.useSystemd = manager.isSystemdAvailable()

	return manager
}

// isSystemdAvailable checks if systemd is available
func (l *LinuxServerManager) isSystemdAvailable() bool {
	cmd := exec.Command("which", "systemctl")
	return cmd.Run() == nil
}

// GetHealthStatus returns Linux server health status
func (l *LinuxServerManager) GetHealthStatus(ctx context.Context) (*types.HealthStatus, error) {
	components := make(map[string]types.ComponentHealth)

	// Get CPU health
	if err := l.getCPUHealth(ctx, components); err != nil {
		return nil, types.NewServiceError("health", "cpu", err, "failed to get CPU health")
	}

	// Get memory health
	if err := l.getMemoryHealth(ctx, components); err != nil {
		return nil, types.NewServiceError("health", "memory", err, "failed to get memory health")
	}

	// Get disk health
	if err := l.getDiskHealth(ctx, components); err != nil {
		return nil, types.NewServiceError("health", "disk", err, "failed to get disk health")
	}

	// Get network health
	if err := l.getNetworkHealth(ctx, components); err != nil {
		return nil, types.NewServiceError("health", "network", err, "failed to get network health")
	}

	// Calculate overall health
	overall := l.CalculateOverallHealth(components)

	// Get system uptime
	uptime, _ := l.getSystemUptime(ctx)

	// Get load average
	loadAverage, _ := l.getLoadAverage(ctx)

	// Generate alerts
	alerts := l.GenerateHealthAlerts(components)

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

// getCPUHealth gets CPU health information from /proc/stat
func (l *LinuxServerManager) getCPUHealth(ctx context.Context, components map[string]types.ComponentHealth) error {
	// Use top command to get CPU usage
	cmd := exec.CommandContext(ctx, "top", "-bn1")
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
		if strings.Contains(line, "%Cpu(s):") {
			// Parse line like: "%Cpu(s):  5.9 us,  2.9 sy,  0.0 ni, 91.2 id,  0.0 wa,  0.0 hi,  0.0 si,  0.0 st"
			parts := strings.Split(line, ",")
			if len(parts) >= 4 {
				idlePart := strings.TrimSpace(parts[3])
				if strings.Contains(idlePart, "id") {
					idleStr := strings.Fields(idlePart)[0]
					if idle, err := strconv.ParseFloat(idleStr, 64); err == nil {
						cpuUsage = 100.0 - idle
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

// getMemoryHealth gets memory health information from /proc/meminfo
func (l *LinuxServerManager) getMemoryHealth(ctx context.Context, components map[string]types.ComponentHealth) error {
	cmd := exec.CommandContext(ctx, "cat", "/proc/meminfo")
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
	memInfo := make(map[string]uint64)

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			key := strings.TrimSuffix(parts[0], ":")
			if value, err := strconv.ParseUint(parts[1], 10, 64); err == nil {
				memInfo[key] = value * 1024 // Convert from KB to bytes
			}
		}
	}

	total := memInfo["MemTotal"]
	available := memInfo["MemAvailable"]
	if available == 0 {
		// Fallback calculation for older systems
		free := memInfo["MemFree"]
		buffers := memInfo["Buffers"]
		cached := memInfo["Cached"]
		available = free + buffers + cached
	}

	used := total - available
	memUsage := float64(used) / float64(total) * 100

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

// getDiskHealth gets disk health information using df command
func (l *LinuxServerManager) getDiskHealth(ctx context.Context, components map[string]types.ComponentHealth) error {
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
func (l *LinuxServerManager) getNetworkHealth(ctx context.Context, components map[string]types.ComponentHealth) error {
	// Check network interfaces
	cmd := exec.CommandContext(ctx, "ip", "link", "show")
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
		if strings.Contains(line, "state UP") {
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

// getSystemUptime gets Linux system uptime
func (l *LinuxServerManager) getSystemUptime(ctx context.Context) (time.Duration, error) {
	cmd := exec.CommandContext(ctx, "cat", "/proc/uptime")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	fields := strings.Fields(string(output))
	if len(fields) < 1 {
		return 0, fmt.Errorf("invalid uptime format")
	}

	uptimeSeconds, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, err
	}

	uptime := time.Duration(uptimeSeconds) * time.Second
	return uptime, nil
}

// getLoadAverage gets Linux load average
func (l *LinuxServerManager) getLoadAverage(ctx context.Context) ([]float64, error) {
	cmd := exec.CommandContext(ctx, "cat", "/proc/loadavg")
	output, err := cmd.Output()
	if err != nil {
		return []float64{0, 0, 0}, err
	}

	fields := strings.Fields(string(output))
	if len(fields) < 3 {
		return []float64{0, 0, 0}, fmt.Errorf("invalid loadavg format")
	}

	loadAvg := make([]float64, 3)
	for i := 0; i < 3; i++ {
		if val, err := strconv.ParseFloat(fields[i], 64); err == nil {
			loadAvg[i] = val
		}
	}

	return loadAvg, nil
}

// ListServices returns Linux services information
func (l *LinuxServerManager) ListServices(ctx context.Context) ([]*types.ServiceInfo, error) {
	if l.useSystemd {
		return l.listSystemdServices(ctx)
	}
	return l.listSysVServices(ctx)
}

// listSystemdServices lists systemd services
func (l *LinuxServerManager) listSystemdServices(ctx context.Context) ([]*types.ServiceInfo, error) {
	cmd := exec.CommandContext(ctx, "systemctl", "list-units", "--type=service", "--no-pager", "--plain")
	output, err := cmd.Output()
	if err != nil {
		return nil, types.NewServiceError("services", "list", err, "failed to list systemd services")
	}

	lines := strings.Split(string(output), "\n")
	services := make([]*types.ServiceInfo, 0)

	for _, line := range lines {
		if strings.TrimSpace(line) == "" || strings.Contains(line, "UNIT") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 4 {
			unitName := fields[0]
			activeState := fields[2]
			subState := fields[3]

			// Extract service name (remove .service suffix)
			serviceName := strings.TrimSuffix(unitName, ".service")

			var status types.ServiceStatus
			if activeState == "active" && subState == "running" {
				status = types.ServiceStatusRunning
			} else if activeState == "inactive" {
				status = types.ServiceStatusStopped
			} else {
				status = types.ServiceStatusUnknown
			}

			service := &types.ServiceInfo{
				Name:        serviceName,
				DisplayName: serviceName,
				Status:      status,
				StartType:   types.StartTypeAutomatic, // Would need additional query
				PID:         0,                        // Would need additional query
				Memory:      0,                        // Would need additional query
				CPU:         0,                        // Would need additional query
				Uptime:      0,                        // Would need additional query
				Description: unitName,
				Path:        "",
			}

			services = append(services, service)
		}
	}

	return services, nil
}

// listSysVServices lists SysV init services (fallback)
func (l *LinuxServerManager) listSysVServices(ctx context.Context) ([]*types.ServiceInfo, error) {
	cmd := exec.CommandContext(ctx, "service", "--status-all")
	output, err := cmd.Output()
	if err != nil {
		return nil, types.NewServiceError("services", "list", err, "failed to list SysV services")
	}

	lines := strings.Split(string(output), "\n")
	services := make([]*types.ServiceInfo, 0)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse lines like: " [ + ]  service-name"
		if strings.Contains(line, "]") {
			parts := strings.Split(line, "]")
			if len(parts) >= 2 {
				statusPart := strings.TrimSpace(parts[0])
				serviceName := strings.TrimSpace(parts[1])

				var status types.ServiceStatus
				if strings.Contains(statusPart, "+") {
					status = types.ServiceStatusRunning
				} else if strings.Contains(statusPart, "-") {
					status = types.ServiceStatusStopped
				} else {
					status = types.ServiceStatusUnknown
				}

				service := &types.ServiceInfo{
					Name:        serviceName,
					DisplayName: serviceName,
					Status:      status,
					StartType:   types.StartTypeManual,
					PID:         0,
					Memory:      0,
					CPU:         0,
					Uptime:      0,
					Description: serviceName,
					Path:        "",
				}

				services = append(services, service)
			}
		}
	}

	return services, nil
}

// ControlService controls a Linux service
func (l *LinuxServerManager) ControlService(ctx context.Context, serviceName string, action types.ServiceAction) error {
	if l.useSystemd {
		return l.controlSystemdService(ctx, serviceName, action)
	}
	return l.controlSysVService(ctx, serviceName, action)
}

// controlSystemdService controls a systemd service
func (l *LinuxServerManager) controlSystemdService(ctx context.Context, serviceName string, action types.ServiceAction) error {
	var cmd *exec.Cmd

	switch action {
	case types.ServiceActionStart:
		cmd = exec.CommandContext(ctx, "systemctl", "start", serviceName)
	case types.ServiceActionStop:
		cmd = exec.CommandContext(ctx, "systemctl", "stop", serviceName)
	case types.ServiceActionRestart:
		cmd = exec.CommandContext(ctx, "systemctl", "restart", serviceName)
	case types.ServiceActionReload:
		cmd = exec.CommandContext(ctx, "systemctl", "reload", serviceName)
	default:
		return types.NewServiceError(serviceName, string(action), nil, "unsupported service action")
	}

	if err := cmd.Run(); err != nil {
		return types.NewServiceError(serviceName, string(action), err, "failed to control systemd service")
	}

	return nil
}

// controlSysVService controls a SysV init service
func (l *LinuxServerManager) controlSysVService(ctx context.Context, serviceName string, action types.ServiceAction) error {
	var cmd *exec.Cmd

	switch action {
	case types.ServiceActionStart:
		cmd = exec.CommandContext(ctx, "service", serviceName, "start")
	case types.ServiceActionStop:
		cmd = exec.CommandContext(ctx, "service", serviceName, "stop")
	case types.ServiceActionRestart:
		cmd = exec.CommandContext(ctx, "service", serviceName, "restart")
	case types.ServiceActionReload:
		cmd = exec.CommandContext(ctx, "service", serviceName, "reload")
	default:
		return types.NewServiceError(serviceName, string(action), nil, "unsupported service action")
	}

	if err := cmd.Run(); err != nil {
		return types.NewServiceError(serviceName, string(action), err, "failed to control SysV service")
	}

	return nil
}

// GetActiveUsers returns active Linux user sessions
func (l *LinuxServerManager) GetActiveUsers(ctx context.Context) ([]*types.UserSession, error) {
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

// GetServiceLogs returns Linux service logs
func (l *LinuxServerManager) GetServiceLogs(ctx context.Context, serviceName string, tail bool) (*types.LogStream, error) {
	var cmd *exec.Cmd

	if l.useSystemd {
		if tail {
			cmd = exec.CommandContext(ctx, "journalctl", "-u", serviceName, "-n", "10", "--no-pager")
		} else {
			cmd = exec.CommandContext(ctx, "journalctl", "-u", serviceName, "--no-pager")
		}
	} else {
		// Fallback to syslog
		cmd = exec.CommandContext(ctx, "grep", serviceName, "/var/log/syslog")
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

// ConfigureAlerts configures Linux-specific alert settings
func (l *LinuxServerManager) ConfigureAlerts(ctx context.Context, config *types.AlertConfig) error {
	// Linux-specific alert configuration would go here
	// This could involve systemd, cron, or other monitoring tools
	return nil
}

// BackupConfiguration backs up Linux server configuration
func (l *LinuxServerManager) BackupConfiguration(ctx context.Context, backupPath string) error {
	// Linux-specific configuration backup would go here
	// This could involve /etc, systemd units, cron jobs, etc.
	return nil
}
