package managers

import (
	"context"
	"testing"
	"time"

	"suppercommand/internal/managers/remote"
	"suppercommand/internal/types"
)

func TestRemoteManager(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Run("create_manager", func(t *testing.T) {
		factory := remote.NewFactory()
		manager, err := factory.CreateManager()
		if err != nil {
			t.Fatalf("Failed to create remote manager: %v", err)
		}
		if manager == nil {
			t.Fatal("Manager should not be nil")
		}
	})

	t.Run("server_management", func(t *testing.T) {
		factory := remote.NewFactory()
		manager, err := factory.CreateManager()
		if err != nil {
			t.Fatalf("Failed to create manager: %v", err)
		}

		// List servers (should be empty initially)
		servers, err := manager.ListServers(ctx)
		if err != nil {
			t.Errorf("Failed to list servers: %v", err)
		}
		if servers == nil {
			t.Error("Servers should not be nil")
		}

		// Add a mock server
		config := &types.ServerConfig{
			Name:     "test-server",
			Host:     "localhost",
			Port:     22,
			Username: "testuser",
			Password: "testpass",
			Tags:     map[string]string{"env": "test"},
		}

		err = manager.AddServer(ctx, config)
		if err != nil {
			t.Errorf("Failed to add server: %v", err)
		}

		// List servers again
		servers, err = manager.ListServers(ctx)
		if err != nil {
			t.Errorf("Failed to list servers after adding: %v", err)
		}
		if len(servers) == 0 {
			t.Error("Should have at least one server after adding")
		}

		// Get server
		server, err := manager.GetServer(ctx, "test-server")
		if err != nil {
			t.Errorf("Failed to get server: %v", err)
		}
		if server == nil {
			t.Error("Server should not be nil")
		}

		// Update server
		config.Tags["updated"] = "true"
		err = manager.UpdateServer(ctx, "test-server", config)
		if err != nil {
			t.Errorf("Failed to update server: %v", err)
		}

		// Remove server
		err = manager.RemoveServer(ctx, "test-server")
		if err != nil {
			t.Errorf("Failed to remove server: %v", err)
		}
	})

	t.Run("command_execution", func(t *testing.T) {
		factory := remote.NewFactory()
		manager, err := factory.CreateManager()
		if err != nil {
			t.Fatalf("Failed to create manager: %v", err)
		}

		// Add a mock server for testing
		config := &types.ServerConfig{
			Name:     "test-server",
			Host:     "localhost",
			Port:     22,
			Username: "testuser",
			Password: "testpass",
		}

		err = manager.AddServer(ctx, config)
		if err != nil {
			t.Errorf("Failed to add server: %v", err)
		}

		// Execute command (will use mock implementation)
		result, err := manager.ExecuteCommand(ctx, "test-server", "echo 'test'")
		if err != nil {
			t.Errorf("Failed to execute command: %v", err)
		}
		if result == nil {
			t.Error("Command result should not be nil")
		}

		// Execute script
		script := "#!/bin/bash\necho 'test script'"
		result, err = manager.ExecuteScript(ctx, "test-server", script)
		if err != nil {
			t.Errorf("Failed to execute script: %v", err)
		}
		if result == nil {
			t.Error("Script result should not be nil")
		}

		// Clean up
		manager.RemoveServer(ctx, "test-server")
	})

	t.Run("file_operations", func(t *testing.T) {
		factory := remote.NewFactory()
		manager, err := factory.CreateManager()
		if err != nil {
			t.Fatalf("Failed to create manager: %v", err)
		}

		// Add a mock server
		config := &types.ServerConfig{
			Name:     "test-server",
			Host:     "localhost",
			Port:     22,
			Username: "testuser",
			Password: "testpass",
		}

		err = manager.AddServer(ctx, config)
		if err != nil {
			t.Errorf("Failed to add server: %v", err)
		}

		// Upload file (mock)
		err = manager.UploadFile(ctx, "test-server", "/tmp/test.txt", "/remote/test.txt")
		if err != nil {
			t.Errorf("Failed to upload file: %v", err)
		}

		// Download file (mock)
		err = manager.DownloadFile(ctx, "test-server", "/remote/test.txt", "/tmp/downloaded.txt")
		if err != nil {
			t.Errorf("Failed to download file: %v", err)
		}

		// Clean up
		manager.RemoveServer(ctx, "test-server")
	})
}

func TestClusterManager(t *testing.T) {
	t.Run("cluster_operations", func(t *testing.T) {
		clusterManager := remote.NewClusterManager()

		// Create cluster
		err := clusterManager.CreateCluster("test-cluster", "Test cluster",
			[]string{"server1", "server2"}, map[string]string{"env": "test"})
		if err != nil {
			t.Errorf("Failed to create cluster: %v", err)
		}

		// List clusters
		clusters := clusterManager.ListClusters()
		if len(clusters) == 0 {
			t.Error("Should have at least one cluster")
		}

		// Get cluster
		cluster, err := clusterManager.GetCluster("test-cluster")
		if err != nil {
			t.Errorf("Failed to get cluster: %v", err)
		}
		if cluster == nil {
			t.Error("Cluster should not be nil")
		}

		// Update cluster
		err = clusterManager.UpdateCluster("test-cluster", "Updated description",
			[]string{"server1", "server2", "server3"}, map[string]string{"env": "test", "updated": "true"})
		if err != nil {
			t.Errorf("Failed to update cluster: %v", err)
		}

		// Delete cluster
		err = clusterManager.DeleteCluster("test-cluster")
		if err != nil {
			t.Errorf("Failed to delete cluster: %v", err)
		}
	})

	t.Run("cluster_validation", func(t *testing.T) {
		clusterManager := remote.NewClusterManager()

		// Test invalid cluster creation
		err := clusterManager.CreateCluster("", "Empty name cluster", []string{}, nil)
		if err == nil {
			t.Error("Should fail to create cluster with empty name")
		}

		err = clusterManager.CreateCluster("test", "No servers", []string{}, nil)
		if err == nil {
			t.Error("Should fail to create cluster with no servers")
		}
	})
}

func TestConfigSyncManager(t *testing.T) {
	ctx := context.Background()

	t.Run("sync_profile_management", func(t *testing.T) {
		// Create mock remote manager
		factory := remote.NewFactory()
		remoteManager, err := factory.CreateManager()
		if err != nil {
			t.Fatalf("Failed to create remote manager: %v", err)
		}

		syncManager := remote.NewConfigSyncManager(remoteManager)

		// Create sync profile
		profile := &remote.SyncProfile{
			Name:        "test-sync",
			Description: "Test sync profile",
			SourcePath:  "/tmp/source",
			TargetPath:  "/tmp/target",
			Servers:     []string{"server1"},
		}

		err = syncManager.CreateSyncProfile(profile)
		if err != nil {
			t.Errorf("Failed to create sync profile: %v", err)
		}

		// List profiles
		profiles := syncManager.ListSyncProfiles()
		if len(profiles) == 0 {
			t.Error("Should have at least one sync profile")
		}

		// Get profile
		retrievedProfile := syncManager.GetSyncProfile("test-sync")
		if retrievedProfile == nil {
			t.Error("Should be able to retrieve sync profile")
		}

		// Update profile
		profile.Description = "Updated description"
		err = syncManager.UpdateSyncProfile(profile)
		if err != nil {
			t.Errorf("Failed to update sync profile: %v", err)
		}

		// Execute sync (mock)
		err = syncManager.ExecuteSync(ctx, "test-sync")
		if err != nil {
			t.Errorf("Failed to execute sync: %v", err)
		}

		// Get sync stats
		stats := syncManager.GetSyncStats()
		if stats == nil {
			t.Error("Sync stats should not be nil")
		}

		// Delete profile
		err = syncManager.DeleteSyncProfile("test-sync")
		if err != nil {
			t.Errorf("Failed to delete sync profile: %v", err)
		}
	})

	t.Run("sync_validation", func(t *testing.T) {
		factory := remote.NewFactory()
		remoteManager, err := factory.CreateManager()
		if err != nil {
			t.Fatalf("Failed to create remote manager: %v", err)
		}

		syncManager := remote.NewConfigSyncManager(remoteManager)

		// Test invalid sync profile
		invalidProfile := &remote.SyncProfile{
			Name:        "",
			Description: "Invalid profile",
			SourcePath:  "",
			TargetPath:  "",
			Servers:     []string{},
		}

		err = syncManager.CreateSyncProfile(invalidProfile)
		if err == nil {
			t.Error("Should fail to create invalid sync profile")
		}
	})
}

func TestRemoteMonitoring(t *testing.T) {
	ctx := context.Background()

	t.Run("server_monitoring", func(t *testing.T) {
		factory := remote.NewFactory()
		manager, err := factory.CreateManager()
		if err != nil {
			t.Fatalf("Failed to create manager: %v", err)
		}

		// Add a mock server
		config := &types.ServerConfig{
			Name:     "test-server",
			Host:     "localhost",
			Port:     22,
			Username: "testuser",
			Password: "testpass",
		}

		err = manager.AddServer(ctx, config)
		if err != nil {
			t.Errorf("Failed to add server: %v", err)
		}

		// Check server health
		health, err := manager.CheckServerHealth(ctx, "test-server")
		if err != nil {
			t.Errorf("Failed to check server health: %v", err)
		}
		if health == nil {
			t.Error("Server health should not be nil")
		}

		// Get server metrics
		metrics, err := manager.GetServerMetrics(ctx, "test-server")
		if err != nil {
			t.Errorf("Failed to get server metrics: %v", err)
		}
		if metrics == nil {
			t.Error("Server metrics should not be nil")
		}

		// Clean up
		manager.RemoveServer(ctx, "test-server")
	})

	t.Run("connection_testing", func(t *testing.T) {
		factory := remote.NewFactory()
		manager, err := factory.CreateManager()
		if err != nil {
			t.Fatalf("Failed to create manager: %v", err)
		}

		// Test connection to mock server
		config := &types.ServerConfig{
			Name:     "test-server",
			Host:     "localhost",
			Port:     22,
			Username: "testuser",
			Password: "testpass",
		}

		err = manager.TestConnection(ctx, config)
		if err != nil {
			t.Logf("Connection test failed as expected for mock server: %v", err)
		}
	})
}

func TestRemoteErrorHandling(t *testing.T) {
	ctx := context.Background()
	factory := remote.NewFactory()
	manager, err := factory.CreateManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	t.Run("invalid_server_operations", func(t *testing.T) {
		// Try to get non-existent server
		_, err := manager.GetServer(ctx, "non-existent")
		if err == nil {
			t.Error("Should fail to get non-existent server")
		}

		// Try to execute command on non-existent server
		_, err = manager.ExecuteCommand(ctx, "non-existent", "echo test")
		if err == nil {
			t.Error("Should fail to execute command on non-existent server")
		}

		// Try to remove non-existent server
		err = manager.RemoveServer(ctx, "non-existent")
		if err == nil {
			t.Error("Should fail to remove non-existent server")
		}
	})

	t.Run("invalid_configurations", func(t *testing.T) {
		// Try to add server with invalid config
		invalidConfig := &types.ServerConfig{
			Name: "", // Empty name
			Host: "localhost",
			Port: 22,
		}

		err := manager.AddServer(ctx, invalidConfig)
		if err == nil {
			t.Error("Should fail to add server with invalid config")
		}

		// Try to add server with invalid port
		invalidConfig = &types.ServerConfig{
			Name: "test",
			Host: "localhost",
			Port: -1, // Invalid port
		}

		err = manager.AddServer(ctx, invalidConfig)
		if err == nil {
			t.Error("Should fail to add server with invalid port")
		}
	})
}
