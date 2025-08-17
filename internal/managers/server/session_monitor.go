package server

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"suppercommand/internal/types"
)

// SessionMonitor monitors user sessions and tracks changes
type SessionMonitor struct {
	manager         types.ServerManager
	sessions        map[string]*types.UserSession
	sessionHistory  []types.SessionEvent
	mutex           sync.RWMutex
	alertManager    *AlertManager
	monitorInterval time.Duration
	maxHistory      int
}

// NewSessionMonitor creates a new session monitor
func NewSessionMonitor(manager types.ServerManager, alertManager *AlertManager) *SessionMonitor {
	return &SessionMonitor{
		manager:         manager,
		sessions:        make(map[string]*types.UserSession),
		sessionHistory:  make([]types.SessionEvent, 0),
		alertManager:    alertManager,
		monitorInterval: 30 * time.Second,
		maxHistory:      1000,
	}
}

// StartMonitoring starts continuous session monitoring
func (sm *SessionMonitor) StartMonitoring(ctx context.Context) error {
	ticker := time.NewTicker(sm.monitorInterval)
	defer ticker.Stop()

	// Initial scan
	if err := sm.scanSessions(ctx); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := sm.scanSessions(ctx); err != nil {
				// Log error but continue monitoring
				continue
			}
		}
	}
}

// scanSessions scans for session changes
func (sm *SessionMonitor) scanSessions(ctx context.Context) error {
	// Get current sessions
	currentSessions, err := sm.manager.GetActiveUsers(ctx)
	if err != nil {
		return err
	}

	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Create map of current sessions for easy lookup
	currentMap := make(map[string]*types.UserSession)
	for _, session := range currentSessions {
		key := sm.getSessionKey(session)
		currentMap[key] = session
	}

	// Check for new sessions
	for key, session := range currentMap {
		if _, exists := sm.sessions[key]; !exists {
			sm.recordSessionEvent("login", session, map[string]interface{}{
				"remote_host": session.RemoteHost,
				"terminal":    session.Terminal,
			})
		}
	}

	// Check for ended sessions
	for key, session := range sm.sessions {
		if _, exists := currentMap[key]; !exists {
			sm.recordSessionEvent("logout", session, map[string]interface{}{
				"duration": time.Since(session.LoginTime).String(),
			})
		}
	}

	// Update sessions map
	sm.sessions = currentMap

	// Check for suspicious activity
	sm.checkSuspiciousActivity(currentSessions)

	return nil
}

// recordSessionEvent records a session event
func (sm *SessionMonitor) recordSessionEvent(eventType string, session *types.UserSession, details map[string]interface{}) {
	event := types.SessionEvent{
		Type:      eventType,
		Username:  session.Username,
		SessionID: session.SessionID,
		Timestamp: time.Now(),
		Details:   details,
	}

	sm.sessionHistory = append(sm.sessionHistory, event)

	// Keep history within limits
	if len(sm.sessionHistory) > sm.maxHistory {
		sm.sessionHistory = sm.sessionHistory[len(sm.sessionHistory)-sm.maxHistory:]
	}
}

// checkSuspiciousActivity checks for suspicious session activity
func (sm *SessionMonitor) checkSuspiciousActivity(sessions []*types.UserSession) {
	// Check for multiple concurrent sessions from same user
	userSessions := make(map[string]int)
	for _, session := range sessions {
		userSessions[session.Username]++
	}

	for username, count := range userSessions {
		if count > 3 { // More than 3 concurrent sessions
			sm.generateSecurityAlert("multiple_sessions", username, map[string]interface{}{
				"session_count": count,
				"threshold":     3,
			})
		}
	}

	// Check for sessions from unusual locations (if remote host info available)
	for _, session := range sessions {
		if session.RemoteHost != "" && sm.isUnusualLocation(session.RemoteHost) {
			sm.generateSecurityAlert("unusual_location", session.Username, map[string]interface{}{
				"remote_host": session.RemoteHost,
				"session_id":  session.SessionID,
			})
		}
	}

	// Check for long idle sessions
	for _, session := range sessions {
		if session.IdleTime > 4*time.Hour { // Idle for more than 4 hours
			sm.generateSecurityAlert("long_idle_session", session.Username, map[string]interface{}{
				"idle_time":  session.IdleTime.String(),
				"session_id": session.SessionID,
			})
		}
	}
}

// generateSecurityAlert generates a security-related alert
func (sm *SessionMonitor) generateSecurityAlert(alertType, username string, details map[string]interface{}) {
	// Create a health alert for security issues
	alert := types.HealthAlert{
		ID:           alertType + "_" + username + "_" + time.Now().Format("20060102150405"),
		Level:        types.HealthLevelWarning,
		Component:    "Security",
		Message:      sm.formatSecurityMessage(alertType, username, details),
		Timestamp:    time.Now(),
		Acknowledged: false,
		Value:        1.0,
		Threshold:    1.0,
	}

	// Add to alert manager if available
	if sm.alertManager != nil {
		sm.alertManager.AddAlert(&alert)
	}
	if sm.alertManager != nil {
		// In a real implementation, we'd add this to the alert manager
		// For now, we'll just record it
	}
}

// formatSecurityMessage formats security alert messages
func (sm *SessionMonitor) formatSecurityMessage(alertType, username string, details map[string]interface{}) string {
	switch alertType {
	case "multiple_sessions":
		return fmt.Sprintf("User %s has %v concurrent sessions (threshold: %v)",
			username, details["session_count"], details["threshold"])
	case "unusual_location":
		return fmt.Sprintf("User %s logged in from unusual location: %v",
			username, details["remote_host"])
	case "long_idle_session":
		return fmt.Sprintf("User %s has been idle for %v",
			username, details["idle_time"])
	default:
		return fmt.Sprintf("Security alert for user %s: %s", username, alertType)
	}
}

// isUnusualLocation checks if a remote host is from an unusual location
func (sm *SessionMonitor) isUnusualLocation(remoteHost string) bool {
	// In a real implementation, this would check against:
	// - Known good IP ranges
	// - Geolocation data
	// - Historical login patterns
	// For now, we'll use simple heuristics

	// Check for private IP ranges (these are usually safe)
	if strings.HasPrefix(remoteHost, "192.168.") ||
		strings.HasPrefix(remoteHost, "10.") ||
		strings.HasPrefix(remoteHost, "172.") {
		return false
	}

	// For demo purposes, consider any external IP as potentially unusual
	return true
}

// getSessionKey creates a unique key for a session
func (sm *SessionMonitor) getSessionKey(session *types.UserSession) string {
	return session.Username + "_" + session.SessionID
}

// GetSessionHistory returns session history
func (sm *SessionMonitor) GetSessionHistory() []types.SessionEvent {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	// Return a copy to avoid race conditions
	history := make([]types.SessionEvent, len(sm.sessionHistory))
	copy(history, sm.sessionHistory)
	return history
}

// GetSessionStats returns session statistics
func (sm *SessionMonitor) GetSessionStats() map[string]interface{} {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	// Count events by type
	eventCounts := make(map[string]int)
	for _, event := range sm.sessionHistory {
		eventCounts[event.Type]++
	}

	// Count unique users
	uniqueUsers := make(map[string]bool)
	for _, session := range sm.sessions {
		uniqueUsers[session.Username] = true
	}

	return map[string]interface{}{
		"active_sessions":  len(sm.sessions),
		"unique_users":     len(uniqueUsers),
		"total_events":     len(sm.sessionHistory),
		"login_events":     eventCounts["login"],
		"logout_events":    eventCounts["logout"],
		"monitor_interval": sm.monitorInterval.String(),
		"history_size":     sm.maxHistory,
	}
}

// SetMonitorInterval sets the monitoring interval
func (sm *SessionMonitor) SetMonitorInterval(interval time.Duration) {
	sm.monitorInterval = interval
}

// ClearHistory clears session history
func (sm *SessionMonitor) ClearHistory() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.sessionHistory = make([]types.SessionEvent, 0)
}
