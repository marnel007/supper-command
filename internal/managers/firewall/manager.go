package firewall

import (
	"context"
	"fmt"
	"suppercommand/internal/types"
	"suppercommand/internal/utils"
	"time"
)

// Factory creates firewall managers based on the current platform
type Factory struct{}

// NewFactory creates a new firewall manager factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateManager creates a firewall manager for the current platform
func (f *Factory) CreateManager() (types.FirewallManager, error) {
	platform := utils.GetCurrentPlatform()

	switch platform {
	case types.PlatformWindows:
		return NewWindowsFirewallManager(), nil
	case types.PlatformLinux:
		return NewLinuxFirewallManager(), nil
	case types.PlatformDarwin:
		return NewDarwinFirewallManager(), nil
	default:
		return nil, types.NewFirewallError("create_manager", platform, nil, "unsupported platform")
	}
}

// BaseFirewallManager provides common functionality for all firewall managers
type BaseFirewallManager struct {
	platform types.Platform
}

// NewBaseFirewallManager creates a new base firewall manager
func NewBaseFirewallManager(platform types.Platform) *BaseFirewallManager {
	return &BaseFirewallManager{
		platform: platform,
	}
}

// GetPlatform returns the platform this manager is for
func (b *BaseFirewallManager) GetPlatform() types.Platform {
	return b.platform
}

// ValidateRule validates a firewall rule
func (b *BaseFirewallManager) ValidateRule(rule *types.FirewallRule) error {
	if rule == nil {
		return types.NewFirewallError("validate", b.platform, nil, "rule cannot be nil")
	}

	return rule.Validate()
}

// GenerateRuleID generates a unique ID for a firewall rule
func (b *BaseFirewallManager) GenerateRuleID(rule *types.FirewallRule) string {
	// Simple ID generation based on rule properties
	// In a real implementation, this might use UUIDs or other unique identifiers
	return fmt.Sprintf("%s_%s_%s_%s",
		rule.Direction,
		rule.Protocol,
		rule.LocalPort,
		rule.Action)
}

// MockFirewallManager provides a mock implementation for testing
type MockFirewallManager struct {
	*BaseFirewallManager
	rules   []*types.FirewallRule
	enabled bool
}

// NewMockFirewallManager creates a new mock firewall manager
func NewMockFirewallManager() *MockFirewallManager {
	return &MockFirewallManager{
		BaseFirewallManager: NewBaseFirewallManager(utils.GetCurrentPlatform()),
		rules:               make([]*types.FirewallRule, 0),
		enabled:             true,
	}
}

// GetStatus returns the mock firewall status
func (m *MockFirewallManager) GetStatus(ctx context.Context) (*types.FirewallStatus, error) {
	return &types.FirewallStatus{
		Enabled:     m.enabled,
		Platform:    m.platform,
		RuleCount:   len(m.rules),
		LastUpdated: time.Now(),
		Profile:     "mock",
		Version:     "1.0.0",
	}, nil
}

// ListRules returns the mock firewall rules
func (m *MockFirewallManager) ListRules(ctx context.Context) ([]*types.FirewallRule, error) {
	return m.rules, nil
}

// AddRule adds a rule to the mock firewall
func (m *MockFirewallManager) AddRule(ctx context.Context, rule *types.FirewallRule) error {
	if err := m.ValidateRule(rule); err != nil {
		return err
	}

	rule.ID = m.GenerateRuleID(rule)
	m.rules = append(m.rules, rule)
	return nil
}

// RemoveRule removes a rule from the mock firewall
func (m *MockFirewallManager) RemoveRule(ctx context.Context, ruleID string) error {
	for i, rule := range m.rules {
		if rule.ID == ruleID {
			m.rules = append(m.rules[:i], m.rules[i+1:]...)
			return nil
		}
	}
	return types.NewFirewallError("remove_rule", m.platform, nil, "rule not found")
}

// EnableFirewall enables the mock firewall
func (m *MockFirewallManager) EnableFirewall(ctx context.Context) error {
	m.enabled = true
	return nil
}

// DisableFirewall disables the mock firewall
func (m *MockFirewallManager) DisableFirewall(ctx context.Context) error {
	m.enabled = false
	return nil
}

// BackupRules backs up the mock firewall rules
func (m *MockFirewallManager) BackupRules(ctx context.Context, filepath string) error {
	// Mock implementation - would write rules to file in real implementation
	return nil
}

// RestoreRules restores the mock firewall rules
func (m *MockFirewallManager) RestoreRules(ctx context.Context, filepath string) error {
	// Mock implementation - would read rules from file in real implementation
	return nil
}
