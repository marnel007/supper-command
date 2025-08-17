package server

import (
	"context"
	"fmt"
	"strings"

	"suppercommand/internal/managers/server"
	"suppercommand/internal/types"
)

// SimpleServerCommand is a basic server command without circular dependencies
type SimpleServerCommand struct {
	name        string
	description string
	usage       string
	manager     types.ServerManager
}

// NewServerCommand creates a new simple server command
func NewServerCommand() *SimpleServerCommand {
	factory := server.NewFactory()
	manager, err := factory.CreateManager()
	if err != nil {
		// Continue with nil manager, will show error in execution
	}

	return &SimpleServerCommand{
		name:        "server",
		description: "Server management and monitoring",
		usage:       "server [health|services|users|alerts|backup] [options]",
		manager:     manager,
	}
}

// GetName returns the command name
func (s *SimpleServerCommand) GetName() string {
	return s.name
}

// GetDescription returns the command description
func (s *SimpleServerCommand) GetDescription() string {
	return s.description
}

// GetUsage returns the command usage
func (s *SimpleServerCommand) GetUsage() string {
	return s.usage
}

// Execute executes the server command
func (s *SimpleServerCommand) Execute(ctx context.Context, args []string) error {
	if s.manager == nil {
		fmt.Println("Server manager not available")
		return fmt.Errorf("server manager not initialized")
	}

	if len(args) == 0 {
		return s.showHelp()
	}

	switch args[0] {
	case "health":
		return s.checkHealth(ctx)
	case "services":
		return s.manageServices(ctx, args[1:])
	case "users":
		return s.listUsers(ctx)
	case "alerts":
		return s.manageAlerts(ctx, args[1:])
	case "backup":
		return s.manageBackup(ctx, args[1:])
	case "session":
		return s.manageSessions(ctx, args[1:])
	case "help", "--help", "-h":
		return s.showHelp()
	default:
		fmt.Printf("Unknown subcommand: %s\n", args[0])
		return s.showHelp()
	}
}

// checkHealth checks server health status
func (s *SimpleServerCommand) checkHealth(ctx context.Context) error {
	fmt.Println("Checking server health...")

	health, err := s.manager.GetHealthStatus(ctx)
	if err != nil {
		fmt.Printf("Error checking health: %v\n", err)
		return err
	}

	fmt.Printf("Overall Health: %s\n", health.Overall)
	fmt.Printf("Uptime: %s\n", health.Uptime)
	fmt.Printf("Last Check: %s\n", health.Timestamp.Format("2006-01-02 15:04:05"))

	if len(health.Components) > 0 {
		fmt.Printf("\nComponent Health:\n")
		for name, component := range health.Components {
			fmt.Printf("- %s: %s (%.1f%s)\n", name, component.Status, component.Value, component.Unit)
			if component.Message != "" {
				fmt.Printf("  %s\n", component.Message)
			}
		}
	}

	if len(health.Alerts) > 0 {
		fmt.Printf("\nActive Alerts:\n")
		for _, alert := range health.Alerts {
			fmt.Printf("- %s: %s\n", alert.Component, alert.Message)
		}
	}

	return nil
}

// manageServices manages system services
func (s *SimpleServerCommand) manageServices(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return s.listServices(ctx)
	}

	switch args[0] {
	case "list":
		return s.listServices(ctx)
	case "start":
		if len(args) < 2 {
			fmt.Println("Usage: server services start <service_name>")
			return nil
		}
		return s.controlService(ctx, args[1], types.ServiceActionStart)
	case "stop":
		if len(args) < 2 {
			fmt.Println("Usage: server services stop <service_name>")
			return nil
		}
		return s.controlService(ctx, args[1], types.ServiceActionStop)
	case "restart":
		if len(args) < 2 {
			fmt.Println("Usage: server services restart <service_name>")
			return nil
		}
		return s.controlService(ctx, args[1], types.ServiceActionRestart)
	default:
		fmt.Printf("Unknown services subcommand: %s\n", args[0])
		return nil
	}
}

// listServices lists all system services
func (s *SimpleServerCommand) listServices(ctx context.Context) error {
	services, err := s.manager.ListServices(ctx)
	if err != nil {
		fmt.Printf("Error listing services: %v\n", err)
		return err
	}

	if len(services) == 0 {
		fmt.Println("No services found")
		return nil
	}

	fmt.Printf("System Services (%d):\n", len(services))
	for i, service := range services {
		fmt.Printf("%d. %s - %s (%s)\n", i+1, service.Name, service.Status, service.StartType)
		if service.Description != "" {
			fmt.Printf("   %s\n", service.Description)
		}
	}

	return nil
}

// controlService controls a system service
func (s *SimpleServerCommand) controlService(ctx context.Context, serviceName string, action types.ServiceAction) error {
	fmt.Printf("Performing %s on service: %s\n", action, serviceName)

	err := s.manager.ControlService(ctx, serviceName, action)
	if err != nil {
		fmt.Printf("Error controlling service: %v\n", err)
		return err
	}

	fmt.Printf("Service %s %s successfully\n", serviceName, action)
	return nil
}

// listUsers lists active users
func (s *SimpleServerCommand) listUsers(ctx context.Context) error {
	users, err := s.manager.GetActiveUsers(ctx)
	if err != nil {
		fmt.Printf("Error listing users: %v\n", err)
		return err
	}

	if len(users) == 0 {
		fmt.Println("No active users found")
		return nil
	}

	fmt.Printf("Active Users (%d):\n", len(users))
	for i, user := range users {
		fmt.Printf("%d. %s (Terminal: %s, Login: %s)\n",
			i+1,
			user.Username,
			user.Terminal,
			user.LoginTime.Format("2006-01-02 15:04:05"))
		if user.RemoteHost != "" {
			fmt.Printf("   Remote: %s\n", user.RemoteHost)
		}
	}

	return nil
}

// manageAlerts manages server alerts
func (s *SimpleServerCommand) manageAlerts(ctx context.Context, args []string) error {
	if len(args) == 0 {
		fmt.Println("Usage: server alerts [list|clear] [options]")
		return nil
	}

	switch args[0] {
	case "list":
		fmt.Println("Alert listing not implemented in simple version")
		return nil
	case "clear":
		fmt.Println("Alert clearing not implemented in simple version")
		return nil
	default:
		fmt.Printf("Unknown alerts subcommand: %s\n", args[0])
		return nil
	}
}

// manageBackup manages server backups
func (s *SimpleServerCommand) manageBackup(ctx context.Context, args []string) error {
	if len(args) == 0 {
		fmt.Println("Usage: server backup [list|create|restore] [options]")
		return nil
	}

	switch args[0] {
	case "list":
		fmt.Println("Backup listing not implemented in simple version")
		return nil
	case "create":
		fmt.Println("Backup creation not implemented in simple version")
		return nil
	case "restore":
		fmt.Println("Backup restoration not implemented in simple version")
		return nil
	default:
		fmt.Printf("Unknown backup subcommand: %s\n", args[0])
		return nil
	}
}

// manageSessions manages user sessions
func (s *SimpleServerCommand) manageSessions(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return s.listUsers(ctx) // Default to listing users
	}

	switch args[0] {
	case "list":
		return s.listUsers(ctx)
	case "kill":
		fmt.Println("Session termination not implemented in simple version")
		return nil
	default:
		fmt.Printf("Unknown session subcommand: %s\n", args[0])
		return nil
	}
}

// showHelp shows command help
func (s *SimpleServerCommand) showHelp() error {
	help := `
Server Management Command

Usage: server [command] [options]

Commands:
  health              Check server health status
  services [cmd]      Manage system services
    list              List all services
    start <name>      Start a service
    stop <name>       Stop a service
    restart <name>    Restart a service
  users               List active users
  session [cmd]       Manage user sessions
    list              List active sessions
  alerts [cmd]        Manage server alerts (not implemented)
  backup [cmd]        Manage server backups (not implemented)
  help                Show this help message

Examples:
  server health       # Check server health
  server services list # List all services
  server users        # List active users
  server services start "Print Spooler"  # Start a service
`
	fmt.Println(strings.TrimSpace(help))
	return nil
}

// Name returns the command name (alias for GetName)
func (s *SimpleServerCommand) Name() string {
	return s.GetName()
}

// Description returns the command description (alias for GetDescription)
func (s *SimpleServerCommand) Description() string {
	return s.GetDescription()
}

// Usage returns the command usage (alias for GetUsage)
func (s *SimpleServerCommand) Usage() string {
	return s.GetUsage()
}
