package remote

import (
	"context"
	"fmt"
	"strings"

	"suppercommand/internal/managers/remote"
	"suppercommand/internal/types"
)

// SimpleRemoteCommand is a basic remote command without circular dependencies
type SimpleRemoteCommand struct {
	name        string
	description string
	usage       string
	manager     types.RemoteManager
}

// NewRemoteCommand creates a new simple remote command
func NewRemoteCommand() *SimpleRemoteCommand {
	factory := remote.NewFactory()
	manager, err := factory.CreateManager()
	if err != nil {
		// Continue with nil manager, will show error in execution
	}

	return &SimpleRemoteCommand{
		name:        "remote",
		description: "Remote server management",
		usage:       "remote [list|add|remove|exec] [options]",
		manager:     manager,
	}
}

// GetName returns the command name
func (r *SimpleRemoteCommand) GetName() string {
	return r.name
}

// GetDescription returns the command description
func (r *SimpleRemoteCommand) GetDescription() string {
	return r.description
}

// GetUsage returns the command usage
func (r *SimpleRemoteCommand) GetUsage() string {
	return r.usage
}

// Execute executes the remote command
func (r *SimpleRemoteCommand) Execute(ctx context.Context, args []string) error {
	if r.manager == nil {
		fmt.Println("Remote manager not available")
		return fmt.Errorf("remote manager not initialized")
	}

	if len(args) == 0 {
		return r.showHelp()
	}

	switch args[0] {
	case "list":
		return r.listServers(ctx)
	case "add":
		return r.addServer(ctx, args[1:])
	case "remove":
		return r.removeServer(ctx, args[1:])
	case "exec":
		return r.executeCommand(ctx, args[1:])
	case "cluster":
		return r.manageCluster(ctx, args[1:])
	case "sync":
		return r.manageSync(ctx, args[1:])
	case "help", "--help", "-h":
		return r.showHelp()
	default:
		fmt.Printf("Unknown subcommand: %s\n", args[0])
		return r.showHelp()
	}
}

// listServers lists all configured remote servers
func (r *SimpleRemoteCommand) listServers(ctx context.Context) error {
	servers, err := r.manager.ListServers(ctx)
	if err != nil {
		fmt.Printf("Error listing servers: %v\n", err)
		return err
	}

	if len(servers) == 0 {
		fmt.Println("No remote servers configured")
		return nil
	}

	fmt.Printf("Remote Servers (%d):\n", len(servers))
	for i, server := range servers {
		fmt.Printf("%d. %s (%s:%d) - %s\n",
			i+1,
			server.Config.Name,
			server.Config.Host,
			server.Config.Port,
			server.Status)
	}

	return nil
}

// addServer adds a new remote server
func (r *SimpleRemoteCommand) addServer(ctx context.Context, args []string) error {
	if len(args) < 2 {
		fmt.Println("Usage: remote add <name> <user@host> [--port <port>] [--key <keyfile>] [--password <password>]")
		return nil
	}

	name := args[0]
	userHost := args[1]

	// Parse user@host
	parts := strings.Split(userHost, "@")
	if len(parts) != 2 {
		fmt.Println("Invalid format. Use: user@host")
		return nil
	}

	config := &types.ServerConfig{
		Name:     name,
		Host:     parts[1],
		Port:     22, // Default SSH port
		Username: parts[0],
		Password: "demo", // Simplified for demo
	}

	fmt.Printf("Adding server: %s (%s@%s:%d)\n", name, config.Username, config.Host, config.Port)

	err := r.manager.AddServer(ctx, config)
	if err != nil {
		fmt.Printf("Error adding server: %v\n", err)
		return err
	}

	fmt.Printf("Server '%s' added successfully\n", name)
	return nil
}

// removeServer removes a remote server
func (r *SimpleRemoteCommand) removeServer(ctx context.Context, args []string) error {
	if len(args) < 1 {
		fmt.Println("Usage: remote remove <name>")
		return nil
	}

	name := args[0]
	fmt.Printf("Removing server: %s\n", name)

	err := r.manager.RemoveServer(ctx, name)
	if err != nil {
		fmt.Printf("Error removing server: %v\n", err)
		return err
	}

	fmt.Printf("Server '%s' removed successfully\n", name)
	return nil
}

// executeCommand executes a command on a remote server
func (r *SimpleRemoteCommand) executeCommand(ctx context.Context, args []string) error {
	if len(args) < 2 {
		fmt.Println("Usage: remote exec <server> <command>")
		return nil
	}

	serverName := args[0]
	command := strings.Join(args[1:], " ")

	fmt.Printf("Executing on %s: %s\n", serverName, command)

	result, err := r.manager.ExecuteCommand(ctx, serverName, command)
	if err != nil {
		fmt.Printf("Error executing command: %v\n", err)
		return err
	}

	fmt.Printf("Output:\n%s\n", result.Output)
	if result.Error != "" {
		fmt.Printf("Error: %s\n", result.Error)
	}
	fmt.Printf("Exit Code: %d\n", result.ExitCode)

	return nil
}

// manageCluster manages server clusters
func (r *SimpleRemoteCommand) manageCluster(ctx context.Context, args []string) error {
	if len(args) == 0 {
		fmt.Println("Usage: remote cluster [list|create|delete] [options]")
		return nil
	}

	switch args[0] {
	case "list":
		fmt.Println("Cluster management not implemented in simple version")
		return nil
	case "create":
		fmt.Println("Cluster creation not implemented in simple version")
		return nil
	case "delete":
		fmt.Println("Cluster deletion not implemented in simple version")
		return nil
	default:
		fmt.Printf("Unknown cluster subcommand: %s\n", args[0])
		return nil
	}
}

// manageSync manages configuration synchronization
func (r *SimpleRemoteCommand) manageSync(ctx context.Context, args []string) error {
	if len(args) == 0 {
		fmt.Println("Usage: remote sync [list|create|execute] [options]")
		return nil
	}

	switch args[0] {
	case "list":
		fmt.Println("Sync profile management not implemented in simple version")
		return nil
	case "create":
		fmt.Println("Sync profile creation not implemented in simple version")
		return nil
	case "execute":
		fmt.Println("Sync execution not implemented in simple version")
		return nil
	default:
		fmt.Printf("Unknown sync subcommand: %s\n", args[0])
		return nil
	}
}

// showHelp shows command help
func (r *SimpleRemoteCommand) showHelp() error {
	help := `
Remote Server Management Command

Usage: remote [command] [options]

Commands:
  list                List all configured remote servers
  add <name> <user@host>  Add a new remote server
  remove <name>       Remove a remote server
  exec <server> <cmd> Execute command on remote server
  cluster [cmd]       Manage server clusters (not implemented)
  sync [cmd]          Manage configuration sync (not implemented)
  help                Show this help message

Examples:
  remote list         # List all servers
  remote add web1 admin@192.168.1.10  # Add server
  remote exec web1 "uptime"  # Execute command
  remote remove web1  # Remove server
`
	fmt.Println(strings.TrimSpace(help))
	return nil
}

// Name returns the command name (alias for GetName)
func (r *SimpleRemoteCommand) Name() string {
	return r.GetName()
}

// Description returns the command description (alias for GetDescription)
func (r *SimpleRemoteCommand) Description() string {
	return r.GetDescription()
}

// Usage returns the command usage (alias for GetUsage)
func (r *SimpleRemoteCommand) Usage() string {
	return r.GetUsage()
}
