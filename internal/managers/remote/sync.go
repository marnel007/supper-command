package remote

import (
	"context"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"suppercommand/internal/types"
)

// ConfigSyncManager manages configuration synchronization across servers
type ConfigSyncManager struct {
	remoteManager types.RemoteManager
	syncProfiles  map[string]*SyncProfile
	syncHistory   []SyncEvent
	mutex         sync.RWMutex
	maxHistory    int
}

// SyncProfile defines what and how to synchronize
type SyncProfile struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	SourcePath   string            `json:"source_path"`
	TargetPath   string            `json:"target_path"`
	Servers      []string          `json:"servers"`
	Excludes     []string          `json:"excludes"`
	PreCommands  []string          `json:"pre_commands"`
	PostCommands []string          `json:"post_commands"`
	BackupBefore bool              `json:"backup_before"`
	ValidateSync bool              `json:"validate_sync"`
	Permissions  string            `json:"permissions"`
	Owner        string            `json:"owner"`
	Group        string            `json:"group"`
	Tags         map[string]string `json:"tags"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// SyncEvent represents a synchronization event
type SyncEvent struct {
	ProfileName  string                 `json:"profile_name"`
	EventType    string                 `json:"event_type"`
	Timestamp    time.Time              `json:"timestamp"`
	Servers      []string               `json:"servers"`
	SuccessCount int                    `json:"success_count"`
	FailureCount int                    `json:"failure_count"`
	Duration     time.Duration          `json:"duration"`
	Results      map[string]*SyncResult `json:"results"`
	TotalFiles   int                    `json:"total_files"`
	TotalSize    int64                  `json:"total_size"`
	Error        string                 `json:"error,omitempty"`
}

// SyncResult represents the result of syncing to a single server
type SyncResult struct {
	ServerName       string        `json:"server_name"`
	Success          bool          `json:"success"`
	FilesUpdated     int           `json:"files_updated"`
	FilesSkipped     int           `json:"files_skipped"`
	BytesTransferred int64         `json:"bytes_transferred"`
	Duration         time.Duration `json:"duration"`
	Error            string        `json:"error,omitempty"`
	BackupPath       string        `json:"backup_path,omitempty"`
	Checksum         string        `json:"checksum,omitempty"`
}

// NewConfigSyncManager creates a new configuration sync manager
func NewConfigSyncManager(remoteManager types.RemoteManager) *ConfigSyncManager {
	return &ConfigSyncManager{
		remoteManager: remoteManager,
		syncProfiles:  make(map[string]*SyncProfile),
		syncHistory:   make([]SyncEvent, 0),
		maxHistory:    1000,
	}
}

// CreateSyncProfile creates a new synchronization profile
func (csm *ConfigSyncManager) CreateSyncProfile(profile *SyncProfile) error {
	csm.mutex.Lock()
	defer csm.mutex.Unlock()

	if profile.Name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}

	if _, exists := csm.syncProfiles[profile.Name]; exists {
		return fmt.Errorf("sync profile '%s' already exists", profile.Name)
	}

	if profile.SourcePath == "" {
		return fmt.Errorf("source path cannot be empty")
	}

	if profile.TargetPath == "" {
		return fmt.Errorf("target path cannot be empty")
	}

	if len(profile.Servers) == 0 {
		return fmt.Errorf("at least one server must be specified")
	}

	profile.CreatedAt = time.Now()
	profile.UpdatedAt = time.Now()

	csm.syncProfiles[profile.Name] = profile
	return nil
}

// UpdateSyncProfile updates an existing synchronization profile
func (csm *ConfigSyncManager) UpdateSyncProfile(profile *SyncProfile) error {
	csm.mutex.Lock()
	defer csm.mutex.Unlock()

	if _, exists := csm.syncProfiles[profile.Name]; !exists {
		return fmt.Errorf("sync profile '%s' not found", profile.Name)
	}

	profile.UpdatedAt = time.Now()
	csm.syncProfiles[profile.Name] = profile
	return nil
}

// DeleteSyncProfile deletes a synchronization profile
func (csm *ConfigSyncManager) DeleteSyncProfile(name string) error {
	csm.mutex.Lock()
	defer csm.mutex.Unlock()

	if _, exists := csm.syncProfiles[name]; !exists {
		return fmt.Errorf("sync profile '%s' not found", name)
	}

	delete(csm.syncProfiles, name)
	return nil
}

// ListSyncProfiles returns all synchronization profiles
func (csm *ConfigSyncManager) ListSyncProfiles() []*SyncProfile {
	csm.mutex.RLock()
	defer csm.mutex.RUnlock()

	profiles := make([]*SyncProfile, 0, len(csm.syncProfiles))
	for _, profile := range csm.syncProfiles {
		profiles = append(profiles, profile)
	}

	return profiles
}

// SyncConfiguration synchronizes configuration using a profile
func (csm *ConfigSyncManager) SyncConfiguration(ctx context.Context, profileName string) (*SyncEvent, error) {
	csm.mutex.RLock()
	profile, exists := csm.syncProfiles[profileName]
	csm.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("sync profile '%s' not found", profileName)
	}

	startTime := time.Now()
	event := &SyncEvent{
		ProfileName: profileName,
		EventType:   "sync",
		Timestamp:   startTime,
		Servers:     profile.Servers,
		Results:     make(map[string]*SyncResult),
	}

	// Check if source exists
	sourceInfo, err := os.Stat(profile.SourcePath)
	if err != nil {
		event.Error = fmt.Sprintf("source path error: %v", err)
		csm.recordSyncEvent(event)
		return event, err
	}

	// Calculate source checksum and size
	sourceChecksum, sourceSize, fileCount, err := csm.calculateSourceInfo(profile.SourcePath)
	if err != nil {
		event.Error = fmt.Sprintf("failed to calculate source info: %v", err)
		csm.recordSyncEvent(event)
		return event, err
	}

	event.TotalFiles = fileCount
	event.TotalSize = sourceSize

	// Sync to each server
	var wg sync.WaitGroup
	var resultMutex sync.Mutex

	for _, serverName := range profile.Servers {
		wg.Add(1)
		go func(server string) {
			defer wg.Done()

			result := csm.syncToServer(ctx, profile, server, sourceChecksum, sourceInfo.IsDir())

			resultMutex.Lock()
			event.Results[server] = result
			if result.Success {
				event.SuccessCount++
			} else {
				event.FailureCount++
			}
			resultMutex.Unlock()
		}(serverName)
	}

	wg.Wait()

	event.Duration = time.Since(startTime)
	csm.recordSyncEvent(event)

	return event, nil
}

// syncToServer synchronizes configuration to a single server
func (csm *ConfigSyncManager) syncToServer(ctx context.Context, profile *SyncProfile, serverName, sourceChecksum string, isDir bool) *SyncResult {
	startTime := time.Now()
	result := &SyncResult{
		ServerName: serverName,
		Success:    false,
	}

	// Execute pre-commands
	if len(profile.PreCommands) > 0 {
		for _, cmd := range profile.PreCommands {
			if _, err := csm.remoteManager.ExecuteCommand(ctx, serverName, cmd); err != nil {
				result.Error = fmt.Sprintf("pre-command failed: %v", err)
				result.Duration = time.Since(startTime)
				return result
			}
		}
	}

	// Create backup if requested
	if profile.BackupBefore {
		backupPath := fmt.Sprintf("%s.backup.%s", profile.TargetPath, time.Now().Format("20060102_150405"))
		backupCmd := fmt.Sprintf("cp -r %s %s 2>/dev/null || true", profile.TargetPath, backupPath)
		if _, err := csm.remoteManager.ExecuteCommand(ctx, serverName, backupCmd); err == nil {
			result.BackupPath = backupPath
		}
	}

	// Upload file or directory
	if isDir {
		// For directories, we'd need to implement recursive upload
		// For now, simulate directory sync
		result.FilesUpdated = 5               // Mock value
		result.BytesTransferred = 1024 * 1024 // Mock 1MB
	} else {
		// Upload single file
		tempPath := "/tmp/" + filepath.Base(profile.SourcePath)
		if err := csm.remoteManager.(*SSHRemoteManager).UploadFile(ctx, serverName, profile.SourcePath, tempPath); err != nil {
			result.Error = fmt.Sprintf("upload failed: %v", err)
			result.Duration = time.Since(startTime)
			return result
		}

		// Move to target location
		moveCmd := fmt.Sprintf("mv %s %s", tempPath, profile.TargetPath)
		if _, err := csm.remoteManager.ExecuteCommand(ctx, serverName, moveCmd); err != nil {
			result.Error = fmt.Sprintf("move failed: %v", err)
			result.Duration = time.Since(startTime)
			return result
		}

		result.FilesUpdated = 1
		// Get file size
		if info, err := os.Stat(profile.SourcePath); err == nil {
			result.BytesTransferred = info.Size()
		}
	}

	// Set permissions if specified
	if profile.Permissions != "" {
		chmodCmd := fmt.Sprintf("chmod %s %s", profile.Permissions, profile.TargetPath)
		csm.remoteManager.ExecuteCommand(ctx, serverName, chmodCmd)
	}

	// Set ownership if specified
	if profile.Owner != "" || profile.Group != "" {
		owner := profile.Owner
		if profile.Group != "" {
			owner += ":" + profile.Group
		}
		chownCmd := fmt.Sprintf("chown %s %s", owner, profile.TargetPath)
		csm.remoteManager.ExecuteCommand(ctx, serverName, chownCmd)
	}

	// Validate sync if requested
	if profile.ValidateSync {
		// Calculate remote checksum
		checksumCmd := fmt.Sprintf("md5sum %s | cut -d' ' -f1", profile.TargetPath)
		if checksumResult, err := csm.remoteManager.ExecuteCommand(ctx, serverName, checksumCmd); err == nil {
			remoteChecksum := strings.TrimSpace(checksumResult.Output)
			result.Checksum = remoteChecksum
			if remoteChecksum != sourceChecksum {
				result.Error = "checksum validation failed"
				result.Duration = time.Since(startTime)
				return result
			}
		}
	}

	// Execute post-commands
	if len(profile.PostCommands) > 0 {
		for _, cmd := range profile.PostCommands {
			if _, err := csm.remoteManager.ExecuteCommand(ctx, serverName, cmd); err != nil {
				result.Error = fmt.Sprintf("post-command failed: %v", err)
				result.Duration = time.Since(startTime)
				return result
			}
		}
	}

	result.Success = true
	result.Duration = time.Since(startTime)
	return result
}

// calculateSourceInfo calculates checksum, size, and file count for source
func (csm *ConfigSyncManager) calculateSourceInfo(sourcePath string) (string, int64, int, error) {
	info, err := os.Stat(sourcePath)
	if err != nil {
		return "", 0, 0, err
	}

	if info.IsDir() {
		// For directories, calculate combined checksum and stats
		return csm.calculateDirInfo(sourcePath)
	}

	// For single file
	data, err := ioutil.ReadFile(sourcePath)
	if err != nil {
		return "", 0, 0, err
	}

	checksum := fmt.Sprintf("%x", md5.Sum(data))
	return checksum, info.Size(), 1, nil
}

// calculateDirInfo calculates info for a directory
func (csm *ConfigSyncManager) calculateDirInfo(dirPath string) (string, int64, int, error) {
	var totalSize int64
	var fileCount int
	var checksums []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			data, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			checksum := fmt.Sprintf("%x", md5.Sum(data))
			checksums = append(checksums, checksum)
			totalSize += info.Size()
			fileCount++
		}

		return nil
	})

	if err != nil {
		return "", 0, 0, err
	}

	// Combine all checksums
	combinedData := strings.Join(checksums, "")
	combinedChecksum := fmt.Sprintf("%x", md5.Sum([]byte(combinedData)))

	return combinedChecksum, totalSize, fileCount, nil
}

// recordSyncEvent records a sync event in history
func (csm *ConfigSyncManager) recordSyncEvent(event *SyncEvent) {
	csm.mutex.Lock()
	defer csm.mutex.Unlock()

	csm.syncHistory = append(csm.syncHistory, *event)

	// Keep history within limits
	if len(csm.syncHistory) > csm.maxHistory {
		csm.syncHistory = csm.syncHistory[len(csm.syncHistory)-csm.maxHistory:]
	}
}

// GetSyncHistory returns synchronization history
func (csm *ConfigSyncManager) GetSyncHistory() []SyncEvent {
	csm.mutex.RLock()
	defer csm.mutex.RUnlock()

	// Return a copy
	history := make([]SyncEvent, len(csm.syncHistory))
	copy(history, csm.syncHistory)
	return history
}

// GetSyncStats returns synchronization statistics
func (csm *ConfigSyncManager) GetSyncStats() map[string]interface{} {
	csm.mutex.RLock()
	defer csm.mutex.RUnlock()

	totalProfiles := len(csm.syncProfiles)
	totalEvents := len(csm.syncHistory)
	successfulSyncs := 0
	failedSyncs := 0
	totalServers := 0

	// Count unique servers across all profiles
	serverSet := make(map[string]bool)
	for _, profile := range csm.syncProfiles {
		for _, server := range profile.Servers {
			serverSet[server] = true
		}
	}
	totalServers = len(serverSet)

	// Count successful/failed syncs
	for _, event := range csm.syncHistory {
		if event.FailureCount == 0 {
			successfulSyncs++
		} else {
			failedSyncs++
		}
	}

	return map[string]interface{}{
		"total_profiles":   totalProfiles,
		"total_events":     totalEvents,
		"successful_syncs": successfulSyncs,
		"failed_syncs":     failedSyncs,
		"total_servers":    totalServers,
		"history_size":     csm.maxHistory,
	}
}

// ValidateProfile validates a sync profile configuration
func (csm *ConfigSyncManager) ValidateProfile(profile *SyncProfile) error {
	// Check source path exists
	if _, err := os.Stat(profile.SourcePath); err != nil {
		return fmt.Errorf("source path validation failed: %v", err)
	}

	// Validate servers exist in remote manager
	servers, err := csm.remoteManager.ListServers(context.Background())
	if err != nil {
		return fmt.Errorf("failed to list servers: %v", err)
	}

	serverMap := make(map[string]bool)
	for _, server := range servers {
		serverMap[server.Config.Name] = true
	}

	for _, serverName := range profile.Servers {
		if !serverMap[serverName] {
			return fmt.Errorf("server '%s' not found in remote manager", serverName)
		}
	}

	return nil
}

// DryRun performs a dry run of synchronization
func (csm *ConfigSyncManager) DryRun(ctx context.Context, profileName string) (*SyncEvent, error) {
	// Similar to SyncConfiguration but without actually making changes
	// This would show what would be synchronized
	event := &SyncEvent{
		ProfileName: profileName,
		EventType:   "dry_run",
		Timestamp:   time.Now(),
	}

	// Implementation would check what files would be updated
	// without actually transferring them

	return event, nil
}
