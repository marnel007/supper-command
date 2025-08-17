package filesystem

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"suppercommand/internal/commands"

	"github.com/fatih/color"
)

// RmCommand removes files
type RmCommand struct {
	*commands.BaseCommand
}

// NewRmCommand creates a new rm command
func NewRmCommand() *RmCommand {
	return &RmCommand{
		BaseCommand: commands.NewBaseCommand(
			"rm",
			"Remove files and directories",
			"rm [-r] [-f] <file1> [file2] ...",
			[]string{"windows", "linux", "darwin"},
			false,
		),
	}
}

// Execute removes one or more files or directories
func (r *RmCommand) Execute(ctx context.Context, args *commands.Arguments) (*commands.Result, error) {
	startTime := time.Now()

	if len(args.Raw) == 0 {
		return &commands.Result{
			Output:   "Usage: rm [-r] [-f] <file1> [file2] ...\n",
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	// Parse flags
	recursive := false
	force := false
	var targets []string

	for _, arg := range args.Raw {
		switch arg {
		case "-r", "--recursive":
			recursive = true
		case "-f", "--force":
			force = true
		case "-rf", "-fr":
			recursive = true
			force = true
		default:
			targets = append(targets, arg)
		}
	}

	if len(targets) == 0 {
		return &commands.Result{
			Output:   "Error: No files or directories specified\n",
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	var output string
	hasErrors := false
	successCount := 0

	for _, target := range targets {
		// Expand glob patterns
		matches, err := filepath.Glob(target)
		if err != nil {
			output += color.New(color.FgRed).Sprintf("rm: %s: Invalid pattern: %v\n", target, err)
			hasErrors = true
			continue
		}

		if len(matches) == 0 {
			if !force {
				output += color.New(color.FgRed).Sprintf("rm: %s: No such file or directory\n", target)
				hasErrors = true
			}
			continue
		}

		for _, match := range matches {
			err := r.removeTarget(match, recursive, force)
			if err != nil {
				output += color.New(color.FgRed).Sprintf("rm: %s: %v\n", match, err)
				hasErrors = true
			} else {
				output += color.New(color.FgGreen).Sprintf("🗑️  Removed: %s\n", match)
				successCount++
			}
		}
	}

	// Summary
	if successCount > 0 {
		output += fmt.Sprintf("\n✅ Successfully removed %d item%s\n",
			successCount,
			map[bool]string{true: "", false: "s"}[successCount == 1])
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

// removeTarget removes a single file or directory
func (r *RmCommand) removeTarget(target string, recursive, force bool) error {
	info, err := os.Stat(target)
	if err != nil {
		if os.IsNotExist(err) && force {
			return nil // Ignore non-existent files with -f
		}
		return err
	}

	if info.IsDir() {
		if !recursive {
			return fmt.Errorf("is a directory (use -r to remove directories)")
		}
		return os.RemoveAll(target)
	}

	return os.Remove(target)
}
