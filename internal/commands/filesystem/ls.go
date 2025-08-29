package filesystem

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"suppercommand/internal/commands"

	"github.com/fatih/color"
)

// LsCommand lists directory contents
type LsCommand struct {
	*commands.BaseCommand
}

// NewLsCommand creates a new ls command
func NewLsCommand() *LsCommand {
	return &LsCommand{
		BaseCommand: commands.NewBaseCommand(
			"ls",
			"List directory contents with rich formatting",
			"ls [-a] [-l] [-h] [directory|pattern]",
			[]string{"windows", "linux", "darwin"},
			false,
		),
	}
}

// Execute lists directory contents
func (l *LsCommand) Execute(ctx context.Context, args *commands.Arguments) (*commands.Result, error) {
	startTime := time.Now()

	// Parse arguments
	dir := "."
	showAll := false
	showLong := false
	showHuman := false
	pattern := "*"

	for _, arg := range args.Raw {
		switch arg {
		case "-a", "--all":
			showAll = true
		case "-l", "--long":
			showLong = true
		case "-h", "--human-readable":
			showHuman = true
		case "-la", "-al":
			showLong = true
			showAll = true
		default:
			if !strings.HasPrefix(arg, "-") {
				if strings.Contains(arg, "*") || strings.Contains(arg, "?") {
					pattern = filepath.Base(arg)
					dir = filepath.Dir(arg)
					if dir == "." {
						dir = "."
					}
				} else {
					dir = arg
				}
			}
		}
	}

	// Read directory
	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		return &commands.Result{
			Output:   "",
			Error:    err,
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, err
	}

	// Filter entries
	var filteredEntries []os.FileInfo
	for _, entry := range entries {
		// Skip hidden files unless -a flag is used
		if !showAll && strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		// Match pattern
		matched, err := filepath.Match(pattern, entry.Name())
		if err == nil && matched {
			filteredEntries = append(filteredEntries, entry)
		}
	}

	var output strings.Builder

	// Color functions
	dirColor := color.New(color.FgCyan, color.Bold).SprintFunc()
	fileColor := color.New(color.FgWhite).SprintFunc()
	exeColor := color.New(color.FgGreen, color.Bold).SprintFunc()
	hiddenColor := color.New(color.FgHiBlack).SprintFunc()
	sizeColor := color.New(color.FgYellow).SprintFunc()
	dateColor := color.New(color.FgBlue).SprintFunc()

	if showLong {
		// Long format with details
		output.WriteString(fmt.Sprintf("📁 Directory: %s\n", color.New(color.FgCyan).Sprint(dir)))
		output.WriteString("═══════════════════════════════════════════════════════════════\n")

		totalSize := int64(0)
		totalFiles := 0
		totalDirs := 0

		for _, entry := range filteredEntries {
			// Permissions
			mode := entry.Mode()
			perms := mode.String()

			// Size
			size := entry.Size()
			totalSize += size
			sizeStr := formatSize(size, showHuman)

			// Modified time
			modTime := entry.ModTime().Format("Jan 02 15:04")

			// Name with colors
			name := entry.Name()
			var coloredName string

			if entry.IsDir() {
				coloredName = dirColor(name) + "/"
				totalDirs++
			} else if isExecutable(filepath.Join(dir, name)) {
				coloredName = exeColor(name) + "*"
				totalFiles++
			} else if strings.HasPrefix(name, ".") {
				coloredName = hiddenColor(name)
				totalFiles++
			} else {
				coloredName = fileColor(name)
				totalFiles++
			}

			output.WriteString(fmt.Sprintf("%s %8s %s %s\n",
				perms,
				sizeColor(sizeStr),
				dateColor(modTime),
				coloredName))
		}

		output.WriteString("═══════════════════════════════════════════════════════════════\n")
		output.WriteString(fmt.Sprintf("📊 Total: %d files, %d directories, %s\n",
			totalFiles, totalDirs, sizeColor(formatSize(totalSize, showHuman))))
	} else {
		// Simple format
		for _, entry := range filteredEntries {
			name := entry.Name()

			if entry.IsDir() {
				output.WriteString(dirColor(name) + "/\n")
			} else if isExecutable(filepath.Join(dir, name)) {
				output.WriteString(exeColor(name) + "*\n")
			} else if strings.HasPrefix(name, ".") {
				output.WriteString(hiddenColor(name) + "\n")
			} else {
				output.WriteString(fileColor(name) + "\n")
			}
		}
	}

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// formatSize formats file size in human readable format
func formatSize(size int64, human bool) string {
	if !human {
		return fmt.Sprintf("%d", size)
	}

	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}

	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// isExecutable checks if a file is executable
func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	// On Windows, check for .exe, .bat, .cmd extensions
	if strings.HasSuffix(strings.ToLower(path), ".exe") ||
		strings.HasSuffix(strings.ToLower(path), ".bat") ||
		strings.HasSuffix(strings.ToLower(path), ".cmd") {
		return true
	}

	// On Unix-like systems, check execute permission
	mode := info.Mode()
	return mode&0111 != 0
}
