package firewall

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
	"time"

	"suppercommand/internal/types"
)

// WindowsFirewallManager manages Windows Firewall using netsh commands
type WindowsFirewallManager struct {
	*BaseFirewallManager
}

// NewWindowsFirewallManager creates a new Windows firewall manager
func NewWindowsFirewallManager() *WindowsFirewallManager {
	return &WindowsFirewallManager{
		BaseFirewallManager: NewBaseFirewallManager(types.PlatformWindows),
	}
}

// GetStatus returns the current Windows firewall status
func (w *WindowsFirewallManager) GetStatus(ctx context.Context) (*types.FirewallStatus, error) {
	cmd := exec.CommandContext(ctx, "netsh", "advfirewall", "show", "allprofiles", "state")
	output, err := cmd.Output()
	if err != nil {
		return nil, types.NewFirewallError("get_status", types.PlatformWindows, err, "failed to get firewall status")
	}

	enabled := strings.Contains(string(output), "State                                 ON")

	// Get rule count
	rules, err := w.ListRules(ctx)
	ruleCount := 0
	if err == nil {
		ruleCount = len(rules)
	}

	return &types.FirewallStatus{
		Enabled:     enabled,
		Platform:    types.PlatformWindows,
		RuleCount:   ruleCount,
		LastUpdated: time.Now(),
		Profile:     "Windows Defender Firewall",
		Version:     w.getFirewallVersion(),
	}, nil
}

// ListRules returns all Windows firewall rules
func (w *WindowsFirewallManager) ListRules(ctx context.Context) ([]*types.FirewallRule, error) {
	cmd := exec.CommandContext(ctx, "netsh", "advfirewall", "firewall", "show", "rule", "name=all", "verbose")
	output, err := cmd.Output()
	if err != nil {
		return nil, types.NewFirewallError("list_rules", types.PlatformWindows, err, "failed to list firewall rules")
	}

	return w.parseWindowsRules(string(output))
}

// AddRule adds a new Windows firewall rule
func (w *WindowsFirewallManager) AddRule(ctx context.Context, rule *types.FirewallRule) error {
	if err := w.ValidateRule(rule); err != nil {
		return err
	}

	args := w.buildNetshAddCommand(rule)
	cmd := exec.CommandContext(ctx, "netsh", args...)

	if err := cmd.Run(); err != nil {
		return types.NewFirewallError("add_rule", types.PlatformWindows, err,
			fmt.Sprintf("failed to add rule '%s'", rule.Name))
	}

	return nil
}

// RemoveRule removes a Windows firewall rule by name
func (w *WindowsFirewallManager) RemoveRule(ctx context.Context, ruleID string) error {
	// In Windows, we typically remove rules by name
	cmd := exec.CommandContext(ctx, "netsh", "advfirewall", "firewall", "delete", "rule", fmt.Sprintf("name=%s", ruleID))

	if err := cmd.Run(); err != nil {
		return types.NewFirewallError("remove_rule", types.PlatformWindows, err,
			fmt.Sprintf("failed to remove rule '%s'", ruleID))
	}

	return nil
}

// EnableFirewall enables the Windows firewall
func (w *WindowsFirewallManager) EnableFirewall(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "netsh", "advfirewall", "set", "allprofiles", "state", "on")

	if err := cmd.Run(); err != nil {
		return types.NewFirewallError("enable", types.PlatformWindows, err, "failed to enable firewall")
	}

	return nil
}

// DisableFirewall disables the Windows firewall
func (w *WindowsFirewallManager) DisableFirewall(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "netsh", "advfirewall", "set", "allprofiles", "state", "off")

	if err := cmd.Run(); err != nil {
		return types.NewFirewallError("disable", types.PlatformWindows, err, "failed to disable firewall")
	}

	return nil
}

// BackupRules backs up Windows firewall rules to a file
func (w *WindowsFirewallManager) BackupRules(ctx context.Context, filepath string) error {
	rules, err := w.ListRules(ctx)
	if err != nil {
		return types.NewFirewallError("backup", types.PlatformWindows, err, "failed to get rules for backup")
	}

	data, err := json.MarshalIndent(rules, "", "  ")
	if err != nil {
		return types.NewFirewallError("backup", types.PlatformWindows, err, "failed to marshal rules")
	}

	if err := ioutil.WriteFile(filepath, data, 0644); err != nil {
		return types.NewFirewallError("backup", types.PlatformWindows, err, "failed to write backup file")
	}

	return nil
}

// RestoreRules restores Windows firewall rules from a file
func (w *WindowsFirewallManager) RestoreRules(ctx context.Context, filepath string) error {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return types.NewFirewallError("restore", types.PlatformWindows, err, "failed to read backup file")
	}

	var rules []*types.FirewallRule
	if err := json.Unmarshal(data, &rules); err != nil {
		return types.NewFirewallError("restore", types.PlatformWindows, err, "failed to parse backup file")
	}

	// Add each rule
	for _, rule := range rules {
		if err := w.AddRule(ctx, rule); err != nil {
			// Log error but continue with other rules
			continue
		}
	}

	return nil
}

// parseWindowsRules parses the output of netsh firewall rule listing
func (w *WindowsFirewallManager) parseWindowsRules(output string) ([]*types.FirewallRule, error) {
	var rules []*types.FirewallRule

	// Split output into rule blocks
	ruleBlocks := strings.Split(output, "Rule Name:")

	for _, block := range ruleBlocks[1:] { // Skip first empty block
		rule := w.parseWindowsRuleBlock(block)
		if rule != nil {
			rules = append(rules, rule)
		}
	}

	return rules, nil
}

// parseWindowsRuleBlock parses a single Windows firewall rule block
func (w *WindowsFirewallManager) parseWindowsRuleBlock(block string) *types.FirewallRule {
	rule := &types.FirewallRule{
		Created: time.Now(),
		Enabled: true,
	}

	lines := strings.Split(block, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "Rule Name:") {
			rule.Name = strings.TrimSpace(strings.TrimPrefix(line, "Rule Name:"))
			rule.ID = rule.Name // Use name as ID for Windows
		} else if strings.Contains(line, "Direction:") {
			if strings.Contains(line, "In") {
				rule.Direction = types.DirectionInbound
			} else {
				rule.Direction = types.DirectionOutbound
			}
		} else if strings.Contains(line, "Action:") {
			if strings.Contains(line, "Allow") {
				rule.Action = types.ActionAllow
			} else {
				rule.Action = types.ActionBlock
			}
		} else if strings.Contains(line, "Protocol:") {
			protocol := strings.ToLower(strings.TrimSpace(strings.Split(line, ":")[1]))
			switch protocol {
			case "tcp":
				rule.Protocol = types.ProtocolTCP
			case "udp":
				rule.Protocol = types.ProtocolUDP
			case "icmp":
				rule.Protocol = types.ProtocolICMP
			default:
				rule.Protocol = types.ProtocolAny
			}
		} else if strings.Contains(line, "Local Port:") {
			rule.LocalPort = strings.TrimSpace(strings.Split(line, ":")[1])
		} else if strings.Contains(line, "Remote Port:") {
			rule.RemotePort = strings.TrimSpace(strings.Split(line, ":")[1])
		} else if strings.Contains(line, "Local IP:") {
			rule.LocalIP = strings.TrimSpace(strings.Split(line, ":")[1])
		} else if strings.Contains(line, "Remote IP:") {
			rule.RemoteIP = strings.TrimSpace(strings.Split(line, ":")[1])
		}
	}

	// Only return rule if it has required fields
	if rule.Name != "" && rule.Direction != "" && rule.Action != "" {
		return rule
	}

	return nil
}

// buildNetshAddCommand builds the netsh command arguments for adding a rule
func (w *WindowsFirewallManager) buildNetshAddCommand(rule *types.FirewallRule) []string {
	args := []string{"advfirewall", "firewall", "add", "rule"}

	args = append(args, fmt.Sprintf("name=%s", rule.Name))
	args = append(args, fmt.Sprintf("dir=%s", w.mapDirection(rule.Direction)))
	args = append(args, fmt.Sprintf("action=%s", w.mapAction(rule.Action)))

	if rule.Protocol != "" {
		args = append(args, fmt.Sprintf("protocol=%s", w.mapProtocol(rule.Protocol)))
	}

	if rule.LocalPort != "" {
		args = append(args, fmt.Sprintf("localport=%s", rule.LocalPort))
	}

	if rule.RemotePort != "" {
		args = append(args, fmt.Sprintf("remoteport=%s", rule.RemotePort))
	}

	if rule.LocalIP != "" {
		args = append(args, fmt.Sprintf("localip=%s", rule.LocalIP))
	}

	if rule.RemoteIP != "" {
		args = append(args, fmt.Sprintf("remoteip=%s", rule.RemoteIP))
	}

	return args
}

// mapDirection maps internal direction to Windows netsh direction
func (w *WindowsFirewallManager) mapDirection(direction types.Direction) string {
	switch direction {
	case types.DirectionInbound:
		return "in"
	case types.DirectionOutbound:
		return "out"
	default:
		return "in"
	}
}

// mapAction maps internal action to Windows netsh action
func (w *WindowsFirewallManager) mapAction(action types.Action) string {
	switch action {
	case types.ActionAllow:
		return "allow"
	case types.ActionBlock:
		return "block"
	default:
		return "allow"
	}
}

// mapProtocol maps internal protocol to Windows netsh protocol
func (w *WindowsFirewallManager) mapProtocol(protocol types.Protocol) string {
	switch protocol {
	case types.ProtocolTCP:
		return "TCP"
	case types.ProtocolUDP:
		return "UDP"
	case types.ProtocolICMP:
		return "ICMPv4"
	case types.ProtocolAny:
		return "any"
	default:
		return "any"
	}
}

// getFirewallVersion gets the Windows firewall version
func (w *WindowsFirewallManager) getFirewallVersion() string {
	// This is a simplified version - in reality you might query the actual version
	return "Windows Defender Firewall"
}
