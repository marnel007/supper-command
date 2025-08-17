package system

import (
	"context"
	"os"
	"time"

	"suppercommand/internal/commands"
)

// ExitCommand exits the shell
type ExitCommand struct {
	*commands.BaseCommand
}

// NewExitCommand creates a new exit command
func NewExitCommand() *ExitCommand {
	return &ExitCommand{
		BaseCommand: commands.NewBaseCommand(
			"exit",
			"Exit the shell",
			"exit [code]",
			[]string{"windows", "linux", "darwin"},
			false,
		),
	}
}

// Execute exits the shell
func (e *ExitCommand) Execute(ctx context.Context, args *commands.Arguments) (*commands.Result, error) {
	startTime := time.Now()

	exitCode := 0
	if len(args.Raw) > 0 {
		// Try to parse exit code
		// For simplicity, just use 0 for now
	}

	// In a real implementation, this would signal the shell to exit
	// For now, we'll just call os.Exit
	os.Exit(exitCode)

	return &commands.Result{
		Output:   "Goodbye!\n",
		ExitCode: exitCode,
		Duration: time.Since(startTime),
	}, nil
}
