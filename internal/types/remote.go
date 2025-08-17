package types

import (
	"context"
	"fmt"
	"time"
)

// RemoteManager defines the interface for remote server management operations
type RemoteManager interface {
	AddServer(ctx context.Context, config *ServerConfig) error
	RemoveServer(ctx context.Context, serverID string) error
	UpdateServer(ctx context.Context, serverID string, config *ServerConfig) error
	ListServers(ctx context.Context) ([]*ServerInfo, error)
	GetServer(ctx context.Context, serverID string) (*ServerInfo, error)
	ExecuteCommand(ctx context.Context, serverID string, command string) (*RemoteResult, error)
	ExecuteScript(ctx context.Context, serverID string, script string) (*RemoteResult, error)
	ExecuteCommandOnCluster(ctx context.Context, clusterID string, command string) ([]*RemoteResult, error)
	UploadFile(ctx context.Context, serverID string, localPath string, remotePath string) error
	DownloadFile(ctx context.Context, serverID string, remotePath string, localPath string) error
	TestConnection(ctx context.Context, config *ServerConfig) error
	CheckServerHealth(ctx context.Context, serverID string) (*HealthStatus, error)
	GetServerMetrics(ctx context.Context, serverID string) (*PerformanceMetrics, error)
	GetClusterStatus(ctx context.Context) (*ClusterStatus, error)
	SyncConfiguration(ctx context.Context, serverIDs []string) error
	MonitorCluster(ctx context.Context, realtime bool) (*ClusterMonitor, error)
}

// ServerConfig contains configuration for a remote server
type ServerConfig struct {
	ID         string        `json:"id"`
	Name       string        `json:"name"`
	Host       string        `json:"host"`
	Port       int           `json:"port"`
	Username   string        `json:"username"`
	AuthMethod AuthMethod    `json:"auth_method"`
	KeyPath    string        `json:"key_path,omitempty"`
	Password   string        `json:"password,omitempty"` // Should be encrypted
	Tags       []string      `json:"tags"`
	Created    time.Time     `json:"created"`
	LastSeen   time.Time     `json:"last_seen"`
	Enabled    bool          `json:"enabled"`
	Timeout    time.Duration `json:"timeout"`
}

// ServerInfo contains information about a remote server
type ServerInfo struct {
	Config      *ServerConfig       `json:"config"`
	Status      ServerStatus        `json:"status"`
	Health      *HealthStatus       `json:"health,omitempty"`
	Performance *PerformanceMetrics `json:"performance,omitempty"`
	LastError   string              `json:"last_error,omitempty"`
	LastChecked time.Time           `json:"last_checked"`
	Version     string              `json:"version,omitempty"`
	Platform    Platform            `json:"platform"`
}

// ServerStatus represents the connection status of a remote server
type ServerStatus string

const (
	ServerStatusOnline     ServerStatus = "online"
	ServerStatusOffline    ServerStatus = "offline"
	ServerStatusConnecting ServerStatus = "connecting"
	ServerStatusError      ServerStatus = "error"
	ServerStatusUnknown    ServerStatus = "unknown"
	ServerStatusDegraded   ServerStatus = "degraded"
)

// RemoteResult contains the result of a remote command execution
type RemoteResult struct {
	ServerID   string        `json:"server_id"`
	ServerName string        `json:"server_name"`
	Command    string        `json:"command"`
	Output     string        `json:"output"`
	Error      string        `json:"error,omitempty"`
	ExitCode   int           `json:"exit_code"`
	Duration   time.Duration `json:"duration"`
	Timestamp  time.Time     `json:"timestamp"`
	Success    bool          `json:"success"`
}

// ClusterStatus contains the overall status of the server cluster
type ClusterStatus struct {
	TotalServers   int            `json:"total_servers"`
	OnlineServers  int            `json:"online_servers"`
	OfflineServers int            `json:"offline_servers"`
	ErrorServers   int            `json:"error_servers"`
	Servers        []*ServerInfo  `json:"servers"`
	LastUpdated    time.Time      `json:"last_updated"`
	OverallHealth  HealthLevel    `json:"overall_health"`
	Alerts         []ClusterAlert `json:"alerts"`
}

// ClusterMonitor provides real-time monitoring of the cluster
type ClusterMonitor struct {
	Status         *ClusterStatus         `json:"status"`
	Metrics        map[string]interface{} `json:"metrics"`
	Events         []ClusterEvent         `json:"events"`
	StartTime      time.Time              `json:"start_time"`
	UpdateInterval time.Duration          `json:"update_interval"`
}

// ClusterAlert represents an alert for the entire cluster
type ClusterAlert struct {
	ID           string      `json:"id"`
	Level        HealthLevel `json:"level"`
	Message      string      `json:"message"`
	ServerID     string      `json:"server_id,omitempty"`
	ServerName   string      `json:"server_name,omitempty"`
	Timestamp    time.Time   `json:"timestamp"`
	Acknowledged bool        `json:"acknowledged"`
}

// ClusterEvent represents an event in the cluster
type ClusterEvent struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"` // server_added, server_removed, status_change, etc.
	ServerID  string                 `json:"server_id,omitempty"`
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// SyncResult contains the result of a configuration synchronization
type SyncResult struct {
	ServerID   string        `json:"server_id"`
	ServerName string        `json:"server_name"`
	Success    bool          `json:"success"`
	FilesSync  int           `json:"files_synced"`
	Error      string        `json:"error,omitempty"`
	Duration   time.Duration `json:"duration"`
	Timestamp  time.Time     `json:"timestamp"`
}

// ConnectionPool manages SSH connections to remote servers
type ConnectionPool interface {
	GetConnection(ctx context.Context, serverID string) (Connection, error)
	ReleaseConnection(serverID string, conn Connection) error
	CloseAll() error
}

// Connection represents an SSH connection to a remote server
type Connection interface {
	Execute(ctx context.Context, command string) (*RemoteResult, error)
	IsAlive() bool
	Close() error
	GetServerID() string
}

// Validate validates the server configuration
func (c *ServerConfig) Validate() error {
	if c.Name == "" {
		return NewRemoteError("", c.Host, "validate", nil, "server name is required")
	}
	if c.Host == "" {
		return NewRemoteError(c.ID, "", "validate", nil, "server host is required")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return NewRemoteError(c.ID, c.Host, "validate", nil, "invalid port number")
	}
	if c.Username == "" {
		return NewRemoteError(c.ID, c.Host, "validate", nil, "username is required")
	}
	if c.AuthMethod == AuthMethodKey && c.KeyPath == "" {
		return NewRemoteError(c.ID, c.Host, "validate", nil, "key path is required for key authentication")
	}
	if c.AuthMethod == AuthMethodPassword && c.Password == "" {
		return NewRemoteError(c.ID, c.Host, "validate", nil, "password is required for password authentication")
	}
	return nil
}

// IsOnline returns true if the server is online
func (s *ServerInfo) IsOnline() bool {
	return s.Status == ServerStatusOnline
}

// GetConnectionString returns a connection string for the server
func (s *ServerInfo) GetConnectionString() string {
	if s.Config.Port == 22 {
		return fmt.Sprintf("%s@%s", s.Config.Username, s.Config.Host)
	}
	return fmt.Sprintf("%s@%s:%d", s.Config.Username, s.Config.Host, s.Config.Port)
}

// GetHealthSummary returns a summary of the cluster health
func (c *ClusterStatus) GetHealthSummary() string {
	if c.TotalServers == 0 {
		return "No servers configured"
	}

	healthyPercent := float64(c.OnlineServers) / float64(c.TotalServers) * 100

	switch {
	case healthyPercent >= 90:
		return fmt.Sprintf("Cluster is healthy (%d/%d servers online)", c.OnlineServers, c.TotalServers)
	case healthyPercent >= 70:
		return fmt.Sprintf("Cluster has some issues (%d/%d servers online)", c.OnlineServers, c.TotalServers)
	default:
		return fmt.Sprintf("Cluster has serious issues (%d/%d servers online)", c.OnlineServers, c.TotalServers)
	}
}
