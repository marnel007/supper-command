# SuperShell Quick Reference Card

## 🔥 Firewall Management
```bash
firewall status                    # Check firewall status
firewall enable                    # Enable firewall (admin required)
firewall disable                   # Disable firewall (admin required)
firewall rules list                # List all firewall rules
help firewall                     # Detailed help with examples
```

## ⚡ Performance Monitoring
```bash
perf analyze                      # Analyze system performance
perf monitor                      # Real-time monitoring
perf report                       # Generate performance report
perf baseline create <name>       # Create performance baseline
perf baseline list                # List all baselines
perf baseline delete <name>       # Delete baseline
help perf                        # Detailed help with examples
```

## 🖥️ Server Management
```bash
server health                     # Check server health
server services list              # List all services
server services start "<name>"    # Start service (admin required)
server services stop "<name>"     # Stop service (admin required)
server services restart "<name>"  # Restart service (admin required)
server users                      # List active users
server session list               # List active sessions
help server                      # Detailed help with examples
```

## 🌐 Remote Server Management
```bash
remote list                       # List configured servers
remote add <name> <user@host>     # Add remote server
remote add <name> <user@host> --port 2222  # Add with custom port
remote exec <server> "<cmd>"      # Execute command remotely
remote remove <name>              # Remove server configuration
help remote                      # Detailed help with examples
```

## 🌐 Network Tools
```bash
ping <host>                       # Test connectivity
ping -c 5 <host>                  # Send 5 ping packets
tracert <host>                    # Trace route to destination
nslookup <domain>                 # DNS lookup
nslookup <domain> MX              # Get MX records
ipconfig                         # Show network configuration
netstat                          # Show network connections
netstat -tcp :80                 # Show TCP connections on port 80
portscan <host>                  # Scan common ports
speedtest                        # Test internet speed
sniff -c 10                      # Capture 10 packets
wget <url>                       # Download file
arp -a                           # Show ARP table
```

## 📁 File System Operations
```bash
ls                               # List files
ls -la                          # List with details
dir                             # Windows-style listing
pwd                             # Show current directory
cd <path>                       # Change directory
cat <file>                      # Display file contents
cp <src> <dest>                 # Copy file
mv <old> <new>                  # Move/rename file
rm <file>                       # Delete file
rm *.tmp                        # Delete with wildcard
mkdir <dir>                     # Create directory
rmdir <dir>                     # Remove directory
```

## ⚙️ System Information
```bash
whoami                          # Current user
hostname                        # System hostname
ver                            # SuperShell version
ver -v                         # Detailed version info
sysinfo                        # System information
sysinfo -v                     # Detailed system info
killtask <process>             # Terminate process
killtask -f <process>          # Force terminate
clear                          # Clear screen
echo "<text>"                  # Print text
```

## 🔍 Help & Discovery
```bash
help                           # Show all commands
help <command>                 # Detailed help for specific command
<command> help                 # Command-specific help
lookup <topic>                 # Find related commands
lookup -m                      # Interactive command menu
lookup -t <category>           # Commands by category
lookup -s <term>               # Similar commands
```

## 🧠 Smart History System
```bash
history                        # Show recent command history
history smart "<query>"        # Intelligent search (e.g., "git commit")
history patterns               # Show detected usage patterns
history suggest                # Get context-aware suggestions
history timeline               # Visual timeline of commands
history stats                  # Comprehensive usage statistics
history export json            # Export history as JSON
history export csv             # Export history as CSV
history add "<command>"        # Manually add command
history clear                  # Clear all history
help history                   # Detailed help with examples
```

## 🚀 Quick Start Guide
```bash
# 1. Build SuperShell
go build ./cmd/supershell

# 2. Run interactive shell
./supershell.exe

# 3. Run single commands
./supershell.exe -c "command args"

# 4. Test all new management commands
firewall status
perf analyze
server health
remote add test admin@localhost
```

## 🧪 Common Workflows

### System Health Check
```bash
firewall status && perf analyze && server health && sysinfo
```

### Network Diagnostics
```bash
ipconfig && ping 8.8.8.8 && tracert google.com && speedtest
```

### Performance Monitoring
```bash
perf baseline create before && perf analyze && perf report
```

### Smart History Usage
```bash
history smart "network" && history patterns && history suggest
```

### Service Management
```bash
server services list && server health && server users
```

## 💡 Pro Tips & Shortcuts

### Tab Completion
- Press `Tab` to auto-complete commands and subcommands
- Works for all command categories and options

### Command Chaining
```bash
# Multiple commands in sequence
command1 && command2 && command3
```

### Getting Help
```bash
# Multiple ways to access help
help firewall                  # Comprehensive help
firewall help                  # Quick help
lookup firewall                # Find related commands
```

### Admin Privileges
Commands requiring administrator privileges:
- `firewall enable/disable`
- `server services start/stop/restart`
- Some system monitoring features

### Development Mode
```bash
# Mock connections for testing (current default)
remote add test admin@localhost  # Uses mock SSH
```

## 🔧 Troubleshooting Quick Fixes

| Issue | Solution |
|-------|----------|
| Command not found | `help \| grep <command>` or `lookup <command>` |
| Permission denied | Run as administrator/sudo |
| Network timeout | Check connectivity with `ping 8.8.8.8` |
| Service won't start | Check with `server health` |
| Remote connection fails | Verify with `remote list` (mock mode) |

## 📊 Command Categories

| 🔥 **Security** | ⚡ **Performance** | 🖥️ **System** | 🌐 **Network** |
|----------------|-------------------|---------------|----------------|
| firewall | perf | server | ping |
| | | sysinfo | tracert |
| | | killtask | nslookup |
| | | whoami | netstat |

| 📁 **Files** | 🌐 **Remote** | 🔍 **Help** | 🧠 **History** |
|-------------|--------------|-------------|----------------|
| ls/dir | remote | help | history |
| cat | | lookup | |
| cp/mv/rm | | | |
| mkdir | | | |

---

**💡 Remember:** Use `help <command>` for detailed examples and use cases for any command!