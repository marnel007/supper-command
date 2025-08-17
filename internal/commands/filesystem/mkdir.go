package filesystem

import (
	"context"
	"fmt"
	"os"
	"time"

	"suppercommand/internal/commands"

	"github.com/fatih/color"
)

// MkdirCommand creates directories
type MkdirCommand struct {
	*commands.BaseCommand
}

// NewMkdirCommand creates a new mkdir command
func NewMkdirCommand() *MkdirCommand {
	return &MkdirCommand{
		BaseCommand: commands.NewBaseCommand(
			"mkdir",
			"Create directories",
			"mkdir [-p] <directory1> [directory2] ...",
			[]string{"windows", "linux", "darwin"},
			false,
		),
	}
}

// Execute creates one or more directories
func (m *MkdirCommand) Execute(ctx context.Context, args *commands.Arguments) (*commands.Result, error) {
	startTime := time.Now()

	if len(args.Raw) == 0 {
		return &commands.Result{
			Output:   "Usage: mkdir [-p] <directory1> [directory2] ...\n",
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	// Parse flags
	createParents := false
	var directories []string

	for _, arg := range args.Raw {
		if arg == "-p" || arg == "--parents" {
			createParents = true
		} else {
			directories = append(directories, arg)
		}
	}

	if len(directories) == 0 {
		return &commands.Result{
			Output:   "Error: No directories specified\n",
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	var output string
	hasErrors := false
	successCount := 0

	for _, dir := range directories {
		var err error

		if createParents {
			err = os.MkdirAll(dir, 0755)
		} else {
			err = os.Mkdir(dir, 0755)
		}

		if err != nil {
			if os.IsExist(err) {
				output += color.New(color.FgYellow).Sprintf("mkdir: %s: Directory already exists\n", dir)
			} else {
				output += color.New(color.FgRed).Sprintf("mkdir: %s: %v\n", dir, err)
				hasErrors = true
			}
		} else {
			output += color.New(color.FgGreen).Sprintf("✅ Created directory: %s\n", dir)
			successCount++
		}
	}

	// Summary
	if successCount > 0 {
		output += fmt.Sprintf("\n📁 Successfully created %d director%s\n",
			successCount,
			map[bool]string{true: "y", false: "ies"}[successCount == 1])
	}

	exitCode := 0
	if hasErrors {
		exitCode = 1
	}

	return &commands.Result{
		Output:   output,
		ExitCode: exitCode,
		Duration: time.Since(startTime),
	}, nil
}
