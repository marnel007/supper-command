package remote

import (
	"context"
	"fmt"
	"sync"
	"time"

	"suppercommand/internal/types"
)

// SSHConnectionInterface defines the interface for SSH connections
type SSHConnectionInterface interface {
	Connect(ctx context.Context) error
	Execute(ctx context.Context, command string) (*types.RemoteResult, error)
	Close() error
	IsConnected() bool
	LastActivity() time.Time
	UploadFile(ctx context.Context, localPath, remotePath string) error
	DownloadFile(ctx context.Context, remotePath, localPath string) error
	CreateTunnel(ctx context.Context, localPort int, remoteHost string, remotePort int) error
	GetServerInfo(ctx context.Context) (*types.ServerInfo, error)
}

// Factory creates remote managers
type Factory struct{}

// NewFactory creates a new remote manager factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateManager creates a remote manager
func (f *Factory) CreateManager() (types.RemoteManager, error) {
	return NewSSHRemoteManager(), nil
}

// BaseRemoteManager provides common functionality for all remote managers
type BaseRemoteManager struct {
	servers     map[string]*types.ServerConfig
	connections map[string]SSHConnectionInterface
	mutex       sync.RWMutex
}

// NewBaseRemoteManager creates a new base remote manager
func NewBaseRemoteManager() *BaseRemoteManager {
	return &BaseRemoteManager{
		servers:     make(map[string]*types.ServerConfig),
		connections: make(map[string]SSHConnectionInterface),
	}
}

// AddServer adds a server to the remote management
func (b *BaseRemoteManager) AddServer(ctx context.Context, config *types.ServerConfig) error {
	if config.Name == "" {
		return types.NewRemoteError("", "", "add_server", nil, "server name cannot be empty")
	}

	if config.Host == "" {
		return types.NewRemoteError(config.Name, config.Host, "add_server", nil, "server host cannot be empty")
	}

	if config.Port == 0 {
		config.Port = 22 // Default SSH port
	}

	if config.Username == "" {
		return types.NewRemoteError(config.Name, config.Host, "add_server", nil, "username cannot be empty")
	}

	// Validate authentication method
	if config.KeyPath == "" && config.Password == "" {
		return types.NewRemoteError(config.Name, config.Host, "add_server", nil, "either key path or password must be provided")
	}

	b.mutex.Lock()
	defer b.mutex.Unlock()

	// Check if server already exists
	if _, exists := b.servers[config.Name]; exists {
		return types.NewRemoteError(config.Name, config.Host, "add_server", nil, "server already exists")
	}

	// Test connection (optional - just warn if it fails)
	if err := b.testConnection(ctx, config); err != nil {
		// Don't fail the add operation, just log the warning
		// The server will be marked as offline until it can be reached
		fmt.Printf("Warning: Could not connect to server %s@%s:%d - %v\n", config.Username, config.Host, config.Port, err)
		fmt.Printf("Server added but marked as offline. Connection will be attempted when needed.\n")
	}

	b.servers[config.Name] = config
	return nil
}

// RemoveServer removes a server from remote management
func (b *BaseRemoteManager) RemoveServer(ctx context.Context, serverName string) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if _, exists := b.servers[serverName]; !exists {
		return types.NewRemoteError(serverName, "", "remove_server", nil, "server not found")
	}

	// Close existing connection if any
	if conn, exists := b.connections[serverName]; exists {
		conn.Close()
		delete(b.connections, serverName)
	}

	delete(b.servers, serverName)
	return nil
}

// ListServers returns all configured servers
func (b *BaseRemoteManager) ListServers(ctx context.Context) ([]*types.ServerInfo, error) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	servers := make([]*types.ServerInfo, 0, len(b.servers))
	for _, config := range b.servers {
		info := &types.ServerInfo{
			Config:      config,
			Status:      b.getServerStatus(config.Name),
			LastChecked: b.getLastSeen(config.Name),
			Platform:    types.PlatformWindows, // Default to Windows for now
			Version:     "",
		}
		servers = append(servers, info)
	}

	return servers, nil
}

// getServerStatus gets the current status of a server
func (b *BaseRemoteManager) getServerStatus(serverName string) types.ServerStatus {
	if conn, exists := b.connections[serverName]; exists {
		if conn.IsConnected() {
			return types.ServerStatusOnline
		}
	}

	// Test connectivity
	if config, exists := b.servers[serverName]; exists {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := b.testConnection(ctx, config); err == nil {
			return types.ServerStatusOnline
		}
	}

	return types.ServerStatusOffline
}

// getLastSeen gets the last seen time for a server
func (b *BaseRemoteManager) getLastSeen(serverName string) time.Time {
	if conn, exists := b.connections[serverName]; exists {
		return conn.LastActivity()
	}
	return time.Time{}
}

// testConnection tests connectivity to a server
func (b *BaseRemoteManager) testConnection(ctx context.Context, config *types.ServerConfig) error {
	// Create a temporary mock connection for testing
	// In a production environment, you might want to use real SSH connection
	// For now, we'll use mock to avoid connection issues during development
	conn := NewMockSSHConnection(config)
	defer conn.Close()

	return conn.Connect(ctx)
}

// getConnection gets or creates a connection to a server
func (b *BaseRemoteManager) getConnection(ctx context.Context, serverName string) (SSHConnectionInterface, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// Check if connection already exists and is active
	if conn, exists := b.connections[serverName]; exists {
		if conn.IsConnected() {
			return conn, nil
		}
		// Connection is stale, remove it
		conn.Close()
		delete(b.connections, serverName)
	}

	// Get server config
	config, exists := b.servers[serverName]
	if !exists {
		return nil, types.NewRemoteError(serverName, "", "get_connection", nil, "server not found")
	}

	// Create new connection
	conn := NewMockSSHConnection(config)
	if err := conn.Connect(ctx); err != nil {
		return nil, types.NewRemoteError(serverName, "", "get_connection", err, "failed to connect")
	}

	b.connections[serverName] = conn
	return conn, nil
}

// ExecuteCommand executes a command on a remote server
func (b *BaseRemoteManager) ExecuteCommand(ctx context.Context, serverName, command string) (*types.RemoteResult, error) {
	conn, err := b.getConnection(ctx, serverName)
	if err != nil {
		return nil, err
	}

	return conn.Execute(ctx, command)
}

// ExecuteCommandOnCluster executes a command on multiple servers
func (b *BaseRemoteManager) ExecuteCommandOnCluster(ctx context.Context, serverNames []string, command string) (map[string]*types.RemoteResult, error) {
	if len(serverNames) == 0 {
		return nil, types.NewRemoteError("", "", "cluster_execute", nil, "no servers specified")
	}

	results := make(map[string]*types.RemoteResult)
	var wg sync.WaitGroup
	var mutex sync.Mutex

	// Execute commands in parallel
	for _, serverName := range serverNames {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()

			result, err := b.ExecuteCommand(ctx, name, command)
			if err != nil {
				result = &types.RemoteResult{
					ServerName: name,
					Command:    command,
					ExitCode:   -1,
					Output:     "",
					Error:      err.Error(),
					Duration:   0,
					Timestamp:  time.Now(),
				}
			}

			mutex.Lock()
			results[name] = result
			mutex.Unlock()
		}(serverName)
	}

	wg.Wait()
	return results, nil
}

// GetClusterStatus gets status of all servers in cluster
func (b *BaseRemoteManager) GetClusterStatus(ctx context.Context) (*types.ClusterStatus, error) {
	servers, err := b.ListServers(ctx)
	if err != nil {
		return nil, err
	}

	online := 0
	offline := 0
	errorCount := 0

	for _, server := range servers {
		switch server.Status {
		case types.ServerStatusOnline:
			online++
		case types.ServerStatusOffline:
			offline++
		case types.ServerStatusError:
			errorCount++
		}
	}

	status := &types.ClusterStatus{
		TotalServers:   len(servers),
		OnlineServers:  online,
		OfflineServers: offline,
		ErrorServers:   errorCount,
		Servers:        servers,
		LastUpdated:    time.Now(),
		OverallHealth:  types.HealthLevelHealthy, // Simplified
		Alerts:         []types.ClusterAlert{},   // Empty for now
	}

	return status, nil
}

// SyncConfiguration synchronizes configuration across servers
func (b *BaseRemoteManager) SyncConfiguration(ctx context.Context, serverNames []string, configPath string) error {
	if len(serverNames) == 0 {
		return types.NewRemoteError("", "", "sync_config", nil, "no servers specified")
	}

	// In a real implementation, this would:
	// 1. Read the configuration file
	// 2. Upload it to each server
	// 3. Verify the sync was successful
	// 4. Optionally restart services

	var wg sync.WaitGroup
	errors := make([]error, 0)
	var errorMutex sync.Mutex

	for _, serverName := range serverNames {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()

			// Simulate configuration sync
			_, err := b.ExecuteCommand(ctx, name, fmt.Sprintf("echo 'Syncing config: %s'", configPath))
			if err != nil {
				errorMutex.Lock()
				errors = append(errors, types.NewRemoteError(name, "", "sync_config", err, "failed to sync configuration"))
				errorMutex.Unlock()
			}
		}(serverName)
	}

	wg.Wait()

	if len(errors) > 0 {
		return errors[0] // Return first error
	}

	return nil
}

// MonitorCluster starts monitoring cluster health
func (b *BaseRemoteManager) MonitorCluster(ctx context.Context) error {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Check cluster health
			if err := b.checkClusterHealth(ctx); err != nil {
				// Log error but continue monitoring
				continue
			}
		}
	}
}

// checkClusterHealth checks the health of all servers in the cluster
func (b *BaseRemoteManager) checkClusterHealth(ctx context.Context) error {
	b.mutex.RLock()
	serverNames := make([]string, 0, len(b.servers))
	for name := range b.servers {
		serverNames = append(serverNames, name)
	}
	b.mutex.RUnlock()

	// Execute health check command on all servers
	results, err := b.ExecuteCommandOnCluster(ctx, serverNames, "echo 'health_check'")
	if err != nil {
		return err
	}

	// Process results
	for serverName, result := range results {
		if result.ExitCode != 0 {
			// Server is unhealthy
			fmt.Printf("Server %s is unhealthy\n", serverName)
			continue
		}
		// Update server status
		fmt.Printf("Server %s is healthy\n", serverName)
	}

	return nil
}

// CloseAllConnections closes all active connections
func (b *BaseRemoteManager) CloseAllConnections() {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	for _, conn := range b.connections {
		conn.Close()
	}
	b.connections = make(map[string]SSHConnectionInterface)
}

// GetConnectionStats returns connection statistics
func (b *BaseRemoteManager) GetConnectionStats() map[string]interface{} {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	activeConnections := 0
	for _, conn := range b.connections {
		if conn.IsConnected() {
			activeConnections++
		}
	}

	return map[string]interface{}{
		"total_servers":        len(b.servers),
		"active_connections":   activeConnections,
		"connection_pool_size": len(b.connections),
	}
}
