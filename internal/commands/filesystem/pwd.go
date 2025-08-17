package filesystem

import (
	"context"
	"os"
	"time"

	"suppercommand/internal/commands"
)

// PwdCommand shows the current working directory
type PwdCommand struct {
	*commands.BaseCommand
}

// NewPwdCommand creates a new pwd command
func NewPwdCommand() *PwdCommand {
	return &PwdCommand{
		BaseCommand: commands.NewBaseCommand(
			"pwd",
			"Print the current working directory",
			"pwd",
			[]string{"windows", "linux", "darwin"},
			false,
		),
	}
}

// Execute shows the current working directory
func (p *PwdCommand) Execute(ctx context.Context, args *commands.Arguments) (*commands.Result, error) {
	startTime := time.Now()

	cwd, err := os.Getwd()
	if err != nil {
		return &commands.Result{
			Output:   "",
			Error:    err,
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, err
	}

	return &commands.Result{
		Output:   cwd + "\n",
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}
