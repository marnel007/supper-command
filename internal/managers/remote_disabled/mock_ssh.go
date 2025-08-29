package remote

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"suppercommand/internal/types"
)

// MockSSHConnection provides a mock SSH connection for testing and development
type MockSSHConnection struct {
	config       *types.ServerConfig
	connected    bool
	lastActivity time.Time
	mutex        sync.RWMutex
}

// NewMockSSHConnection creates a new mock SSH connection
func NewMockSSHConnection(config *types.ServerConfig) *MockSSHConnection {
	return &MockSSHConnection{
		config:       config,
		connected:    false,
		lastActivity: time.Now(),
	}
}

// Connect simulates establishing SSH connection
func (m *MockSSHConnection) Connect(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Simulate connection delay
	time.Sleep(100 * time.Millisecond)

	// For mock purposes, always succeed if basic validation passes
	if m.config.Host == "" || m.config.Username == "" {
		return fmt.Errorf("invalid connection parameters")
	}

	m.connected = true
	m.lastActivity = time.Now()
	return nil
}

// ExecuteCommand simulates executing a command on the remote server
func (m *MockSSHConnection) ExecuteCommand(ctx context.Context, command string) (*types.RemoteResult, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.connected {
		return nil, fmt.Errorf("not connected to server")
	}

	startTime := time.Now()
	m.lastActivity = time.Now()

	// Simulate command execution
	result := &types.RemoteResult{
		ServerName: m.config.Name,
		Command:    command,
		ExitCode:   0,
		Duration:   time.Since(startTime),
		Timestamp:  startTime,
	}

	// Mock different command responses
	switch {
	case strings.Contains(command, "echo"):
		// Echo commands return the echoed text
		parts := strings.Fields(command)
		if len(parts) > 1 {
			result.Output = strings.Join(parts[1:], " ")
		}
	case strings.Contains(command, "uname -s"):
		// Return mock platform
		result.Output = "Linux"
	case strings.Contains(command, "uname -r"):
		// Return mock kernel version
		result.Output = "5.4.0-mock"
	case strings.Contains(command, "uptime -s"):
		// Return mock boot time
		bootTime := time.Now().Add(-24 * time.Hour)
		result.Output = bootTime.Format("2006-01-02 15:04:05")
	case strings.Contains(command, "cat /proc/loadavg"):
		// Return mock load average
		result.Output = "0.5 0.7 0.9 1/123 12345"
	case strings.Contains(command, "health_check"):
		// Health check command
		result.Output = "OK"
	case strings.Contains(command, "ls"):
		// Mock directory listing
		result.Output = "file1.txt\nfile2.txt\ndirectory1/\n"
	case strings.Contains(command, "ps"):
		// Mock process listing
		result.Output = "PID  COMMAND\n1234 nginx\n5678 mysql\n"
	case strings.Contains(command, "df"):
		// Mock disk usage
		result.Output = "Filesystem     1K-blocks    Used Available Use% Mounted on\n/dev/sda1       10485760 5242880   5242880  50% /\n"
	case strings.Contains(command, "free"):
		// Mock memory usage
		result.Output = "              total        used        free      shared  buff/cache   available\nMem:        8192000     4096000     2048000           0     2048000     4096000\n"
	case strings.Contains(command, "whoami"):
		// Return configured username
		result.Output = m.config.Username
	case strings.Contains(command, "hostname"):
		// Return configured hostname
		result.Output = m.config.Host
	case strings.Contains(command, "date"):
		// Return current date
		result.Output = time.Now().Format("Mon Jan 2 15:04:05 MST 2006")
	case strings.Contains(command, "pwd"):
		// Return mock current directory
		result.Output = "/home/" + m.config.Username
	default:
		// For unknown commands, return generic success
		result.Output = fmt.Sprintf("Mock execution of: %s", command)
	}

	// Simulate some commands that might fail
	if strings.Contains(command, "fail") || strings.Contains(command, "error") {
		result.ExitCode = 1
		result.Error = "Command failed (mock error)"
	}

	// Add some realistic delay
	time.Sleep(50 * time.Millisecond)
	result.Duration = time.Since(startTime)

	return result, nil
}

// UploadFile simulates uploading a file to the remote server
func (m *MockSSHConnection) UploadFile(ctx context.Context, localPath, remotePath string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.connected {
		return fmt.Errorf("not connected to server")
	}

	// Check if local file exists
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		return fmt.Errorf("local file does not exist: %s", localPath)
	}

	// Simulate upload delay
	time.Sleep(200 * time.Millisecond)
	m.lastActivity = time.Now()

	return nil
}

// DownloadFile simulates downloading a file from the remote server
func (m *MockSSHConnection) DownloadFile(ctx context.Context, remotePath, localPath string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.connected {
		return fmt.Errorf("not connected to server")
	}

	// Create mock file content
	content := fmt.Sprintf("Mock file content from %s:%s\nDownloaded at: %s\n",
		m.config.Host, remotePath, time.Now().Format(time.RFC3339))

	// Write to local file
	if err := ioutil.WriteFile(localPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write local file: %w", err)
	}

	// Simulate download delay
	time.Sleep(150 * time.Millisecond)
	m.lastActivity = time.Now()

	return nil
}

// CreateTunnel simulates creating an SSH tunnel
func (m *MockSSHConnection) CreateTunnel(ctx context.Context, localPort int, remoteHost string, remotePort int) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.connected {
		return fmt.Errorf("not connected to server")
	}

	// For mock purposes, just log the tunnel creation
	fmt.Printf("Mock SSH tunnel created: localhost:%s -> %s:%s:%d\n",
		localPort, m.config.Host, remoteHost, remotePort)

	m.lastActivity = time.Now()
	return nil
}

// IsConnected checks if the mock connection is active
func (m *MockSSHConnection) IsConnected() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.connected
}

// LastActivity returns the last activity time
func (m *MockSSHConnection) LastActivity() time.Time {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.lastActivity
}

// Close closes the mock SSH connection
func (m *MockSSHConnection) Close() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.connected = false
	return nil
}

// GetServerInfo gets mock information about the remote server
func (m *MockSSHConnection) GetServerInfo(ctx context.Context) (*types.ServerInfo, error) {
	if !m.IsConnected() {
		return nil, fmt.Errorf("not connected to server")
	}

	info := &types.ServerInfo{
		Config:      m.config,
		Status:      types.ServerStatusOnline,
		LastChecked: m.lastActivity,
		Platform:    types.PlatformLinux, // Mock as Linux
		Version:     "5.4.0-mock",
	}

	return info, nil
}

// MockSSHConnectionFactory creates mock SSH connections for testing
func MockSSHConnectionFactory(config *types.ServerConfig) *MockSSHConnection {
	return NewMockSSHConnection(config)
}

// Execute executes a command and returns the result
func (m *MockSSHConnection) Execute(ctx context.Context, command string) (*types.RemoteResult, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.connected {
		return nil, fmt.Errorf("not connected to server")
	}

	m.lastActivity = time.Now()

	// Simulate command execution
	result := &types.RemoteResult{
		ServerID:   m.config.Name,
		ServerName: m.config.Name,
		Command:    command,
		Output:     fmt.Sprintf("Mock output for: %s", command),
		Error:      "",
		ExitCode:   0,
		Duration:   time.Millisecond * 100,
		Timestamp:  time.Now(),
	}

	return result, nil
}
