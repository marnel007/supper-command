# SuperShell Command Guide

This comprehensive guide provides detailed examples and usage information for all SuperShell commands, including the new management commands for firewall, performance monitoring, server management, and remote administration.

## 📋 Table of Contents
- [🔥 Firewall Management](#-firewall-management)
- [⚡ Performance Monitoring](#-performance-monitoring)
- [🖥️ Server Management](#️-server-management)
- [🌐 Remote Server Management](#-remote-server-management)
- [🧠 Smart History System](#-smart-history-system)
- [🌐 Network Commands](#-network-commands)
- [📁 File System Commands](#-file-system-commands)
- [⚙️ System Commands](#️-system-commands)
- [🧪 Testing Workflows](#-testing-workflows)
- [🔧 Troubleshooting](#-troubleshooting)

## 🔥 Firewall Management

### Basic Usage
```bash
# Check firewall status
firewall status

# Enable/disable firewall (requires admin)
firewall enable
firewall disable

# Manage firewall rules
firewall rules list
firewall rules
```

### Examples
```bash
# Security audit - check firewall configuration
firewall status

# Enable firewall for security compliance
firewall enable

# List all firewall rules for troubleshooting
firewall rules list

# Get detailed help
help firewall
```

### Use Cases
- **Security Management**: Monitor and control system firewall settings
- **Rule Management**: View and manage firewall rules for applications
- **Compliance Checking**: Verify firewall status for security audits
- **Troubleshooting**: Diagnose network connectivity issues

---

## ⚡ Performance Monitoring

### Basic Usage
```bash
# Analyze current performance
perf analyze

# Start real-time monitoring
perf monitor

# Generate performance report
perf report

# Manage baselines
perf baseline create <name>
perf baseline list
perf baseline delete <name>
```

### Examples
```bash
# Quick performance check
perf analyze

# Create baseline for comparison
perf baseline create production-baseline

# Generate comprehensive report
perf report

# Monitor system in real-time
perf monitor

# List all saved baselines
perf baseline list

# Get detailed help
help perf
```

### Use Cases
- **Performance Monitoring**: Track system resource usage over time
- **Bottleneck Detection**: Identify CPU, memory, disk, or network bottlenecks
- **Capacity Planning**: Understand system limits and plan for scaling
- **Troubleshooting**: Diagnose performance issues and slowdowns

---

## 🖥️ Server Management

### Basic Usage
```bash
# Check server health
server health

# Manage services
server services list
server services start "<service_name>"
server services stop "<service_name>"
server services restart "<service_name>"

# Monitor users
server users
server session list
```

### Examples
```bash
# Health check dashboard
server health

# List all system services
server services list

# Start a specific service
server services start "Print Spooler"

# Stop Windows Update service
server services stop "Windows Update"

# Restart DNS service
server services restart "DNS Client"

# Check active users
server users

# Get detailed help
help server
```

### Use Cases
- **System Administration**: Monitor and manage server components
- **Service Management**: Control Windows/Linux services
- **User Monitoring**: Track active users and sessions
- **Health Monitoring**: Get real-time server health status

---

## 🌐 Remote Server Management

### Basic Usage
```bash
# List configured servers
remote list

# Add new server
remote add <name> <user@host>

# Execute commands
remote exec <server> "<command>"

# Remove server
remote remove <name>
```

### Examples
```bash
# List all configured remote servers
remote list

# Add servers with different configurations
remote add web1 admin@192.168.1.10
remote add db1 root@db.example.com --port 2222
remote add app1 deploy@app.com --key ~/.ssh/deploy_key

# Execute commands on remote servers
remote exec web1 "uptime"
remote exec db1 "df -h"
remote exec web1 "systemctl status nginx"

# Remove old server configuration
remote remove old-server

# Get detailed help
help remote
```

### Use Cases
- **Remote Administration**: Manage multiple servers from one location
- **Command Execution**: Run commands across multiple servers
- **Server Monitoring**: Check status and health of remote systems
- **Deployment Management**: Execute deployment scripts remotely

---

## 🧠 Smart History System

### Basic Usage
```bash
# Show recent command history
history

# Intelligent search
history smart "<query>"

# Pattern analysis
history patterns

# Smart suggestions
history suggest

# Timeline view
history timeline

# Usage statistics
history stats

# Export history
history export <format>
```

### Smart Search Examples
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
history smart filesystem
```

### Pattern Recognition
```bash
# View detected patterns
history patterns

# Example output:
# 🧠 Command Patterns
# 🔄 Pattern 1: Git Workflow
#    Description: Common Git version control workflow
#    Usage: 8 times, last used 08/17 14:45
#    Commands: git status → git add → git commit
#
# 🔄 Pattern 2: System Monitoring
#    Description: Frequently used command: perf
#    Usage: 12 times, last used 08/17 15:30
```

### Context-Aware Suggestions
```bash
# Get smart suggestions based on context
history suggest

# Example output:
# 💡 Smart Suggestions
# 💡 Suggestion 1: git log --oneline -10
# 💡 Suggestion 2: ls -la
# 💡 Suggestion 3: perf analyze
# 💡 Suggestion 4: server health
# 💡 Suggestion 5: history stats
```

### Timeline & Analytics
```bash
# Visual timeline view
history timeline

# Example output:
# 📅 Command Timeline
# 📅 Sunday, August 17, 2025 (13 commands)
#   ├─ 14:42 history ✓
#   ├─ 14:42 ls -la ✓
#   ├─ 14:42 git status ✓
#   └─ 14:43 perf analyze ✓

# Comprehensive statistics
history stats

# Example output:
# 📊 Command Statistics
# 📊 Overview
#    Total Commands: 25
#    Success Rate: 96.0%
#    Most Used: history (8 times)
#
# 🏆 Top Commands
#    1. history      [██████████░░░░░] (8)
#    2. git          [████░░░░░░░░░░░] (4)
#    3. ls           [███░░░░░░░░░░░░] (3)
```

### Export & Management
```bash
# Export in different formats
history export json       # For programmatic analysis
history export csv        # For spreadsheet analysis
history export txt        # Human-readable format

# Manual management
history add "docker ps -a"    # Add command manually
history clear                 # Clear all history (with confirmation)

# Get detailed help
help history
```

### Advanced Features

#### Automatic Tracking
- All commands automatically recorded with metadata
- Execution time, exit codes, working directory tracked
- Duration and performance metrics captured
- Cross-session persistence maintained

#### Smart Categorization
Commands are automatically classified into categories:
- 🗂️ **Filesystem**: ls, cd, cp, mv, rm, mkdir
- 🌐 **Network**: ping, wget, curl, ssh, tracert
- ⚙️ **Management**: firewall, server, perf, remote
- 🔧 **Development**: git, docker, npm, yarn
- 📊 **Monitoring**: ps, top, netstat, htop
- 🔍 **Search**: grep, find, locate, which

#### Intelligent Features
- **Pattern Learning**: System learns from your usage patterns
- **Context Awareness**: Suggestions adapt to directory and time
- **Fuzzy Matching**: Flexible search with typo tolerance
- **Visual Analytics**: Color-coded output with progress bars

### Use Cases

#### For Developers
```bash
# Track development workflows
history smart "git"
history patterns          # See Git workflow patterns

# Find complex commands you used before
history smart "docker compose"
history smart "build deployment"

# Get productivity insights
history stats
history suggest
```

#### For System Administrators
```bash
# Monitor command usage for compliance
history stats
history timeline
history export csv        # Export for audit reports

# Find troubleshooting commands
history smart "network troubleshoot"
history smart "service restart"
history smart "performance issue"

# Build knowledge base
history patterns
history export json       # Share with team
```

#### For Power Users
```bash
# Build personal command library
history smart "backup"
history smart "automation"

# Optimize workflows
history patterns          # Identify common sequences
history suggest           # Get efficiency suggestions

# Analyze productivity
history stats             # See usage patterns
history timeline          # Track daily activities
```

### Privacy & Storage
- **Local Storage**: All data stored in `~/.supershell_history.json`
- **No Cloud Sync**: Complete privacy, no external data transmission
- **Configurable Limits**: Automatic cleanup (keeps last 1000 commands)
- **JSON Format**: Human-readable and easily portable format

---

## 🧪 Complete Testing Workflow

### Quick Test All Commands
```bash
# Build SuperShell
go build ./cmd/supershell

# Test all new commands
./supershell.exe -c "firewall status"
./supershell.exe -c "perf analyze"
./supershell.exe -c "server health"
./supershell.exe -c "remote list"

# Test help system
./supershell.exe -c "help firewall"
./supershell.exe -c "help perf"
./supershell.exe -c "help server"
./supershell.exe -c "help remote"
```

### Interactive Shell Testing
```bash
# Start interactive shell
./supershell.exe

# Test commands interactively
firewall status
perf analyze
server health
remote list
help firewall
exit
```

### Advanced Examples
```bash
# Performance monitoring workflow
perf baseline create before-optimization
perf analyze
# ... make system changes ...
perf analyze
perf baseline create after-optimization

# Server management workflow
server health
server services list | grep -i "stopped"
server services start "Print Spooler"
server users

# Remote server workflow
remote add prod-web admin@prod.example.com
remote exec prod-web "uptime"
remote exec prod-web "free -h"
remote exec prod-web "df -h"
remote remove prod-web
```

---

## 📊 Expected Output Examples

### Firewall Status
```
Firewall Enabled: true
Profile: Windows Defender Firewall
Platform: windows
Rule Count: 15
Last Updated: 2025-08-17 12:47:03
```

### Performance Analysis
```
Performance Analysis Results:
Overall Health: healthy
Timestamp: 2025-08-17 12:48:47

Bottlenecks Found:
- High memory usage detected (85%)

Suggestions:
- Memory: Consider closing unused applications
```

### Server Health
```
Overall Health: healthy
Uptime: 2d 14h 30m
Last Check: 2025-08-17 12:49:29

Component Health:
- CPU: healthy (21.1%)
  CPU usage is normal
- Memory: healthy (59.8%)
  Memory usage is normal
- Disk: healthy (11.6%)
  Disk usage is normal
```

### Remote Server List
```
Remote Servers (3):
1. web1 (admin@192.168.1.10:22) - connected
2. db1 (root@db.example.com:2222) - connected
3. app1 (deploy@app.com:22) - disconnected
```

---

## 🔧 Troubleshooting

### Common Issues

**Firewall commands require admin privileges:**
```bash
# Run as administrator or use sudo on Linux
sudo ./supershell.exe -c "firewall enable"
```

**Remote server connection issues:**
```bash
# Check SSH connectivity first
ssh user@hostname

# Verify server configuration
remote list
```

**Performance monitoring not showing data:**
```bash
# Ensure sufficient permissions for system monitoring
# Some metrics require elevated privileges
```

### Getting Help
```bash
# General help
help

# Command-specific help
help firewall
help perf
help server
help remote

# Command usage
firewall help
perf help
server help
remote help
```

---

## 🚀 Advanced Features

### Tab Completion
All commands support tab completion for subcommands and options:
```bash
firewall <TAB>          # Shows: status, enable, disable, rules, help
perf baseline <TAB>     # Shows: create, list, delete
server services <TAB>   # Shows: list, start, stop, restart
```

### Command Chaining
```bash
# Check firewall and server health together
./supershell.exe -c "firewall status" && ./supershell.exe -c "server health"
```

### Scripting Support
```bash
# Use in scripts
if ./supershell.exe -c "firewall status" | grep -q "Enabled: true"; then
    echo "Firewall is enabled"
fi
```

---

## 🌐 Network Commands

### Basic Network Tools
```bash
# Network connectivity
ping google.com                    # Test connectivity
ping -c 5 8.8.8.8                 # Send 5 ping packets
tracert google.com                 # Trace route to destination

# DNS operations
nslookup google.com                # DNS lookup
nslookup google.com MX             # Get MX records
nslookup example.com -s 8.8.8.8    # Use specific DNS server

# Network information
ipconfig                           # Show network configuration
netstat                           # Show network connections
netstat -tcp :80                  # Show TCP connections on port 80
arp -a                            # Show ARP table

# Network scanning
portscan google.com               # Scan common ports
portscan 192.168.1.1 -p 80,443,22 # Scan specific ports
speedtest                         # Test internet speed
```

### Advanced Network Tools
```bash
# Packet capture and analysis
sniff -c 10                       # Capture 10 packets
sniff -p HTTP -v                  # Capture HTTP packets with details
sniff --port 80 -c 5              # Capture packets on port 80

# File transfer
wget https://example.com/file.zip  # Download file
wget -v https://api.github.com/users # Download with verbose output

# Network discovery
netdiscover                       # Discover network devices
route                            # Show routing table
```

---

## 📁 File System Commands

### Basic File Operations
```bash
# Directory operations
ls                               # List files
ls -la                          # List with details
dir                             # Windows-style directory listing
pwd                             # Show current directory
cd /path/to/directory           # Change directory

# File operations
cat filename.txt                # Display file contents
cp source.txt destination.txt   # Copy file
mv oldname.txt newname.txt      # Move/rename file
rm filename.txt                 # Delete file
rm *.tmp                        # Delete files with wildcard

# Directory management
mkdir new_directory             # Create directory
rmdir directory_name            # Remove directory
```

### Advanced File Operations
```bash
# File searching and manipulation
find . -name "*.txt"            # Find text files
grep "pattern" filename.txt     # Search within files
head -n 10 filename.txt         # Show first 10 lines
tail -n 10 filename.txt         # Show last 10 lines
```

---

## ⚙️ System Commands

### System Information
```bash
# Basic system info
whoami                          # Current user
hostname                        # System hostname
ver                            # SuperShell version
ver -v                         # Detailed version info
sysinfo                        # System information
sysinfo -v                     # Detailed system info

# Process management
killtask notepad               # Terminate notepad processes
killtask -f chrome             # Force terminate Chrome
killtask -t explorer           # Terminate with child processes
```

### System Utilities
```bash
# Help and discovery
help                           # Show all commands
help <command>                 # Detailed help for specific command
lookup network                 # Find network-related commands
lookup -m                      # Interactive command menu
lookup -t security             # Security-related commands

# System maintenance
clear                          # Clear screen
echo "Hello World"             # Print text
winupdate                      # Windows update operations (Windows only)
```

---

## 🧪 Testing Workflows

### Complete System Check
```bash
# Comprehensive system assessment
firewall status
perf analyze
server health
sysinfo -v
netstat | head -20
```

### Network Diagnostics
```bash
# Network troubleshooting workflow
ipconfig
ping 8.8.8.8
tracert google.com
nslookup google.com
speedtest
netstat -tcp
```

### Performance Monitoring Workflow
```bash
# Performance baseline and monitoring
perf baseline create before-optimization
perf analyze
perf monitor &
# ... make system changes ...
perf analyze
perf baseline create after-optimization
perf report
```

### Remote Server Management Workflow
```bash
# Remote administration workflow
remote add prod-web admin@prod.example.com
remote add prod-db root@db.example.com
remote list
remote exec prod-web "uptime"
remote exec prod-web "df -h"
remote exec prod-db "systemctl status mysql"
remote remove old-server
```

---

## 🔧 Troubleshooting

### Common Issues and Solutions

**Command Not Found:**
```bash
# Check if command is available
help | grep command_name
lookup command_name
```

**Permission Denied:**
```bash
# Run as administrator (Windows) or with sudo (Linux)
# Some commands require elevated privileges:
# - firewall enable/disable
# - server services start/stop/restart
# - System service management
```

**Network Connection Issues:**
```bash
# Diagnose network problems
ping 8.8.8.8                  # Test basic connectivity
ipconfig                      # Check network configuration
netstat -tcp                  # Check active connections
tracert google.com            # Trace network path
```

**Performance Issues:**
```bash
# Diagnose performance problems
perf analyze                  # Quick performance check
server health                 # Check system health
sysinfo -v                   # Detailed system information
```

**Remote Server Issues:**
```bash
# Troubleshoot remote connections
remote list                   # Check configured servers
# Note: Current implementation uses mock SSH for development
# Real SSH connections require proper SSH setup
```

### Getting Help
```bash
# Multiple ways to get help
help                          # General help
help <command>                # Detailed command help
<command> help                # Command-specific help
lookup <topic>                # Find related commands
lookup -m                     # Interactive help menu
```

### Debug Mode
```bash
# For troubleshooting command issues
./supershell.exe -c "command args"  # Single command execution
# Check logs for detailed error information
```

---

## 🎯 Best Practices

### Security
- Always verify firewall status before making network changes
- Use SSH keys instead of passwords for remote connections
- Regularly monitor system health and performance
- Keep track of active users and services

### Performance
- Create performance baselines before making system changes
- Monitor system resources regularly
- Use appropriate timeouts for network operations
- Clean up temporary files and processes

### Administration
- Document server configurations and changes
- Use descriptive names for remote servers
- Regularly backup important configurations
- Test commands in development before production use

---

## 📊 Command Categories Summary

| Category | Commands | Purpose |
|----------|----------|---------|
| **Firewall** | `firewall status/enable/disable/rules` | Security management |
| **Performance** | `perf analyze/monitor/report/baseline` | System optimization |
| **Server** | `server health/services/users` | System administration |
| **Remote** | `remote add/list/exec/remove` | Remote management |
| **Network** | `ping/tracert/nslookup/netstat/portscan` | Network diagnostics |
| **Files** | `ls/cat/cp/mv/rm/mkdir` | File operations |
| **System** | `sysinfo/whoami/hostname/killtask` | System information |
| **Help** | `help/lookup` | Documentation and discovery |

This comprehensive guide covers all SuperShell commands with practical examples and troubleshooting information. Each command is designed to be intuitive and provide valuable system management capabilities.