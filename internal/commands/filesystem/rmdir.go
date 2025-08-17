package filesystem

import (
	"context"
	"os"
	"time"

	"suppercommand/internal/commands"
)

// RmdirCommand removes directories
type RmdirCommand struct {
	*commands.BaseCommand
}

// NewRmdirCommand creates a new rmdir command
func NewRmdirCommand() *RmdirCommand {
	return &RmdirCommand{
		BaseCommand: commands.NewBaseCommand(
			"rmdir",
			"Remove empty directories",
			"rmdir <directory>...",
			[]string{"windows", "linux", "darwin"},
			false,
		),
	}
}

// Execute removes the specified directories
func (r *RmdirCommand) Execute(ctx context.Context, args *commands.Arguments) (*commands.Result, error) {
	startTime := time.Now()

	if len(args.Raw) == 0 {
		return &commands.Result{
			Output:   "Usage: rmdir <directory>...\n",
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	for _, dir := range args.Raw {
		err := os.Remove(dir)
		if err != nil {
			return &commands.Result{
				Output:   "",
				Error:    err,
				ExitCode: 1,
				Duration: time.Since(startTime),
			}, err
		}
	}

	return &commands.Result{
		Output:   "",
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}
