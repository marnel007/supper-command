package system

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"time"

	"suppercommand/internal/commands"
)

// ClearCommand clears the terminal screen
type ClearCommand struct {
	*commands.BaseCommand
}

// NewClearCommand creates a new clear command
func NewClearCommand() *ClearCommand {
	return &ClearCommand{
		BaseCommand: commands.NewBaseCommand(
			"clear",
			"Clear the terminal screen",
			"clear",
			[]string{"windows", "linux", "darwin"},
			false,
		),
	}
}

// Execute clears the terminal screen
func (c *ClearCommand) Execute(ctx context.Context, args *commands.Arguments) (*commands.Result, error) {
	startTime := time.Now()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
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
