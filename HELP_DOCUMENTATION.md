# SuperShell Complete Help Documentation

## 📖 Overview

SuperShell is a powerful, cross-platform command-line shell with advanced system management capabilities. This documentation provides comprehensive help for all commands, organized by category with detailed examples and use cases.

## 🎯 Quick Navigation

- [🔥 Security & Firewall Commands](#-security--firewall-commands)
- [⚡ Performance Monitoring Commands](#-performance-monitoring-commands)
- [🖥️ Server Management Commands](#️-server-management-commands)
- [🌐 Remote Administration Commands](#-remote-administration-commands)
- [🌐 Network Tools](#-network-tools)
- [📁 File Operations](#-file-operations)
- [⚙️ System Information](#️-system-information)
- [🔍 Help & Discovery](#-help--discovery)
- [🧠 Smart History System](#-smart-history-system)
- [🚀 FastCP File Transfer](#-fastcp-file-transfer)

---

## 🔥 Security & Firewall Commands

### `firewall` - Firewall Management

**Purpose:** Manage system firewall settings and security policies.

**Syntax:**
```bash
firewall [command] [options]
```

**Commands:**
- `status` - Show current firewall status and configuration
- `enable` - Enable the system firewall (requires admin privileges)
- `disable` - Disable the system firewall (requires admin privileges)
- `rules list` - List all configured firewall rules
- `help` - Show detailed help

**Examples:**
```bash
# Check firewall status
firewall status

# Enable Windows Defender Firewall
firewall enable

# Disable firewall (use with caution)
firewall disable

# List all firewall rules
firewall rules list
```

**Use Cases:**
- Security compliance checking
- Network troubleshooting
- System hardening
- Audit preparation

**Platform Support:**
- ✅ Windows (Windows Defender Firewall)
- 🔄 Linux (iptables/ufw - planned)
- 🔄 macOS (pfctl - planned)

---

## ⚡ Performance Monitoring Commands

### `perf` - Performance Analysis

**Purpose:** Monitor system performance and analyze resource usage.

**Syntax:**
```bash
perf [command] [options]
```

**Commands:**
- `analyze` - Perform comprehensive system performance analysis
- `monitor` - Start real-time performance monitoring
- `report` - Generate detailed performance report
- `baseline create <name>` - Create performance baseline
- `baseline list` - List all saved baselines
- `baseline delete <name>` - Delete a baseline
- `help` - Show detailed help

**Examples:**
```bash
# Quick performance analysis
perf analyze

# Create baseline before optimization
perf baseline create before-optimization

# Generate comprehensive report
perf report

# Start real-time monitoring
perf monitor

# List all baselines
perf baseline list
```

**Metrics Monitored:**
- CPU usage and load averages
- Memory utilization and availability
- Disk I/O operations and utilization
- Network bandwidth and packet rates
- System responsiveness

**Use Cases:**
- Performance troubleshooting
- Capacity planning
- System optimization
- Bottleneck identification
- Historical trend analysis

---

## 🖥️ Server Management Commands

### `server` - System Administration

**Purpose:** Comprehensive server management and system monitoring.

**Syntax:**
```bash
server [command] [options]
```

**Commands:**
- `health` - Check overall server health status
- `services list` - List all system services
- `services start <name>` - Start a service (requires admin)
- `services stop <name>` - Stop a service (requires admin)
- `services restart <name>` - Restart a service (requires admin)
- `users` - List active users and sessions
- `session list` - List active user sessions
- `help` - Show detailed help

**Examples:**
```bash
# Check server health
server health

# List all services
server services list

# Start Print Spooler service
server services start "Print Spooler"

# Stop Windows Update service
server services stop "Windows Update"

# Restart DNS Client service
server services restart "DNS Client"

# List active users
server users
```

**Health Components:**
- CPU utilization
- Memory consumption
- Disk usage across drives
- Network adapter status
- Critical service health
- System uptime

**Use Cases:**
- System administration
- Service management
- User monitoring
- Health monitoring
- Troubleshooting

---

## 🌐 Remote Administration Commands

### `remote` - Remote Server Management

**Purpose:** Manage remote servers and execute commands via SSH.

**Syntax:**
```bash
remote [command] [options]
```

**Commands:**
- `list` - List all configured remote servers
- `add <name> <user@host>` - Add new remote server
- `remove <name>` - Remove server configuration
- `exec <server> "<command>"` - Execute command on remote server
- `help` - Show detailed help

**Examples:**
```bash
# List configured servers
remote list

# Add servers with different configurations
remote add web1 admin@192.168.1.10
remote add db1 root@database.local --port 2222
remote add app1 deploy@app.example.com --key ~/.ssh/deploy_key

# Execute commands remotely
remote exec web1 "uptime"
remote exec db1 "df -h"
remote exec web1 "systemctl status nginx"

# Remove server
remote remove old-server
```

**Connection Methods:**
- SSH key authentication (recommended)
- Password authentication
- Custom SSH ports
- Connection pooling

**Use Cases:**
- Remote administration
- Distributed system management
- Command execution across servers
- Deployment automation
- Configuration management

**Note:** Current implementation uses mock SSH for development. Real SSH support planned.

---

## 🌐 Network Tools

### Network Connectivity

#### `ping` - Test Network Connectivity
```bash
ping <host>                    # Basic connectivity test
ping -c 5 8.8.8.8             # Send 5 packets
ping -t 2000 example.com       # 2 second timeout
```

#### `tracert` - Trace Network Route
```bash
tracert google.com             # Trace route to Google
tracert -m 15 8.8.8.8          # Max 15 hops
```

### DNS Operations

#### `nslookup` - DNS Queries
```bash
nslookup google.com            # Basic DNS lookup
nslookup google.com MX         # Get MX records
nslookup example.com -s 8.8.8.8  # Use specific DNS server
```

### Network Information

#### `ipconfig` - Network Configuration
```bash
ipconfig                       # Show network interfaces
ipconfig /all                  # Detailed information (Windows)
```

#### `netstat` - Network Connections
```bash
netstat                        # Show all connections
netstat -tcp                   # TCP connections only
netstat -tcp :80               # Connections on port 80
netstat --csv                  # Export as CSV
```

#### `arp` - ARP Table Management
```bash
arp -a                         # Show all ARP entries
arp 192.168.1.1               # Show specific entry
arp -d 192.168.1.100          # Delete ARP entry
```

### Network Scanning

#### `portscan` - Port Scanning
```bash
portscan google.com            # Scan common ports
portscan 192.168.1.1 -p 80,443,22  # Specific ports
portscan example.com --top-ports 100  # Top 100 ports
```

#### `netdiscover` - Network Discovery
```bash
netdiscover                    # Discover local network devices
```

### Network Analysis

#### `sniff` - Packet Capture
```bash
sniff -c 10                    # Capture 10 packets
sniff -p HTTP -v               # HTTP packets with details
sniff --port 80 --save capture.pcap  # Save to file
```

#### `speedtest` - Internet Speed Test
```bash
speedtest                      # Full speed test
speedtest -s                   # Simple output
speedtest --download-only      # Download only
```

### File Transfer

#### `wget` - Download Files
```bash
wget https://example.com/file.zip     # Download file
wget -v https://api.github.com/data   # Verbose output
```

---

## 📁 File Operations

### Directory Navigation

#### `ls` / `dir` - List Directory Contents
```bash
ls                             # List files
ls -la                         # Detailed listing
dir                           # Windows-style listing
```

#### `pwd` - Print Working Directory
```bash
pwd                           # Show current directory
```

#### `cd` - Change Directory
```bash
cd /path/to/directory         # Change directory
cd ..                         # Go up one level
cd ~                          # Go to home directory
```

### File Management

#### `cat` - Display File Contents
```bash
cat filename.txt              # Display file
cat file1.txt file2.txt       # Display multiple files
```

#### `cp` - Copy Files
```bash
cp source.txt destination.txt  # Copy file
cp -r source_dir dest_dir     # Copy directory recursively
```

#### `mv` - Move/Rename Files
```bash
mv oldname.txt newname.txt    # Rename file
mv file.txt /new/location/    # Move file
```

#### `rm` - Remove Files
```bash
rm filename.txt               # Delete file
rm *.tmp                      # Delete with wildcard
rm -r directory               # Delete directory recursively
```

### Directory Management

#### `mkdir` - Create Directories
```bash
mkdir new_directory           # Create directory
mkdir -p path/to/new/dir      # Create parent directories
```

#### `rmdir` - Remove Directories
```bash
rmdir empty_directory         # Remove empty directory
rmdir -r directory_tree       # Remove directory tree
```

---

## ⚙️ System Information

### User Information

#### `whoami` - Current User
```bash
whoami                        # Show current user
```

#### `hostname` - System Name
```bash
hostname                      # Show system hostname
```

### System Details

#### `sysinfo` - System Information
```bash
sysinfo                       # Basic system info
sysinfo -v                    # Detailed information
sysinfo --cpu                 # CPU information only
sysinfo --memory              # Memory information only
```

#### `ver` - Version Information
```bash
ver                           # SuperShell version
ver -v                        # Detailed version info
```

### Process Management

#### `killtask` - Terminate Processes
```bash
killtask notepad              # Terminate notepad
killtask -f chrome            # Force terminate Chrome
killtask -t explorer          # Terminate with children
killtask 1234                 # Terminate by PID
```

### System Utilities

#### `clear` - Clear Screen
```bash
clear                         # Clear terminal screen
```

#### `echo` - Print Text
```bash
echo "Hello World"            # Print text
echo $PATH                    # Print environment variable
```

#### `winupdate` - Windows Updates (Windows only)
```bash
winupdate                     # Check for updates
winupdate install             # Install updates
```

---

## 🔍 Help & Discovery

### `help` - Command Help System

**Purpose:** Comprehensive help and documentation system.

**Syntax:**
```bash
help [command]
```

**Examples:**
```bash
help                          # Show all commands
help firewall                 # Detailed help for firewall
help perf                     # Performance command help
```

### `lookup` - Command Discovery

**Purpose:** Interactive command discovery and search system.

**Syntax:**
```bash
lookup [options] [query]
```

**Options:**
- `-m, --menu` - Interactive dropdown menu
- `-s, --similar` - Find similar commands
- `-c, --categories` - Show command categories
- `-t, --task <task>` - Task-based suggestions

**Examples:**
```bash
lookup network                # Find network commands
lookup -m                     # Interactive menu
lookup -t security            # Security-related commands
lookup -s net                 # Similar to 'net'
```

### `exit` - Exit Shell
```bash
exit                          # Exit SuperShell
```

---

## 🧠 Smart History System

### `history` - AI-Powered Command History

**Purpose:** Intelligent command history management with AI-powered search, pattern recognition, and analytics.

**Syntax:**
```bash
history [command] [options]
```

**Commands:**
- `[no args]` - Show recent command history (last 20 commands)
- `smart <query>` - Intelligent search using natural language
- `patterns` - Display detected usage patterns and workflows
- `suggest` - Get context-aware command suggestions
- `timeline` - Show chronological timeline view
- `stats` - Display comprehensive usage statistics
- `export <format>` - Export history (json, csv, txt)
- `add <command>` - Manually add command to history
- `clear` - Clear all command history

**Smart Search Examples:**
```bash
# Natural language search
history smart "backup files"
history smart "git commit"
history smart "network diagnostics"
history smart "system monitoring"
history smart "file operations"

# Category-based search
history smart network
history smart security
history smart development
```

**Pattern Recognition:**
```bash
# View detected patterns
history patterns

# Shows patterns like:
# • Sequential workflows (commands used together)
# • Frequency patterns (most used commands)
# • Time-based patterns (commands used at specific times)
```

**Context-Aware Suggestions:**
```bash
# Get smart suggestions
history suggest

# Provides suggestions based on:
# • Current working directory context
# • Recent command patterns
# • Time of day preferences
# • Historical usage patterns
```

**Timeline & Analytics:**
```bash
# Visual timeline
history timeline
# Shows:
# • Chronological command execution
# • Success/failure indicators
# • Execution timestamps
# • Directory context

# Comprehensive statistics
history stats
# Includes:
# • Total commands and success rate
# • Most frequently used commands
# • Activity patterns by hour/day
# • Command categorization breakdown
# • Visual usage charts
```

**Export & Backup:**
```bash
# Export in different formats
history export json       # JSON for programmatic analysis
history export csv        # CSV for spreadsheet analysis
history export txt        # Human-readable text format

# Manual management
history add "docker ps -a"    # Add command manually
history clear                 # Clear all history (with confirmation)
```

**Advanced Features:**

**🎯 Automatic Tracking:**
- All commands automatically recorded with metadata
- Execution time, exit codes, working directory
- Duration tracking and performance metrics
- Cross-session persistence

**🏷️ Smart Categorization:**
- Commands auto-classified into categories:
  - 🗂️ Filesystem: ls, cd, cp, mv, rm
  - 🌐 Network: ping, wget, curl, ssh
  - ⚙️ Management: firewall, server, perf
  - 🔧 Development: git, docker, npm
  - 📊 Monitoring: ps, top, netstat

**🧠 Intelligent Features:**
- **Pattern Learning:** System learns from usage patterns
- **Context Awareness:** Suggestions adapt to directory and time
- **Fuzzy Matching:** Flexible search with typo tolerance
- **Visual Analytics:** Color-coded output with progress bars

**Use Cases:**

**For Developers:**
```bash
# Track Git workflows
history smart "git"
history patterns          # See Git workflow patterns

# Find complex commands
history smart "docker compose"
history smart "build"
```

**For System Administrators:**
```bash
# Monitor command usage
history stats
history timeline

# Find troubleshooting commands
history smart "network troubleshoot"
history smart "service restart"

# Export for compliance
history export csv
```

**For Power Users:**
```bash
# Build personal command library
history smart "backup"
history suggest           # Get productivity suggestions

# Analyze usage patterns
history patterns
history stats
```

**Storage & Privacy:**
- **Local Storage:** All data in `~/.supershell_history.json`
- **No Cloud Sync:** Complete privacy, no external data
- **Configurable Limits:** Automatic cleanup (1000 commands default)
- **JSON Format:** Human-readable and portable

**Examples in Action:**

```bash
# Basic usage
history                   # View recent commands
   1 14:42:12 ls -la ✓
   2 14:42:15 git status ✓
   3 14:42:18 ping google.com ✓

# Smart search
history smart git
🔍 Smart History Search
🎯 Query: git
📊 Found 5 matches
   2 08/17 14:42 git status ✓
   5 08/17 14:45 git commit -m "Update" ✓

# Pattern analysis
history patterns
🧠 Command Patterns
🔄 Pattern 1: Git Workflow
   Description: Common Git version control workflow
   Usage: 8 times, last used 08/17 14:45
   Commands: git status → git add → git commit

# Smart suggestions
history suggest
💡 Smart Suggestions
💡 Suggestion 1: git log --oneline -10
💡 Suggestion 2: ls -la
💡 Suggestion 3: perf analyze
```

The Smart History system transforms your command history from a simple list into an intelligent assistant that learns from your usage patterns and helps improve your productivity.

---

## 🚀 FastCP File Transfer

### Ultra-Fast File Transfer System

#### `fastcp-send` - Send Files
```bash
fastcp-send file.txt server:/path/  # Send file
fastcp-send -e file.txt server:/    # With encryption
```

#### `fastcp-recv` - Receive Files
```bash
fastcp-recv server:/file.txt ./     # Receive file
fastcp-recv -v server:/data ./      # Verbose mode
```

#### `fastcp-backup` - Create Backups
```bash
fastcp-backup /important/data backup.fcb  # Create backup
fastcp-backup -c /data backup.fcb         # With compression
```

#### `fastcp-restore` - Restore Backups
```bash
fastcp-restore backup.fcb /restore/path   # Restore backup
fastcp-restore -v backup.fcb ./           # Verbose restore
```

#### `fastcp-dedup` - Deduplication
```bash
fastcp-dedup /data/directory          # Deduplicate files
fastcp-dedup -r /data                 # Recursive deduplication
```

**Features:**
- Ultra-fast transfer speeds
- Built-in encryption
- Compression support
- Resume capability
- Deduplication
- Cloud storage integration

---

## 🎯 Usage Patterns

### Getting Started
```bash
# 1. Build SuperShell
go build ./cmd/supershell

# 2. Run interactive shell
./supershell.exe

# 3. Get help
help
lookup -m
```

### System Administration Workflow
```bash
# Check system health
server health
firewall status
perf analyze

# Manage services
server services list
server services restart "DNS Client"

# Monitor performance
perf baseline create daily-check
perf monitor
```

### Network Troubleshooting
```bash
# Basic connectivity
ping 8.8.8.8
tracert google.com

# Network configuration
ipconfig
netstat -tcp

# Advanced diagnostics
portscan target.com
speedtest
```

### File Management
```bash
# Navigate and explore
ls -la
pwd
cd /important/directory

# File operations
cp important.txt backup.txt
rm *.tmp
mkdir new_project
```

---

## 🔧 Advanced Features

### Tab Completion
- Press `Tab` to auto-complete commands
- Works for subcommands and options
- Context-aware suggestions

### Command Chaining
```bash
# Multiple commands
firewall status && perf analyze && server health
```

### Help Integration
```bash
# Multiple help methods
help command           # Comprehensive help
command help          # Quick help
lookup command        # Find related commands
```

### Cross-Platform Support
- Windows (primary support)
- Linux (planned/partial)
- macOS (planned/partial)

---

## 📞 Support & Troubleshooting

### Common Issues

**Command Not Found:**
```bash
help | grep command_name
lookup command_name
```

**Permission Denied:**
- Run as administrator (Windows)
- Use sudo (Linux/macOS)
- Check user privileges

**Network Issues:**
```bash
ping 8.8.8.8          # Test connectivity
ipconfig              # Check configuration
```

### Getting More Help
- Use `help <command>` for detailed information
- Use `lookup -m` for interactive discovery
- Check command-specific help with `<command> help`

---

**💡 Pro Tip:** Use `lookup -m` to explore commands interactively, and `help <command>` for comprehensive documentation with examples!