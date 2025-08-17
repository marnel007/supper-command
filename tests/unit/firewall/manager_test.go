package firewall

import (
	"context"
	"testing"
	"time"

	"suppercommand/internal/managers/firewall"
	"suppercommand/internal/types"
)

func TestFirewallManagerFactory(t *testing.T) {
	factory := firewall.NewFactory()
	if factory == nil {
		t.Fatal("Factory should not be nil")
	}

	manager, err := factory.CreateManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	if manager == nil {
		t.Fatal("Manager should not be nil")
	}
}

func TestFirewallRuleValidation(t *testing.T) {
	tests := []struct {
		name    string
		rule    *types.FirewallRule
		wantErr bool
	}{
		{
			name: "valid rule",
			rule: &types.FirewallRule{
				Name:      "test-rule",
				Action:    types.ActionAllow,
				Direction: types.DirectionInbound,
				Protocol:  types.ProtocolTCP,
				LocalPort: "80",
				Enabled:   true,
			},
			wantErr: false,
		},
		{
			name: "missing name",
			rule: &types.FirewallRule{
				Action:    types.ActionAllow,
				Direction: types.DirectionInbound,
				Protocol:  types.ProtocolTCP,
				LocalPort: "80",
				Enabled:   true,
			},
			wantErr: true,
		},
		{
			name: "missing direction",
			rule: &types.FirewallRule{
				Name:      "test-rule",
				Action:    types.ActionAllow,
				Protocol:  types.ProtocolTCP,
				LocalPort: "80",
				Enabled:   true,
			},
			wantErr: true,
		},
		{
			name: "missing action",
			rule: &types.FirewallRule{
				Name:      "test-rule",
				Direction: types.DirectionInbound,
				Protocol:  types.ProtocolTCP,
				LocalPort: "80",
				Enabled:   true,
			},
			wantErr: true,
		},
		{
			name: "missing protocol",
			rule: &types.FirewallRule{
				Name:      "test-rule",
				Action:    types.ActionAllow,
				Direction: types.DirectionInbound,
				LocalPort: "80",
				Enabled:   true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFirewallStatus(t *testing.T) {
	factory := firewall.NewFactory()
	manager, err := factory.CreateManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	ctx := context.Background()
	status, err := manager.GetStatus(ctx)
	if err != nil {
		t.Fatalf("Failed to get status: %v", err)
	}

	if status == nil {
		t.Fatal("Status should not be nil")
	}

	if status.Platform == "" {
		t.Error("Platform should not be empty")
	}

	if status.LastUpdated.IsZero() {
		t.Error("LastUpdated should not be zero")
	}
}

func TestFirewallRuleLifecycle(t *testing.T) {
	factory := firewall.NewFactory()
	manager, err := factory.CreateManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	ctx := context.Background()

	// Create test rule
	rule := &types.FirewallRule{
		Name:      "test-rule-lifecycle",
		Action:    types.ActionAllow,
		Direction: types.DirectionInbound,
		Protocol:  types.ProtocolTCP,
		LocalPort: "8080",
		Enabled:   true,
		Created:   time.Now(),
	}

	// Add rule
	err = manager.AddRule(ctx, rule)
	if err != nil {
		t.Fatalf("Failed to add rule: %v", err)
	}

	// List rules to verify addition
	rules, err := manager.ListRules(ctx)
	if err != nil {
		t.Fatalf("Failed to list rules: %v", err)
	}

	found := false
	var ruleID string
	for _, r := range rules {
		if r.Name == rule.Name {
			found = true
			ruleID = r.ID
			break
		}
	}

	if !found {
		t.Error("Rule should be found in the list")
	}

	// Remove rule
	if ruleID != "" {
		err = manager.RemoveRule(ctx, ruleID)
		if err != nil {
			t.Fatalf("Failed to remove rule: %v", err)
		}
	}
}

func TestFirewallBackupRestore(t *testing.T) {
	factory := firewall.NewFactory()
	manager, err := factory.CreateManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	ctx := context.Background()
	backupFile := "/tmp/firewall_test_backup.json"

	// Create backup
	err = manager.BackupRules(ctx, backupFile)
	if err != nil {
		t.Fatalf("Failed to backup rules: %v", err)
	}

	// Restore backup
	err = manager.RestoreRules(ctx, backupFile)
	if err != nil {
		t.Fatalf("Failed to restore rules: %v", err)
	}
}

func TestFirewallEnableDisable(t *testing.T) {
	factory := firewall.NewFactory()
	manager, err := factory.CreateManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	ctx := context.Background()

	// Get initial status
	initialStatus, err := manager.GetStatus(ctx)
	if err != nil {
		t.Fatalf("Failed to get initial status: %v", err)
	}

	// Test enable
	err = manager.EnableFirewall(ctx)
	if err != nil {
		t.Fatalf("Failed to enable firewall: %v", err)
	}

	// Test disable
	err = manager.DisableFirewall(ctx)
	if err != nil {
		t.Fatalf("Failed to disable firewall: %v", err)
	}

	// Restore initial state
	if initialStatus.Enabled {
		manager.EnableFirewall(ctx)
	}
}

func BenchmarkFirewallStatus(b *testing.B) {
	factory := firewall.NewFactory()
	manager, err := factory.CreateManager()
	if err != nil {
		b.Fatalf("Failed to create manager: %v", err)
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.GetStatus(ctx)
		if err != nil {
			b.Fatalf("Failed to get status: %v", err)
		}
	}
}

func BenchmarkFirewallListRules(b *testing.B) {
	factory := firewall.NewFactory()
	manager, err := factory.CreateManager()
	if err != nil {
		b.Fatalf("Failed to create manager: %v", err)
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.ListRules(ctx)
		if err != nil {
			b.Fatalf("Failed to list rules: %v", err)
		}
	}
}
