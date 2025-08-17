package filesystem

import (
	"context"
	"strings"
	"time"

	"suppercommand/internal/commands"
)

// EchoCommand displays text
type EchoCommand struct {
	*commands.BaseCommand
}

// NewEchoCommand creates a new echo command
func NewEchoCommand() *EchoCommand {
	return &EchoCommand{
		BaseCommand: commands.NewBaseCommand(
			"echo",
			"Display text to the console",
			"echo <text>",
			[]string{"windows", "linux", "darwin"},
			false,
		),
	}
}

// Execute displays the provided text
func (e *EchoCommand) Execute(ctx context.Context, args *commands.Arguments) (*commands.Result, error) {
	startTime := time.Now()

	output := strings.Join(args.Raw, " ") + "\n"

	return &commands.Result{
		Output:   output,
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}
