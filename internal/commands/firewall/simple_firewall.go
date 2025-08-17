package firewall

import (
	"context"
	"fmt"
	"strings"

	"suppercommand/internal/managers/firewall"
	"suppercommand/internal/types"
)

// SimpleFirewallCommand is a basic firewall command without circular dependencies
type SimpleFirewallCommand struct {
	name        string
	description string
	usage       string
	manager     types.FirewallManager
}

// NewSimpleFirewallCommand creates a new simple firewall command
func NewSimpleFirewallCommand() *SimpleFirewallCommand {
	factory := firewall.NewFactory()
	manager, err := factory.CreateManager()
	if err != nil {
		// Use a mock manager if creation fails
		manager = &firewall.MockFirewallManager{}
	}

	return &SimpleFirewallCommand{
		name:        "firewall",
		description: "Manage Windows firewall settings",
		usage:       "firewall [status|enable|disable|rules] [options]",
		manager:     manager,
	}
}

// GetName returns the command name
func (f *SimpleFirewallCommand) GetName() string {
	return f.name
}

// GetDescription returns the command description
func (f *SimpleFirewallCommand) GetDescription() string {
	return f.description
}

// GetUsage returns the command usage
func (f *SimpleFirewallCommand) GetUsage() string {
	return f.usage
}

// Execute executes the firewall command
func (f *SimpleFirewallCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return f.showHelp()
	}

	switch args[0] {
	case "status":
		return f.showStatus(ctx)
	case "enable":
		return f.enableFirewall(ctx)
	case "disable":
		return f.disableFirewall(ctx)
	case "rules":
		return f.manageRules(ctx, args[1:])
	case "help", "--help", "-h":
		return f.showHelp()
	default:
		fmt.Printf("Unknown subcommand: %s\n", args[0])
		return f.showHelp()
	}
}

// showStatus shows the current firewall status
func (f *SimpleFirewallCommand) showStatus(ctx context.Context) error {
	status, err := f.manager.GetStatus(ctx)
	if err != nil {
		fmt.Printf("Error getting firewall status: %v\n", err)
		return err
	}

	fmt.Printf("Firewall Enabled: %t\n", status.Enabled)
	fmt.Printf("Profile: %s\n", status.Profile)
	fmt.Printf("Platform: %s\n", status.Platform)
	fmt.Printf("Rule Count: %d\n", status.RuleCount)
	fmt.Printf("Last Updated: %s\n", status.LastUpdated.Format("2006-01-02 15:04:05"))

	return nil
}

// enableFirewall enables the firewall
func (f *SimpleFirewallCommand) enableFirewall(ctx context.Context) error {
	fmt.Println("Enabling firewall...")

	err := f.manager.EnableFirewall(ctx)
	if err != nil {
		fmt.Printf("Failed to enable firewall: %v\n", err)
		return err
	}

	fmt.Println("Firewall enabled successfully")
	return nil
}

// disableFirewall disables the firewall
func (f *SimpleFirewallCommand) disableFirewall(ctx context.Context) error {
	fmt.Println("Disabling firewall...")

	err := f.manager.DisableFirewall(ctx)
	if err != nil {
		fmt.Printf("Failed to disable firewall: %v\n", err)
		return err
	}

	fmt.Println("Firewall disabled successfully")
	return nil
}

// manageRules manages firewall rules
func (f *SimpleFirewallCommand) manageRules(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return f.listRules(ctx)
	}

	switch args[0] {
	case "list":
		return f.listRules(ctx)
	case "add":
		fmt.Println("Rule addition not implemented in simple version")
		return nil
	case "remove":
		fmt.Println("Rule removal not implemented in simple version")
		return nil
	default:
		fmt.Printf("Unknown rules subcommand: %s\n", args[0])
		return nil
	}
}

// listRules lists firewall rules
func (f *SimpleFirewallCommand) listRules(ctx context.Context) error {
	rules, err := f.manager.ListRules(ctx)
	if err != nil {
		fmt.Printf("Error listing firewall rules: %v\n", err)
		return err
	}

	if len(rules) == 0 {
		fmt.Println("No firewall rules found")
		return nil
	}

	fmt.Printf("Firewall Rules (%d):\n", len(rules))
	for i, rule := range rules {
		fmt.Printf("%d. %s - %s (%s)\n", i+1, rule.Name, rule.Action, rule.Direction)
		if rule.Description != "" {
			fmt.Printf("   Description: %s\n", rule.Description)
		}
	}

	return nil
}

// showHelp shows command help
func (f *SimpleFirewallCommand) showHelp() error {
	help := `
Firewall Management Command

Usage: firewall [command] [options]

Commands:
  status              Show firewall status
  enable              Enable the firewall
  disable             Disable the firewall
  rules [list]        Manage firewall rules
  help                Show this help message

Examples:
  firewall status     # Show current firewall status
  firewall enable     # Enable the firewall
  firewall disable    # Disable the firewall
  firewall rules list # List all firewall rules
`
	fmt.Println(strings.TrimSpace(help))
	return nil
}

// Name returns the command name (alias for GetName)
func (f *SimpleFirewallCommand) Name() string {
	return f.GetName()
}

// Description returns the command description (alias for GetDescription)
func (f *SimpleFirewallCommand) Description() string {
	return f.GetDescription()
}

// Usage returns the command usage (alias for GetUsage)
func (f *SimpleFirewallCommand) Usage() string {
	return f.GetUsage()
}
