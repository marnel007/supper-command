package remote

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"suppercommand/internal/types"
)

// ClusterMonitor monitors the health and status of remote server clusters
type ClusterMonitor struct {
	remoteManager   types.RemoteManager
	clusterManager  *ClusterManager
	monitoringTasks map[string]*MonitoringTask
	alerts          []MonitoringAlert
	metrics         map[string]*ServerMetrics
	mutex           sync.RWMutex
	running         bool
	stopChan        chan struct{}
}

// MonitoringTask represents a monitoring task
type MonitoringTask struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Servers     []string          `json:"servers"`
	Checks      []HealthCheck     `json:"checks"`
	Interval    time.Duration     `json:"interval"`
	Timeout     time.Duration     `json:"timeout"`
	Enabled     bool              `json:"enabled"`
	Tags        map[string]string `json:"tags"`
	CreatedAt   time.Time         `json:"created_at"`
	LastRun     time.Time         `json:"last_run"`
	NextRun     time.Time         `json:"next_run"`
}

// HealthCheck represents a health check command
type HealthCheck struct {
	Name           string            `json:"name"`
	Command        string            `json:"command"`
	ExpectedExit   int               `json:"expected_exit"`
	ExpectedOutput string            `json:"expected_output,omitempty"`
	Timeout        time.Duration     `json:"timeout"`
	Critical       bool              `json:"critical"`
	Tags           map[string]string `json:"tags"`
}

// ServerMetrics represents metrics for a server
type ServerMetrics struct {
	ServerName      string                 `json:"server_name"`
	LastUpdate      time.Time              `json:"last_update"`
	Status          types.ServerStatus     `json:"status"`
	ResponseTime    time.Duration          `json:"response_time"`
	Uptime          time.Duration          `json:"uptime"`
	LoadAverage     []float64              `json:"load_average"`
	CPUUsage        float64                `json:"cpu_usage"`
	MemoryUsage     float64                `json:"memory_usage"`
	DiskUsage       float64                `json:"disk_usage"`
	NetworkRx       int64                  `json:"network_rx"`
	NetworkTx       int64                  `json:"network_tx"`
	ProcessCount    int                    `json:"process_count"`
	ConnectionCount int                    `json:"connection_count"`
	CustomMetrics   map[string]interface{} `json:"custom_metrics"`
}

// MonitoringAlert represents a monitoring alert
type MonitoringAlert struct {
	ID         string                 `json:"id"`
	Level      types.HealthLevel      `json:"level"`
	ServerName string                 `json:"server_name"`
	CheckName  string                 `json:"check_name"`
	Message    string                 `json:"message"`
	Timestamp  time.Time              `json:"timestamp"`
	Resolved   bool                   `json:"resolved"`
	ResolvedAt time.Time              `json:"resolved_at,omitempty"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// NewClusterMonitor creates a new cluster monitor
func NewClusterMonitor(remoteManager types.RemoteManager, clusterManager *ClusterManager) *ClusterMonitor {
	return &ClusterMonitor{
		remoteManager:   remoteManager,
		clusterManager:  clusterManager,
		monitoringTasks: make(map[string]*MonitoringTask),
		alerts:          make([]MonitoringAlert, 0),
		metrics:         make(map[string]*ServerMetrics),
		stopChan:        make(chan struct{}),
	}
}

// StartMonitoring starts the monitoring system
func (cm *ClusterMonitor) StartMonitoring(ctx context.Context) error {
	cm.mutex.Lock()
	if cm.running {
		cm.mutex.Unlock()
		return fmt.Errorf("monitoring is already running")
	}
	cm.running = true
	cm.mutex.Unlock()

	// Start monitoring goroutine
	go cm.monitoringLoop(ctx)

	return nil
}

// StopMonitoring stops the monitoring system
func (cm *ClusterMonitor) StopMonitoring() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if cm.running {
		close(cm.stopChan)
		cm.running = false
	}
}

// monitoringLoop is the main monitoring loop
func (cm *ClusterMonitor) monitoringLoop(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second) // Check every 10 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-cm.stopChan:
			return
		case <-ticker.C:
			cm.runScheduledTasks(ctx)
		}
	}
}

// runScheduledTasks runs monitoring tasks that are due
func (cm *ClusterMonitor) runScheduledTasks(ctx context.Context) {
	cm.mutex.RLock()
	tasks := make([]*MonitoringTask, 0)
	now := time.Now()

	for _, task := range cm.monitoringTasks {
		if task.Enabled && (task.NextRun.IsZero() || now.After(task.NextRun)) {
			tasks = append(tasks, task)
		}
	}
	cm.mutex.RUnlock()

	// Run tasks
	for _, task := range tasks {
		go cm.runMonitoringTask(ctx, task)
	}
}

// runMonitoringTask runs a single monitoring task
func (cm *ClusterMonitor) runMonitoringTask(ctx context.Context, task *MonitoringTask) {
	startTime := time.Now()

	// Update task timing
	cm.mutex.Lock()
	task.LastRun = startTime
	task.NextRun = startTime.Add(task.Interval)
	cm.mutex.Unlock()

	// Create task context with timeout
	taskCtx, cancel := context.WithTimeout(ctx, task.Timeout)
	defer cancel()

	// Run checks on each server
	var wg sync.WaitGroup
	for _, serverName := range task.Servers {
		wg.Add(1)
		go func(server string) {
			defer wg.Done()
			cm.runServerChecks(taskCtx, task, server)
		}(serverName)
	}

	wg.Wait()
}

// runServerChecks runs health checks on a single server
func (cm *ClusterMonitor) runServerChecks(ctx context.Context, task *MonitoringTask, serverName string) {
	startTime := time.Now()
	_ = startTime // Mark as used to avoid compiler warning

	// Initialize or update server metrics
	cm.mutex.Lock()
	if _, exists := cm.metrics[serverName]; !exists {
		cm.metrics[serverName] = &ServerMetrics{
			ServerName:    serverName,
			CustomMetrics: make(map[string]interface{}),
		}
	}
	metrics := cm.metrics[serverName]
	cm.mutex.Unlock()

	// Test basic connectivity
	connectResult, err := cm.remoteManager.ExecuteCommand(ctx, serverName, "echo 'connectivity_test'")
	if err != nil || connectResult.ExitCode != 0 {
		metrics.Status = types.ServerStatusOffline
		metrics.LastUpdate = time.Now()
		cm.generateAlert(types.HealthLevelCritical, serverName, "connectivity",
			"Server is not responding", map[string]interface{}{
				"error": err,
			})
		return
	}

	metrics.Status = types.ServerStatusOnline
	metrics.ResponseTime = connectResult.Duration

	// Run each health check
	for _, check := range task.Checks {
		cm.runHealthCheck(ctx, serverName, &check, metrics)
	}

	// Collect system metrics
	cm.collectSystemMetrics(ctx, serverName, metrics)

	metrics.LastUpdate = time.Now()
}

// runHealthCheck runs a single health check
func (cm *ClusterMonitor) runHealthCheck(ctx context.Context, serverName string, check *HealthCheck, metrics *ServerMetrics) {
	checkCtx, cancel := context.WithTimeout(ctx, check.Timeout)
	defer cancel()

	result, err := cm.remoteManager.ExecuteCommand(checkCtx, serverName, check.Command)
	if err != nil {
		if check.Critical {
			cm.generateAlert(types.HealthLevelCritical, serverName, check.Name,
				fmt.Sprintf("Health check failed: %v", err), map[string]interface{}{
					"command": check.Command,
					"error":   err.Error(),
				})
		}
		return
	}

	// Check exit code
	if result.ExitCode != check.ExpectedExit {
		level := types.HealthLevelWarning
		if check.Critical {
			level = types.HealthLevelCritical
		}
		cm.generateAlert(level, serverName, check.Name,
			fmt.Sprintf("Health check exit code mismatch: expected %d, got %d",
				check.ExpectedExit, result.ExitCode), map[string]interface{}{
				"command":       check.Command,
				"expected_exit": check.ExpectedExit,
				"actual_exit":   result.ExitCode,
				"output":        result.Output,
			})
		return
	}

	// Check output if specified
	if check.ExpectedOutput != "" && !strings.Contains(result.Output, check.ExpectedOutput) {
		level := types.HealthLevelWarning
		if check.Critical {
			level = types.HealthLevelCritical
		}
		cm.generateAlert(level, serverName, check.Name,
			"Health check output mismatch", map[string]interface{}{
				"command":         check.Command,
				"expected_output": check.ExpectedOutput,
				"actual_output":   result.Output,
			})
		return
	}

	// Store custom metrics if the check returns structured data
	if strings.HasPrefix(result.Output, "METRIC:") {
		cm.parseCustomMetrics(result.Output, metrics)
	}
}

// collectSystemMetrics collects standard system metrics
func (cm *ClusterMonitor) collectSystemMetrics(ctx context.Context, serverName string, metrics *ServerMetrics) {
	// Collect load average
	if result, err := cm.remoteManager.ExecuteCommand(ctx, serverName, "cat /proc/loadavg"); err == nil {
		if loadAvg := parseLoadAverage(result.Output); len(loadAvg) >= 3 {
			metrics.LoadAverage = loadAvg[:3]
		}
	}

	// Collect CPU usage
	if result, err := cm.remoteManager.ExecuteCommand(ctx, serverName,
		"top -bn1 | grep 'Cpu(s)' | awk '{print $2}' | cut -d'%' -f1"); err == nil {
		if cpuUsage, err := parseFloat(strings.TrimSpace(result.Output)); err == nil {
			metrics.CPUUsage = 100.0 - cpuUsage // Convert idle to usage
		}
	}

	// Collect memory usage
	if result, err := cm.remoteManager.ExecuteCommand(ctx, serverName,
		"free | grep Mem | awk '{printf \"%.1f\", $3/$2 * 100.0}'"); err == nil {
		if memUsage, err := parseFloat(strings.TrimSpace(result.Output)); err == nil {
			metrics.MemoryUsage = memUsage
		}
	}

	// Collect disk usage
	if result, err := cm.remoteManager.ExecuteCommand(ctx, serverName,
		"df / | tail -1 | awk '{print $5}' | cut -d'%' -f1"); err == nil {
		if diskUsage, err := parseFloat(strings.TrimSpace(result.Output)); err == nil {
			metrics.DiskUsage = diskUsage
		}
	}

	// Collect process count
	if result, err := cm.remoteManager.ExecuteCommand(ctx, serverName, "ps aux | wc -l"); err == nil {
		if processCount, err := parseInt(strings.TrimSpace(result.Output)); err == nil {
			metrics.ProcessCount = processCount
		}
	}

	// Collect uptime
	if result, err := cm.remoteManager.ExecuteCommand(ctx, serverName, "cat /proc/uptime"); err == nil {
		if uptime := parseUptime(result.Output); uptime > 0 {
			metrics.Uptime = uptime
		}
	}
}

// parseCustomMetrics parses custom metrics from command output
func (cm *ClusterMonitor) parseCustomMetrics(output string, metrics *ServerMetrics) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "METRIC:") {
			parts := strings.SplitN(line[7:], "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				// Try to parse as number
				if floatVal, err := parseFloat(value); err == nil {
					metrics.CustomMetrics[key] = floatVal
				} else {
					metrics.CustomMetrics[key] = value
				}
			}
		}
	}
}

// generateAlert generates a monitoring alert
func (cm *ClusterMonitor) generateAlert(level types.HealthLevel, serverName, checkName, message string, metadata map[string]interface{}) {
	alert := MonitoringAlert{
		ID:         fmt.Sprintf("%s_%s_%d", serverName, checkName, time.Now().Unix()),
		Level:      level,
		ServerName: serverName,
		CheckName:  checkName,
		Message:    message,
		Timestamp:  time.Now(),
		Resolved:   false,
		Metadata:   metadata,
	}

	cm.mutex.Lock()
	cm.alerts = append(cm.alerts, alert)

	// Keep only last 1000 alerts
	if len(cm.alerts) > 1000 {
		cm.alerts = cm.alerts[len(cm.alerts)-1000:]
	}
	cm.mutex.Unlock()
}

// CreateMonitoringTask creates a new monitoring task
func (cm *ClusterMonitor) CreateMonitoringTask(task *MonitoringTask) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if task.Name == "" {
		return fmt.Errorf("task name cannot be empty")
	}

	if _, exists := cm.monitoringTasks[task.Name]; exists {
		return fmt.Errorf("monitoring task '%s' already exists", task.Name)
	}

	if len(task.Servers) == 0 {
		return fmt.Errorf("task must have at least one server")
	}

	if len(task.Checks) == 0 {
		return fmt.Errorf("task must have at least one health check")
	}

	task.CreatedAt = time.Now()
	task.NextRun = time.Now() // Run immediately
	cm.monitoringTasks[task.Name] = task

	return nil
}

// GetMonitoringTasks returns all monitoring tasks
func (cm *ClusterMonitor) GetMonitoringTasks() []*MonitoringTask {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	tasks := make([]*MonitoringTask, 0, len(cm.monitoringTasks))
	for _, task := range cm.monitoringTasks {
		tasks = append(tasks, task)
	}

	return tasks
}

// GetServerMetrics returns metrics for all servers
func (cm *ClusterMonitor) GetServerMetrics() map[string]*ServerMetrics {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	// Return a copy
	metrics := make(map[string]*ServerMetrics)
	for k, v := range cm.metrics {
		metrics[k] = v
	}

	return metrics
}

// GetAlerts returns monitoring alerts
func (cm *ClusterMonitor) GetAlerts(resolved bool) []MonitoringAlert {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	alerts := make([]MonitoringAlert, 0)
	for _, alert := range cm.alerts {
		if alert.Resolved == resolved {
			alerts = append(alerts, alert)
		}
	}

	return alerts
}

// ResolveAlert resolves a monitoring alert
func (cm *ClusterMonitor) ResolveAlert(alertID string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	for i, alert := range cm.alerts {
		if alert.ID == alertID {
			cm.alerts[i].Resolved = true
			cm.alerts[i].ResolvedAt = time.Now()
			return nil
		}
	}

	return fmt.Errorf("alert '%s' not found", alertID)
}

// Helper functions for parsing metrics
func parseLoadAverage(output string) []float64 {
	fields := strings.Fields(output)
	if len(fields) < 3 {
		return nil
	}

	loadAvg := make([]float64, 3)
	for i := 0; i < 3; i++ {
		if val, err := parseFloat(fields[i]); err == nil {
			loadAvg[i] = val
		}
	}

	return loadAvg
}

func parseInt(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}

func parseUptime(output string) time.Duration {
	fields := strings.Fields(output)
	if len(fields) < 1 {
		return 0
	}

	if seconds, err := parseFloat(fields[0]); err == nil {
		return time.Duration(seconds) * time.Second
	}

	return 0
}
