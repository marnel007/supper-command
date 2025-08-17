package filesystem

import (
	"context"
	"os"
	"time"

	"suppercommand/internal/commands"
)

// MvCommand moves/renames files
type MvCommand struct {
	*commands.BaseCommand
}

// NewMvCommand creates a new mv command
func NewMvCommand() *MvCommand {
	return &MvCommand{
		BaseCommand: commands.NewBaseCommand(
			"mv",
			"Move/rename files",
			"mv <source> <destination>",
			[]string{"windows", "linux", "darwin"},
			false,
		),
	}
}

// Execute moves a file from source to destination
func (m *MvCommand) Execute(ctx context.Context, args *commands.Arguments) (*commands.Result, error) {
	startTime := time.Now()

	if len(args.Raw) < 2 {
		return &commands.Result{
			Output:   "Usage: mv <source> <destination>\n",
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	source := args.Raw[0]
	dest := args.Raw[1]

	err := os.Rename(source, dest)
	if err != nil {
		return &commands.Result{
			Output:   "",
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
