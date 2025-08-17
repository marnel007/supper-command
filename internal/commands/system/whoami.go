package system

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"runtime"
	"time"

	"suppercommand/internal/commands"

	"github.com/fatih/color"
)

// WhoamiCommand shows current user information
type WhoamiCommand struct {
	*commands.BaseCommand
}

// NewWhoamiCommand creates a new whoami command
func NewWhoamiCommand() *WhoamiCommand {
	return &WhoamiCommand{
		BaseCommand: commands.NewBaseCommand(
			"whoami",
			"Display current user information",
			"whoami [-v|--verbose]",
			[]string{"windows", "linux", "darwin"},
			false,
		),
	}
}

// Execute shows current user information
func (w *WhoamiCommand) Execute(ctx context.Context, args *commands.Arguments) (*commands.Result, error) {
	startTime := time.Now()

	verbose := false
	for _, arg := range args.Raw {
		if arg == "-v" || arg == "--verbose" {
			verbose = true
		}
	}

	currentUser, err := user.Current()
	if err != nil {
		return &commands.Result{
			Output:   fmt.Sprintf("Error getting user information: %v\n", err),
			Error:    err,
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, err
	}

	if !verbose {
		// Simple output - just username
		return &commands.Result{
			Output:   currentUser.Username + "\n",
			ExitCode: 0,
			Duration: time.Since(startTime),
		}, nil
	}

	// Verbose output with detailed information
	var output string

	output += color.New(color.FgCyan, color.Bold).Sprint("👤 USER INFORMATION\n")
	output += "═══════════════════════════════════════════════════════════════\n\n"

	// Basic user info
	output += color.New(color.FgGreen, color.Bold).Sprint("📋 Identity\n")
	output += fmt.Sprintf("  Username:     %s\n", color.New(color.FgWhite, color.Bold).Sprint(currentUser.Username))
	output += fmt.Sprintf("  User ID:      %s\n", currentUser.Uid)
	output += fmt.Sprintf("  Group ID:     %s\n", currentUser.Gid)
	output += fmt.Sprintf("  Display Name: %s\n", currentUser.Name)
	output += fmt.Sprintf("  Home Dir:     %s\n", currentUser.HomeDir)
	output += "\n"

	// Environment info
	output += color.New(color.FgBlue, color.Bold).Sprint("🌍 Environment\n")
	if hostname, err := os.Hostname(); err == nil {
		output += fmt.Sprintf("  Hostname:     %s\n", hostname)
	}
	output += fmt.Sprintf("  OS:           %s\n", runtime.GOOS)
	output += fmt.Sprintf("  Architecture: %s\n", runtime.GOARCH)

	// Working directory
	if wd, err := os.Getwd(); err == nil {
		output += fmt.Sprintf("  Working Dir:  %s\n", wd)
	}
	output += "\n"

	// Process info
	output += color.New(color.FgMagenta, color.Bold).Sprint("⚙️  Process\n")
	output += fmt.Sprintf("  Process ID:   %d\n", os.Getpid())
	output += fmt.Sprintf("  Parent PID:   %d\n", os.Getppid())

	if exe, err := os.Executable(); err == nil {
		output += fmt.Sprintf("  Executable:   %s\n", exe)
	}
	output += "\n"

	// Environment variables (key ones)
	output += color.New(color.FgYellow, color.Bold).Sprint("🔧 Key Environment Variables\n")
	envVars := map[string]string{
		"PATH":        os.Getenv("PATH"),
		"HOME":        os.Getenv("HOME"),
		"USER":        os.Getenv("USER"),
		"USERNAME":    os.Getenv("USERNAME"),
		"USERPROFILE": os.Getenv("USERPROFILE"),
		"SHELL":       os.Getenv("SHELL"),
		"TERM":        os.Getenv("TERM"),
	}

	for key, value := range envVars {
		if value != "" {
			if key == "PATH" {
				// Truncate PATH for readability
				if len(value) > 80 {
					value = value[:77] + "..."
				}
			}
			output += fmt.Sprintf("  %-12s %s\n", key+":", value)
		}
	}

	output += "\n═══════════════════════════════════════════════════════════════\n"
	output += color.New(color.FgHiBlack).Sprintf("Generated at %s\n", time.Now().Format("2006-01-02 15:04:05"))

	return &commands.Result{
		Output:   output,
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}
