package commands

import (
	"fmt"
	"sort"
	"strings"
)

// HelpSystem provides comprehensive help functionality
type HelpSystem struct {
	registry         *Registry
	completionEngine *CompletionEngine
}

// NewHelpSystem creates a new help system
func NewHelpSystem(registry *Registry) *HelpSystem {
	return &HelpSystem{
		registry:         registry,
		completionEngine: NewCompletionEngine(registry),
	}
}

// ShowGeneralHelp displays general help information
func (h *HelpSystem) ShowGeneralHelp() string {
	var help strings.Builder

	help.WriteString("SuperShell - Advanced Command Line Interface\n")
	help.WriteString("═══════════════════════════════════════════\n\n")

	help.WriteString("USAGE:\n")
	help.WriteString("    <command> [subcommand] [options]\n\n")

	help.WriteString("AVAILABLE COMMANDS:\n")

	categories := h.completionEngine.GetCommandsByCategory()
	categoryNames := make([]string, 0, len(categories))
	for category := range categories {
		categoryNames = append(categoryNames, category)
	}
	sort.Strings(categoryNames)

	for _, category := range categoryNames {
		commands := categories[category]
		help.WriteString(fmt.Sprintf("\n%s:\n", category))

		for _, cmd := range commands {
			if command, err := h.registry.Get(cmd); err == nil {
				help.WriteString(fmt.Sprintf("    %-15s %s\n", cmd, command.Description()))
			}
		}
	}

	help.WriteString("\nGLOBAL OPTIONS:\n")
	help.WriteString("    --help, -h      Show help information\n")
	help.WriteString("    --version, -v   Show version information\n")
	help.WriteString("    --json          Output in JSON format (where supported)\n")
	help.WriteString("    --verbose       Enable verbose output\n")
	help.WriteString("    --quiet         Suppress non-essential output\n\n")

	help.WriteString("EXAMPLES:\n")
	help.WriteString("    firewall status                    # Show firewall status\n")
	help.WriteString("    performance analyze                # Analyze system performance\n")
	help.WriteString("    server health --watch              # Monitor server health\n")
	help.WriteString("    remote add web1 user@host --key ~/.ssh/id_rsa\n")
	help.WriteString("    remote cluster exec web-tier \"uptime\"\n\n")

	help.WriteString("For detailed help on any command, use:\n")
	help.WriteString("    <command> --help\n\n")

	help.WriteString("NEW FEATURES:\n")
	help.WriteString("🔥 Firewall Management    - Cross-platform firewall control\n")
	help.WriteString("📊 Performance Analysis   - System performance monitoring and optimization\n")
	help.WriteString("🖥️  Server Management     - Local server health and service management\n")
	help.WriteString("🌐 Remote Management      - SSH-based remote server operations\n")
	help.WriteString("🏗️  Cluster Operations    - Multi-server cluster management\n")
	help.WriteString("🔄 Config Synchronization - Configuration sync across servers\n\n")

	return help.String()
}

// ShowCommandHelp displays help for a specific command
func (h *HelpSystem) ShowCommandHelp(commandName string) string {
	// Try to get contextual help first
	if help := h.completionEngine.GetContextualHelp(commandName); help != "" {
		return help
	}

	// Fallback to command's own help
	if cmd, err := h.registry.Get(commandName); err == nil {
		var help strings.Builder
		help.WriteString(fmt.Sprintf("Command: %s\n", cmd.Name()))
		help.WriteString(strings.Repeat("=", len(cmd.Name())+9) + "\n\n")
		help.WriteString(fmt.Sprintf("Description: %s\n", cmd.Description()))
		help.WriteString(fmt.Sprintf("Usage: %s\n", cmd.Usage()))

		if platforms := cmd.SupportedPlatforms(); len(platforms) > 0 {
			help.WriteString(fmt.Sprintf("Platforms: %s\n", strings.Join(platforms, ", ")))
		}

		if cmd.RequiresElevation() {
			help.WriteString("⚠️  Requires elevated privileges\n")
		}

		return help.String()
	}

	// Command not found, suggest similar commands
	similar := h.completionEngine.GetSimilarCommands(commandName, 3)
	if len(similar) > 0 {
		return fmt.Sprintf("Command '%s' not found.\n\nDid you mean:\n    %s\n\nUse 'help' to see all available commands.",
			commandName, strings.Join(similar, "\n    "))
	}

	return fmt.Sprintf("Command '%s' not found. Use 'help' to see all available commands.", commandName)
}

// ShowQuickReference displays a quick reference guide
func (h *HelpSystem) ShowQuickReference() string {
	var ref strings.Builder

	ref.WriteString("SuperShell Quick Reference\n")
	ref.WriteString("═════════════════════════\n\n")

	ref.WriteString("🔥 FIREWALL MANAGEMENT\n")
	ref.WriteString("    firewall status                     # Show firewall status\n")
	ref.WriteString("    firewall enable                     # Enable firewall\n")
	ref.WriteString("    firewall rules list                 # List firewall rules\n")
	ref.WriteString("    firewall rules add --port 80 --protocol tcp --action allow\n\n")

	ref.WriteString("📊 PERFORMANCE ANALYSIS\n")
	ref.WriteString("    performance analyze                 # Analyze system performance\n")
	ref.WriteString("    performance monitor --duration 60s  # Monitor for 60 seconds\n")
	ref.WriteString("    performance optimize --auto         # Auto-optimize system\n")
	ref.WriteString("    performance baseline create         # Create performance baseline\n\n")

	ref.WriteString("🖥️  SERVER MANAGEMENT\n")
	ref.WriteString("    server health                       # Show server health\n")
	ref.WriteString("    server services list                # List all services\n")
	ref.WriteString("    server services start nginx         # Start nginx service\n")
	ref.WriteString("    server users --watch                # Monitor user sessions\n\n")

	ref.WriteString("🌐 REMOTE MANAGEMENT\n")
	ref.WriteString("    remote add web1 user@host --key ~/.ssh/id_rsa\n")
	ref.WriteString("    remote list --status                # List servers with status\n")
	ref.WriteString("    remote exec web1 \"uptime\"           # Execute command on server\n")
	ref.WriteString("    remote health --all                 # Check health of all servers\n\n")

	ref.WriteString("🏗️  CLUSTER OPERATIONS\n")
	ref.WriteString("    remote cluster create web-tier web1,web2,web3\n")
	ref.WriteString("    remote cluster exec web-tier \"systemctl status nginx\"\n")
	ref.WriteString("    remote cluster health web-tier      # Check cluster health\n\n")

	ref.WriteString("🔄 CONFIG SYNCHRONIZATION\n")
	ref.WriteString("    remote sync create nginx-config /etc/nginx/ /etc/nginx/ web1,web2\n")
	ref.WriteString("    remote sync run nginx-config        # Run synchronization\n")
	ref.WriteString("    remote sync history                 # Show sync history\n\n")

	ref.WriteString("💡 TIPS\n")
	ref.WriteString("    • Use --json for machine-readable output\n")
	ref.WriteString("    • Use --help with any command for detailed help\n")
	ref.WriteString("    • Use Tab completion for command suggestions\n")
	ref.WriteString("    • Use --watch for real-time monitoring\n")
	ref.WriteString("    • Commands support both short (-h) and long (--help) options\n\n")

	return ref.String()
}

// ShowFeatureOverview displays an overview of new features
func (h *HelpSystem) ShowFeatureOverview() string {
	var overview strings.Builder

	overview.WriteString("SuperShell New Features Overview\n")
	overview.WriteString("═══════════════════════════════\n\n")

	overview.WriteString("🔥 FIREWALL MANAGEMENT\n")
	overview.WriteString("Cross-platform firewall management with unified interface:\n")
	overview.WriteString("• Windows: netsh integration with Windows Firewall\n")
	overview.WriteString("• Linux: iptables and ufw support with sudo handling\n")
	overview.WriteString("• macOS: pfctl integration with system firewall\n")
	overview.WriteString("• Rule management, backup/restore, status monitoring\n\n")

	overview.WriteString("📊 PERFORMANCE ANALYSIS\n")
	overview.WriteString("Comprehensive system performance monitoring and optimization:\n")
	overview.WriteString("• Real-time metrics collection (CPU, memory, disk, network)\n")
	overview.WriteString("• Intelligent bottleneck detection and analysis\n")
	overview.WriteString("• Automated optimization with safety checks\n")
	overview.WriteString("• Baseline management and trend analysis\n")
	overview.WriteString("• Historical performance tracking\n\n")

	overview.WriteString("🖥️  SERVER MANAGEMENT\n")
	overview.WriteString("Local server health monitoring and service management:\n")
	overview.WriteString("• System health monitoring with configurable alerts\n")
	overview.WriteString("• Service discovery and control (start/stop/restart)\n")
	overview.WriteString("• User session monitoring with security insights\n")
	overview.WriteString("• Real-time log streaming and analysis\n")
	overview.WriteString("• Configuration backup and restore\n\n")

	overview.WriteString("🌐 REMOTE MANAGEMENT\n")
	overview.WriteString("SSH-based remote server operations with enterprise features:\n")
	overview.WriteString("• SSH key and password authentication\n")
	overview.WriteString("• Connection pooling and reuse for performance\n")
	overview.WriteString("• Parallel command execution across multiple servers\n")
	overview.WriteString("• Comprehensive health monitoring\n")
	overview.WriteString("• File transfer and synchronization\n\n")

	overview.WriteString("🏗️  CLUSTER OPERATIONS\n")
	overview.WriteString("Multi-server cluster management and coordination:\n")
	overview.WriteString("• Logical server grouping and organization\n")
	overview.WriteString("• Cluster-wide command execution with result aggregation\n")
	overview.WriteString("• Health monitoring across entire clusters\n")
	overview.WriteString("• Success rate tracking and failure analysis\n")
	overview.WriteString("• Configurable concurrency and retry logic\n\n")

	overview.WriteString("🔄 CONFIG SYNCHRONIZATION\n")
	overview.WriteString("Configuration file synchronization across servers:\n")
	overview.WriteString("• File and directory synchronization profiles\n")
	overview.WriteString("• Backup before sync with rollback capabilities\n")
	overview.WriteString("• Checksum validation and integrity checking\n")
	overview.WriteString("• Pre/post command execution hooks\n")
	overview.WriteString("• Permission and ownership management\n\n")

	overview.WriteString("🛡️  ENTERPRISE FEATURES\n")
	overview.WriteString("• Cross-platform compatibility (Windows, Linux, macOS)\n")
	overview.WriteString("• Rich output formatting with color-coded status\n")
	overview.WriteString("• JSON output support for automation\n")
	overview.WriteString("• Comprehensive error handling and validation\n")
	overview.WriteString("• Connection pooling and resource optimization\n")
	overview.WriteString("• Detailed logging and metrics collection\n")
	overview.WriteString("• Auto-completion and contextual help\n\n")

	return overview.String()
}

// GetCommandSyntax returns syntax help for a command
func (h *HelpSystem) GetCommandSyntax(commandName string) string {
	if cmd, err := h.registry.Get(commandName); err == nil {
		return cmd.Usage()
	}
	return fmt.Sprintf("Syntax not available for command: %s", commandName)
}

// SearchCommands searches for commands matching a query
func (h *HelpSystem) SearchCommands(query string) []string {
	query = strings.ToLower(query)
	matches := make([]string, 0)

	for _, cmdName := range h.registry.List() {
		if cmd, err := h.registry.Get(cmdName); err == nil {
			// Search in command name
			if strings.Contains(strings.ToLower(cmdName), query) {
				matches = append(matches, cmdName)
				continue
			}

			// Search in description
			if strings.Contains(strings.ToLower(cmd.Description()), query) {
				matches = append(matches, cmdName)
				continue
			}
		}
	}

	sort.Strings(matches)
	return matches
}
