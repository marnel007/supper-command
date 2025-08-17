package types

import (
	"context"
	"time"
)

// FirewallManager defines the interface for firewall management operations
type FirewallManager interface {
	GetStatus(ctx context.Context) (*FirewallStatus, error)
	ListRules(ctx context.Context) ([]*FirewallRule, error)
	AddRule(ctx context.Context, rule *FirewallRule) error
	RemoveRule(ctx context.Context, ruleID string) error
	EnableFirewall(ctx context.Context) error
	DisableFirewall(ctx context.Context) error
	BackupRules(ctx context.Context, filepath string) error
	RestoreRules(ctx context.Context, filepath string) error
}

// FirewallStatus represents the current firewall status
type FirewallStatus struct {
	Enabled     bool      `json:"enabled"`
	Platform    Platform  `json:"platform"`
	RuleCount   int       `json:"rule_count"`
	LastUpdated time.Time `json:"last_updated"`
	Profile     string    `json:"profile,omitempty"`
	Version     string    `json:"version,omitempty"`
}

// FirewallRule represents a firewall rule
type FirewallRule struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Direction   Direction `json:"direction"`
	Action      Action    `json:"action"`
	Protocol    Protocol  `json:"protocol"`
	LocalPort   string    `json:"local_port,omitempty"`
	RemotePort  string    `json:"remote_port,omitempty"`
	LocalIP     string    `json:"local_ip,omitempty"`
	RemoteIP    string    `json:"remote_ip,omitempty"`
	Enabled     bool      `json:"enabled"`
	Created     time.Time `json:"created"`
	Description string    `json:"description,omitempty"`
}

// FirewallRuleBuilder helps build firewall rules
type FirewallRuleBuilder struct {
	rule *FirewallRule
}

// NewFirewallRuleBuilder creates a new firewall rule builder
func NewFirewallRuleBuilder() *FirewallRuleBuilder {
	return &FirewallRuleBuilder{
		rule: &FirewallRule{
			Enabled: true,
			Created: time.Now(),
		},
	}
}

// WithName sets the rule name
func (b *FirewallRuleBuilder) WithName(name string) *FirewallRuleBuilder {
	b.rule.Name = name
	return b
}

// WithDirection sets the rule direction
func (b *FirewallRuleBuilder) WithDirection(direction Direction) *FirewallRuleBuilder {
	b.rule.Direction = direction
	return b
}

// WithAction sets the rule action
func (b *FirewallRuleBuilder) WithAction(action Action) *FirewallRuleBuilder {
	b.rule.Action = action
	return b
}

// WithProtocol sets the rule protocol
func (b *FirewallRuleBuilder) WithProtocol(protocol Protocol) *FirewallRuleBuilder {
	b.rule.Protocol = protocol
	return b
}

// WithLocalPort sets the local port
func (b *FirewallRuleBuilder) WithLocalPort(port string) *FirewallRuleBuilder {
	b.rule.LocalPort = port
	return b
}

// WithRemotePort sets the remote port
func (b *FirewallRuleBuilder) WithRemotePort(port string) *FirewallRuleBuilder {
	b.rule.RemotePort = port
	return b
}

// WithLocalIP sets the local IP address
func (b *FirewallRuleBuilder) WithLocalIP(ip string) *FirewallRuleBuilder {
	b.rule.LocalIP = ip
	return b
}

// WithRemoteIP sets the remote IP address
func (b *FirewallRuleBuilder) WithRemoteIP(ip string) *FirewallRuleBuilder {
	b.rule.RemoteIP = ip
	return b
}

// WithDescription sets the rule description
func (b *FirewallRuleBuilder) WithDescription(description string) *FirewallRuleBuilder {
	b.rule.Description = description
	return b
}

// Build creates the firewall rule
func (b *FirewallRuleBuilder) Build() *FirewallRule {
	return b.rule
}

// Validate validates the firewall rule
func (r *FirewallRule) Validate() error {
	if r.Name == "" {
		return NewFirewallError("validate", "", nil, "rule name is required")
	}
	if r.Direction == "" {
		return NewFirewallError("validate", "", nil, "rule direction is required")
	}
	if r.Action == "" {
		return NewFirewallError("validate", "", nil, "rule action is required")
	}
	if r.Protocol == "" {
		return NewFirewallError("validate", "", nil, "rule protocol is required")
	}
	return nil
}
