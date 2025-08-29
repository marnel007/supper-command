package remote

import (
	"context"
	"testing"

	"suppercommand/internal/managers/remote_disabled"
	"suppercommand/internal/types"
)

func TestRemoteManagerFactory(t *testing.T) {
	factory := remote.NewFactory()
	if factory == nil {
		t.Fatal("Factory should not be nil")
	}

	manager, err := factory.CreateManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	if manager == nil {
		t.Fatal("Manager should not be nil")
	}
}

func TestRemoteServerConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		config  *types.ServerConfig
		wantErr bool
	}{
		{
			name: "valid SSH key config",
			config: &types.ServerConfig{
				Name:     "test-server",
				Host:     "192.168.1.100",
				Port:     22,
				Username: "admin",
				KeyPath:  "/path/to/key",
				Tags:     map[string]string{"env": "test"},
			},
			wantErr: false,
		},
		{
			name: "valid password config",
			config: &types.ServerConfig{
				Name:     "test-server",
				Host:     "192.168.1.100",
				Port:     22,
				Username: "admin",
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "missing host",
			config: &types.ServerConfig{
				Name:     "test-server",
				Port:     22,
				Username: "admin",
				KeyPath:  "/path/to/key",
			},
			wantErr: true,
		},
		{
			name: "missing authentication",
			config: &types.ServerConfig{
				Name:     "test-server",
				Host:     "192.168.1.100",
				Port:     22,
				Username: "admin",
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			config: &types.ServerConfig{
				Name:     "test-server",
				Host:     "192.168.1.100",
				Port:     99999,
				Username: "admin",
				KeyPath:  "/path/to/key",
			},
			wantErr: true,
		},
	}

	factory := remote.NewFactory()
	manager, err := factory.CreateManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.AddServer(ctx, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddServer() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Clean up if server was added successfully
			if err == nil {
				manager.RemoveServer(ctx, tt.config.Name)
			}
		})
	}
}

func TestRemoteServerLifecycle(t *testing.T) {
	factory := remote.NewFactory()
	manager, err := factory.CreateManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	ctx := context.Background()

	// Test server configuration
	config := &types.ServerConfig{
		Name:     "test-lifecycle-server",
		Host:     "localhost",
		Port:     22,
		Username: "testuser",
		Password: "testpass",
		Tags:     map[string]string{"env": "test", "role": "web"},
	}

	// Add server
	err = manager.AddServer(ctx, config)
	if err != nil {
		t.Fatalf("Failed to add server: %v", err)
	}

	// List servers
	servers, err := manager.ListServers(ctx)
	if err != nil {
		t.Fatalf("Failed to list servers: %v", err)
	}

	found := false
	for _, server := range servers {
		if server.Name == config.Name {
			found = true
			if server.Host != config.Host {
				t.Errorf("Expected host %s, got %s", config.Host, server.Host)
			}
			if server.Port != config.Port {
				t.Errorf("Expected port %d, got %d", config.Port, server.Port)
			}
			if server.Username != config.Username {
				t.Errorf("Expected username %s, got %s", config.Username, server.Username)
			}
			break
		}
	}

	if !found {
		t.Error("Server should be found in the list")
	}

	// Remove server
	err = manager.RemoveServer(ctx, config.Name)
	if err != nil {
		t.Fatalf("Failed to remove server: %v", err)
	}

	// Verify removal
	servers, err = manager.ListServers(ctx)
	if err != nil {
		t.Fatalf("Failed to list servers after removal: %v", err)
	}

	for _, server := range servers {
		if server.Name == config.Name {
			t.Error("Server should not be found after removal")
		}
	}
}

func TestRemoteCommandExecution(t *testing.T) {
	factory := remote.NewFactory()
	manager, err := factory.CreateManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	ctx := context.Background()

	// Add test server
	config := &types.ServerConfig{
		Name:     "test-exec-server",
		Host:     "localhost",
		Port:     22,
		Username: "testuser",
		Password: "testpass",
	}

	err = manager.AddServer(ctx, config)
	if err != nil {
		t.Fatalf("Failed to add server: %v", err)
	}
	defer manager.RemoveServer(ctx, config.Name)

	// Execute command
	result, err := manager.ExecuteCommand(ctx, config.Name, "echo 'test command'")
	if err != nil {
		t.Fatalf("Failed to execute command: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if result.ServerName != config.Name {
		t.Errorf("Expected server name %s, got %s", config.Name, result.ServerName)
	}

	if result.Command != "echo 'test command'" {
		t.Errorf("Expected command 'echo 'test command'', got %s", result.Command)
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}

	if result.Duration <= 0 {
		t.Error("Duration should be positive")
	}

	if result.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
}

func TestRemoteClusterOperations(t *testing.T) {
	factory := remote.NewFactory()
	manager, err := factory.CreateManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	ctx := context.Background()

	// Add test servers
	servers := []string{"server1", "server2", "server3"}
	for _, serverName := range servers {
		config := &types.ServerConfig{
			Name:     serverName,
			Host:     "localhost",
			Port:     22,
			Username: "testuser",
			Password: "testpass",
		}

		err = manager.AddServer(ctx, config)
		if err != nil {
			t.Fatalf("Failed to add server %s: %v", serverName, err)
		}
		defer manager.RemoveServer(ctx, serverName)
	}

	// Execute command on cluster
	results, err := manager.ExecuteCommandOnCluster(ctx, servers, "echo 'cluster test'")
	if err != nil {
		t.Fatalf("Failed to execute command on cluster: %v", err)
	}

	if len(results) != len(servers) {
		t.Errorf("Expected %d results, got %d", len(servers), len(results))
	}

	for _, serverName := range servers {
		result, exists := results[serverName]
		if !exists {
			t.Errorf("Result for server %s should exist", serverName)
			continue
		}

		if result.ServerName != serverName {
			t.Errorf("Expected server name %s, got %s", serverName, result.ServerName)
		}

		if result.ExitCode != 0 {
			t.Errorf("Expected exit code 0 for server %s, got %d", serverName, result.ExitCode)
		}
	}
}

func TestRemoteClusterStatus(t *testing.T) {
	factory := remote.NewFactory()
	manager, err := factory.CreateManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	ctx := context.Background()

	// Get cluster status (should work even with no servers)
	status, err := manager.GetClusterStatus(ctx)
	if err != nil {
		t.Fatalf("Failed to get cluster status: %v", err)
	}

	if status == nil {
		t.Fatal("Status should not be nil")
	}
}

func TestRemoteConfigurationSync(t *testing.T) {
	factory := remote.NewFactory()
	manager, err := factory.CreateManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	ctx := context.Background()

	// Add test server
	config := &types.ServerConfig{
		Name:     "test-sync-server",
		Host:     "localhost",
		Port:     22,
		Username: "testuser",
		Password: "testpass",
	}

	err = manager.AddServer(ctx, config)
	if err != nil {
		t.Fatalf("Failed to add server: %v", err)
	}
	defer manager.RemoveServer(ctx, config.Name)

	// Test configuration sync
	servers := []string{config.Name}
	configPath := "/tmp/test-config"

	err = manager.SyncConfiguration(ctx, servers, configPath)
	if err != nil {
		t.Fatalf("Failed to sync configuration: %v", err)
	}
}

func TestRemoteConnectionStats(t *testing.T) {
	factory := remote.NewFactory()
	manager, err := factory.CreateManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	stats := manager.GetConnectionStats()
	if stats == nil {
		t.Fatal("Stats should not be nil")
	}

	// Check for expected stats keys
	expectedKeys := []string{"total_servers", "active_connections", "connection_pool_size"}
	for _, key := range expectedKeys {
		if _, exists := stats[key]; !exists {
			t.Errorf("Stats should contain key %s", key)
		}
	}
}

func TestRemoteClusterManager(t *testing.T) {
	clusterManager := remote.NewClusterManager()
	if clusterManager == nil {
		t.Fatal("Cluster manager should not be nil")
	}

	// Test cluster creation
	clusterName := "test-cluster"
	description := "Test cluster for unit tests"
	servers := []string{"server1", "server2", "server3"}
	tags := map[string]string{"env": "test", "type": "web"}

	err := clusterManager.CreateCluster(clusterName, description, servers, tags)
	if err != nil {
		t.Fatalf("Failed to create cluster: %v", err)
	}

	// Test cluster listing
	clusters := clusterManager.ListClusters()
	found := false
	for _, cluster := range clusters {
		if cluster.Name == clusterName {
			found = true
			if cluster.Description != description {
				t.Errorf("Expected description %s, got %s", description, cluster.Description)
			}
			if len(cluster.Servers) != len(servers) {
				t.Errorf("Expected %d servers, got %d", len(servers), len(cluster.Servers))
			}
			break
		}
	}

	if !found {
		t.Error("Cluster should be found in the list")
	}

	// Test cluster deletion
	err = clusterManager.DeleteCluster(clusterName)
	if err != nil {
		t.Fatalf("Failed to delete cluster: %v", err)
	}

	// Verify deletion
	clusters = clusterManager.ListClusters()
	for _, cluster := range clusters {
		if cluster.Name == clusterName {
			t.Error("Cluster should not exist after deletion")
		}
	}
}

func BenchmarkRemoteCommandExecution(b *testing.B) {
	factory := remote.NewFactory()
	manager, err := factory.CreateManager()
	if err != nil {
		b.Fatalf("Failed to create manager: %v", err)
	}

	ctx := context.Background()

	// Add test server
	config := &types.ServerConfig{
		Name:     "bench-server",
		Host:     "localhost",
		Port:     22,
		Username: "testuser",
		Password: "testpass",
	}

	err = manager.AddServer(ctx, config)
	if err != nil {
		b.Fatalf("Failed to add server: %v", err)
	}
	defer manager.RemoveServer(ctx, config.Name)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.ExecuteCommand(ctx, config.Name, "echo 'benchmark test'")
		if err != nil {
			b.Fatalf("Failed to execute command: %v", err)
		}
	}
}

func BenchmarkRemoteClusterExecution(b *testing.B) {
	factory := remote.NewFactory()
	manager, err := factory.CreateManager()
	if err != nil {
		b.Fatalf("Failed to create manager: %v", err)
	}

	ctx := context.Background()

	// Add test servers
	servers := []string{"bench-server1", "bench-server2", "bench-server3"}
	for _, serverName := range servers {
		config := &types.ServerConfig{
			Name:     serverName,
			Host:     "localhost",
			Port:     22,
			Username: "testuser",
			Password: "testpass",
		}

		err = manager.AddServer(ctx, config)
		if err != nil {
			b.Fatalf("Failed to add server %s: %v", serverName, err)
		}
		defer manager.RemoveServer(ctx, serverName)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.ExecuteCommandOnCluster(ctx, servers, "echo 'cluster benchmark'")
		if err != nil {
			b.Fatalf("Failed to execute command on cluster: %v", err)
		}
	}
}
