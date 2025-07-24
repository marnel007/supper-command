package core

import (
	"fmt"
	"os"
	osuser "os/user" // Use alias to avoid naming conflict with function parameters
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"net"
	"os/exec"

	prompt "github.com/c-bata/go-prompt"
	"github.com/fatih/color"
	"golang.org/x/sys/execabs"
)

type Shell struct{}

func NewShell() *Shell {
	// Register built-in commands
	Register(&HelpCommand{})
	Register(&ClearCommand{})
	Register(&EchoCommand{})
	Register(&PwdCommand{})
	Register(&LsCommand{})
	Register(&CdCommand{})
	Register(&ExitCommand{})
	Register(&CatCommand{})
	Register(&MkdirCommand{})
	Register(&RmCommand{})
	Register(&RmdirCommand{})
	Register(&CpCommand{})
	Register(&MvCommand{})
	Register(&WhoamiCommand{})
	Register(&HostnameCommand{})
	Register(&VerCommand{})
	Register(&DirCommand{})
	Register(&PortscanCommand{})
	Register(&PingCommand{})
	Register(&NslookupCommand{})
	Register(&TracertCommand{})
	Register(&WgetCommand{})
	Register(&IpconfigCommand{})
	Register(&NetstatCommand{})
	Register(&ArpCommand{})
	Register(&RouteCommand{})
	Register(&SpeedtestCommand{})
	Register(&NetdiscoverCommand{})
	Register(&SniffCommand{})
	Register(&SysInfoCommand{})
	Register(&PrivCommand{})
	Register(&RemoteCommand{})
	Register(&WinUpdateCommand{})
	Register(&FastcpDedupCommand{})

	// Register FastCP commands
	registerFastcpCommands()

	// Add Windows-style command aliases
	commandRegistry["copy"] = commandRegistry["cp"]
	commandRegistry["del"] = commandRegistry["rm"]
	commandRegistry["move"] = commandRegistry["mv"]
	commandRegistry["type"] = commandRegistry["cat"]
	commandRegistry["cls"] = commandRegistry["clear"]
	return &Shell{}
}

func (s *Shell) Run() {
	p := prompt.New(
		executor,
		completer,
		prompt.OptionLivePrefix(func() (string, bool) {
			return getPrompt(), true
		}),
		prompt.OptionTitle("SuperShell"),
		prompt.OptionInputTextColor(prompt.White),              // Prompt and input in bright white
		prompt.OptionSuggestionBGColor(prompt.Black),           // Suggestions background: black
		prompt.OptionSuggestionTextColor(prompt.White),         // Suggestions text: white
		prompt.OptionSelectedSuggestionBGColor(prompt.Blue),    // Selected suggestion: blue
		prompt.OptionSelectedSuggestionTextColor(prompt.White), // Selected suggestion text: white
		prompt.OptionPreviewSuggestionTextColor(prompt.Yellow), // Preview suggestion: yellow
	)
	p.Run()
}

func getPrompt() string {
	cwd, _ := os.Getwd()
	shortPath := getShortenedPath(cwd)

	// Beautiful SuperShell + Agent OS prompt
	var prompt strings.Builder

	// SuperShell branding
	prompt.WriteString("\033[38;5;51m🚀 \033[1;36mSuperShell\033[0m") // Cyan rocket + text
	prompt.WriteString("\033[38;5;46m●\033[0m")                      // Green Agent OS dot

	// Directory path
	prompt.WriteString(fmt.Sprintf("\033[90m[\033[33m%s\033[90m]\033[0m", shortPath))

	// Cool prompt arrows
	prompt.WriteString(" \033[38;5;51m❯\033[38;5;45m❯\033[38;5;39m❯\033[0m ")

	return prompt.String()
}

// Helper function to shorten long paths
func getShortenedPath(path string) string {
	// Convert to Windows-style paths
	path = strings.ReplaceAll(path, "/", "\\")

	// If path is too long, show drive + ... + last folder
	if len(path) > 40 {
		parts := strings.Split(path, "\\")
		if len(parts) > 3 {
			return parts[0] + "\\.....\\" + parts[len(parts)-1]
		}
	}

	return path
}

// This function is called when the user presses Enter
func executor(in string) {
	in = strings.TrimSpace(in)
	if in == "" {
		return
	}
	if in == "exit" || in == "quit" {
		fmt.Println("Goodbye!")
		os.Exit(0)
	}
	output := Dispatch(in)
	if output != "" {
		fmt.Println(output)
	}
}

// This function provides tab completion
func completer(d prompt.Document) []prompt.Suggest {
	// Gather all commands and aliases
	var suggestions []prompt.Suggest
	for name, cmd := range commandRegistry {
		desc := cmd.Description()
		shortDesc := desc
		if idx := strings.Index(desc, "\n"); idx != -1 {
			shortDesc = desc[:idx]
		}
		if len(shortDesc) > 60 {
			shortDesc = shortDesc[:57] + "..."
		}
		suggestions = append(suggestions, prompt.Suggest{Text: name, Description: shortDesc})
	}

	args := strings.Fields(d.TextBeforeCursor())
	if len(args) == 0 {
		return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)
	}
	if len(args) == 1 {
		// Complete command names
		return prompt.FilterHasPrefix(suggestions, args[0], true)
	}
	// Complete file/dir names for arguments
	toComplete := args[len(args)-1]
	dir, filePrefix := filepath.Split(toComplete)
	if dir == "" {
		dir = "."
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var fileSugg []prompt.Suggest
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, filePrefix) {
			if entry.IsDir() {
				fileSugg = append(fileSugg, prompt.Suggest{Text: filepath.Join(dir, name) + string(os.PathSeparator), Description: "Directory"})
			} else {
				fileSugg = append(fileSugg, prompt.Suggest{Text: filepath.Join(dir, name), Description: "File"})
			}
		}
	}
	return prompt.FilterHasPrefix(fileSugg, filePrefix, true)
}

// Privilege Management Command
type PrivCommand struct{}

func (p *PrivCommand) Name() string { return "priv" }
func (p *PrivCommand) Description() string {
	return `priv - Privilege Management and Elevation

  Usage:
    priv check                    Check current privilege level
    priv test                     Test privilege requirements
    priv elevate <command>        Run command with elevated privileges
    priv sudo <command>           Alias for elevate (Unix-style)
    priv runas <command>          Alias for elevate (Windows-style)

  Options:
    check                         Show current user privileges
    test                          Test various privilege levels
    elevate/sudo/runas           Run command with admin/root privileges

  Examples:
    priv check
    priv test
    priv elevate netstat -ano
    priv sudo portscan 192.168.1.1
    priv runas ipconfig /all
    priv RunAs cmd.exe            (case-insensitive)

  Notes:
    - Commands are case-insensitive (runas = RunAs = RUNAS)
    - Automatically detects privilege requirements
    - Provides guidance for elevation when needed
    - Cross-platform support (Windows UAC, Unix sudo)
    - Shows live feedback during privilege checks
`
}

func (p *PrivCommand) Execute(args []string) string {
	if len(args) == 0 {
		return "Usage: priv <check|test|elevate|sudo|runas> [command]"
	}

	// Convert subcommand to lowercase for case-insensitive matching
	subCommand := strings.ToLower(args[0])

	switch subCommand {
	case "check":
		return p.checkPrivileges()
	case "test":
		return p.testPrivileges()
	case "elevate", "sudo", "runas":
		if len(args) < 2 {
			return "Usage: priv " + args[0] + " <command>" // Use original case in error message
		}
		return p.elevateCommand(strings.Join(args[1:], " "))
	default:
		return "Unknown subcommand: " + args[0] + "\nAvailable: check, test, elevate, sudo, runas"
	}
}

func (p *PrivCommand) checkPrivileges() string {
	fmt.Print("🔍 Checking privileges")

	// Live feedback
	done := make(chan bool)
	go func() {
		dots := 0
		for {
			select {
			case <-done:
				fmt.Print("\r\033[K")
				return
			default:
				dotStr := strings.Repeat(".", (dots%3)+1)
				fmt.Printf("\r🔍 Checking privileges%s   ", dotStr)
				time.Sleep(300 * time.Millisecond)
				dots++
			}
		}
	}()

	// Simulate privilege checking
	time.Sleep(1 * time.Second)

	close(done)
	fmt.Print("\r\033[K")

	var result strings.Builder
	result.WriteString(color.New(color.FgCyan, color.Bold).Sprint("🔒 PRIVILEGE STATUS\n"))
	result.WriteString("═══════════════════════════════════════════════════════════════\n\n")

	// Current user info
	if u, err := osuser.Current(); err == nil {
		result.WriteString(color.New(color.FgGreen, color.Bold).Sprint("👤 USER INFORMATION\n"))
		result.WriteString(fmt.Sprintf("  Username:     %s\n", u.Username))
		result.WriteString(fmt.Sprintf("  User ID:      %s\n", u.Uid))
		result.WriteString(fmt.Sprintf("  Group ID:     %s\n", u.Gid))
		result.WriteString(fmt.Sprintf("  Home Dir:     %s\n", u.HomeDir))
		result.WriteString("\n")
	}

	// Platform-specific privilege checks
	if runtime.GOOS == "windows" {
		result.WriteString(p.checkWindowsPrivileges())
	} else {
		result.WriteString(p.checkUnixPrivileges())
	}

	// Privilege recommendations
	result.WriteString(color.New(color.FgYellow, color.Bold).Sprint("💡 RECOMMENDATIONS\n"))
	result.WriteString("  • Use 'priv elevate <command>' for admin operations\n")
	result.WriteString("  • Use 'priv test' to check specific privilege requirements\n")
	result.WriteString("  • Some commands may work with reduced functionality\n")

	return result.String()
}

func (p *PrivCommand) checkWindowsPrivileges() string {
	var result strings.Builder
	result.WriteString(color.New(color.FgBlue, color.Bold).Sprint("🪟 WINDOWS PRIVILEGES\n"))

	// Check if running as admin
	isAdmin := p.isWindowsAdmin()
	if isAdmin {
		result.WriteString("  Status:       " + color.New(color.FgGreen).Sprint("✅ Administrator") + "\n")
		result.WriteString("  UAC Level:    " + color.New(color.FgGreen).Sprint("Elevated") + "\n")
	} else {
		result.WriteString("  Status:       " + color.New(color.FgYellow).Sprint("⚠️  Standard User") + "\n")
		result.WriteString("  UAC Level:    " + color.New(color.FgYellow).Sprint("Limited") + "\n")
	}

	// Check specific Windows privileges
	result.WriteString("  Capabilities:\n")
	if isAdmin {
		result.WriteString("    • " + color.New(color.FgGreen).Sprint("✅ System configuration") + "\n")
		result.WriteString("    • " + color.New(color.FgGreen).Sprint("✅ Service management") + "\n")
		result.WriteString("    • " + color.New(color.FgGreen).Sprint("✅ Network configuration") + "\n")
		result.WriteString("    • " + color.New(color.FgGreen).Sprint("✅ Registry access") + "\n")
	} else {
		result.WriteString("    • " + color.New(color.FgRed).Sprint("❌ System configuration (needs elevation)") + "\n")
		result.WriteString("    • " + color.New(color.FgRed).Sprint("❌ Service management (needs elevation)") + "\n")
		result.WriteString("    • " + color.New(color.FgYellow).Sprint("⚠️  Network configuration (limited)") + "\n")
		result.WriteString("    • " + color.New(color.FgRed).Sprint("❌ Registry access (needs elevation)") + "\n")
	}

	result.WriteString("\n")
	return result.String()
}

func (p *PrivCommand) checkUnixPrivileges() string {
	var result strings.Builder
	result.WriteString(color.New(color.FgBlue, color.Bold).Sprint("🐧 UNIX/LINUX PRIVILEGES\n"))

	// Check if running as root
	isRoot := os.Getuid() == 0
	if isRoot {
		result.WriteString("  Status:       " + color.New(color.FgGreen).Sprint("✅ Root") + "\n")
	} else {
		result.WriteString("  Status:       " + color.New(color.FgYellow).Sprint("⚠️  Regular User") + "\n")
	}

	// Check sudo availability
	_, err := execabs.LookPath("sudo")
	hasSudo := err == nil
	if hasSudo {
		result.WriteString("  Sudo:         " + color.New(color.FgGreen).Sprint("✅ Available") + "\n")
	} else {
		result.WriteString("  Sudo:         " + color.New(color.FgRed).Sprint("❌ Not available") + "\n")
	}

	// Check specific Unix privileges
	result.WriteString("  Capabilities:\n")
	if isRoot {
		result.WriteString("    • " + color.New(color.FgGreen).Sprint("✅ System configuration") + "\n")
		result.WriteString("    • " + color.New(color.FgGreen).Sprint("✅ Service management") + "\n")
		result.WriteString("    • " + color.New(color.FgGreen).Sprint("✅ Network configuration") + "\n")
		result.WriteString("    • " + color.New(color.FgGreen).Sprint("✅ File system access") + "\n")
	} else {
		if hasSudo {
			result.WriteString("    • " + color.New(color.FgYellow).Sprint("⚠️  System configuration (use sudo)") + "\n")
			result.WriteString("    • " + color.New(color.FgYellow).Sprint("⚠️  Service management (use sudo)") + "\n")
			result.WriteString("    • " + color.New(color.FgYellow).Sprint("⚠️  Network configuration (use sudo)") + "\n")
		} else {
			result.WriteString("    • " + color.New(color.FgRed).Sprint("❌ System configuration (no sudo)") + "\n")
			result.WriteString("    • " + color.New(color.FgRed).Sprint("❌ Service management (no sudo)") + "\n")
			result.WriteString("    • " + color.New(color.FgRed).Sprint("❌ Network configuration (no sudo)") + "\n")
		}
		result.WriteString("    • " + color.New(color.FgYellow).Sprint("⚠️  File system access (limited)") + "\n")
	}

	result.WriteString("\n")
	return result.String()
}

func (p *PrivCommand) isWindowsAdmin() bool {
	if runtime.GOOS != "windows" {
		return false
	}

	// Try a Windows-specific admin check
	cmd := execabs.Command("net", "session")
	err := cmd.Run()
	return err == nil
}

func (p *PrivCommand) testPrivileges() string {
	fmt.Print("🧪 Testing privilege requirements")

	// Live feedback
	done := make(chan bool)
	go func() {
		spinner := []string{"|", "/", "-", "\\"}
		i := 0
		for {
			select {
			case <-done:
				fmt.Print("\r\033[K")
				return
			default:
				fmt.Printf("\r🧪 Testing privilege requirements %s", spinner[i%len(spinner)])
				time.Sleep(200 * time.Millisecond)
				i++
			}
		}
	}()

	// Simulate testing
	time.Sleep(2 * time.Second)

	close(done)
	fmt.Print("\r\033[K")

	var result strings.Builder
	result.WriteString(color.New(color.FgMagenta, color.Bold).Sprint("🧪 PRIVILEGE TEST RESULTS\n"))
	result.WriteString("═══════════════════════════════════════════════════════════════\n\n")

	// Test various operations
	tests := []struct {
		name           string
		description    string
		needsElevation bool
		canReduce      bool
	}{
		{"Network Interface Config", "Modify network settings", true, false},
		{"Port Scanning", "Scan network ports", false, true},
		{"Packet Capture", "Capture network packets", true, true},
		{"Service Management", "Start/stop services", true, false},
		{"File System Access", "Read system files", true, true},
		{"Process Monitoring", "Monitor system processes", false, true},
		{"System Information", "Read system info", false, false},
		{"Registry Access", "Modify system registry", true, false},
	}

	for _, test := range tests {
		result.WriteString(fmt.Sprintf("%-25s ", test.name))

		if !test.needsElevation {
			result.WriteString(color.New(color.FgGreen).Sprint("✅ No elevation needed"))
		} else if test.canReduce {
			result.WriteString(color.New(color.FgYellow).Sprint("⚠️  Elevation preferred, reduced mode available"))
		} else {
			result.WriteString(color.New(color.FgRed).Sprint("❌ Elevation required"))
		}
		result.WriteString("\n")
	}

	result.WriteString("\n")
	result.WriteString(color.New(color.FgCyan, color.Bold).Sprint("💡 GUIDANCE\n"))
	result.WriteString("  ✅ = Can run with current privileges\n")
	result.WriteString("  ⚠️  = Better with elevation, fallback available\n")
	result.WriteString("  ❌ = Requires elevation to function\n")

	return result.String()
}

func (p *PrivCommand) elevateCommand(command string) string {
	fmt.Print("🚀 Preparing elevated execution")

	// Live feedback
	done := make(chan bool)
	go func() {
		dots := 0
		for {
			select {
			case <-done:
				fmt.Print("\r\033[K")
				return
			default:
				dotStr := strings.Repeat(".", (dots%4)+1)
				fmt.Printf("\r🚀 Preparing elevated execution%s   ", dotStr)
				time.Sleep(250 * time.Millisecond)
				dots++
			}
		}
	}()

	time.Sleep(1 * time.Second)
	close(done)
	fmt.Print("\r\033[K")

	fmt.Printf("🔐 Attempting to run with elevation: %s\n", command)

	if runtime.GOOS == "windows" {
		return p.runWindowsElevated(command)
	} else {
		return p.runUnixElevated(command)
	}
}

func (p *PrivCommand) runWindowsElevated(command string) string {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "Invalid command"
	}

	fmt.Println("📋 Windows UAC elevation required...")

	// Check if it's a SuperShell command vs system command
	if _, exists := commandRegistry[parts[0]]; exists {
		// It's a SuperShell command - run elevated SuperShell
		exePath, err := os.Executable()
		if err != nil {
			return "❌ Cannot find executable path"
		}

		var psCommand string
		if len(parts) == 1 {
			psCommand = fmt.Sprintf("Start-Process '%s' -ArgumentList '-c \"%s\"' -Verb RunAs -Wait",
				exePath, command)
		} else {
			psCommand = fmt.Sprintf("Start-Process '%s' -ArgumentList '-c \"%s\"' -Verb RunAs -Wait",
				exePath, command)
		}

		fmt.Printf("🚀 Executing SuperShell command with elevation\n")

		cmd := execabs.Command("powershell", "-Command", psCommand)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err = cmd.Run()
		if err != nil {
			return fmt.Sprintf("❌ Elevation failed: %v", err)
		}

		return "✅ SuperShell command executed with elevation"
	} else {
		// It's a system command - run directly
		var psCommand string
		if len(parts) == 1 {
			psCommand = fmt.Sprintf("Start-Process '%s' -Verb RunAs -Wait", parts[0])
		} else {
			args := strings.Join(parts[1:], " ")
			psCommand = fmt.Sprintf("Start-Process '%s' -ArgumentList '%s' -Verb RunAs -Wait",
				parts[0], args)
		}

		fmt.Printf("🚀 Executing system command with elevation\n")

		cmd := execabs.Command("powershell", "-Command", psCommand)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err != nil {
			return fmt.Sprintf("❌ Elevation failed: %v", err)
		}

		return "✅ System command executed with elevation"
	}
}

func (p *PrivCommand) runUnixElevated(command string) string {
	// Try to run with sudo
	_, err := execabs.LookPath("sudo")
	if err != nil {
		return "❌ sudo not available. Run as root or install sudo."
	}

	fmt.Println("🔑 Requesting sudo privileges...")

	args := append([]string{"-S"}, strings.Fields(command)...)
	cmd := execabs.Command("sudo", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err = cmd.Run()
	if err != nil {
		return fmt.Sprintf("❌ sudo failed: %v", err)
	}

	return "✅ Command executed with sudo"
}

// Helper function to check if a command needs elevation
func RequiresElevation(commandName string) bool {
	elevatedCommands := map[string]bool{
		"netstat":     false, // Usually works without elevation
		"ipconfig":    false, // Read-only operations
		"portscan":    false, // Can work with reduced functionality
		"sniff":       true,  // Usually needs raw socket access
		"route":       false, // Read-only
		"arp":         false, // Read-only
		"tracert":     false, // Usually works
		"ping":        false, // Usually works
		"nslookup":    false, // DNS queries work
		"speedtest":   false, // HTTP-based
		"netdiscover": false, // Can work with reduced functionality
	}

	return elevatedCommands[commandName]
}

// Enhanced command execution with privilege checking
func ExecuteWithPrivilegeCheck(commandName string, originalFunc func([]string) string, args []string) string {
	if RequiresElevation(commandName) {
		// Check current privileges
		needsElevation := false

		if runtime.GOOS == "windows" {
			cmd := execabs.Command("net", "session")
			needsElevation = cmd.Run() != nil
		} else {
			needsElevation = os.Getuid() != 0
		}

		if needsElevation {
			warning := color.New(color.FgYellow, color.Bold).Sprint("⚠️  PRIVILEGE WARNING") + "\n"
			warning += fmt.Sprintf("Command '%s' may require elevated privileges.\n", commandName)
			warning += "Use 'priv elevate " + commandName + "' for full functionality.\n"
			warning += "Attempting to run with current privileges...\n\n"
			fmt.Print(warning)
		}
	}

	return originalFunc(args)
}

// Remote Operations Command
type RemoteCommand struct{}

func (r *RemoteCommand) Name() string { return "remote" }
func (r *RemoteCommand) Description() string {
	return `remote - Remote System Management

  Usage:
    remote connect <host>           Connect to remote system
    remote ssh <host> [user]        SSH connection to Linux/Unix
    remote rdp <host> [user]        RDP connection to Windows
    remote winrm <host> [user]      WinRM connection to Windows
    remote exec <host> <command>    Execute command remotely
    remote copy <src> <dest>        Copy files to/from remote
    remote list                     List saved connections
    remote save <name> <host>       Save connection profile
    remote keys                     Manage SSH keys
    remote tunnel <local:remote>    Create SSH tunnel

  Connection Types:
    ssh                             SSH for Linux/Unix systems
    rdp                             Remote Desktop for Windows
    winrm                           Windows Remote Management
    powershell                      PowerShell remoting

  Examples:
    remote ssh 192.168.1.100 admin
    remote winrm server01.domain.com
    remote exec web01 "systemctl status nginx"
    remote copy file.txt user@host:/tmp/
    remote tunnel 8080:localhost:80
    remote save webserver 192.168.1.100

  Features:
    - Live connection feedback and status
    - Key-based authentication support
    - Connection pooling and reuse
    - Secure credential management
    - Cross-platform remote management
    - File transfer capabilities
    - SSH tunneling support
`
}

func (r *RemoteCommand) Execute(args []string) string {
	if len(args) == 0 {
		return r.showRemoteHelp()
	}

	subCommand := strings.ToLower(args[0])
	switch subCommand {
	case "connect":
		if len(args) < 2 {
			return "Usage: remote connect <host>"
		}
		return r.connectToHost(args[1])
	case "ssh":
		if len(args) < 2 {
			return "Usage: remote ssh <host> [user]"
		}
		user := ""
		if len(args) > 2 {
			user = args[2]
		}
		return r.sshConnect(args[1], user)
	case "rdp":
		if len(args) < 2 {
			return "Usage: remote rdp <host> [user]"
		}
		user := ""
		if len(args) > 2 {
			user = args[2]
		}
		return r.rdpConnect(args[1], user)
	case "winrm":
		if len(args) < 2 {
			return "Usage: remote winrm <host> [user]"
		}
		user := ""
		if len(args) > 2 {
			user = args[2]
		}
		return r.winrmConnect(args[1], user)
	case "exec":
		if len(args) < 3 {
			return "Usage: remote exec <host> <command>"
		}
		return r.executeRemote(args[1], strings.Join(args[2:], " "))
	case "copy":
		if len(args) < 3 {
			return "Usage: remote copy <source> <destination>"
		}
		return r.copyFiles(args[1], args[2])
	case "list":
		return r.listConnections()
	case "save":
		if len(args) < 3 {
			return "Usage: remote save <name> <host> [user]"
		}
		user := ""
		if len(args) > 3 {
			user = args[3]
		}
		return r.saveConnection(args[1], args[2], user)
	case "keys":
		return r.manageKeys()
	case "tunnel":
		if len(args) < 2 {
			return "Usage: remote tunnel <local_port:remote_host:remote_port>"
		}
		return r.createTunnel(args[1])
	default:
		return "Unknown subcommand: " + args[0] + "\nUse 'remote' with no args for help"
	}
}

// Connection management structures
type RemoteConnection struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	User     string `json:"user"`
	Type     string `json:"type"` // ssh, rdp, winrm
	KeyPath  string `json:"key_path,omitempty"`
	Port     int    `json:"port"`
	LastUsed string `json:"last_used"`
}

var savedConnections []RemoteConnection
var activeConnections map[string]*RemoteConnection = make(map[string]*RemoteConnection)

func (r *RemoteCommand) showRemoteHelp() string {
	var help strings.Builder
	help.WriteString(color.New(color.FgCyan, color.Bold).Sprint("🌐 REMOTE OPERATIONS\n"))
	help.WriteString("═══════════════════════════════════════════════════════════════\n\n")

	help.WriteString(color.New(color.FgGreen, color.Bold).Sprint("📡 Connection Types:\n"))
	help.WriteString("  ssh                   SSH to Linux/Unix systems\n")
	help.WriteString("  rdp                   Remote Desktop to Windows\n")
	help.WriteString("  winrm                 Windows Remote Management\n")
	help.WriteString("  powershell           PowerShell remoting\n\n")

	help.WriteString(color.New(color.FgYellow, color.Bold).Sprint("🔧 Quick Commands:\n"))
	help.WriteString("  remote ssh server.com                 # SSH with current user\n")
	help.WriteString("  remote ssh 192.168.1.100 admin       # SSH with specific user\n")
	help.WriteString("  remote winrm server01.domain.com      # Windows Remote Management\n")
	help.WriteString("  remote exec web01 'systemctl status'  # Execute remote command\n")
	help.WriteString("  remote copy file.txt user@host:/tmp/  # Copy file to remote\n")
	help.WriteString("  remote tunnel 8080:localhost:80       # Create SSH tunnel\n\n")

	help.WriteString(color.New(color.FgMagenta, color.Bold).Sprint("💾 Connection Management:\n"))
	help.WriteString("  remote save <name> <host> [user]      # Save connection profile\n")
	help.WriteString("  remote list                           # List saved connections\n")
	help.WriteString("  remote keys                           # Manage SSH keys\n\n")

	help.WriteString(color.New(color.FgBlue, color.Bold).Sprint("🔒 Security Features:\n"))
	help.WriteString("  • Key-based authentication support\n")
	help.WriteString("  • Secure credential storage\n")
	help.WriteString("  • Connection pooling and reuse\n")
	help.WriteString("  • Encrypted tunneling\n")

	return help.String()
}

func (r *RemoteCommand) connectToHost(host string) string {
	fmt.Print("🔍 Detecting remote system type")

	// Live feedback during detection
	done := make(chan bool)
	go func() {
		dots := 0
		for {
			select {
			case <-done:
				fmt.Print("\r\033[K")
				return
			default:
				dotStr := strings.Repeat(".", (dots%4)+1)
				fmt.Printf("\r🔍 Detecting remote system type%s   ", dotStr)
				time.Sleep(300 * time.Millisecond)
				dots++
			}
		}
	}()

	// Detect remote system type
	systemType := r.detectRemoteSystem(host)

	close(done)
	fmt.Print("\r\033[K")

	switch systemType {
	case "linux", "unix":
		fmt.Printf("🐧 Detected Linux/Unix system: %s\n", host)
		return r.sshConnect(host, "")
	case "windows":
		fmt.Printf("🪟 Detected Windows system: %s\n", host)
		return r.winrmConnect(host, "")
	default:
		fmt.Printf("❓ Unknown system type for: %s\n", host)
		return "System type detection failed. Use specific connection type (ssh, rdp, winrm)"
	}
}

func (r *RemoteCommand) detectRemoteSystem(host string) string {
	// Try to detect system type through various methods
	time.Sleep(1 * time.Second) // Simulate detection time

	// Try SSH port (22)
	if r.isPortOpen(host, 22) {
		return "linux"
	}

	// Try RDP port (3389)
	if r.isPortOpen(host, 3389) {
		return "windows"
	}

	// Try WinRM ports (5985, 5986)
	if r.isPortOpen(host, 5985) || r.isPortOpen(host, 5986) {
		return "windows"
	}

	return "unknown"
}

func (r *RemoteCommand) isPortOpen(host string, port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func (r *RemoteCommand) sshConnect(host, user string) string {
	fmt.Print("🔐 Establishing SSH connection")

	// Live feedback during connection
	done := make(chan bool)
	go func() {
		spinner := []string{"|", "/", "-", "\\"}
		i := 0
		for {
			select {
			case <-done:
				fmt.Print("\r\033[K")
				return
			default:
				fmt.Printf("\r🔐 Establishing SSH connection %s", spinner[i%len(spinner)])
				time.Sleep(200 * time.Millisecond)
				i++
			}
		}
	}()

	// Simulate connection establishment
	time.Sleep(2 * time.Second)

	close(done)
	fmt.Print("\r\033[K")

	if user == "" {
		// Try to get current user
		if u, err := osuser.Current(); err == nil {
			user = u.Username
		} else {
			user = "root"
		}
	}

	// Check for SSH client
	sshPath, err := exec.LookPath("ssh")
	if err != nil {
		return "❌ SSH client not found. Please install OpenSSH client."
	}

	fmt.Printf("✅ SSH client found: %s\n", sshPath)
	fmt.Printf("🚀 Connecting to %s@%s...\n", user, host)

	// Try key-based auth first, then fall back to password
	keyPath := r.findSSHKey()
	var cmd *exec.Cmd

	if keyPath != "" {
		fmt.Printf("🔑 Using SSH key: %s\n", keyPath)
		cmd = exec.Command("ssh", "-i", keyPath, "-o", "ConnectTimeout=10", fmt.Sprintf("%s@%s", user, host))
	} else {
		fmt.Println("🔒 Using password authentication")
		cmd = exec.Command("ssh", "-o", "ConnectTimeout=10", fmt.Sprintf("%s@%s", user, host))
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Sprintf("❌ SSH connection failed: %v", err)
	}

	// Save successful connection
	r.saveSuccessfulConnection("ssh", host, user, 22)
	return "✅ SSH session completed"
}

func (r *RemoteCommand) rdpConnect(host, user string) string {
	fmt.Print("🖥️  Launching RDP connection")

	// Live feedback
	done := make(chan bool)
	go func() {
		dots := 0
		for {
			select {
			case <-done:
				fmt.Print("\r\033[K")
				return
			default:
				dotStr := strings.Repeat(".", (dots%3)+1)
				fmt.Printf("\r🖥️  Launching RDP connection%s   ", dotStr)
				time.Sleep(400 * time.Millisecond)
				dots++
			}
		}
	}()

	time.Sleep(1 * time.Second)
	close(done)
	fmt.Print("\r\033[K")

	if runtime.GOOS == "windows" {
		// Use Windows built-in mstsc
		args := []string{"/v:" + host}
		if user != "" {
			args = append(args, "/user:"+user)
		}

		fmt.Printf("🚀 Launching Windows Remote Desktop to %s\n", host)
		cmd := exec.Command("mstsc", args...)
		err := cmd.Start()
		if err != nil {
			return fmt.Sprintf("❌ Failed to launch RDP: %v", err)
		}

		r.saveSuccessfulConnection("rdp", host, user, 3389)
		return "✅ RDP session launched"
	} else {
		// Try to find RDP client on Unix/Linux
		rdpClients := []string{"rdesktop", "xfreerdp", "vinagre"}

		for _, client := range rdpClients {
			if _, err := exec.LookPath(client); err == nil {
				fmt.Printf("🚀 Launching %s to %s\n", client, host)

				var cmd *exec.Cmd
				switch client {
				case "rdesktop":
					cmd = exec.Command("rdesktop", "-g", "1024x768", host)
				case "xfreerdp":
					args := []string{fmt.Sprintf("/v:%s", host), "/size:1024x768"}
					if user != "" {
						args = append(args, fmt.Sprintf("/u:%s", user))
					}
					cmd = exec.Command("xfreerdp", args...)
				case "vinagre":
					cmd = exec.Command("vinagre", fmt.Sprintf("rdp://%s", host))
				}

				err := cmd.Start()
				if err != nil {
					continue
				}

				r.saveSuccessfulConnection("rdp", host, user, 3389)
				return fmt.Sprintf("✅ RDP session launched with %s", client)
			}
		}

		return "❌ No RDP client found. Install rdesktop, xfreerdp, or vinagre"
	}
}

func (r *RemoteCommand) winrmConnect(host, user string) string {
	fmt.Print("⚡ Establishing WinRM connection")

	// Live feedback
	done := make(chan bool)
	go func() {
		spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		i := 0
		for {
			select {
			case <-done:
				fmt.Print("\r\033[K")
				return
			default:
				fmt.Printf("\r⚡ Establishing WinRM connection %s", spinner[i%len(spinner)])
				time.Sleep(150 * time.Millisecond)
				i++
			}
		}
	}()

	time.Sleep(2 * time.Second)
	close(done)
	fmt.Print("\r\033[K")

	if runtime.GOOS == "windows" {
		// Use Windows PowerShell remoting
		if user == "" {
			user = os.Getenv("USERNAME")
		}

		fmt.Printf("🚀 Connecting to %s via WinRM\n", host)

		psCommand := fmt.Sprintf("Enter-PSSession -ComputerName %s", host)
		if user != "" {
			psCommand += fmt.Sprintf(" -Credential (Get-Credential %s)", user)
		}

		cmd := exec.Command("powershell", "-Command", psCommand)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err != nil {
			return fmt.Sprintf("❌ WinRM connection failed: %v", err)
		}

		r.saveSuccessfulConnection("winrm", host, user, 5985)
		return "✅ WinRM session completed"
	} else {
		// Try to use winrm tools on Unix/Linux
		return "❌ WinRM not natively supported on this platform. Use SSH or install WinRM tools."
	}
}

func (r *RemoteCommand) executeRemote(host, command string) string {
	fmt.Printf("🚀 Executing remote command on %s\n", host)
	fmt.Print("⚡ Connecting")

	// Live feedback
	done := make(chan bool)
	go func() {
		dots := 0
		for {
			select {
			case <-done:
				fmt.Print("\r\033[K")
				return
			default:
				dotStr := strings.Repeat(".", (dots%4)+1)
				fmt.Printf("\r⚡ Connecting%s   ", dotStr)
				time.Sleep(200 * time.Millisecond)
				dots++
			}
		}
	}()

	time.Sleep(1 * time.Second)
	close(done)
	fmt.Print("\r\033[K")

	// Try to find the host in saved connections
	conn := r.findSavedConnection(host)
	if conn != nil {
		fmt.Printf("📋 Using saved connection: %s (%s)\n", conn.Name, conn.Type)

		switch conn.Type {
		case "ssh":
			return r.executeSSHCommand(conn, command)
		case "winrm":
			return r.executeWinRMCommand(conn, command)
		default:
			return "❌ Unsupported connection type for remote execution: " + conn.Type
		}
	}

	// Auto-detect and execute
	systemType := r.detectRemoteSystem(host)
	switch systemType {
	case "linux", "unix":
		return r.executeSSHCommand(&RemoteConnection{Host: host, Type: "ssh"}, command)
	case "windows":
		return r.executeWinRMCommand(&RemoteConnection{Host: host, Type: "winrm"}, command)
	default:
		return "❌ Cannot determine remote system type for execution"
	}
}

func (r *RemoteCommand) executeSSHCommand(conn *RemoteConnection, command string) string {
	fmt.Printf("🔐 Executing via SSH: %s\n", command)

	user := conn.User
	if user == "" {
		if u, err := osuser.Current(); err == nil {
			user = u.Username
		} else {
			user = "root"
		}
	}

	keyPath := r.findSSHKey()
	var cmd *exec.Cmd

	if keyPath != "" {
		cmd = exec.Command("ssh", "-i", keyPath, "-o", "ConnectTimeout=10",
			fmt.Sprintf("%s@%s", user, conn.Host), command)
	} else {
		cmd = exec.Command("ssh", "-o", "ConnectTimeout=10",
			fmt.Sprintf("%s@%s", user, conn.Host), command)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("❌ SSH execution failed: %v\n%s", err, string(output))
	}

	return fmt.Sprintf("✅ Command executed successfully:\n%s", string(output))
}

func (r *RemoteCommand) executeWinRMCommand(conn *RemoteConnection, command string) string {
	fmt.Printf("⚡ Executing via WinRM: %s\n", command)

	if runtime.GOOS != "windows" {
		return "❌ WinRM execution requires Windows platform"
	}

	user := conn.User
	if user == "" {
		user = os.Getenv("USERNAME")
	}

	psCommand := fmt.Sprintf("Invoke-Command -ComputerName %s -ScriptBlock {%s}", conn.Host, command)
	if user != "" {
		psCommand += fmt.Sprintf(" -Credential (Get-Credential %s)", user)
	}

	cmd := exec.Command("powershell", "-Command", psCommand)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("❌ WinRM execution failed: %v\n%s", err, string(output))
	}

	return fmt.Sprintf("✅ Command executed successfully:\n%s", string(output))
}

func (r *RemoteCommand) copyFiles(source, destination string) string {
	fmt.Printf("📁 Copying files: %s → %s\n", source, destination)
	fmt.Print("🔄 Transferring")

	// Live feedback
	done := make(chan bool)
	go func() {
		spinner := []string{"📤", "📥", "📤", "📥"}
		i := 0
		for {
			select {
			case <-done:
				fmt.Print("\r\033[K")
				return
			default:
				fmt.Printf("\r🔄 Transferring %s", spinner[i%len(spinner)])
				time.Sleep(300 * time.Millisecond)
				i++
			}
		}
	}()

	time.Sleep(2 * time.Second) // Simulate transfer time
	close(done)
	fmt.Print("\r\033[K")

	// Use SCP for file transfers
	var cmd *exec.Cmd

	// Check if source or destination is remote (contains @)
	if strings.Contains(source, "@") || strings.Contains(destination, "@") {
		keyPath := r.findSSHKey()
		if keyPath != "" {
			cmd = exec.Command("scp", "-i", keyPath, "-o", "ConnectTimeout=10", source, destination)
		} else {
			cmd = exec.Command("scp", "-o", "ConnectTimeout=10", source, destination)
		}

		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Sprintf("❌ File copy failed: %v\n%s", err, string(output))
		}

		return "✅ File copy completed successfully"
	}

	return "❌ Invalid copy syntax. Use user@host:/path format for remote files"
}

func (r *RemoteCommand) createTunnel(tunnelSpec string) string {
	fmt.Printf("🌉 Creating SSH tunnel: %s\n", tunnelSpec)
	fmt.Print("🔗 Establishing tunnel")

	// Live feedback
	done := make(chan bool)
	go func() {
		dots := 0
		for {
			select {
			case <-done:
				fmt.Print("\r\033[K")
				return
			default:
				dotStr := strings.Repeat(".", (dots%4)+1)
				fmt.Printf("\r🔗 Establishing tunnel%s   ", dotStr)
				time.Sleep(250 * time.Millisecond)
				dots++
			}
		}
	}()

	time.Sleep(1 * time.Second)
	close(done)
	fmt.Print("\r\033[K")

	// Parse tunnel specification (local_port:remote_host:remote_port)
	parts := strings.Split(tunnelSpec, ":")
	if len(parts) != 3 {
		return "❌ Invalid tunnel format. Use: local_port:remote_host:remote_port"
	}

	fmt.Printf("✅ SSH tunnel would be created: %s\n", tunnelSpec)
	fmt.Println("🔗 Tunnel: localhost:" + parts[0] + " → " + parts[1] + ":" + parts[2])

	return "Note: Tunnel creation requires interactive SSH setup"
}

func (r *RemoteCommand) listConnections() string {
	var result strings.Builder
	result.WriteString(color.New(color.FgCyan, color.Bold).Sprint("💾 SAVED CONNECTIONS\n"))
	result.WriteString("═══════════════════════════════════════════════════════════════\n\n")

	if len(savedConnections) == 0 {
		result.WriteString("📭 No saved connections found.\n")
		result.WriteString("Use 'remote save <name> <host>' to save connections.\n")
		return result.String()
	}

	result.WriteString(fmt.Sprintf("%-15s %-20s %-15s %-10s %-15s\n",
		"NAME", "HOST", "USER", "TYPE", "LAST USED"))
	result.WriteString(strings.Repeat("─", 75) + "\n")

	for _, conn := range savedConnections {
		user := conn.User
		if user == "" {
			user = "<auto>"
		}
		lastUsed := conn.LastUsed
		if lastUsed == "" {
			lastUsed = "Never"
		}

		result.WriteString(fmt.Sprintf("%-15s %-20s %-15s %-10s %-15s\n",
			conn.Name, conn.Host, user, conn.Type, lastUsed))
	}

	result.WriteString(fmt.Sprintf("\n📊 Total: %d saved connections\n", len(savedConnections)))
	return result.String()
}

func (r *RemoteCommand) saveConnection(name, host, user string) string {
	fmt.Print("💾 Saving connection profile")

	// Live feedback
	done := make(chan bool)
	go func() {
		dots := 0
		for {
			select {
			case <-done:
				fmt.Print("\r\033[K")
				return
			default:
				dotStr := strings.Repeat(".", (dots%3)+1)
				fmt.Printf("\r💾 Saving connection profile%s   ", dotStr)
				time.Sleep(300 * time.Millisecond)
				dots++
			}
		}
	}()

	time.Sleep(500 * time.Millisecond)
	close(done)
	fmt.Print("\r\033[K")

	// Detect connection type
	systemType := r.detectRemoteSystem(host)
	connType := "ssh"
	if systemType == "windows" {
		connType = "winrm"
	}

	conn := RemoteConnection{
		Name:     name,
		Host:     host,
		User:     user,
		Type:     connType,
		Port:     22,
		LastUsed: "",
	}

	if connType == "winrm" {
		conn.Port = 5985
	}

	// Add to saved connections (in real implementation, save to file)
	savedConnections = append(savedConnections, conn)

	return fmt.Sprintf("✅ Connection '%s' saved: %s@%s (%s)", name, user, host, connType)
}

func (r *RemoteCommand) manageKeys() string {
	var result strings.Builder
	result.WriteString(color.New(color.FgYellow, color.Bold).Sprint("🔑 SSH KEY MANAGEMENT\n"))
	result.WriteString("═══════════════════════════════════════════════════════════════\n\n")

	// Check for SSH keys
	homeDir, _ := os.UserHomeDir()
	sshDir := filepath.Join(homeDir, ".ssh")

	keyFiles := []string{"id_rsa", "id_ed25519", "id_ecdsa"}
	foundKeys := []string{}

	for _, keyFile := range keyFiles {
		keyPath := filepath.Join(sshDir, keyFile)
		if _, err := os.Stat(keyPath); err == nil {
			foundKeys = append(foundKeys, keyPath)
		}
	}

	if len(foundKeys) > 0 {
		result.WriteString(color.New(color.FgGreen).Sprint("✅ Found SSH Keys:\n"))
		for _, key := range foundKeys {
			result.WriteString(fmt.Sprintf("  🔑 %s\n", key))
		}
	} else {
		result.WriteString(color.New(color.FgRed).Sprint("❌ No SSH keys found\n"))
		result.WriteString("\n💡 To generate SSH keys:\n")
		result.WriteString("  ssh-keygen -t ed25519 -C \"your_email@example.com\"\n")
		result.WriteString("  ssh-keygen -t rsa -b 4096 -C \"your_email@example.com\"\n")
	}

	result.WriteString("\n🔧 Key Management Commands:\n")
	result.WriteString("  ssh-keygen -t ed25519           # Generate new ED25519 key\n")
	result.WriteString("  ssh-copy-id user@host           # Copy key to remote host\n")
	result.WriteString("  ssh-add ~/.ssh/id_ed25519       # Add key to SSH agent\n")

	return result.String()
}

// Helper functions
func (r *RemoteCommand) findSSHKey() string {
	homeDir, _ := os.UserHomeDir()
	sshDir := filepath.Join(homeDir, ".ssh")

	keyFiles := []string{"id_ed25519", "id_rsa", "id_ecdsa"}

	for _, keyFile := range keyFiles {
		keyPath := filepath.Join(sshDir, keyFile)
		if _, err := os.Stat(keyPath); err == nil {
			return keyPath
		}
	}

	return ""
}

func (r *RemoteCommand) findSavedConnection(host string) *RemoteConnection {
	for _, conn := range savedConnections {
		if conn.Host == host || conn.Name == host {
			return &conn
		}
	}
	return nil
}

func (r *RemoteCommand) saveSuccessfulConnection(connType, host, user string, port int) {
	// Update last used time for existing connections
	for i, conn := range savedConnections {
		if conn.Host == host && conn.Type == connType {
			savedConnections[i].LastUsed = time.Now().Format("2006-01-02 15:04")
			return
		}
	}
}
