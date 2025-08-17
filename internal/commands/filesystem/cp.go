package filesystem

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"suppercommand/internal/commands"

	"github.com/fatih/color"
)

// CpCommand copies files and directories
type CpCommand struct {
	*commands.BaseCommand
}

// NewCpCommand creates a new cp command
func NewCpCommand() *CpCommand {
	return &CpCommand{
		BaseCommand: commands.NewBaseCommand(
			"cp",
			"Copy files and directories",
			"cp [-r] [-v] <source> <destination>",
			[]string{"windows", "linux", "darwin"},
			false,
		),
	}
}

// Execute copies files or directories
func (c *CpCommand) Execute(ctx context.Context, args *commands.Arguments) (*commands.Result, error) {
	startTime := time.Now()

	if len(args.Raw) < 2 {
		return &commands.Result{
			Output:   "Usage: cp [-r] [-v] <source> <destination>\n",
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	// Parse flags
	recursive := false
	verbose := false
	var sources []string
	var destination string

	for i, arg := range args.Raw {
		switch arg {
		case "-r", "--recursive":
			recursive = true
		case "-v", "--verbose":
			verbose = true
		default:
			if i == len(args.Raw)-1 {
				destination = arg
			} else if !strings.HasPrefix(arg, "-") {
				sources = append(sources, arg)
			}
		}
	}

	if len(sources) == 0 {
		return &commands.Result{
			Output:   "Error: No source files specified\n",
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	if destination == "" {
		return &commands.Result{
			Output:   "Error: No destination specified\n",
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	var output strings.Builder
	hasErrors := false
	successCount := 0

	// Check if destination is a directory
	destInfo, err := os.Stat(destination)
	isDestDir := err == nil && destInfo.IsDir()

	for _, source := range sources {
		// Expand glob patterns
		matches, err := filepath.Glob(source)
		if err != nil {
			output.WriteString(color.New(color.FgRed).Sprintf("cp: %s: Invalid pattern: %v\n", source, err))
			hasErrors = true
			continue
		}

		if len(matches) == 0 {
			output.WriteString(color.New(color.FgRed).Sprintf("cp: %s: No such file or directory\n", source))
			hasErrors = true
			continue
		}

		for _, match := range matches {
			var destPath string
			if isDestDir || len(sources) > 1 {
				destPath = filepath.Join(destination, filepath.Base(match))
			} else {
				destPath = destination
			}

			err := c.copyItem(match, destPath, recursive, verbose, &output)
			if err != nil {
				output.WriteString(color.New(color.FgRed).Sprintf("cp: %s: %v\n", match, err))
				hasErrors = true
			} else {
				successCount++
				if verbose {
					output.WriteString(color.New(color.FgGreen).Sprintf("✅ Copied: %s → %s\n", match, destPath))
				}
			}
		}
	}

	// Summary
	if successCount > 0 {
		output.WriteString(fmt.Sprintf("\n📁 Successfully copied %d item%s\n",
			successCount,
			map[bool]string{true: "", false: "s"}[successCount == 1]))
	}

	exitCode := 0
	if hasErrors {
		exitCode = 1
	}

	return &commands.Result{
		Output:   output.String(),
		ExitCode: exitCode,
		Duration: time.Since(startTime),
	}, nil
}

// copyItem copies a single file or directory
func (c *CpCommand) copyItem(src, dest string, recursive, verbose bool, output *strings.Builder) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if srcInfo.IsDir() {
		if !recursive {
			return fmt.Errorf("is a directory (use -r to copy directories)")
		}
		return c.copyDirectory(src, dest, verbose, output)
	}

	return c.copyFile(src, dest)
}

// copyFile copies a single file
func (c *CpCommand) copyFile(src, dest string) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create destination file
	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy contents
	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}

	// Copy permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dest, srcInfo.Mode())
}

// copyDirectory copies a directory recursively
func (c *CpCommand) copyDirectory(src, dest string, verbose bool, output *strings.Builder) error {
	// Create destination directory
	err := os.MkdirAll(dest, 0755)
	if err != nil {
		return err
	}

	// Read source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		if entry.IsDir() {
			err = c.copyDirectory(srcPath, destPath, verbose, output)
		} else {
			err = c.copyFile(srcPath, destPath)
		}

		if err != nil {
			return err
		}

		if verbose {
			output.WriteString(color.New(color.FgBlue).Sprintf("  📄 %s → %s\n", srcPath, destPath))
		}
	}

	return nil
}
