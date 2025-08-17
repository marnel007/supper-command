package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"suppercommand/internal/types"
)

// AlertManager manages server health alerts and notifications
type AlertManager struct {
	config  *types.AlertConfig
	alerts  []types.HealthAlert
	history []types.HealthAlert
}

// NewAlertManager creates a new alert manager
func NewAlertManager() *AlertManager {
	return &AlertManager{
		config:  getDefaultAlertConfig(),
		alerts:  make([]types.HealthAlert, 0),
		history: make([]types.HealthAlert, 0),
	}
}

// getDefaultAlertConfig returns default alert configuration
func getDefaultAlertConfig() *types.AlertConfig {
	return &types.AlertConfig{
		Enabled: true,
		Thresholds: map[string]types.AlertThreshold{
			"CPU": {
				Metric:   "CPU",
				Warning:  70.0,
				Critical: 90.0,
				Unit:     "%",
				Enabled:  true,
			},
			"Memory": {
				Metric:   "Memory",
				Warning:  80.0,
				Critical: 95.0,
				Unit:     "%",
				Enabled:  true,
			},
			"Disk": {
				Metric:   "Disk",
				Warning:  85.0,
				Critical: 95.0,
				Unit:     "%",
				Enabled:  true,
			},
			"LoadAverage": {
				Metric:   "LoadAverage",
				Warning:  2.0,
				Critical: 5.0,
				Unit:     "load",
				Enabled:  true,
			},
		},
		Notifications: []types.NotificationConfig{
			{
				Type:    "log",
				Target:  "/var/log/supershell-alerts.log",
				Enabled: true,
				Settings: map[string]string{
					"format": "json",
				},
			},
		},
		CheckInterval: 30 * time.Second,
	}
}

// ProcessHealthStatus processes health status and generates alerts
func (a *AlertManager) ProcessHealthStatus(ctx context.Context, health *types.HealthStatus) error {
	if !a.config.Enabled {
		return nil
	}

	newAlerts := make([]types.HealthAlert, 0)

	// Check each component against thresholds
	for componentName, component := range health.Components {
		threshold, exists := a.config.Thresholds[componentName]
		if !exists || !threshold.Enabled {
			continue
		}

		// Check if alert should be generated
		if component.Status == types.HealthLevelCritical || component.Status == types.HealthLevelWarning {
			alert := types.HealthAlert{
				ID:           fmt.Sprintf("%s_%d", componentName, time.Now().Unix()),
				Level:        component.Status,
				Component:    componentName,
				Message:      component.Message,
				Timestamp:    time.Now(),
				Acknowledged: false,
				Value:        component.Value,
				Threshold:    component.Threshold,
			}

			// Check if this is a new alert (not already active)
			if !a.isAlertActive(alert) {
				newAlerts = append(newAlerts, alert)
				a.alerts = append(a.alerts, alert)
			}
		}
	}

	// Send notifications for new alerts
	for _, alert := range newAlerts {
		if err := a.sendNotifications(ctx, alert); err != nil {
			// Log error but don't fail the entire operation
			continue
		}
	}

	// Clean up old alerts
	a.cleanupOldAlerts()

	return nil
}

// isAlertActive checks if an alert is already active
func (a *AlertManager) isAlertActive(newAlert types.HealthAlert) bool {
	for _, existingAlert := range a.alerts {
		if existingAlert.Component == newAlert.Component &&
			existingAlert.Level == newAlert.Level &&
			!existingAlert.Acknowledged &&
			time.Since(existingAlert.Timestamp) < 5*time.Minute {
			return true
		}
	}
	return false
}

// sendNotifications sends notifications for an alert
func (a *AlertManager) sendNotifications(ctx context.Context, alert types.HealthAlert) error {
	for _, notification := range a.config.Notifications {
		if !notification.Enabled {
			continue
		}

		switch notification.Type {
		case "log":
			if err := a.sendLogNotification(alert, notification); err != nil {
				return err
			}
		case "webhook":
			if err := a.sendWebhookNotification(ctx, alert, notification); err != nil {
				return err
			}
		case "email":
			if err := a.sendEmailNotification(ctx, alert, notification); err != nil {
				return err
			}
		}
	}

	return nil
}

// sendLogNotification sends a log notification
func (a *AlertManager) sendLogNotification(alert types.HealthAlert, notification types.NotificationConfig) error {
	logFile := notification.Target
	if logFile == "" {
		logFile = "/var/log/supershell-alerts.log"
	}

	// Create log entry
	logEntry := map[string]interface{}{
		"timestamp": alert.Timestamp.Format(time.RFC3339),
		"level":     string(alert.Level),
		"component": alert.Component,
		"message":   alert.Message,
		"value":     alert.Value,
		"threshold": alert.Threshold,
		"alert_id":  alert.ID,
	}

	// Format based on settings
	format := notification.Settings["format"]
	var logLine string

	if format == "json" {
		data, err := json.Marshal(logEntry)
		if err != nil {
			return err
		}
		logLine = string(data) + "\n"
	} else {
		// Default text format
		logLine = fmt.Sprintf("[%s] %s %s: %s (value: %.2f, threshold: %.2f)\n",
			alert.Timestamp.Format("2006-01-02 15:04:05"),
			strings.ToUpper(string(alert.Level)),
			alert.Component,
			alert.Message,
			alert.Value,
			alert.Threshold)
	}

	// Append to log file
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(logLine)
	return err
}

// sendWebhookNotification sends a webhook notification (placeholder)
func (a *AlertManager) sendWebhookNotification(ctx context.Context, alert types.HealthAlert, notification types.NotificationConfig) error {
	// Webhook implementation would go here
	// This would make HTTP POST request to the webhook URL
	return nil
}

// sendEmailNotification sends an email notification (placeholder)
func (a *AlertManager) sendEmailNotification(ctx context.Context, alert types.HealthAlert, notification types.NotificationConfig) error {
	// Email implementation would go here
	// This would use SMTP to send email notifications
	return nil
}

// cleanupOldAlerts removes old alerts from the active list
func (a *AlertManager) cleanupOldAlerts() {
	cutoff := time.Now().Add(-24 * time.Hour) // Keep alerts for 24 hours

	activeAlerts := make([]types.HealthAlert, 0)
	for _, alert := range a.alerts {
		if alert.Timestamp.After(cutoff) {
			activeAlerts = append(activeAlerts, alert)
		} else {
			// Move to history
			a.history = append(a.history, alert)
		}
	}

	a.alerts = activeAlerts

	// Keep only last 1000 historical alerts
	if len(a.history) > 1000 {
		a.history = a.history[len(a.history)-1000:]
	}
}

// AddAlert adds a new alert to the active alerts list
func (a *AlertManager) AddAlert(alert *types.HealthAlert) {
	if !a.isAlertActive(*alert) {
		a.alerts = append(a.alerts, *alert)
		a.history = append(a.history, *alert)
	}
}

// GetActiveAlerts returns currently active alerts
func (a *AlertManager) GetActiveAlerts() []types.HealthAlert {
	return a.alerts
}

// GetAlertHistory returns historical alerts
func (a *AlertManager) GetAlertHistory() []types.HealthAlert {
	return a.history
}

// AcknowledgeAlert acknowledges an alert
func (a *AlertManager) AcknowledgeAlert(alertID string) error {
	for i, alert := range a.alerts {
		if alert.ID == alertID {
			a.alerts[i].Acknowledged = true
			return nil
		}
	}

	return fmt.Errorf("alert not found: %s", alertID)
}

// UpdateConfig updates the alert configuration
func (a *AlertManager) UpdateConfig(config *types.AlertConfig) error {
	a.config = config
	return nil
}

// GetConfig returns the current alert configuration
func (a *AlertManager) GetConfig() *types.AlertConfig {
	return a.config
}

// SaveConfig saves the alert configuration to a file
func (a *AlertManager) SaveConfig(filepath string) error {
	data, err := json.MarshalIndent(a.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath, data, 0644)
}

// LoadConfig loads alert configuration from a file
func (a *AlertManager) LoadConfig(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, a.config)
}

// GetAlertStats returns statistics about alerts
func (a *AlertManager) GetAlertStats() map[string]interface{} {
	activeCount := len(a.alerts)
	historyCount := len(a.history)

	// Count by level
	criticalCount := 0
	warningCount := 0
	acknowledgedCount := 0

	for _, alert := range a.alerts {
		switch alert.Level {
		case types.HealthLevelCritical:
			criticalCount++
		case types.HealthLevelWarning:
			warningCount++
		}

		if alert.Acknowledged {
			acknowledgedCount++
		}
	}

	return map[string]interface{}{
		"active_alerts":       activeCount,
		"historical_alerts":   historyCount,
		"critical_alerts":     criticalCount,
		"warning_alerts":      warningCount,
		"acknowledged_alerts": acknowledgedCount,
		"config_enabled":      a.config.Enabled,
		"check_interval":      a.config.CheckInterval.String(),
	}
}
