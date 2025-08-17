package firewall

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"suppercommand/internal/types"
)

// DarwinFirewallManager manages macOS firewall using pfctl
type DarwinFirewallManager struct {
	*BaseFirewallManager
}

// NewDarwinFirewallManager creates a new Darwin firewall manager
func NewDarwinFirewallManager() *DarwinFirewallManager {
	return &DarwinFirewallManager{
		BaseFirewallManager: NewBaseFirewallManager(types.PlatformDarwin),
	}
}

// GetStatus returns the current macOS firewall status
func (d *DarwinFirewallManager) GetStatus(ctx context.Context) (*types.FirewallStatus, error) {
	// Check if pfctl is enabled
	cmd := exec.CommandContext(ctx, "pfctl", "-s", "info")
	output, err := cmd.Output()
	if err != nil {
		return nil, types.NewFirewallError("get_status", types.PlatformDarwin, err, "failed to get pfctl status")
	}

	enabled := strings.Contains(string(output), "Status: Enabled")

	// Get rule count
	rules, err := d.ListRules(ctx)
	ruleCount := 0
	if err == nil {
		ruleCount = len(rules)
	}

	return &types.FirewallStatus{
		Enabled:     enabled,
		Platform:    types.PlatformDarwin,
		RuleCount:   ruleCount,
		LastUpdated: time.Now(),
		Profile:     "pfctl (Packet Filter)",
		Version:     d.getPfctlVersion(),
	}, nil
}

// ListRules returns all macOS firewall rules
func (d *DarwinFirewallManager) ListRules(ctx context.Context) ([]*types.FirewallRule, error) {
	cmd := exec.CommandContext(ctx, "pfctl", "-s", "rules")
	output, err := cmd.Output()
	if err != nil {
		return nil, types.NewFirewallError("list_rules", types.PlatformDarwin, err, "failed to list pfctl rules")
	}

	return d.parsePfctlRules(string(output))
}

// AddRule adds a new macOS firewall rule
func (d *DarwinFirewallManager) AddRule(ctx context.Context, rule *types.FirewallRule) error {
	if err := d.ValidateRule(rule); err != nil {
		return err
	}

	// For macOS, we would typically add rules to a pf.conf file and reload
	// This is a simplified implementation
	return types.NewFirewallError("add_rule", types.PlatformDarwin, nil,
		"adding rules on macOS requires manual pf.conf configuration")
}

// RemoveRule removes a macOS firewall rule
func (d *DarwinFirewallManager) RemoveRule(ctx context.Context, ruleID string) error {
	return types.NewFirewallError("remove_rule", types.PlatformDarwin, nil,
		"removing rules on macOS requires manual pf.conf configuration")
}

// EnableFirewall enables the macOS firewall
func (d *DarwinFirewallManager) EnableFirewall(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "pfctl", "-e")

	if err := cmd.Run(); err != nil {
		return types.NewFirewallError("enable", types.PlatformDarwin, err, "failed to enable pfctl")
	}

	return nil
}

// DisableFirewall disables the macOS firewall
func (d *DarwinFirewallManager) DisableFirewall(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "pfctl", "-d")

	if err := cmd.Run(); err != nil {
		return types.NewFirewallError("disable", types.PlatformDarwin, err, "failed to disable pfctl")
	}

	return nil
}

// BackupRules backs up macOS firewall rules to a file
func (d *DarwinFirewallManager) BackupRules(ctx context.Context, filepath string) error {
	rules, err := d.ListRules(ctx)
	if err != nil {
		return types.NewFirewallError("backup", types.PlatformDarwin, err, "failed to get rules for backup")
	}

	data, err := json.MarshalIndent(rules, "", "  ")
	if err != nil {
		return types.NewFirewallError("backup", types.PlatformDarwin, err, "failed to marshal rules")
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return types.NewFirewallError("backup", types.PlatformDarwin, err, "failed to write backup file")
	}

	return nil
}

// RestoreRules restores macOS firewall rules from a file
func (d *DarwinFirewallManager) RestoreRules(ctx context.Context, filepath string) error {
	return types.NewFirewallError("restore", types.PlatformDarwin, nil,
		"restoring rules on macOS requires manual pf.conf configuration")
}

// parsePfctlRules parses pfctl rules output
func (d *DarwinFirewallManager) parsePfctlRules(output string) ([]*types.FirewallRule, error) {
	var rules []*types.FirewallRule

	lines := strings.Split(output, "\n")

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		rule := d.parsePfctlRuleLine(line, i)
		if rule != nil {
			rules = append(rules, rule)
		}
	}

	return rules, nil
}

// parsePfctlRuleLine parses a single pfctl rule line
func (d *DarwinFirewallManager) parsePfctlRuleLine(line string, index int) *types.FirewallRule {
	rule := &types.FirewallRule{
		ID:      fmt.Sprintf("pf-rule-%d", index),
		Name:    fmt.Sprintf("pfctl-rule-%d", index),
		Created: time.Now(),
		Enabled: true,
	}

	// Basic parsing of pfctl rules
	// This is simplified - real pfctl rules are more complex

	if strings.Contains(line, "pass") {
		rule.Action = types.ActionAllow
	} else if strings.Contains(line, "block") {
		rule.Action = types.ActionBlock
	}

	if strings.Contains(line, "in") {
		rule.Direction = types.DirectionInbound
	} else if strings.Contains(line, "out") {
		rule.Direction = types.DirectionOutbound
	}

	if strings.Contains(line, "tcp") {
		rule.Protocol = types.ProtocolTCP
	} else if strings.Contains(line, "udp") {
		rule.Protocol = types.ProtocolUDP
	} else if strings.Contains(line, "icmp") {
		rule.Protocol = types.ProtocolICMP
	}

	return rule
}

// getPfctlVersion gets the pfctl version
func (d *DarwinFirewallManager) getPfctlVersion() string {
	// pfctl doesn't have a version flag, so we return a generic version
	return "macOS pfctl"
}
