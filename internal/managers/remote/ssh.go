package remote

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"strings"
	"sync"
	"time"

	"suppercommand/internal/types"

	"golang.org/x/crypto/ssh"
)

// SSHConnection represents an SSH connection to a remote server
type SSHConnection struct {
	config       *types.ServerConfig
	client       *ssh.Client
	session      *ssh.Session
	connected    bool
	lastActivity time.Time
	mutex        sync.RWMutex
}

// NewSSHConnection creates a new SSH connection
func NewSSHConnection(config *types.ServerConfig) *SSHConnection {
	return &SSHConnection{
		config:       config,
		connected:    false,
		lastActivity: time.Now(),
	}
}

// Connect establishes SSH connection to the remote server
func (s *SSHConnection) Connect(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.connected && s.client != nil {
		return nil // Already connected
	}

	// Create SSH client configuration
	sshConfig, err := s.createSSHConfig()
	if err != nil {
		return fmt.Errorf("failed to create SSH config: %w", err)
	}

	// Set connection timeout
	dialer := &net.Dialer{
		Timeout: 10 * time.Second,
	}

	// Connect to the remote server
	address := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return fmt.Errorf("failed to dial %s: %w", address, err)
	}

	// Create SSH connection
	sshConn, chans, reqs, err := ssh.NewClientConn(conn, address, sshConfig)
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to create SSH connection: %w", err)
	}

	s.client = ssh.NewClient(sshConn, chans, reqs)
	s.connected = true
	s.lastActivity = time.Now()

	return nil
}

// createSSHConfig creates SSH client configuration
func (s *SSHConnection) createSSHConfig() (*ssh.ClientConfig, error) {
	config := &ssh.ClientConfig{
		User:            s.config.Username,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // In production, use proper host key verification
		Timeout:         10 * time.Second,
	}

	// Configure authentication method
	if s.config.KeyPath != "" {
		// Use key-based authentication
		auth, err := s.createKeyAuth()
		if err != nil {
			return nil, fmt.Errorf("failed to create key auth: %w", err)
		}
		config.Auth = []ssh.AuthMethod{auth}
	} else if s.config.Password != "" {
		// Use password authentication
		config.Auth = []ssh.AuthMethod{
			ssh.Password(s.config.Password),
		}
	} else {
		return nil, fmt.Errorf("no authentication method configured")
	}

	return config, nil
}

// createKeyAuth creates key-based authentication
func (s *SSHConnection) createKeyAuth() (ssh.AuthMethod, error) {
	// Read private key file
	keyData, err := ioutil.ReadFile(s.config.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file %s: %w", s.config.KeyPath, err)
	}

	// Parse private key (simplified - no passphrase support for now)
	var signer ssh.Signer
	signer, err = ssh.ParsePrivateKey(keyData)

	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return ssh.PublicKeys(signer), nil
}

// ExecuteCommand executes a command on the remote server
func (s *SSHConnection) ExecuteCommand(ctx context.Context, command string) (*types.RemoteResult, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.connected || s.client == nil {
		return nil, fmt.Errorf("not connected to server")
	}

	startTime := time.Now()

	// Create a new session for this command
	session, err := s.client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Execute command with context timeout
	done := make(chan error, 1)
	var output []byte

	go func() {
		output, err = session.CombinedOutput(command)
		done <- err
	}()

	select {
	case <-ctx.Done():
		session.Signal(ssh.SIGTERM)
		return &types.RemoteResult{
			ServerName: s.config.Name,
			Command:    command,
			ExitCode:   -1,
			Output:     "",
			Error:      "command timed out",
			Duration:   time.Since(startTime),
			Timestamp:  startTime,
		}, ctx.Err()
	case err := <-done:
		s.lastActivity = time.Now()

		result := &types.RemoteResult{
			ServerName: s.config.Name,
			Command:    command,
			Output:     string(output),
			Duration:   time.Since(startTime),
			Timestamp:  startTime,
		}

		if err != nil {
			if exitError, ok := err.(*ssh.ExitError); ok {
				result.ExitCode = exitError.ExitStatus()
			} else {
				result.ExitCode = -1
			}
			result.Error = err.Error()
		} else {
			result.ExitCode = 0
		}

		return result, nil
	}
}

// UploadFile uploads a file to the remote server
func (s *SSHConnection) UploadFile(ctx context.Context, localPath, remotePath string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.connected || s.client == nil {
		return fmt.Errorf("not connected to server")
	}

	// Read local file
	data, err := ioutil.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("failed to read local file: %w", err)
	}

	// Create SCP session
	session, err := s.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Use SCP to upload file
	// This is a simplified implementation - in production, use a proper SCP library
	command := fmt.Sprintf("cat > %s", remotePath)
	stdin, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	go func() {
		defer stdin.Close()
		stdin.Write(data)
	}()

	if err := session.Run(command); err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	s.lastActivity = time.Now()
	return nil
}

// DownloadFile downloads a file from the remote server
func (s *SSHConnection) DownloadFile(ctx context.Context, remotePath, localPath string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.connected || s.client == nil {
		return fmt.Errorf("not connected to server")
	}

	// Create session
	session, err := s.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Read remote file
	output, err := session.Output(fmt.Sprintf("cat %s", remotePath))
	if err != nil {
		return fmt.Errorf("failed to read remote file: %w", err)
	}

	// Write to local file
	if err := ioutil.WriteFile(localPath, output, 0644); err != nil {
		return fmt.Errorf("failed to write local file: %w", err)
	}

	s.lastActivity = time.Now()
	return nil
}

// CreateTunnel creates an SSH tunnel
func (s *SSHConnection) CreateTunnel(ctx context.Context, localPort int, remoteHost string, remotePort int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.connected || s.client == nil {
		return fmt.Errorf("not connected to server")
	}

	// Listen on local port
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", localPort))
	if err != nil {
		return fmt.Errorf("failed to listen on local port: %w", err)
	}

	go func() {
		defer listener.Close()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				conn, err := listener.Accept()
				if err != nil {
					continue
				}

				go s.handleTunnelConnection(conn, remoteHost, remotePort)
			}
		}
	}()

	return nil
}

// handleTunnelConnection handles a tunnel connection
func (s *SSHConnection) handleTunnelConnection(localConn net.Conn, remoteHost string, remotePort int) {
	defer localConn.Close()

	// Connect to remote host through SSH
	remoteConn, err := s.client.Dial("tcp", fmt.Sprintf("%s:%d", remoteHost, remotePort))
	if err != nil {
		return
	}
	defer remoteConn.Close()

	// Copy data between connections
	done := make(chan bool, 2)

	go func() {
		defer func() { done <- true }()
		// Copy from local to remote
		localConn.(*net.TCPConn).ReadFrom(remoteConn)
	}()

	go func() {
		defer func() { done <- true }()
		// Copy from remote to local
		remoteConn.(*net.TCPConn).ReadFrom(localConn)
	}()

	<-done
}

// IsConnected checks if the connection is active
func (s *SSHConnection) IsConnected() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if !s.connected || s.client == nil {
		return false
	}

	// Test connection with a simple command
	session, err := s.client.NewSession()
	if err != nil {
		s.connected = false
		return false
	}
	defer session.Close()

	// Run a simple command to test connectivity
	if err := session.Run("echo test"); err != nil {
		s.connected = false
		return false
	}

	return true
}

// LastActivity returns the last activity time
func (s *SSHConnection) LastActivity() time.Time {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.lastActivity
}

// Close closes the SSH connection
func (s *SSHConnection) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.session != nil {
		s.session.Close()
		s.session = nil
	}

	if s.client != nil {
		err := s.client.Close()
		s.client = nil
		s.connected = false
		return err
	}

	s.connected = false
	return nil
}

// GetServerInfo gets information about the remote server
func (s *SSHConnection) GetServerInfo(ctx context.Context) (*types.ServerInfo, error) {
	if !s.IsConnected() {
		return nil, fmt.Errorf("not connected to server")
	}

	info := &types.ServerInfo{
		Config:      s.config,
		Status:      types.ServerStatusOnline,
		LastChecked: s.lastActivity,
		Platform:    types.PlatformLinux, // Default
		Version:     "",
	}

	// Get platform information
	if result, err := s.ExecuteCommand(ctx, "uname -s"); err == nil && result.ExitCode == 0 {
		platform := strings.TrimSpace(result.Output)
		switch strings.ToLower(platform) {
		case "linux":
			info.Platform = types.PlatformLinux
		case "darwin":
			info.Platform = types.PlatformDarwin
		default:
			info.Platform = types.PlatformLinux // Default to Linux
		}
	}

	// Get version information
	if result, err := s.ExecuteCommand(ctx, "uname -r"); err == nil && result.ExitCode == 0 {
		info.Version = strings.TrimSpace(result.Output)
	}

	// Additional server info could be collected here
	// For now, we'll keep it simple

	return info, nil
}

// parseFloat parses a float64 from string
func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

// Execute executes a command and returns the result (implements SSHConnectionInterface)
func (s *SSHConnection) Execute(ctx context.Context, command string) (*types.RemoteResult, error) {
	return s.ExecuteCommand(ctx, command)
}
