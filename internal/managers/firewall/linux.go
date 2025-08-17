package firewall

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"suppercommand/internal/types"
)

// LinuxFirewallManager manages Linux firewall using ufw and iptables
type LinuxFirewallManager struct {
	*BaseFirewallManager
	useUfw bool // Whether to use ufw or iptables directly
}

// NewLinuxFirewallManager creates a new Linux firewall manager
func NewLinuxFirewallManager() *LinuxFirewallManager {
	manager := &LinuxFirewallManager{
		BaseFirewallManager: NewBaseFirewallManager(types.PlatformLinux),
	}

	// Check if ufw is available
	manager.useUfw = manager.isUfwAvailable()

	return manager
}

// isUfwAvailable checks if ufw is installed and available
func (l *LinuxFirewallManager) isUfwAvailable() bool {
	cmd := exec.Command("which", "ufw")
	return cmd.Run() == nil
}

// GetStatus returns the current Linux firewall status
func (l *LinuxFirewallManager) GetStatus(ctx context.Context) (*types.FirewallStatus, error) {
	if l.useUfw {
		return l.getUfwStatus(ctx)
	}
	return l.getIptablesStatus(ctx)
}

// getUfwStatus gets status using ufw
func (l *LinuxFirewallManager) getUfwStatus(ctx context.Context) (*types.FirewallStatus, error) {
	cmd := exec.CommandContext(ctx, "ufw", "status", "verbose")
	output, err := cmd.Output()
	if err != nil {
		return nil, types.NewFirewallError("get_status", types.PlatformLinux, err, "failed to get ufw status")
	}

	outputStr := string(output)
	enabled := strings.Contains(outputStr, "Status: active")

	// Count rules
	rules, err := l.ListRules(ctx)
	ruleCount := 0
	if err == nil {
		ruleCount = len(rules)
	}

	return &types.FirewallStatus{
		Enabled:     enabled,
		Platform:    types.PlatformLinux,
		RuleCount:   ruleCount,
		LastUpdated: time.Now(),
		Profile:     "UFW (Uncomplicated Firewall)",
		Version:     l.getUfwVersion(),
	}, nil
}

// getIptablesStatus gets status using iptables
func (l *LinuxFirewallManager) getIptablesStatus(ctx context.Context) (*types.FirewallStatus, error) {
	cmd := exec.CommandContext(ctx, "iptables", "-L", "-n")
	output, err := cmd.Output()
	if err != nil {
		return nil, types.NewFirewallError("get_status", types.PlatformLinux, err, "failed to get iptables status")
	}

	// Check if there are any rules (basic detection)
	enabled := !strings.Contains(string(output), "policy ACCEPT") ||
		strings.Count(string(output), "\n") > 10

	rules, err := l.ListRules(ctx)
	ruleCount := 0
	if err == nil {
		ruleCount = len(rules)
	}

	return &types.FirewallStatus{
		Enabled:     enabled,
		Platform:    types.PlatformLinux,
		RuleCount:   ruleCount,
		LastUpdated: time.Now(),
		Profile:     "iptables",
		Version:     l.getIptablesVersion(),
	}, nil
}

// ListRules returns all Linux firewall rules
func (l *LinuxFirewallManager) ListRules(ctx context.Context) ([]*types.FirewallRule, error) {
	if l.useUfw {
		return l.listUfwRules(ctx)
	}
	return l.listIptablesRules(ctx)
}

// listUfwRules lists rules using ufw
func (l *LinuxFirewallManager) listUfwRules(ctx context.Context) ([]*types.FirewallRule, error) {
	cmd := exec.CommandContext(ctx, "ufw", "status", "numbered")
	output, err := cmd.Output()
	if err != nil {
		return nil, types.NewFirewallError("list_rules", types.PlatformLinux, err, "failed to list ufw rules")
	}

	return l.parseUfwRules(string(output))
}

// listIptablesRules lists rules using iptables
func (l *LinuxFirewallManager) listIptablesRules(ctx context.Context) ([]*types.FirewallRule, error) {
	cmd := exec.CommandContext(ctx, "iptables", "-L", "-n", "--line-numbers")
	output, err := cmd.Output()
	if err != nil {
		return nil, types.NewFirewallError("list_rules", types.PlatformLinux, err, "failed to list iptables rules")
	}

	return l.parseIptablesRules(string(output))
}

// AddRule adds a new Linux firewall rule
func (l *LinuxFirewallManager) AddRule(ctx context.Context, rule *types.FirewallRule) error {
	// Check if firewall tools are available
	if err := l.checkFirewallTools(); err != nil {
		return err
	}

	// Validate rule with Linux-specific checks
	if err := l.validateLinuxRule(rule); err != nil {
		return err
	}

	// Check privileges
	if err := l.validatePrivileges("add_rule"); err != nil {
		return err
	}

	// Generate rule ID if not provided
	if rule.ID == "" {
		rule.ID = l.generateRuleID(rule)
	}

	if l.useUfw {
		return l.addUfwRule(ctx, rule)
	}
	return l.addIptablesRule(ctx, rule)
}

// generateRuleID generates a unique ID for a Linux firewall rule
func (l *LinuxFirewallManager) generateRuleID(rule *types.FirewallRule) string {
	// Create a unique ID based on rule properties
	id := fmt.Sprintf("%s_%s_%s",
		rule.Direction,
		rule.Protocol,
		rule.Action)

	if rule.LocalPort != "" {
		id += "_" + rule.LocalPort
	}

	if rule.RemoteIP != "" {
		// Replace dots with underscores for ID
		cleanIP := strings.ReplaceAll(rule.RemoteIP, ".", "_")
		id += "_" + cleanIP
	}

	// Add timestamp to ensure uniqueness
	id += "_" + fmt.Sprintf("%d", time.Now().Unix())

	return id
}

// addUfwRule adds a rule using ufw
func (l *LinuxFirewallManager) addUfwRule(ctx context.Context, rule *types.FirewallRule) error {
	args := l.buildUfwCommand(rule)
	cmd := l.createSudoCommand(ctx, "ufw", args...)

	if err := cmd.Run(); err != nil {
		return types.NewFirewallError("add_rule", types.PlatformLinux, err,
			fmt.Sprintf("failed to add ufw rule '%s'", rule.Name))
	}

	return nil
}

// addIptablesRule adds a rule using iptables
func (l *LinuxFirewallManager) addIptablesRule(ctx context.Context, rule *types.FirewallRule) error {
	args := l.buildIptablesCommand(rule)
	cmd := l.createSudoCommand(ctx, "iptables", args...)

	if err := cmd.Run(); err != nil {
		return types.NewFirewallError("add_rule", types.PlatformLinux, err,
			fmt.Sprintf("failed to add iptables rule '%s'", rule.Name))
	}

	return nil
}

// RemoveRule removes a Linux firewall rule
func (l *LinuxFirewallManager) RemoveRule(ctx context.Context, ruleID string) error {
	if err := l.validatePrivileges("remove_rule"); err != nil {
		return err
	}

	if l.useUfw {
		return l.removeUfwRule(ctx, ruleID)
	}
	return l.removeIptablesRule(ctx, ruleID)
}

// removeUfwRule removes a rule using ufw
func (l *LinuxFirewallManager) removeUfwRule(ctx context.Context, ruleID string) error {
	// For ufw, ruleID might be a number or rule description
	cmd := l.createSudoCommand(ctx, "ufw", "delete", ruleID)

	if err := cmd.Run(); err != nil {
		return types.NewFirewallError("remove_rule", types.PlatformLinux, err,
			fmt.Sprintf("failed to remove ufw rule '%s'", ruleID))
	}

	return nil
}

// removeIptablesRule removes a rule using iptables
func (l *LinuxFirewallManager) removeIptablesRule(ctx context.Context, ruleID string) error {
	// For iptables, ruleID should be in format "chain:line_number"
	parts := strings.Split(ruleID, ":")
	if len(parts) != 2 {
		return types.NewFirewallError("remove_rule", types.PlatformLinux, nil,
			"invalid rule ID format, expected 'chain:line_number'")
	}

	chain := parts[0]
	lineNum := parts[1]

	cmd := l.createSudoCommand(ctx, "iptables", "-D", chain, lineNum)

	if err := cmd.Run(); err != nil {
		return types.NewFirewallError("remove_rule", types.PlatformLinux, err,
			fmt.Sprintf("failed to remove iptables rule '%s'", ruleID))
	}

	return nil
}

// EnableFirewall enables the Linux firewall
func (l *LinuxFirewallManager) EnableFirewall(ctx context.Context) error {
	if err := l.validatePrivileges("enable"); err != nil {
		return err
	}

	if l.useUfw {
		cmd := l.createSudoCommand(ctx, "ufw", "enable")
		if err := cmd.Run(); err != nil {
			return types.NewFirewallError("enable", types.PlatformLinux, err, "failed to enable ufw")
		}
	} else {
		// For iptables, we set default policies to be more restrictive
		commands := [][]string{
			{"iptables", "-P", "INPUT", "DROP"},
			{"iptables", "-P", "FORWARD", "DROP"},
			{"iptables", "-P", "OUTPUT", "ACCEPT"},
		}

		for _, cmdArgs := range commands {
			cmd := l.createSudoCommand(ctx, cmdArgs[0], cmdArgs[1:]...)
			if err := cmd.Run(); err != nil {
				return types.NewFirewallError("enable", types.PlatformLinux, err,
					fmt.Sprintf("failed to set iptables policy: %v", cmdArgs))
			}
		}
	}

	return nil
}

// DisableFirewall disables the Linux firewall
func (l *LinuxFirewallManager) DisableFirewall(ctx context.Context) error {
	if err := l.validatePrivileges("disable"); err != nil {
		return err
	}

	if l.useUfw {
		cmd := l.createSudoCommand(ctx, "ufw", "disable")
		if err := cmd.Run(); err != nil {
			return types.NewFirewallError("disable", types.PlatformLinux, err, "failed to disable ufw")
		}
	} else {
		// For iptables, we set permissive policies
		commands := [][]string{
			{"iptables", "-P", "INPUT", "ACCEPT"},
			{"iptables", "-P", "FORWARD", "ACCEPT"},
			{"iptables", "-P", "OUTPUT", "ACCEPT"},
			{"iptables", "-F"}, // Flush all rules
		}

		for _, cmdArgs := range commands {
			cmd := l.createSudoCommand(ctx, cmdArgs[0], cmdArgs[1:]...)
			if err := cmd.Run(); err != nil {
				return types.NewFirewallError("disable", types.PlatformLinux, err,
					fmt.Sprintf("failed to reset iptables: %v", cmdArgs))
			}
		}
	}

	return nil
}

// BackupRules backs up Linux firewall rules to a file
func (l *LinuxFirewallManager) BackupRules(ctx context.Context, filepath string) error {
	rules, err := l.ListRules(ctx)
	if err != nil {
		return types.NewFirewallError("backup", types.PlatformLinux, err, "failed to get rules for backup")
	}

	data, err := json.MarshalIndent(rules, "", "  ")
	if err != nil {
		return types.NewFirewallError("backup", types.PlatformLinux, err, "failed to marshal rules")
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return types.NewFirewallError("backup", types.PlatformLinux, err, "failed to write backup file")
	}

	return nil
}

// RestoreRules restores Linux firewall rules from a file
func (l *LinuxFirewallManager) RestoreRules(ctx context.Context, filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return types.NewFirewallError("restore", types.PlatformLinux, err, "failed to read backup file")
	}

	var rules []*types.FirewallRule
	if err := json.Unmarshal(data, &rules); err != nil {
		return types.NewFirewallError("restore", types.PlatformLinux, err, "failed to parse backup file")
	}

	// Add each rule
	for _, rule := range rules {
		if err := l.AddRule(ctx, rule); err != nil {
			// Log error but continue with other rules
			continue
		}
	}

	return nil
}

// parseUfwRules parses ufw status output
func (l *LinuxFirewallManager) parseUfwRules(output string) ([]*types.FirewallRule, error) {
	var rules []*types.FirewallRule

	lines := strings.Split(output, "\n")
	ruleRegex := regexp.MustCompile(`\[\s*(\d+)\]\s+(.+)`)

	for _, line := range lines {
		matches := ruleRegex.FindStringSubmatch(line)
		if len(matches) >= 3 {
			rule := l.parseUfwRuleLine(matches[1], matches[2])
			if rule != nil {
				rules = append(rules, rule)
			}
		}
	}

	return rules, nil
}

// parseUfwRuleLine parses a single ufw rule line
func (l *LinuxFirewallManager) parseUfwRuleLine(number, ruleLine string) *types.FirewallRule {
	rule := &types.FirewallRule{
		ID:      number,
		Name:    fmt.Sprintf("ufw-rule-%s", number),
		Created: time.Now(),
		Enabled: true,
	}

	// Parse rule components from the rule line
	// Enhanced parsing for better accuracy
	ruleLine = strings.TrimSpace(ruleLine)

	// Parse action
	if strings.Contains(strings.ToUpper(ruleLine), "ALLOW") {
		rule.Action = types.ActionAllow
	} else if strings.Contains(strings.ToUpper(ruleLine), "DENY") {
		rule.Action = types.ActionBlock
	} else if strings.Contains(strings.ToUpper(ruleLine), "REJECT") {
		rule.Action = types.ActionBlock
	} else {
		rule.Action = types.ActionAllow // Default
	}

	// Parse direction
	if strings.Contains(strings.ToUpper(ruleLine), " IN ") {
		rule.Direction = types.DirectionInbound
	} else if strings.Contains(strings.ToUpper(ruleLine), " OUT ") {
		rule.Direction = types.DirectionOutbound
	} else {
		rule.Direction = types.DirectionInbound // Default
	}

	// Extract port and protocol information with improved regex
	portProtocolRegex := regexp.MustCompile(`(\d+)/(tcp|udp|icmp)`)
	matches := portProtocolRegex.FindStringSubmatch(strings.ToLower(ruleLine))
	if len(matches) >= 3 {
		rule.LocalPort = matches[1]
		switch matches[2] {
		case "tcp":
			rule.Protocol = types.ProtocolTCP
		case "udp":
			rule.Protocol = types.ProtocolUDP
		case "icmp":
			rule.Protocol = types.ProtocolICMP
		default:
			rule.Protocol = types.ProtocolTCP
		}
	} else {
		// Try to extract just port number
		portRegex := regexp.MustCompile(`\b(\d+)\b`)
		portMatches := portRegex.FindStringSubmatch(ruleLine)
		if len(portMatches) >= 2 {
			rule.LocalPort = portMatches[1]
			rule.Protocol = types.ProtocolTCP // Default to TCP
		} else {
			rule.Protocol = types.ProtocolAny
		}
	}

	// Extract IP addresses
	ipRegex := regexp.MustCompile(`(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})`)
	ipMatches := ipRegex.FindAllString(ruleLine, -1)
	if len(ipMatches) > 0 {
		// First IP is usually the source/remote IP
		rule.RemoteIP = ipMatches[0]
		if len(ipMatches) > 1 {
			rule.LocalIP = ipMatches[1]
		}
	}

	// Set description from the original rule line
	rule.Description = strings.TrimSpace(ruleLine)

	return rule
}

// parseIptablesRules parses iptables output
func (l *LinuxFirewallManager) parseIptablesRules(output string) ([]*types.FirewallRule, error) {
	var rules []*types.FirewallRule

	lines := strings.Split(output, "\n")
	currentChain := ""

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "Chain ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				currentChain = parts[1]
			}
			continue
		}

		if strings.Contains(line, "target") && strings.Contains(line, "prot") {
			// Skip header line
			continue
		}

		rule := l.parseIptablesRuleLine(line, currentChain)
		if rule != nil {
			rules = append(rules, rule)
		}
	}

	return rules, nil
}

// parseIptablesRuleLine parses a single iptables rule line
func (l *LinuxFirewallManager) parseIptablesRuleLine(line, chain string) *types.FirewallRule {
	fields := strings.Fields(line)
	if len(fields) < 4 {
		return nil
	}

	rule := &types.FirewallRule{
		Created: time.Now(),
		Enabled: true,
	}

	// Parse fields: num target prot opt source destination [additional]
	if len(fields) >= 1 {
		rule.ID = fmt.Sprintf("%s:%s", chain, fields[0])
		rule.Name = fmt.Sprintf("iptables-%s-%s", chain, fields[0])
	}

	if len(fields) >= 2 {
		target := strings.ToUpper(fields[1])
		switch target {
		case "ACCEPT":
			rule.Action = types.ActionAllow
		case "DROP", "REJECT":
			rule.Action = types.ActionBlock
		default:
			rule.Action = types.ActionAllow // Default for unknown targets
		}
	}

	if len(fields) >= 3 {
		protocol := strings.ToLower(fields[2])
		switch protocol {
		case "tcp":
			rule.Protocol = types.ProtocolTCP
		case "udp":
			rule.Protocol = types.ProtocolUDP
		case "icmp":
			rule.Protocol = types.ProtocolICMP
		case "all":
			rule.Protocol = types.ProtocolAny
		default:
			rule.Protocol = types.ProtocolAny
		}
	}

	// Parse source and destination (fields 4 and 5)
	if len(fields) >= 5 {
		source := fields[4]
		destination := fields[5]

		if source != "0.0.0.0/0" && source != "anywhere" {
			rule.RemoteIP = source
		}

		if destination != "0.0.0.0/0" && destination != "anywhere" {
			rule.LocalIP = destination
		}
	}

	// Parse additional options for port information
	fullLine := strings.Join(fields, " ")

	// Look for dpt: (destination port) or spt: (source port)
	dptRegex := regexp.MustCompile(`dpt:(\d+)`)
	if matches := dptRegex.FindStringSubmatch(fullLine); len(matches) >= 2 {
		rule.LocalPort = matches[1]
	}

	sptRegex := regexp.MustCompile(`spt:(\d+)`)
	if matches := sptRegex.FindStringSubmatch(fullLine); len(matches) >= 2 {
		rule.RemotePort = matches[1]
	}

	// Set direction based on chain
	switch strings.ToUpper(chain) {
	case "INPUT":
		rule.Direction = types.DirectionInbound
	case "OUTPUT":
		rule.Direction = types.DirectionOutbound
	case "FORWARD":
		rule.Direction = types.DirectionInbound // Treat forward as inbound
	default:
		rule.Direction = types.DirectionInbound
	}

	// Set description from the original line
	rule.Description = fmt.Sprintf("iptables %s rule: %s", chain, strings.TrimSpace(line))

	return rule
}

// buildUfwCommand builds ufw command arguments
func (l *LinuxFirewallManager) buildUfwCommand(rule *types.FirewallRule) []string {
	args := []string{}

	// Add action
	switch rule.Action {
	case types.ActionAllow:
		args = append(args, "allow")
	case types.ActionBlock:
		args = append(args, "deny")
	default:
		args = append(args, "allow") // Default to allow
	}

	// Add direction
	if rule.Direction == types.DirectionOutbound {
		args = append(args, "out")
	}

	// Add from/to IP addresses
	if rule.RemoteIP != "" {
		args = append(args, "from", rule.RemoteIP)
	}

	if rule.LocalIP != "" {
		args = append(args, "to", rule.LocalIP)
	}

	// Add port and protocol
	if rule.LocalPort != "" {
		if rule.Protocol != "" && rule.Protocol != types.ProtocolAny {
			portSpec := fmt.Sprintf("%s/%s", rule.LocalPort, strings.ToLower(string(rule.Protocol)))
			args = append(args, portSpec)
		} else {
			// If no protocol specified, default to TCP
			portSpec := fmt.Sprintf("%s/tcp", rule.LocalPort)
			args = append(args, portSpec)
		}
	} else if rule.Protocol != "" && rule.Protocol != types.ProtocolAny {
		// Protocol without port
		args = append(args, "proto", strings.ToLower(string(rule.Protocol)))
	}

	return args
}

// buildIptablesCommand builds iptables command arguments
func (l *LinuxFirewallManager) buildIptablesCommand(rule *types.FirewallRule) []string {
	args := []string{"-A"}

	// Determine chain based on direction
	if rule.Direction == types.DirectionInbound {
		args = append(args, "INPUT")
	} else {
		args = append(args, "OUTPUT")
	}

	// Add protocol
	if rule.Protocol != "" && rule.Protocol != types.ProtocolAny {
		args = append(args, "-p", strings.ToLower(string(rule.Protocol)))
	}

	// Add source IP
	if rule.RemoteIP != "" {
		args = append(args, "-s", rule.RemoteIP)
	}

	// Add destination IP
	if rule.LocalIP != "" {
		args = append(args, "-d", rule.LocalIP)
	}

	// Add ports based on direction
	if rule.LocalPort != "" {
		if rule.Direction == types.DirectionInbound {
			args = append(args, "--dport", rule.LocalPort)
		} else {
			args = append(args, "--sport", rule.LocalPort)
		}
	}

	if rule.RemotePort != "" {
		if rule.Direction == types.DirectionInbound {
			args = append(args, "--sport", rule.RemotePort)
		} else {
			args = append(args, "--dport", rule.RemotePort)
		}
	}

	// Add action
	switch rule.Action {
	case types.ActionAllow:
		args = append(args, "-j", "ACCEPT")
	case types.ActionBlock:
		args = append(args, "-j", "DROP")
	default:
		args = append(args, "-j", "ACCEPT") // Default to accept
	}

	// Add comment if name is provided
	if rule.Name != "" {
		args = append(args, "-m", "comment", "--comment", rule.Name)
	}

	return args
}

// getUfwVersion gets the ufw version
func (l *LinuxFirewallManager) getUfwVersion() string {
	cmd := exec.Command("ufw", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

// getIptablesVersion gets the iptables version
func (l *LinuxFirewallManager) getIptablesVersion() string {
	cmd := exec.Command("iptables", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

// createSudoCommand creates a command with sudo if needed
func (l *LinuxFirewallManager) createSudoCommand(ctx context.Context, command string, args ...string) *exec.Cmd {
	// Check if we're already running as root
	if l.isRunningAsRoot() {
		return exec.CommandContext(ctx, command, args...)
	}

	// Check if sudo is available
	if !l.isSudoAvailable() {
		// If sudo is not available, try running without it
		// This might fail, but we'll let the caller handle the error
		return exec.CommandContext(ctx, command, args...)
	}

	// Prepare sudo command
	sudoArgs := []string{command}
	sudoArgs = append(sudoArgs, args...)
	return exec.CommandContext(ctx, "sudo", sudoArgs...)
}

// isRunningAsRoot checks if the current process is running as root
func (l *LinuxFirewallManager) isRunningAsRoot() bool {
	return os.Geteuid() == 0
}

// isSudoAvailable checks if sudo is available on the system
func (l *LinuxFirewallManager) isSudoAvailable() bool {
	cmd := exec.Command("which", "sudo")
	return cmd.Run() == nil
}

// needsPrivileges checks if an operation needs elevated privileges
func (l *LinuxFirewallManager) needsPrivileges(operation string) bool {
	// Most firewall operations need root privileges
	privilegedOps := map[string]bool{
		"add_rule":    true,
		"remove_rule": true,
		"enable":      true,
		"disable":     true,
		"modify":      true,
	}

	return privilegedOps[operation]
}

// validatePrivileges validates that we have the necessary privileges for an operation
func (l *LinuxFirewallManager) validatePrivileges(operation string) error {
	if !l.needsPrivileges(operation) {
		return nil
	}

	if l.isRunningAsRoot() {
		return nil
	}

	if !l.isSudoAvailable() {
		return types.NewFirewallError(operation, types.PlatformLinux, nil,
			"operation requires root privileges but sudo is not available")
	}

	// Test sudo access
	cmd := exec.Command("sudo", "-n", "true")
	if err := cmd.Run(); err != nil {
		return types.NewFirewallError(operation, types.PlatformLinux, err,
			"operation requires root privileges - please run with sudo or configure passwordless sudo")
	}

	return nil
}

// validateLinuxRule validates a rule for Linux-specific requirements
func (l *LinuxFirewallManager) validateLinuxRule(rule *types.FirewallRule) error {
	// Basic validation first
	if err := rule.Validate(); err != nil {
		return err
	}

	// Linux-specific validations
	if rule.LocalPort != "" {
		if err := l.validatePort(rule.LocalPort); err != nil {
			return types.NewFirewallError("validate", types.PlatformLinux, err,
				fmt.Sprintf("invalid local port '%s'", rule.LocalPort))
		}
	}

	if rule.RemotePort != "" {
		if err := l.validatePort(rule.RemotePort); err != nil {
			return types.NewFirewallError("validate", types.PlatformLinux, err,
				fmt.Sprintf("invalid remote port '%s'", rule.RemotePort))
		}
	}

	if rule.LocalIP != "" {
		if err := l.validateIP(rule.LocalIP); err != nil {
			return types.NewFirewallError("validate", types.PlatformLinux, err,
				fmt.Sprintf("invalid local IP '%s'", rule.LocalIP))
		}
	}

	if rule.RemoteIP != "" {
		if err := l.validateIP(rule.RemoteIP); err != nil {
			return types.NewFirewallError("validate", types.PlatformLinux, err,
				fmt.Sprintf("invalid remote IP '%s'", rule.RemoteIP))
		}
	}

	return nil
}

// validatePort validates a port number or range
func (l *LinuxFirewallManager) validatePort(port string) error {
	// Handle port ranges (e.g., "80:90")
	if strings.Contains(port, ":") {
		parts := strings.Split(port, ":")
		if len(parts) != 2 {
			return fmt.Errorf("invalid port range format")
		}

		for _, part := range parts {
			if err := l.validateSinglePort(part); err != nil {
				return err
			}
		}
		return nil
	}

	return l.validateSinglePort(port)
}

// validateSinglePort validates a single port number
func (l *LinuxFirewallManager) validateSinglePort(port string) error {
	if port == "" {
		return fmt.Errorf("port cannot be empty")
	}

	// Convert to int and validate range
	portNum := 0
	if _, err := fmt.Sscanf(port, "%d", &portNum); err != nil {
		return fmt.Errorf("port must be a number")
	}

	if portNum < 1 || portNum > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}

	return nil
}

// validateIP validates an IP address or CIDR notation
func (l *LinuxFirewallManager) validateIP(ip string) error {
	if ip == "" {
		return fmt.Errorf("IP address cannot be empty")
	}

	// Handle CIDR notation
	if strings.Contains(ip, "/") {
		parts := strings.Split(ip, "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid CIDR format")
		}

		// Validate IP part
		if err := l.validateSingleIP(parts[0]); err != nil {
			return err
		}

		// Validate subnet mask
		mask := 0
		if _, err := fmt.Sscanf(parts[1], "%d", &mask); err != nil {
			return fmt.Errorf("invalid subnet mask")
		}

		if mask < 0 || mask > 32 {
			return fmt.Errorf("subnet mask must be between 0 and 32")
		}

		return nil
	}

	return l.validateSingleIP(ip)
}

// validateSingleIP validates a single IP address
func (l *LinuxFirewallManager) validateSingleIP(ip string) error {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return fmt.Errorf("invalid IP address format")
	}

	for _, part := range parts {
		octet := 0
		if _, err := fmt.Sscanf(part, "%d", &octet); err != nil {
			return fmt.Errorf("invalid IP address octet")
		}

		if octet < 0 || octet > 255 {
			return fmt.Errorf("IP address octet must be between 0 and 255")
		}
	}

	return nil
}

// checkFirewallTools checks if required firewall tools are available
func (l *LinuxFirewallManager) checkFirewallTools() error {
	if l.useUfw {
		if !l.isUfwAvailable() {
			return types.NewFirewallError("check_tools", types.PlatformLinux, nil,
				"ufw is not available on this system")
		}
	} else {
		// Check for iptables
		cmd := exec.Command("which", "iptables")
		if err := cmd.Run(); err != nil {
			return types.NewFirewallError("check_tools", types.PlatformLinux, err,
				"iptables is not available on this system")
		}
	}

	return nil
}
