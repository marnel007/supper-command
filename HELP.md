# SuperShell Help

Welcome to **SuperShell**! Below you'll find a summary of all available commands, their usage, and options.

---

## Table of Contents
- [General Commands](#general-commands)
- [Network Commands](#network-commands)
- [Advanced: Packet Sniffer](#advanced-packet-sniffer)
- [Routing Table](#routing-table)
- [Alias Usage](#alias-usage)

---

## General Commands

| Command     | Description                        |
|-------------|------------------------------------|
| `help`      | Show this help message             |
| `clear`     | Clear the screen                   |
| `echo`      | Print text to the screen           |
| `pwd`       | Print working directory            |
| `ls`        | List directory contents            |
| `cd`        | Change directory                   |
| `cat`       | Show file contents                 |
| `mkdir`     | Create a new directory             |
| `rm`        | Delete a file                      |
| `rmdir`     | Remove a directory                 |
| `cp`        | Copy a file                        |
| `mv`        | Move or rename a file              |
| `whoami`    | Show current user                  |
| `hostname`  | Show the system hostname           |
| `ver`       | Show shell version                 |
| `exit`      | Exit the shell                     |

---

## Network Commands

| Command         | Description                                      |
|-----------------|--------------------------------------------------|
| `ipconfig`      | Show network interfaces and IP addresses          |
| `netstat`       | Show open network connections                    |
| `arp`           | Show the ARP table                               |
| `nslookup`      | Query DNS records for a domain                   |
| `ping`          | Ping a host to test connectivity                 |
| `tracert`       | Trace the route to a host                        |
| `wget`          | Download a file from a URL                       |
| `speedtest`     | Run a Go-native speed test                       |
| `netdiscover`   | Discover live hosts on a subnet                  |
| `portscan`      | Scan TCP ports on a host                         |

---

## Advanced: Packet Sniffer

### sniff - Packet sniffer

```
Usage:
  sniff <iface|index> [file.pcap] [max_packets] [bpf_filter]

Options:
  <iface|index>    Interface name or index to capture from (required)
  [file.pcap]      Optional file to save packets (Wireshark-compatible)
  [max_packets]    Optional max packets to capture (default: 50)
  [bpf_filter]     Optional BPF filter (e.g. "tcp port 443")

Examples:
  sniff 2
  sniff 2 capture.pcap 200
  sniff 2 "" 100 "tcp"
  sniff 2 capture.pcap 100 "tcp port 443 or port 80"
  sniff 2 "" 50 "tcp port 22"

Filter examples:
  "tcp"                      (all TCP traffic)
  "port 80"                  (HTTP)
  "tcp port 443 or port 80"  (HTTPS or HTTP)
  "tcp and port 22"          (SSH)

Notes:
  - Saves to .pcap if file is specified
  - Default max_packets is 50
  - BPF filter is optional
  - Use sniff with no arguments to list interfaces and see this help
```

---

## Routing Table

### route - Show the routing table

```
Usage:
  route

Options:
  (no options yet)

Notes:
  - Shows the system routing table
  - On Windows, uses 'route print'
  - On Unix, uses 'ip route' or 'netstat -rn'
```

---

## Alias Usage

```
alias                # List all aliases
alias <name> <cmd>   # Create or update an alias (e.g. alias ll ls -l)
unalias <name>       # Remove an alias
```

---

Type `help` in the shell to see this message again. 