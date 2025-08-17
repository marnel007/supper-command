package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"suppercommand/internal/commands"
	"suppercommand/pkg/errors"
)

// CdCommand changes the current directory
type CdCommand struct {
	*commands.BaseCommand
}

// NewCdCommand creates a new cd command
func NewCdCommand() *CdCommand {
	return &CdCommand{
		BaseCommand: commands.NewBaseCommand(
			"cd",
			"Change the current directory",
			"cd <directory>",
			[]string{"windows", "linux", "darwin"},
			false,
		),
	}
}

// Execute changes the current directory
func (c *CdCommand) Execute(ctx context.Context, args *commands.Arguments) (*commands.Result, error) {
	startTime := time.Now()

	if len(args.Raw) == 0 {
		// No arguments - go to home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return &commands.Result{
				Output:   "Error: Could not determine home directory\n",
				Error:    err,
				ExitCode: 1,
				Duration: time.Since(startTime),
			}, err
		}
		args.Raw = []string{homeDir}
	}

	targetDir := args.Raw[0]

	// Handle special cases
	switch targetDir {
	case "~":
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return &commands.Result{
				Output:   "Error: Could not determine home directory\n",
				Error:    err,
				ExitCode: 1,
				Duration: time.Since(startTime),
			}, err
		}
		targetDir = homeDir
	case "-":
		// TODO: Implement previous directory functionality
		return &commands.Result{
			Output:   "Previous directory functionality not yet implemented\n",
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(targetDir)
	if err != nil {
		return &commands.Result{
			Output:   "Error: Invalid path\n",
			Error:    err,
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, err
	}

	// Check if directory exists
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &commands.Result{
				Output:   "Error: Directory does not exist\n",
				Error:    errors.NewValidationError("directory does not exist: %s", absPath),
				ExitCode: 1,
				Duration: time.Since(startTime),
			}, nil
		}
		return &commands.Result{
			Output:   "Error: Cannot access directory\n",
			Error:    err,
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, err
	}

	// Check if it's actually a directory
	if !info.IsDir() {
		return &commands.Result{
			Output:   "Error: Not a directory\n",
			Error:    errors.NewValidationError("not a directory: %s", absPath),
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	// Change directory
	err = os.Chdir(absPath)
	if err != nil {
		return &commands.Result{
			Output:   "Error: Permission denied or cannot change directory\n",
			Error:    err,
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, err
	}

	return &commands.Result{
		Output:   "",
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}
