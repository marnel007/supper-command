package commands

import (
	"context"
	"fmt"
	"time"

	"suppercommand/internal/commands/firewall"
	"suppercommand/internal/commands/performance"
	// "suppercommand/internal/commands/remote" // Temporarily disabled due to Go 1.14 compatibility
	"suppercommand/internal/commands/server"
)

// SimpleCommandAdapter adapts simple commands to the Command interface
type SimpleCommandAdapter struct {
	*BaseCommand
	executeFunc func(ctx context.Context, args []string) error
}

// NewFirewallAdapter creates an adapter for the firewall command
func NewFirewallAdapter() Command {
	firewallCmd := firewall.NewSimpleFirewallCommand()
	return &SimpleCommandAdapter{
		BaseCommand: NewBaseCommand(
			firewallCmd.Name(),
			firewallCmd.Description(),
			firewallCmd.Usage(),
			[]string{"windows", "linux", "darwin"},
			false,
		),
		executeFunc: firewallCmd.Execute,
	}
}

// NewPerformanceAdapter creates an adapter for the performance command
func NewPerformanceAdapter() Command {
	perfCmd := performance.NewPerfCommand()
	return &SimpleCommandAdapter{
		BaseCommand: NewBaseCommand(
			perfCmd.Name(),
			perfCmd.Description(),
			perfCmd.Usage(),
			[]string{"windows", "linux", "darwin"},
			false,
		),
		executeFunc: perfCmd.Execute,
	}
}

// NewServerAdapter creates an adapter for the server command
func NewServerAdapter() Command {
	serverCmd := server.NewServerCommand()
	return &SimpleCommandAdapter{
		BaseCommand: NewBaseCommand(
			serverCmd.Name(),
			serverCmd.Description(),
			serverCmd.Usage(),
			[]string{"windows", "linux", "darwin"},
			false,
		),
		executeFunc: serverCmd.Execute,
	}
}

// NewRemoteAdapter creates an adapter for the remote command
func NewRemoteAdapter() Command {
	// Temporarily disabled due to Go 1.14 compatibility
	return &SimpleCommandAdapter{
		BaseCommand: NewBaseCommand(
			"remote",
			"Remote management (temporarily disabled)",
			"remote [disabled]",
			[]string{"windows", "linux", "darwin"},
			false,
		),
		executeFunc: func(ctx context.Context, args []string) error {
			return fmt.Errorf("remote management temporarily disabled due to Go 1.14 compatibility")
		},
	}
}

// Execute adapts the simple command execution to the Command interface
func (a *SimpleCommandAdapter) Execute(ctx context.Context, args *Arguments) (*Result, error) {
	startTime := time.Now()

	// Execute the simple command
	err := a.executeFunc(ctx, args.Raw)

	result := &Result{
		Duration: time.Since(startTime),
		ExitCode: 0,
	}

	if err != nil {
		result.Error = err
		result.ExitCode = 1
	}

	return result, nil
}
