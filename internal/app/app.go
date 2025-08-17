package app

import (
	"context"
	"fmt"
	"time"

	"suppercommand/internal/commands"
	"suppercommand/internal/commands/filesystem"
	"suppercommand/internal/commands/networking"
	"suppercommand/internal/commands/system"
	"suppercommand/internal/config"
	"suppercommand/internal/monitoring"
	"suppercommand/internal/shell"
)

// Application orchestrates the entire shell lifecycle
type Application struct {
	config   *config.Config
	shell    shell.Shell
	registry *commands.Registry
	monitor  monitoring.Monitor
	logger   monitoring.Logger
}

// NewApplication creates a new application instance with dependency injection
func NewApplication() *Application {
	return &Application{}
}

// Initialize sets up all application components
func (a *Application) Initialize(ctx context.Context) error {
	// Load configuration
	loader := config.NewLoader()
	a.config = loader.LoadWithDefaults()

	// Initialize monitoring
	a.logger = monitoring.NewLogger(a.config.Monitoring)
	a.monitor = monitoring.NewMonitor(a.config.Monitoring, a.logger)

	// Initialize command registry
	a.registry = commands.NewRegistry(a.logger)

	// Register built-in commands
	if err := a.registerBuiltinCommands(); err != nil {
		return fmt.Errorf("failed to register builtin commands: %w", err)
	}

	// Initialize shell
	a.shell = shell.NewShell(a.config.Shell, a.registry, a.monitor, a.logger)

	// Initialize shell
	if err := a.shell.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize shell: %w", err)
	}

	a.logger.Info("Application initialized successfully")
	return nil
}

// Run starts the application main loop
func (a *Application) Run(ctx context.Context) error {
	a.logger.Info("Starting SuperShell application")

	// Start monitoring
	if err := a.monitor.Start(ctx); err != nil {
		return fmt.Errorf("failed to start monitoring: %w", err)
	}

	// Run shell
	return a.shell.Run(ctx)
}

// ExecuteCommand executes a single command
func (a *Application) ExecuteCommand(ctx context.Context, input string) (*shell.ExecutionResult, error) {
	if a.shell == nil {
		return nil, fmt.Errorf("shell not initialized")
	}
	return a.shell.ExecuteCommand(ctx, input)
}

// Shutdown gracefully shuts down the application
func (a *Application) Shutdown(ctx context.Context) error {
	a.logger.Info("Shutting down SuperShell application")

	// Create shutdown timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Shutdown components in reverse order
	if a.shell != nil {
		if err := a.shell.Shutdown(shutdownCtx); err != nil {
			a.logger.Error("Failed to shutdown shell", err)
		}
	}

	if a.monitor != nil {
		if err := a.monitor.Stop(shutdownCtx); err != nil {
			a.logger.Error("Failed to stop monitoring", err)
		}
	}

	a.logger.Info("Application shutdown complete")
	return nil
}

// registerBuiltinCommands registers all built-in commands
func (a *Application) registerBuiltinCommands() error {
	// System commands
	systemCommands := []commands.Command{
		system.NewHelpCommand(a.registry),
		system.NewClearCommand(),
		system.NewSysInfoCommand(),
		system.NewWhoamiCommand(),
		system.NewHostnameCommand(),
		system.NewExitCommand(),
		system.NewVerCommand(),
		system.NewHelpHTMLCommand(a.registry),
		system.NewWinUpdateCommand(),
		system.NewKillTaskCommand(),
		system.NewLookupCommand(a.registry),
		system.NewSmartHistoryCommand(a.registry),
	}

	// Filesystem commands
	filesystemCommands := []commands.Command{
		filesystem.NewPwdCommand(),
		filesystem.NewLsCommand(),
		filesystem.NewDirCommand(),
		filesystem.NewEchoCommand(),
		filesystem.NewCdCommand(),
		filesystem.NewCatCommand(),
		filesystem.NewMkdirCommand(),
		filesystem.NewRmCommand(),
		filesystem.NewRmdirCommand(),
		filesystem.NewCpCommand(),
		filesystem.NewMvCommand(),
	}

	// Networking commands
	networkingCommands := []commands.Command{
		networking.NewPingCommand(),
		networking.NewTracertCommand(),
		networking.NewNslookupCommand(),
		networking.NewNetstatCommand(),
		networking.NewPortscanCommand(),
		networking.NewIpconfigCommand(),
		networking.NewWgetCommand(),
		networking.NewArpCommand(),
		networking.NewRouteCommand(),
		networking.NewSpeedtestCommand(),
		networking.NewNetdiscoverCommand(),
		networking.NewSniffCommand(),
		networking.NewFastcpSendCommand(),
		networking.NewFastcpRecvCommand(),
		networking.NewFastcpBackupCommand(),
		networking.NewFastcpRestoreCommand(),
		networking.NewFastcpDedupCommand(),
	}

	// Management commands
	managementCommands := []commands.Command{
		commands.NewFirewallAdapter(),
		commands.NewPerformanceAdapter(),
		commands.NewServerAdapter(),
		commands.NewRemoteAdapter(),
	}

	// Register all commands
	allCommands := [][]commands.Command{
		systemCommands,
		filesystemCommands,
		networkingCommands,
		managementCommands,
	}

	for _, commandGroup := range allCommands {
		for _, cmd := range commandGroup {
			if err := a.registry.Register(cmd); err != nil {
				return err
			}
		}
	}

	return nil
}
