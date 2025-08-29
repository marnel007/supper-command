package remote

import (
	"context"
	"fmt"
	"sync"
	"time"

	"suppercommand/internal/types"
)

// ParallelExecutor executes commands on multiple servers in parallel
type ParallelExecutor struct {
	maxConcurrency int
	timeout        time.Duration
	retryAttempts  int
	retryDelay     time.Duration
}

// NewParallelExecutor creates a new parallel executor
func NewParallelExecutor(maxConcurrency int, timeout time.Duration) *ParallelExecutor {
	return &ParallelExecutor{
		maxConcurrency: maxConcurrency,
		timeout:        timeout,
		retryAttempts:  3,
		retryDelay:     2 * time.Second,
	}
}

// ExecuteOnServers executes a command on multiple servers in parallel
func (pe *ParallelExecutor) ExecuteOnServers(ctx context.Context, remoteManager types.RemoteManager, servers []string, command string) (map[string]*types.RemoteResult, error) {
	if len(servers) == 0 {
		return nil, fmt.Errorf("no servers specified")
	}

	// Create semaphore to limit concurrency
	semaphore := make(chan struct{}, pe.maxConcurrency)
	results := make(map[string]*types.RemoteResult)
	var mutex sync.Mutex
	var wg sync.WaitGroup

	// Create context with timeout
	execCtx, cancel := context.WithTimeout(ctx, pe.timeout)
	defer cancel()

	// Execute on each server
	for _, serverName := range servers {
		wg.Add(1)
		go func(server string) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Execute with retry
			result := pe.executeWithRetry(execCtx, remoteManager, server, command)

			// Store result
			mutex.Lock()
			results[server] = result
			mutex.Unlock()
		}(serverName)
	}

	wg.Wait()
	return results, nil
}

// executeWithRetry executes a command with retry logic
func (pe *ParallelExecutor) executeWithRetry(ctx context.Context, remoteManager types.RemoteManager, serverName, command string) *types.RemoteResult {
	var lastResult *types.RemoteResult
	var lastError error

	for attempt := 0; attempt <= pe.retryAttempts; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-ctx.Done():
				return &types.RemoteResult{
					ServerName: serverName,
					Command:    command,
					ExitCode:   -1,
					Output:     "",
					Error:      "context cancelled during retry",
					Duration:   0,
					Timestamp:  time.Now(),
				}
			case <-time.After(pe.retryDelay):
			}
		}

		result, err := remoteManager.ExecuteCommand(ctx, serverName, command)
		if err != nil {
			lastError = err
			lastResult = &types.RemoteResult{
				ServerName: serverName,
				Command:    command,
				ExitCode:   -1,
				Output:     "",
				Error:      err.Error(),
				Duration:   0,
				Timestamp:  time.Now(),
			}
			continue
		}

		// Success or command failed (but connection worked)
		return result
	}

	// All retries failed
	if lastResult != nil {
		lastResult.Error = fmt.Sprintf("failed after %d attempts: %s", pe.retryAttempts+1, lastResult.Error)
		return lastResult
	}

	return &types.RemoteResult{
		ServerName: serverName,
		Command:    command,
		ExitCode:   -1,
		Output:     "",
		Error:      fmt.Sprintf("failed after %d attempts: %v", pe.retryAttempts+1, lastError),
		Duration:   0,
		Timestamp:  time.Now(),
	}
}

// ExecuteBatch executes multiple commands on multiple servers
func (pe *ParallelExecutor) ExecuteBatch(ctx context.Context, remoteManager types.RemoteManager, batch *BatchExecution) (*BatchResult, error) {
	if len(batch.Commands) == 0 {
		return nil, fmt.Errorf("no commands specified")
	}

	if len(batch.Servers) == 0 {
		return nil, fmt.Errorf("no servers specified")
	}

	startTime := time.Now()
	results := make(map[string]map[string]*types.RemoteResult)

	// Execute each command on all servers
	for i, command := range batch.Commands {
		commandResults, err := pe.ExecuteOnServers(ctx, remoteManager, batch.Servers, command.Command)
		if err != nil {
			return nil, fmt.Errorf("failed to execute command %d: %w", i+1, err)
		}

		// Store results by command name
		commandName := command.Name
		if commandName == "" {
			commandName = fmt.Sprintf("command_%d", i+1)
		}
		results[commandName] = commandResults

		// Check if we should stop on failure
		if batch.StopOnFailure {
			for _, result := range commandResults {
				if result.ExitCode != 0 {
					return &BatchResult{
						BatchName:    batch.Name,
						TotalServers: len(batch.Servers),
						Results:      results,
						StartTime:    startTime,
						Duration:     time.Since(startTime),
						Completed:    false,
						FailedAt:     commandName,
					}, nil
				}
			}
		}
	}

	return &BatchResult{
		BatchName:    batch.Name,
		TotalServers: len(batch.Servers),
		Results:      results,
		StartTime:    startTime,
		Duration:     time.Since(startTime),
		Completed:    true,
	}, nil
}

// BatchExecution represents a batch of commands to execute
type BatchExecution struct {
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	Servers       []string       `json:"servers"`
	Commands      []BatchCommand `json:"commands"`
	StopOnFailure bool           `json:"stop_on_failure"`
	Parallel      bool           `json:"parallel"`
}

// BatchCommand represents a command in a batch
type BatchCommand struct {
	Name        string        `json:"name"`
	Command     string        `json:"command"`
	Description string        `json:"description"`
	Timeout     time.Duration `json:"timeout"`
}

// BatchResult represents the result of a batch execution
type BatchResult struct {
	BatchName    string                                    `json:"batch_name"`
	TotalServers int                                       `json:"total_servers"`
	Results      map[string]map[string]*types.RemoteResult `json:"results"`
	StartTime    time.Time                                 `json:"start_time"`
	Duration     time.Duration                             `json:"duration"`
	Completed    bool                                      `json:"completed"`
	FailedAt     string                                    `json:"failed_at,omitempty"`
}

// GetOverallStats returns overall statistics for the batch execution
func (br *BatchResult) GetOverallStats() map[string]interface{} {
	totalCommands := len(br.Results)
	totalExecutions := totalCommands * br.TotalServers
	successfulExecutions := 0
	failedExecutions := 0

	for _, commandResults := range br.Results {
		for _, result := range commandResults {
			if result.ExitCode == 0 {
				successfulExecutions++
			} else {
				failedExecutions++
			}
		}
	}

	successRate := float64(successfulExecutions) / float64(totalExecutions) * 100

	return map[string]interface{}{
		"total_commands":        totalCommands,
		"total_servers":         br.TotalServers,
		"total_executions":      totalExecutions,
		"successful_executions": successfulExecutions,
		"failed_executions":     failedExecutions,
		"success_rate":          successRate,
		"duration":              br.Duration.String(),
		"completed":             br.Completed,
	}
}

// GetFailedServers returns servers that had any failures
func (br *BatchResult) GetFailedServers() []string {
	failedServers := make(map[string]bool)

	for _, commandResults := range br.Results {
		for serverName, result := range commandResults {
			if result.ExitCode != 0 {
				failedServers[serverName] = true
			}
		}
	}

	servers := make([]string, 0, len(failedServers))
	for server := range failedServers {
		servers = append(servers, server)
	}

	return servers
}

// GetCommandStats returns statistics for each command
func (br *BatchResult) GetCommandStats() map[string]map[string]interface{} {
	stats := make(map[string]map[string]interface{})

	for commandName, commandResults := range br.Results {
		successful := 0
		failed := 0
		totalDuration := time.Duration(0)

		for _, result := range commandResults {
			if result.ExitCode == 0 {
				successful++
			} else {
				failed++
			}
			totalDuration += result.Duration
		}

		avgDuration := totalDuration / time.Duration(len(commandResults))
		successRate := float64(successful) / float64(len(commandResults)) * 100

		stats[commandName] = map[string]interface{}{
			"successful_count": successful,
			"failed_count":     failed,
			"success_rate":     successRate,
			"average_duration": avgDuration.String(),
		}
	}

	return stats
}

// SetRetryConfig sets retry configuration
func (pe *ParallelExecutor) SetRetryConfig(attempts int, delay time.Duration) {
	pe.retryAttempts = attempts
	pe.retryDelay = delay
}

// SetConcurrency sets maximum concurrency
func (pe *ParallelExecutor) SetConcurrency(maxConcurrency int) {
	pe.maxConcurrency = maxConcurrency
}

// SetTimeout sets execution timeout
func (pe *ParallelExecutor) SetTimeout(timeout time.Duration) {
	pe.timeout = timeout
}
