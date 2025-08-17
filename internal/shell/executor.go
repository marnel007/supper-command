package shell

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"suppercommand/internal/commands"
	"suppercommand/internal/commands/system"
	"suppercommand/internal/monitoring"
)

// Executor handles command execution
type Executor struct {
	registry       *commands.Registry
	monitor        monitoring.Monitor
	logger         monitoring.Logger
	historyTracker *system.HistoryTracker
}

// NewExecutor creates a new command executor
func NewExecutor(
	registry *commands.Registry,
	monitor monitoring.Monitor,
	logger monitoring.Logger,
) *Executor {
	return &Executor{
		registry:       registry,
		monitor:        monitor,
		logger:         logger,
		historyTracker: system.NewHistoryTracker(),
	}
}

// Initialize initializes the executor
func (e *Executor) Initialize(ctx context.Context) error {
	e.logger.Info("Command executor initialized")
	return nil
}

// Execute executes a command string
func (e *Executor) Execute(ctx context.Context, input string) (*ExecutionResult, error) {
	startTime := time.Now()

	// Parse input
	input = strings.TrimSpace(input)
	if input == "" {
		return &ExecutionResult{
			Output:   "",
			Duration: time.Since(startTime),
		}, nil
	}

	// Handle exit commands
	if input == "exit" || input == "quit" {
		return &ExecutionResult{
			Output:   "Goodbye!",
			ExitCode: 0,
			Duration: time.Since(startTime),
		}, nil
	}

	// Parse command and arguments
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return &ExecutionResult{
			Output:   "",
			Duration: time.Since(startTime),
		}, nil
	}

	commandName := parts[0]
	args := commands.ParseArguments(parts[1:])

	// Check if command exists in registry first
	_, err := e.registry.Get(commandName)
	if err != nil {
		// Command not found in registry, try external command execution
		e.logger.Debug("Internal command not found, trying external command",
			monitoring.Field{Key: "command", Value: commandName})

		return e.executeExternalCommand(ctx, input, startTime)
	}

	// Execute command through registry
	result, err := e.registry.Execute(ctx, commandName, args)

	duration := time.Since(startTime)
	success := err == nil

	// Record metrics
	e.monitor.RecordCommandExecution(commandName, duration, success)

	if err != nil {
		e.logger.Error("Command execution failed", err,
			monitoring.Field{Key: "command", Value: commandName},
			monitoring.Field{Key: "duration", Value: duration})

		return &ExecutionResult{
			Output:   "",
			Error:    err,
			ExitCode: 1,
			Duration: duration,
		}, err
	}

	e.logger.Debug("Command executed successfully",
		monitoring.Field{Key: "command", Value: commandName},
		monitoring.Field{Key: "duration", Value: duration})

	// Track command in history
	cwd, _ := os.Getwd()
	e.historyTracker.TrackCommand(input, cwd, result.ExitCode, result.Duration)

	// Convert commands.Result to ExecutionResult
	return &ExecutionResult{
		Output:     result.Output,
		Error:      result.Error,
		ExitCode:   result.ExitCode,
		Duration:   result.Duration,
		MemoryUsed: result.MemoryUsed,
		Warnings:   result.Warnings,
	}, nil
}

// Shutdown gracefully shuts down the executor
func (e *Executor) Shutdown(ctx context.Context) error {
	e.logger.Info("Command executor shutdown")
	return nil
}

// executeExternalCommand executes external system commands
func (e *Executor) executeExternalCommand(ctx context.Context, input string, startTime time.Time) (*ExecutionResult, error) {
	var cmd *exec.Cmd

	// Determine the shell to use based on the operating system
	switch runtime.GOOS {
	case "windows":
		// Use PowerShell on Windows for better compatibility
		cmd = exec.CommandContext(ctx, "powershell", "-Command", input)
	case "linux", "darwin":
		// Use bash on Unix-like systems
		cmd = exec.CommandContext(ctx, "bash", "-c", input)
	default:
		return &ExecutionResult{
			Output:   "External command execution not supported on this platform\n",
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	// Execute the command
	output, err := cmd.CombinedOutput()
	duration := time.Since(startTime)

	exitCode := 0
	if err != nil {
		// Try to get exit code from error
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = 1
		}
	}

	// Record metrics for external command
	commandName := strings.Fields(input)[0]
	success := exitCode == 0
	e.monitor.RecordCommandExecution("external:"+commandName, duration, success)

	if success {
		e.logger.Debug("External command executed successfully",
			monitoring.Field{Key: "command", Value: commandName},
			monitoring.Field{Key: "duration", Value: duration})
	} else {
		e.logger.Debug("External command failed",
			monitoring.Field{Key: "command", Value: commandName},
			monitoring.Field{Key: "exit_code", Value: exitCode},
			monitoring.Field{Key: "duration", Value: duration})
	}

	// Track external command in history
	cwd, _ := os.Getwd()
	e.historyTracker.TrackCommand(input, cwd, exitCode, duration)

	return &ExecutionResult{
		Output:   string(output),
		ExitCode: exitCode,
		Duration: duration,
	}, nil
}
