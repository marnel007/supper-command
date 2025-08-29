package remote

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"suppercommand/internal/types"
)

// SSHRemoteManager implements RemoteManager using SSH
type SSHRemoteManager struct {
	*BaseRemoteManager
	connectionPool *ConnectionPool
}

// NewSSHRemoteManager creates a new SSH remote manager
func NewSSHRemoteManager() *SSHRemoteManager {
	return &SSHRemoteManager{
		BaseRemoteManager: NewBaseRemoteManager(),
		connectionPool:    NewConnectionPool(),
	}
}

// ConnectionPool manages SSH connections with pooling and reuse
type ConnectionPool struct {
	connections map[string]*PooledConnection
	mutex       sync.RWMutex
	maxIdle     time.Duration
	maxLife     time.Duration
}

// PooledConnection represents a pooled SSH connection
type PooledConnection struct {
	SSHConnectionInterface
	createdAt time.Time
	lastUsed  time.Time
	useCount  int
	inUse     bool
	mutex     sync.Mutex
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool() *ConnectionPool {
	pool := &ConnectionPool{
		connections: make(map[string]*PooledConnection),
		maxIdle:     5 * time.Minute,
		maxLife:     30 * time.Minute,
	}

	// Start cleanup goroutine
	go pool.cleanup()

	return pool
}

// GetConnection gets a connection from the pool or creates a new one
func (cp *ConnectionPool) GetConnection(ctx context.Context, config *types.ServerConfig) (*PooledConnection, error) {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()

	key := fmt.Sprintf("%s@%s:%d", config.Username, config.Host, config.Port)

	// Check if we have an existing connection
	if conn, exists := cp.connections[key]; exists {
		conn.mutex.Lock()
		defer conn.mutex.Unlock()

		// Check if connection is still valid and not in use
		if !conn.inUse && conn.IsConnected() && time.Since(conn.lastUsed) < cp.maxIdle {
			conn.inUse = true
			conn.lastUsed = time.Now()
			conn.useCount++
			return conn, nil
		}

		// Connection is stale or in use, remove it
		if !conn.inUse {
			conn.Close()
			delete(cp.connections, key)
		}
	}

	// Create new connection
	sshConn := NewSSHConnection(config)
	if err := sshConn.Connect(ctx); err != nil {
		return nil, err
	}

	pooledConn := &PooledConnection{
		SSHConnectionInterface: sshConn,
		createdAt:              time.Now(),
		lastUsed:               time.Now(),
		useCount:               1,
		inUse:                  true,
	}

	cp.connections[key] = pooledConn
	return pooledConn, nil
}

// ReleaseConnection releases a connection back to the pool
func (cp *ConnectionPool) ReleaseConnection(conn *PooledConnection) {
	conn.mutex.Lock()
	defer conn.mutex.Unlock()

	conn.inUse = false
	conn.lastUsed = time.Now()
}

// cleanup removes stale connections from the pool
func (cp *ConnectionPool) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		cp.mutex.Lock()

		for key, conn := range cp.connections {
			conn.mutex.Lock()

			shouldRemove := false

			// Check if connection is too old
			if time.Since(conn.createdAt) > cp.maxLife {
				shouldRemove = true
			}

			// Check if connection has been idle too long
			if !conn.inUse && time.Since(conn.lastUsed) > cp.maxIdle {
				shouldRemove = true
			}

			// Check if connection is no longer valid
			if !conn.IsConnected() {
				shouldRemove = true
			}

			if shouldRemove && !conn.inUse {
				conn.Close()
				delete(cp.connections, key)
			}

			conn.mutex.Unlock()
		}

		cp.mutex.Unlock()
	}
}

// CloseAll closes all connections in the pool
func (cp *ConnectionPool) CloseAll() {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()

	for _, conn := range cp.connections {
		conn.mutex.Lock()
		conn.Close()
		conn.mutex.Unlock()
	}

	cp.connections = make(map[string]*PooledConnection)
}

// GetStats returns connection pool statistics
func (cp *ConnectionPool) GetStats() map[string]interface{} {
	cp.mutex.RLock()
	defer cp.mutex.RUnlock()

	totalConnections := len(cp.connections)
	activeConnections := 0
	idleConnections := 0

	for _, conn := range cp.connections {
		conn.mutex.Lock()
		if conn.inUse {
			activeConnections++
		} else {
			idleConnections++
		}
		conn.mutex.Unlock()
	}

	return map[string]interface{}{
		"total_connections":  totalConnections,
		"active_connections": activeConnections,
		"idle_connections":   idleConnections,
		"max_idle_time":      cp.maxIdle.String(),
		"max_life_time":      cp.maxLife.String(),
	}
}

// ExecuteCommand executes a command using the connection pool
func (sm *SSHRemoteManager) ExecuteCommand(ctx context.Context, serverName, command string) (*types.RemoteResult, error) {
	// Get server config
	sm.mutex.RLock()
	config, exists := sm.servers[serverName]
	sm.mutex.RUnlock()

	if !exists {
		return nil, types.NewRemoteError(serverName, "", "execute", nil, "server not found")
	}

	// Get connection from pool
	conn, err := sm.connectionPool.GetConnection(ctx, config)
	if err != nil {
		return nil, types.NewRemoteError(serverName, "", "execute", err, "failed to get connection")
	}
	defer sm.connectionPool.ReleaseConnection(conn)

	// Execute command
	return conn.Execute(ctx, command)
}

// UploadFile uploads a file to a remote server
func (sm *SSHRemoteManager) UploadFile(ctx context.Context, serverName, localPath, remotePath string) error {
	// Get server config
	sm.mutex.RLock()
	config, exists := sm.servers[serverName]
	sm.mutex.RUnlock()

	if !exists {
		return types.NewRemoteError(serverName, "", "upload", nil, "server not found")
	}

	// Get connection from pool
	conn, err := sm.connectionPool.GetConnection(ctx, config)
	if err != nil {
		return types.NewRemoteError(serverName, "", "upload", err, "failed to get connection")
	}
	defer sm.connectionPool.ReleaseConnection(conn)

	// Upload file
	return conn.UploadFile(ctx, localPath, remotePath)
}

// DownloadFile downloads a file from a remote server
func (sm *SSHRemoteManager) DownloadFile(ctx context.Context, serverName, remotePath, localPath string) error {
	// Get server config
	sm.mutex.RLock()
	config, exists := sm.servers[serverName]
	sm.mutex.RUnlock()

	if !exists {
		return types.NewRemoteError(serverName, "", "download", nil, "server not found")
	}

	// Get connection from pool
	conn, err := sm.connectionPool.GetConnection(ctx, config)
	if err != nil {
		return types.NewRemoteError(serverName, "", "download", err, "failed to get connection")
	}
	defer sm.connectionPool.ReleaseConnection(conn)

	// Download file
	return conn.DownloadFile(ctx, remotePath, localPath)
}

// CreateTunnel creates an SSH tunnel
func (sm *SSHRemoteManager) CreateTunnel(ctx context.Context, serverName, localPort, remoteHost string, remotePort int) error {
	// Get server config
	sm.mutex.RLock()
	config, exists := sm.servers[serverName]
	sm.mutex.RUnlock()

	if !exists {
		return types.NewRemoteError(serverName, "", "tunnel", nil, "server not found")
	}

	// Get connection from pool
	conn, err := sm.connectionPool.GetConnection(ctx, config)
	if err != nil {
		return types.NewRemoteError(serverName, "", "tunnel", err, "failed to get connection")
	}
	// Note: Don't release connection immediately for tunnels as they need to stay open

	// Create tunnel
	localPortInt, err := strconv.Atoi(localPort)
	if err != nil {
		return types.NewRemoteError(serverName, "", "tunnel", err, "invalid local port")
	}
	return conn.CreateTunnel(ctx, localPortInt, remoteHost, remotePort)
}

// GetServerInfo gets detailed information about a server
func (sm *SSHRemoteManager) GetServerInfo(ctx context.Context, serverName string) (*types.ServerInfo, error) {
	// Get server config
	sm.mutex.RLock()
	config, exists := sm.servers[serverName]
	sm.mutex.RUnlock()

	if !exists {
		return nil, types.NewRemoteError(serverName, "", "info", nil, "server not found")
	}

	// Get connection from pool
	conn, err := sm.connectionPool.GetConnection(ctx, config)
	if err != nil {
		return nil, types.NewRemoteError(serverName, "", "info", err, "failed to get connection")
	}
	defer sm.connectionPool.ReleaseConnection(conn)

	// Get server info
	return conn.GetServerInfo(ctx)
}

// CloseAllConnections closes all connections
func (sm *SSHRemoteManager) CloseAllConnections() {
	sm.BaseRemoteManager.CloseAllConnections()
	sm.connectionPool.CloseAll()
}

// GetConnectionStats returns enhanced connection statistics
func (sm *SSHRemoteManager) GetConnectionStats() map[string]interface{} {
	baseStats := sm.BaseRemoteManager.GetConnectionStats()
	poolStats := sm.connectionPool.GetStats()

	// Merge statistics
	stats := make(map[string]interface{})
	for k, v := range baseStats {
		stats[k] = v
	}
	for k, v := range poolStats {
		stats["pool_"+k] = v
	}

	return stats
}

// CheckServerHealth checks the health of a remote server
func (s *SSHRemoteManager) CheckServerHealth(ctx context.Context, serverID string) (*types.HealthStatus, error) {
	// Get server config
	s.mutex.RLock()
	_, exists := s.servers[serverID]
	s.mutex.RUnlock()

	if !exists {
		return nil, types.NewRemoteError(serverID, "", "check_health", nil, "server not found")
	}

	// Create a basic health check by trying to connect and run a simple command
	_, err := s.ExecuteCommand(ctx, serverID, "echo 'health_check'")
	if err != nil {
		return &types.HealthStatus{
			Overall:   types.HealthLevelCritical,
			Timestamp: time.Now(),
			Components: map[string]types.ComponentHealth{
				"Connection": {
					Status:      types.HealthLevelCritical,
					Message:     "Failed to connect to server",
					LastChecked: time.Now(),
				},
			},
		}, nil
	}

	// If command executed successfully, server is healthy
	return &types.HealthStatus{
		Overall:   types.HealthLevelHealthy,
		Timestamp: time.Now(),
		Components: map[string]types.ComponentHealth{
			"Connection": {
				Status:      types.HealthLevelHealthy,
				Message:     "Server is responding",
				LastChecked: time.Now(),
			},
		},
	}, nil
}

// GetServerMetrics gets performance metrics from a remote server
func (s *SSHRemoteManager) GetServerMetrics(ctx context.Context, serverID string) (*types.PerformanceMetrics, error) {
	// Get server config
	s.mutex.RLock()
	config, exists := s.servers[serverID]
	s.mutex.RUnlock()

	if !exists {
		return nil, types.NewRemoteError(serverID, "", "get_metrics", nil, "server not found")
	}

	// Execute commands to get basic metrics
	cpuResult, err := s.ExecuteCommand(ctx, serverID, "top -bn1 | grep 'Cpu(s)' | awk '{print $2}' | cut -d'%' -f1")
	if err != nil {
		return nil, types.NewRemoteError(serverID, config.Host, "get_metrics", err, "failed to get CPU metrics")
	}

	memResult, err := s.ExecuteCommand(ctx, serverID, "free | grep Mem | awk '{printf \"%.2f\", $3/$2 * 100.0}'")
	if err != nil {
		return nil, types.NewRemoteError(serverID, config.Host, "get_metrics", err, "failed to get memory metrics")
	}

	// Parse results (simplified)
	cpuUsage := 0.0
	if cpuResult.Output != "" {
		if parsed, err := strconv.ParseFloat(strings.TrimSpace(cpuResult.Output), 64); err == nil {
			cpuUsage = parsed
		}
	}

	memUsage := 0.0
	if memResult.Output != "" {
		if parsed, err := strconv.ParseFloat(strings.TrimSpace(memResult.Output), 64); err == nil {
			memUsage = parsed
		}
	}

	return &types.PerformanceMetrics{
		Timestamp: time.Now(),
		CPU: types.CPUMetrics{
			Usage:       cpuUsage,
			LoadAverage: []float64{0, 0, 0},  // Simplified
			CoreUsage:   []float64{cpuUsage}, // Single core for simplicity
			Processes:   0,
			Threads:     0,
		},
		Memory: types.MemoryMetrics{
			Total:     0, // Would need additional commands
			Used:      0,
			Available: 0,
			Usage:     memUsage,
			SwapTotal: 0,
			SwapUsed:  0,
			SwapUsage: 0,
			Cached:    0,
			Buffers:   0,
		},
		Disk:    types.DiskMetrics{},    // Would need additional commands
		Network: types.NetworkMetrics{}, // Would need additional commands
	}, nil
}

// ExecuteCommandOnCluster executes a command on all servers in a cluster
func (s *SSHRemoteManager) ExecuteCommandOnCluster(ctx context.Context, clusterID string, command string) ([]*types.RemoteResult, error) {
	// Get cluster information
	clusterManager := NewClusterManager()
	cluster, err := clusterManager.GetCluster(clusterID)
	if err != nil {
		return nil, types.NewRemoteError(clusterID, "", "execute_cluster", err, "cluster not found")
	}

	// Execute command on all servers in parallel
	results := make([]*types.RemoteResult, 0, len(cluster.Servers))
	resultChan := make(chan *types.RemoteResult, len(cluster.Servers))
	errorChan := make(chan error, len(cluster.Servers))

	// Start goroutines for each server
	for _, serverName := range cluster.Servers {
		go func(serverName string) {
			result, err := s.ExecuteCommand(ctx, serverName, command)
			if err != nil {
				errorChan <- err
				return
			}
			resultChan <- result
		}(serverName)
	}

	// Collect results
	for i := 0; i < len(cluster.Servers); i++ {
		select {
		case result := <-resultChan:
			results = append(results, result)
		case err := <-errorChan:
			// Create a failed result for this server
			results = append(results, &types.RemoteResult{
				ServerName: "unknown",
				Command:    command,
				ExitCode:   1,
				Output:     "",
				Error:      err.Error(),
				Duration:   0,
				Timestamp:  time.Now(),
			})
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return results, nil
}

// ExecuteScript executes a script on a remote server
func (s *SSHRemoteManager) ExecuteScript(ctx context.Context, serverID string, script string) (*types.RemoteResult, error) {
	// For now, just execute the script as a command
	// In a real implementation, this would upload the script and execute it
	return s.ExecuteCommand(ctx, serverID, script)
}

// GetServer gets information about a specific server
func (s *SSHRemoteManager) GetServer(ctx context.Context, serverID string) (*types.ServerInfo, error) {
	s.mutex.RLock()
	config, exists := s.servers[serverID]
	s.mutex.RUnlock()

	if !exists {
		return nil, types.NewRemoteError(serverID, "", "get_server", nil, "server not found")
	}

	return &types.ServerInfo{
		Config:      config,
		Status:      s.getServerStatus(serverID),
		LastChecked: s.getLastSeen(serverID),
		Platform:    types.PlatformWindows,
		Version:     "",
	}, nil
}

// UpdateServer updates server configuration
func (s *SSHRemoteManager) UpdateServer(ctx context.Context, serverID string, config *types.ServerConfig) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.servers[serverID]; !exists {
		return types.NewRemoteError(serverID, "", "update_server", nil, "server not found")
	}

	s.servers[serverID] = config
	return nil
}

// TestConnection tests the connection to a server
func (s *SSHRemoteManager) TestConnection(ctx context.Context, config *types.ServerConfig) error {
	// Mock implementation - always succeeds
	fmt.Printf("Mock: Testing connection to %s@%s:%d\n", config.Username, config.Host, config.Port)
	return nil
}

// SyncConfiguration syncs configuration to multiple servers
func (s *SSHRemoteManager) SyncConfiguration(ctx context.Context, serverIDs []string) error {
	// Mock implementation
	fmt.Printf("Mock: Syncing configuration to %d servers\n", len(serverIDs))
	return nil
}

// MonitorCluster monitors cluster status
func (s *SSHRemoteManager) MonitorCluster(ctx context.Context, realtime bool) (*types.ClusterMonitor, error) {
	// Mock implementation
	return &types.ClusterMonitor{
		Status: &types.ClusterStatus{
			TotalServers:   len(s.servers),
			OnlineServers:  len(s.servers),
			OfflineServers: 0,
			ErrorServers:   0,
			Servers:        []*types.ServerInfo{},
			LastUpdated:    time.Now(),
			OverallHealth:  types.HealthLevelHealthy,
			Alerts:         []types.ClusterAlert{},
		},
		Metrics:   map[string]interface{}{},
		Events:    []types.ClusterEvent{},
		StartTime: time.Now(),
	}, nil
}
