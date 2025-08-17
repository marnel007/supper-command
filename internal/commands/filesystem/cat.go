package filesystem

import (
	"context"
	"fmt"
	"os"
	"time"

	"suppercommand/internal/commands"

	"github.com/fatih/color"
)

// CatCommand displays file contents
type CatCommand struct {
	*commands.BaseCommand
}

// NewCatCommand creates a new cat command
func NewCatCommand() *CatCommand {
	return &CatCommand{
		BaseCommand: commands.NewBaseCommand(
			"cat",
			"Display file contents",
			"cat <file1> [file2] ...",
			[]string{"windows", "linux", "darwin"},
			false,
		),
	}
}

// Execute displays the contents of one or more files
func (c *CatCommand) Execute(ctx context.Context, args *commands.Arguments) (*commands.Result, error) {
	startTime := time.Now()

	if len(args.Raw) == 0 {
		return &commands.Result{
			Output:   "Usage: cat <file1> [file2] ...\n",
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	var output string
	hasErrors := false

	for i, filename := range args.Raw {
		// Add separator between files if multiple files
		if i > 0 {
			output += fmt.Sprintf("\n%s\n",
				color.New(color.FgCyan, color.Bold).Sprintf("==> %s <==", filename))
		} else if len(args.Raw) > 1 {
			output += fmt.Sprintf("%s\n",
				color.New(color.FgCyan, color.Bold).Sprintf("==> %s <==", filename))
		}

		// Check if file exists and is readable
		info, err := os.Stat(filename)
		if err != nil {
			if os.IsNotExist(err) {
				output += color.New(color.FgRed).Sprintf("cat: %s: No such file or directory\n", filename)
			} else {
				output += color.New(color.FgRed).Sprintf("cat: %s: %v\n", filename, err)
			}
			hasErrors = true
			continue
		}

		// Check if it's a directory
		if info.IsDir() {
			output += color.New(color.FgRed).Sprintf("cat: %s: Is a directory\n", filename)
			hasErrors = true
			continue
		}

		// Check file size (warn for very large files)
		if info.Size() > 10*1024*1024 { // 10MB
			output += color.New(color.FgYellow).Sprintf("Warning: %s is large (%d bytes). Continue? (y/N): ",
				filename, info.Size())
			// For now, just show a warning and continue
			output += color.New(color.FgYellow).Sprint("Proceeding...\n")
		}

		// Read and display file contents
		content, err := os.ReadFile(filename)
		if err != nil {
			output += color.New(color.FgRed).Sprintf("cat: %s: %v\n", filename, err)
			hasErrors = true
			continue
		}

		output += string(content)

		// Add newline if file doesn't end with one
		if len(content) > 0 && content[len(content)-1] != '\n' {
			output += "\n"
		}
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
