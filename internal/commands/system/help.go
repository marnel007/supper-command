package system

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"suppercommand/internal/commands"
)

// HelpCommand shows available commands
type HelpCommand struct {
	*commands.BaseCommand
	registry *commands.Registry
}

// NewHelpCommand creates a new help command
func NewHelpCommand(registry *commands.Registry) *HelpCommand {
	return &HelpCommand{
		BaseCommand: commands.NewBaseCommand(
			"help",
			"Show available commands and their descriptions",
			"help [command]",
			[]string{"windows", "linux", "darwin"},
			false,
		),
		registry: registry,
	}
}

// Execute shows help information
func (h *HelpCommand) Execute(ctx context.Context, args *commands.Arguments) (*commands.Result, error) {
	startTime := time.Now()

	// Check if help for a specific command is requested
	if len(args.Raw) > 0 {
		return h.showCommandHelp(args.Raw[0], startTime)
	}

	// Show general help
	return h.showGeneralHelp(startTime)
}

// showGeneralHelp displays all available commands
func (h *HelpCommand) showGeneralHelp(startTime time.Time) (*commands.Result, error) {
	var output strings.Builder
	output.WriteString("SuperShell - Available Commands:\n")
	output.WriteString("================================\n\n")

	// Get all command names and sort them
	commandNames := h.registry.List()
	sort.Strings(commandNames)

	for _, name := range commandNames {
		cmd, err := h.registry.Get(name)
		if err != nil {
			continue
		}

		output.WriteString(fmt.Sprintf("  %-12s %s\n", name, cmd.Description()))
	}

	output.WriteString("\nUse 'help <command>' for detailed information about a specific command.\n")

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// showCommandHelp displays detailed help for a specific command
func (h *HelpCommand) showCommandHelp(commandName string, startTime time.Time) (*commands.Result, error) {
	cmd, err := h.registry.Get(commandName)
	if err != nil {
		return &commands.Result{
			Output:   fmt.Sprintf("Command '%s' not found. Use 'help' to see all available commands.\n", commandName),
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("SuperShell - Command Help: %s\n", commandName))
	output.WriteString("================================\n\n")

	output.WriteString(fmt.Sprintf("Name:        %s\n", cmd.Name()))
	output.WriteString(fmt.Sprintf("Description: %s\n", cmd.Description()))
	output.WriteString(fmt.Sprintf("Usage:       %s\n", cmd.Usage()))

	platforms := cmd.SupportedPlatforms()
	if len(platforms) > 0 {
		output.WriteString(fmt.Sprintf("Platforms:   %s\n", strings.Join(platforms, ", ")))
	}

	if cmd.RequiresElevation() {
		output.WriteString("Privileges:  Requires administrator/root privileges\n")
	}

	// Add detailed help for specific commands
	output.WriteString("\n")
	output.WriteString(h.getDetailedHelp(commandName))

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// getDetailedHelp returns detailed help text for specific commands
func (h *HelpCommand) getDetailedHelp(commandName string) string {
	switch commandName {
	case "sniff":
		return `Detailed Options:
  -i, --interface <name>    Network interface to monitor (default: eth0)
  -c, --count <number>      Number of packets to capture (default: 10)
  -p, --protocol <proto>    Filter by protocol (TCP, UDP, HTTP, HTTPS, DNS, SSH, FTP, etc.)
  -s, --source <ip>         Filter by source IP address
  -d, --dest <ip>           Filter by destination IP address
  --port <port>             Filter by port number (matches source or destination)
  -v, --verbose             Show detailed packet information
  --hex                     Display hexadecimal payload dump
  --save <file>             Save capture to file
  --continuous              Continuous capture mode
  -t, --timeout <seconds>   Capture timeout for continuous mode

Examples:
  sniff -c 10                           # Capture 10 packets
  sniff -p HTTP -v                      # Capture HTTP packets with details
  sniff -s 192.168.1.100 --hex         # Capture from specific IP with hex dump
  sniff --port 80 -c 5                 # Capture packets on port 80
  sniff -p TCP -d 8.8.8.8 --save cap.pcap  # Capture TCP to 8.8.8.8 and save
`

	case "wget":
		return `Detailed Options:
  -v, --verbose             Show detailed download progress and information
  <url>                     URL to download from
  [filename]                Optional filename (auto-detected if not provided)

Examples:
  wget https://example.com/file.zip     # Download file
  wget -v https://api.github.com/users  # Download with verbose output
  wget https://example.com/data.json mydata.json  # Download with custom name
`

	case "arp":
		return `Detailed Options:
  -a, --all                 Show all ARP entries
  -d, --delete <ip>         Delete ARP entry for specified IP
  [ip_address]              Show ARP entry for specific IP

Examples:
  arp -a                    # Show all ARP entries
  arp 192.168.1.1           # Show ARP entry for specific IP
  arp -d 192.168.1.100      # Delete ARP entry
`

	case "route":
		return `Detailed Options:
  print, show               Display routing table (default)
  add <dest> <gateway>      Add a new route
  delete <dest> [gateway]   Delete a route
  -4, --ipv4                Show only IPv4 routes
  -6, --ipv6                Show only IPv6 routes

Examples:
  route                     # Show routing table
  route -4                  # Show only IPv4 routes
  route add 10.0.0.0/8 192.168.1.1     # Add route
  route delete 10.0.0.0/8              # Delete route
`

	case "speedtest":
		return `Detailed Options:
  -s, --simple              Simple output format
  -q, --quiet               Minimal output
  --download-only           Test download speed only
  --upload-only             Test upload speed only

Examples:
  speedtest                 # Full speed test
  speedtest -s              # Simple output format
  speedtest --download-only # Download test only
`

	case "portscan":
		return `Detailed Options:
  <host>                    Target host to scan
  [ports]                   Port range (default: common ports)
  -p, --ports <range>       Specific ports (e.g., 80,443,22-25)
  -t, --timeout <ms>        Connection timeout in milliseconds
  --top-ports <n>           Scan top N most common ports

Examples:
  portscan google.com       # Scan common ports
  portscan 192.168.1.1 -p 80,443,22    # Scan specific ports
  portscan example.com --top-ports 100 # Scan top 100 ports
`

	case "ping":
		return `Detailed Options:
  <host>                    Target host to ping
  -c, --count <number>      Number of ping packets to send
  -t, --timeout <ms>        Timeout for each ping in milliseconds
  -i, --interval <ms>       Interval between pings in milliseconds

Examples:
  ping google.com           # Ping with default settings
  ping -c 5 8.8.8.8         # Send 5 ping packets
  ping -t 2000 example.com  # 2 second timeout
`

	case "tracert":
		return `Detailed Options:
  <host>                    Target host to trace route to
  -m, --max-hops <number>   Maximum number of hops (default: 30)
  -t, --timeout <ms>        Timeout for each hop in milliseconds

Examples:
  tracert google.com        # Trace route to Google
  tracert -m 15 8.8.8.8     # Trace with max 15 hops
`

	case "nslookup":
		return `Detailed Options:
  <domain>                  Domain name to lookup
  [record_type]             DNS record type (A, AAAA, MX, NS, TXT, etc.)
  -s, --server <dns_server> Use specific DNS server

Examples:
  nslookup google.com       # Basic DNS lookup
  nslookup google.com MX    # Get MX records
  nslookup example.com -s 8.8.8.8  # Use specific DNS server
`

	case "sysinfo":
		return `Detailed Options:
  -v, --verbose             Show detailed system information
  --cpu                     Show only CPU information
  --memory                  Show only memory information
  --disk                    Show only disk information
  --network                 Show only network information

Examples:
  sysinfo                   # Basic system information
  sysinfo -v                # Detailed system information
  sysinfo --cpu             # CPU information only
`

	case "killtask":
		return `Detailed Options:
  -f, --force               Force terminate processes immediately (SIGKILL on Unix)
  -t, --tree                Terminate process tree including child processes
  <pid>                     Process ID to terminate
  <process_name>            Process name to terminate (e.g., notepad.exe)

Examples:
  killtask 1234             # Terminate process with PID 1234
  killtask notepad          # Terminate all notepad processes
  killtask -f chrome        # Force terminate all Chrome processes
  killtask -t explorer      # Terminate Explorer and all child processes
  killtask 1234 5678 notepad # Terminate multiple processes
`

	case "lookup":
		return `Detailed Options:
  -m, --menu                Show interactive dropdown-style menu
  -s, --similar             Show similar commands using fuzzy matching
  -c, --categories          Show all command categories
  -t, --task <task>         Get task-based command suggestions
  [query]                   Search term for command lookup

Examples:
  lookup -m                 # Interactive menu with dropdown navigation
  lookup ping               # Find commands related to 'ping'
  lookup network            # Find all network-related commands
  lookup -c                 # Show all command categories
  lookup -t network         # Get commands for network tasks
  lookup -t file            # Get commands for file operations
  lookup -s net             # Find similar commands to 'net'
  lookup copy               # Find commands related to copying
  lookup -t security        # Get security-related commands

Task Categories:
  network, file, system, security, monitoring

Interactive Menu Features:
  • Browse by Category - Organized command exploration
  • Task-Based Lookup - Find commands for specific goals
  • Popular Commands - Most commonly used commands
  • Search Commands - Smart search functionality
`

	case "ver":
		return `Detailed Options:
  -v, --verbose             Show detailed version information with features and runtime details

Examples:
  ver                       # Show basic version information
  ver -v                    # Show comprehensive version details with features and system info

Use Cases:
  • Version Checking - Verify SuperShell version for compatibility or support
  • System Information - Get runtime and build information for troubleshooting
  • Feature Discovery - See what features are available in your version
`

	case "help":
		return `Detailed Options:
  [command]                 Get detailed help for a specific command

Examples:
  help                      # Show all available commands with descriptions
  help ping                 # Get detailed help for the ping command
  help sniff                # Get comprehensive help for the sniff command with all options

Use Cases:
  • Command Reference - Quick reference for command syntax and options
  • Learning Tool - Learn about available commands and their capabilities
  • Troubleshooting - Get help when commands aren't working as expected
`

	case "firewall":
		return `Detailed Options:
  status                    Show current firewall status and configuration
  enable                    Enable the system firewall (requires admin privileges)
  disable                   Disable the system firewall (requires admin privileges)
  rules [subcommand]        Manage firewall rules
    list                    List all firewall rules
    add                     Add a new firewall rule (future feature)
    remove                  Remove a firewall rule (future feature)
  help                      Show firewall command help

Examples:
  firewall status           # Check if firewall is enabled and show basic info
  firewall enable           # Enable Windows Defender Firewall
  firewall disable          # Disable Windows Defender Firewall
  firewall rules list       # List all configured firewall rules
  firewall rules            # Same as 'firewall rules list'

Use Cases:
  • Security Management - Monitor and control system firewall settings
  • Rule Management - View and manage firewall rules for applications
  • Compliance Checking - Verify firewall status for security audits
  • Troubleshooting - Diagnose network connectivity issues related to firewall

Platform Support:
  • Windows: Full support via Windows Defender Firewall
  • Linux: Support via iptables/ufw (planned)
  • macOS: Support via pfctl (planned)

Note: Enabling/disabling firewall requires administrator privileges.
`

	case "perf":
		return `Detailed Options:
  analyze                   Perform comprehensive system performance analysis
  monitor                   Start real-time performance monitoring
  report                    Generate detailed performance report
  baseline [subcommand]     Manage performance baselines
    create <name>           Create a new performance baseline
    list                    List all saved baselines
    delete <name>           Delete a performance baseline
  help                      Show performance command help

Examples:
  perf analyze              # Analyze current system performance
  perf monitor              # Start real-time monitoring (Ctrl+C to stop)
  perf report               # Generate comprehensive performance report
  perf baseline create prod-baseline    # Create baseline named 'prod-baseline'
  perf baseline list        # List all saved performance baselines
  perf baseline delete old-baseline     # Delete baseline named 'old-baseline'

Use Cases:
  • Performance Monitoring - Track system resource usage over time
  • Bottleneck Detection - Identify CPU, memory, disk, or network bottlenecks
  • Capacity Planning - Understand system limits and plan for scaling
  • Troubleshooting - Diagnose performance issues and slowdowns
  • Baseline Comparison - Compare current performance against historical data

Metrics Analyzed:
  • CPU Usage - Per-core utilization and load averages
  • Memory Usage - RAM utilization, available memory, swap usage
  • Disk I/O - Read/write operations, disk utilization, queue depth
  • Network I/O - Bandwidth usage, packet rates, connection counts
  • System Load - Overall system health and responsiveness

Output Information:
  • Overall Health Status (healthy/warning/critical)
  • Resource Utilization Percentages
  • Performance Bottlenecks and Warnings
  • Optimization Suggestions
  • Historical Trend Analysis (when baselines available)
`

	case "server":
		return `Detailed Options:
  health                    Check overall server health and component status
  services [subcommand]     Manage system services
    list                    List all system services with status
    start <name>            Start a specific service
    stop <name>             Stop a specific service
    restart <name>          Restart a specific service
  users                     List all active users and sessions
  session [subcommand]      Manage user sessions
    list                    List active user sessions (same as 'users')
    kill <session>          Terminate a user session (future feature)
  alerts [subcommand]       Manage server alerts (future feature)
  backup [subcommand]       Manage server backups (future feature)
  help                      Show server command help

Examples:
  server health             # Check overall server health status
  server services list      # List all system services
  server services start "Print Spooler"    # Start the Print Spooler service
  server services stop "Windows Update"    # Stop Windows Update service
  server services restart "DNS Client"     # Restart DNS Client service
  server users              # List all active users and login times
  server session list       # Same as 'server users'

Use Cases:
  • System Administration - Monitor and manage server components
  • Service Management - Control Windows/Linux services remotely
  • User Monitoring - Track active users and sessions
  • Health Monitoring - Get real-time server health status
  • Troubleshooting - Diagnose service and user-related issues

Health Check Components:
  • CPU Usage - Current processor utilization
  • Memory Usage - RAM consumption and availability
  • Disk Usage - Storage utilization across drives
  • Network Status - Network adapter status and connectivity
  • Service Status - Critical system services health
  • Uptime - System uptime and stability metrics

Service Management:
  • Windows: Full support for Windows services via SC commands
  • Linux: Support for systemd/init services (planned)
  • Service Dependencies - Automatic handling of service dependencies
  • Status Monitoring - Real-time service status updates

Note: Service management operations require administrator privileges.
`

	case "remote":
		return `Detailed Options:
  list                      List all configured remote servers
  add <name> <user@host>    Add a new remote server configuration
    --port <port>           Specify SSH port (default: 22)
    --key <keyfile>         Use SSH key file for authentication
    --password <pass>       Use password authentication (not recommended)
  remove <name>             Remove a remote server configuration
  exec <server> <command>   Execute command on remote server
  cluster [subcommand]      Manage server clusters (future feature)
    list                    List all server clusters
    create <name>           Create a new server cluster
    delete <name>           Delete a server cluster
  sync [subcommand]         Manage configuration synchronization (future feature)
    list                    List sync profiles
    create <profile>        Create sync profile
    execute <profile>       Execute sync profile
  help                      Show remote command help

Examples:
  remote list               # List all configured remote servers
  remote add web1 admin@192.168.1.10       # Add server with default settings
  remote add db1 root@db.example.com --port 2222   # Add server with custom port
  remote add app1 deploy@app.com --key ~/.ssh/deploy_key   # Add server with SSH key
  remote exec web1 "uptime"                 # Check uptime on web1 server
  remote exec db1 "df -h"                   # Check disk usage on db1 server
  remote exec web1 "systemctl status nginx"    # Check nginx status on web1
  remote remove old-server                  # Remove server configuration

Use Cases:
  • Remote Administration - Manage multiple servers from one location
  • Command Execution - Run commands across multiple servers
  • Server Monitoring - Check status and health of remote systems
  • Deployment Management - Execute deployment scripts remotely
  • Configuration Management - Synchronize configurations across servers

Connection Methods:
  • SSH Key Authentication - Secure, passwordless authentication (recommended)
  • Password Authentication - Username/password login (less secure)
  • Custom SSH Ports - Support for non-standard SSH ports
  • Connection Pooling - Efficient connection reuse for multiple commands

Security Features:
  • Encrypted Communications - All traffic encrypted via SSH
  • Key Management - Secure SSH key storage and usage
  • Connection Validation - Verify server identity and connectivity
  • Timeout Handling - Automatic timeout for unresponsive connections

Supported Platforms:
  • Linux Servers - Full SSH support for all Linux distributions
  • Windows Servers - SSH support via OpenSSH or third-party SSH servers
  • macOS Servers - Native SSH support
  • Cloud Instances - AWS, Azure, GCP, and other cloud providers

Note: SSH client must be available on the local system. Windows 10/11 includes OpenSSH by default.
`

	case "history":
		return `🧠 Smart History System - AI-Powered Command History Management

Detailed Options:
  [no args]                 Show recent command history (last 20 commands)
  smart <query>             Intelligent search through command history using natural language
  patterns                  Display detected usage patterns and workflows
  suggest                   Get context-aware command suggestions based on your usage
  timeline                  Show chronological timeline view of command history
  stats                     Display comprehensive usage statistics and analytics
  export <format>           Export history in specified format (json, csv, txt)
  add <command>             Manually add a command to history
  clear                     Clear all command history (with confirmation)

Smart Search Examples:
  history smart "backup files"             # Find all backup-related commands
  history smart "git commit"               # Find Git commit commands
  history smart "network diagnostics"      # Find network troubleshooting commands
  history smart "system monitoring"        # Find system monitoring commands
  history smart "file operations"          # Find file management commands

Pattern Recognition:
  history patterns          # Shows detected patterns like:
                            • Sequential workflows (commands used together)
                            • Frequency patterns (most used commands)
                            • Time-based patterns (commands used at specific times)

Context-Aware Suggestions:
  history suggest           # Provides smart suggestions based on:
                            • Current working directory context
                            • Recent command patterns
                            • Time of day preferences
                            • Historical usage patterns

Timeline & Analytics:
  history timeline          # Visual timeline grouped by date with:
                            • Chronological command execution
                            • Success/failure indicators
                            • Execution time stamps
                            • Directory context

  history stats             # Comprehensive statistics including:
                            • Total commands and success rate
                            • Most frequently used commands
                            • Activity patterns by hour/day
                            • Command categorization breakdown
                            • Visual usage charts and progress bars

Export & Backup:
  history export json       # Export as JSON for programmatic analysis
  history export csv        # Export as CSV for spreadsheet analysis
  history export txt        # Export as human-readable text

Advanced Features:
  • Automatic Tracking - All commands automatically recorded with metadata
  • Smart Categorization - Commands auto-classified (filesystem, network, development, etc.)
  • Intelligent Tagging - Auto-generated relevant tags for each command
  • Context Awareness - Suggestions adapt to directory, time, and usage patterns
  • Pattern Learning - System learns from your usage to provide better recommendations
  • Visual Analytics - Color-coded output with progress bars and statistical insights
  • Cross-Session Persistence - History maintained across shell sessions
  • Performance Optimized - Fast search with indexed data and efficient storage

Use Cases:
  • Command Discovery - Find commands you used before but can't remember
  • Workflow Analysis - Understand your command usage patterns and optimize workflows
  • Knowledge Base - Build a searchable personal command reference
  • Productivity Insights - Analyze your command-line productivity and habits
  • Team Sharing - Export and share command histories with team members
  • Audit Trail - Maintain detailed logs of command execution for compliance
  • Learning Tool - Discover new commands through pattern analysis and suggestions

Storage & Privacy:
  • Local Storage - All data stored locally in ~/.supershell_history.json
  • No Cloud Sync - Complete privacy, no data sent to external servers
  • Configurable Limits - Automatic cleanup (keeps last 1000 commands by default)
  • JSON Format - Human-readable and easily portable format

The Smart History system transforms your command history from a simple list into an intelligent assistant that learns from your usage patterns and helps improve your productivity.
`

	default:
		return "No additional help available for this command.\n"
	}
}
