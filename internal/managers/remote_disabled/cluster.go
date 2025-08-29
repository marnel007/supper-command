package remote

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"suppercommand/internal/types"
)

// ClusterManager manages clusters of remote servers
type ClusterManager struct {
	clusters map[string]*Cluster
	mutex    sync.RWMutex
}

// Cluster represents a group of servers
type Cluster struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Servers     []string          `json:"servers"`
	Tags        map[string]string `json:"tags"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Health      *ClusterHealth    `json:"health"`
	Config      *ClusterConfig    `json:"config"`
}

// ClusterHealth represents the health status of a cluster
type ClusterHealth struct {
	OverallStatus  types.ServerStatus `json:"overall_status"`
	OnlineServers  int                `json:"online_servers"`
	OfflineServers int                `json:"offline_servers"`
	TotalServers   int                `json:"total_servers"`
	LastChecked    time.Time          `json:"last_checked"`
	HealthyPercent float64            `json:"healthy_percent"`
	ResponseTime   time.Duration      `json:"response_time"`
}

// ClusterConfig represents cluster configuration
type ClusterConfig struct {
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	CommandTimeout      time.Duration `json:"command_timeout"`
	MaxConcurrency      int           `json:"max_concurrency"`
	RetryAttempts       int           `json:"retry_attempts"`
	RetryDelay          time.Duration `json:"retry_delay"`
}

// NewClusterManager creates a new cluster manager
func NewClusterManager() *ClusterManager {
	return &ClusterManager{
		clusters: make(map[string]*Cluster),
	}
}

// CreateCluster creates a new cluster
func (cm *ClusterManager) CreateCluster(name, description string, servers []string, tags map[string]string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if name == "" {
		return fmt.Errorf("cluster name cannot be empty")
	}

	if _, exists := cm.clusters[name]; exists {
		return fmt.Errorf("cluster '%s' already exists", name)
	}

	if len(servers) == 0 {
		return fmt.Errorf("cluster must have at least one server")
	}

	cluster := &Cluster{
		Name:        name,
		Description: description,
		Servers:     servers,
		Tags:        tags,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Config:      getDefaultClusterConfig(),
	}

	cm.clusters[name] = cluster
	return nil
}

// DeleteCluster deletes a cluster
func (cm *ClusterManager) DeleteCluster(name string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if _, exists := cm.clusters[name]; !exists {
		return fmt.Errorf("cluster '%s' not found", name)
	}

	delete(cm.clusters, name)
	return nil
}

// AddServerToCluster adds a server to a cluster
func (cm *ClusterManager) AddServerToCluster(clusterName, serverName string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cluster, exists := cm.clusters[clusterName]
	if !exists {
		return fmt.Errorf("cluster '%s' not found", clusterName)
	}

	// Check if server is already in cluster
	for _, server := range cluster.Servers {
		if server == serverName {
			return fmt.Errorf("server '%s' is already in cluster '%s'", serverName, clusterName)
		}
	}

	cluster.Servers = append(cluster.Servers, serverName)
	cluster.UpdatedAt = time.Now()
	return nil
}

// RemoveServerFromCluster removes a server from a cluster
func (cm *ClusterManager) RemoveServerFromCluster(clusterName, serverName string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cluster, exists := cm.clusters[clusterName]
	if !exists {
		return fmt.Errorf("cluster '%s' not found", clusterName)
	}

	// Find and remove server
	for i, server := range cluster.Servers {
		if server == serverName {
			cluster.Servers = append(cluster.Servers[:i], cluster.Servers[i+1:]...)
			cluster.UpdatedAt = time.Now()
			return nil
		}
	}

	return fmt.Errorf("server '%s' not found in cluster '%s'", serverName, clusterName)
}

// ListClusters returns all clusters
func (cm *ClusterManager) ListClusters() []*Cluster {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	clusters := make([]*Cluster, 0, len(cm.clusters))
	for _, cluster := range cm.clusters {
		clusters = append(clusters, cluster)
	}

	// Sort by name
	sort.Slice(clusters, func(i, j int) bool {
		return clusters[i].Name < clusters[j].Name
	})

	return clusters
}

// GetCluster returns a specific cluster
func (cm *ClusterManager) GetCluster(name string) (*Cluster, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	cluster, exists := cm.clusters[name]
	if !exists {
		return nil, fmt.Errorf("cluster '%s' not found", name)
	}

	return cluster, nil
}

// ExecuteOnCluster executes a command on all servers in a cluster
func (cm *ClusterManager) ExecuteOnCluster(ctx context.Context, remoteManager types.RemoteManager, clusterName, command string) (*ClusterExecutionResult, error) {
	cluster, err := cm.GetCluster(clusterName)
	if err != nil {
		return nil, err
	}

	startTime := time.Now()

	// Execute command on all servers in parallel
	results, err := remoteManager.ExecuteCommandOnCluster(ctx, clusterName, command)
	if err != nil {
		return nil, err
	}

	// Analyze results
	successful := 0
	failed := 0
	totalDuration := time.Duration(0)
	resultMap := make(map[string]*types.RemoteResult)

	for _, result := range results {
		if result.ExitCode == 0 {
			successful++
		} else {
			failed++
		}
		totalDuration += result.Duration
		resultMap[result.ServerName] = result
	}

	clusterResult := &ClusterExecutionResult{
		ClusterName:     clusterName,
		Command:         command,
		TotalServers:    len(cluster.Servers),
		SuccessfulCount: successful,
		FailedCount:     failed,
		Results:         resultMap,
		StartTime:       startTime,
		Duration:        time.Since(startTime),
		AverageDuration: totalDuration / time.Duration(len(results)),
	}

	return clusterResult, nil
}

// CheckClusterHealth checks the health of all servers in a cluster
func (cm *ClusterManager) CheckClusterHealth(ctx context.Context, remoteManager types.RemoteManager, clusterName string) (*ClusterHealth, error) {
	cluster, err := cm.GetCluster(clusterName)
	if err != nil {
		return nil, err
	}

	startTime := time.Now()

	// Execute health check command on all servers
	results, err := remoteManager.ExecuteCommandOnCluster(ctx, clusterName, "echo 'health_check'")
	if err != nil {
		return nil, err
	}

	// Analyze health results
	online := 0
	offline := 0

	for _, result := range results {
		if result.ExitCode == 0 {
			online++
		} else {
			offline++
		}
	}

	total := len(cluster.Servers)
	healthyPercent := float64(online) / float64(total) * 100

	var overallStatus types.ServerStatus
	if online == total {
		overallStatus = types.ServerStatusOnline
	} else if online > 0 {
		overallStatus = types.ServerStatusDegraded
	} else {
		overallStatus = types.ServerStatusOffline
	}

	health := &ClusterHealth{
		OverallStatus:  overallStatus,
		OnlineServers:  online,
		OfflineServers: offline,
		TotalServers:   total,
		LastChecked:    time.Now(),
		HealthyPercent: healthyPercent,
		ResponseTime:   time.Since(startTime),
	}

	// Update cluster health
	cm.mutex.Lock()
	cluster.Health = health
	cm.mutex.Unlock()

	return health, nil
}

// UpdateClusterConfig updates cluster configuration
func (cm *ClusterManager) UpdateClusterConfig(clusterName string, config *ClusterConfig) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cluster, exists := cm.clusters[clusterName]
	if !exists {
		return fmt.Errorf("cluster '%s' not found", clusterName)
	}

	cluster.Config = config
	cluster.UpdatedAt = time.Now()
	return nil
}

// GetClusterStats returns statistics for all clusters
func (cm *ClusterManager) GetClusterStats() map[string]interface{} {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	totalClusters := len(cm.clusters)
	totalServers := 0
	healthyClusters := 0
	degradedClusters := 0
	offlineClusters := 0

	for _, cluster := range cm.clusters {
		totalServers += len(cluster.Servers)

		if cluster.Health != nil {
			switch cluster.Health.OverallStatus {
			case types.ServerStatusOnline:
				healthyClusters++
			case types.ServerStatusDegraded:
				degradedClusters++
			case types.ServerStatusOffline:
				offlineClusters++
			}
		}
	}

	return map[string]interface{}{
		"total_clusters":    totalClusters,
		"total_servers":     totalServers,
		"healthy_clusters":  healthyClusters,
		"degraded_clusters": degradedClusters,
		"offline_clusters":  offlineClusters,
	}
}

// ClusterExecutionResult represents the result of executing a command on a cluster
type ClusterExecutionResult struct {
	ClusterName     string                         `json:"cluster_name"`
	Command         string                         `json:"command"`
	TotalServers    int                            `json:"total_servers"`
	SuccessfulCount int                            `json:"successful_count"`
	FailedCount     int                            `json:"failed_count"`
	Results         map[string]*types.RemoteResult `json:"results"`
	StartTime       time.Time                      `json:"start_time"`
	Duration        time.Duration                  `json:"duration"`
	AverageDuration time.Duration                  `json:"average_duration"`
}

// GetSuccessRate returns the success rate as a percentage
func (cer *ClusterExecutionResult) GetSuccessRate() float64 {
	if cer.TotalServers == 0 {
		return 0
	}
	return float64(cer.SuccessfulCount) / float64(cer.TotalServers) * 100
}

// GetFailedServers returns a list of servers that failed
func (cer *ClusterExecutionResult) GetFailedServers() []string {
	failed := make([]string, 0)
	for serverName, result := range cer.Results {
		if result.ExitCode != 0 {
			failed = append(failed, serverName)
		}
	}
	return failed
}

// GetSuccessfulServers returns a list of servers that succeeded
func (cer *ClusterExecutionResult) GetSuccessfulServers() []string {
	successful := make([]string, 0)
	for serverName, result := range cer.Results {
		if result.ExitCode == 0 {
			successful = append(successful, serverName)
		}
	}
	return successful
}

// getDefaultClusterConfig returns default cluster configuration
func getDefaultClusterConfig() *ClusterConfig {
	return &ClusterConfig{
		HealthCheckInterval: 30 * time.Second,
		CommandTimeout:      60 * time.Second,
		MaxConcurrency:      10,
		RetryAttempts:       3,
		RetryDelay:          5 * time.Second,
	}
}
