package system

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"suppercommand/internal/commands"

	"github.com/fatih/color"
)

const (
	SuperShellVersion = "2.0.0"
	SuperShellBuild   = "refactored"
)

// VerCommand shows version information
type VerCommand struct {
	*commands.BaseCommand
}

// NewVerCommand creates a new ver command
func NewVerCommand() *VerCommand {
	return &VerCommand{
		BaseCommand: commands.NewBaseCommand(
			"ver",
			"Display version information",
			"ver [-v|--verbose]",
			[]string{"windows", "linux", "darwin"},
			false,
		),
	}
}

// Execute shows version information
func (v *VerCommand) Execute(ctx context.Context, args *commands.Arguments) (*commands.Result, error) {
	startTime := time.Now()

	verbose := false
	for _, arg := range args.Raw {
		if arg == "-v" || arg == "--verbose" {
			verbose = true
		}
	}

	if !verbose {
		// Simple version output
		return &commands.Result{
			Output:   fmt.Sprintf("SuperShell %s\n", SuperShellVersion),
			ExitCode: 0,
			Duration: time.Since(startTime),
		}, nil
	}

	// Verbose version information
	var output string

	output += color.New(color.FgCyan, color.Bold).Sprint("🚀 SUPERSHELL VERSION INFORMATION\n")
	output += "═══════════════════════════════════════════════════════════════\n\n"

	// Version details
	output += color.New(color.FgGreen, color.Bold).Sprint("📋 Version Details\n")
	output += fmt.Sprintf("  Product:      %s\n", color.New(color.FgWhite, color.Bold).Sprint("SuperShell"))
	output += fmt.Sprintf("  Version:      %s\n", color.New(color.FgCyan).Sprint(SuperShellVersion))
	output += fmt.Sprintf("  Build:        %s\n", color.New(color.FgYellow).Sprint(SuperShellBuild))
	output += fmt.Sprintf("  Architecture: %s\n", SuperShellBuild)
	output += "\n"

	// Runtime information
	output += color.New(color.FgBlue, color.Bold).Sprint("⚡ Runtime Information\n")
	output += fmt.Sprintf("  Go Version:   %s\n", runtime.Version())
	output += fmt.Sprintf("  OS/Arch:      %s/%s\n", runtime.GOOS, runtime.GOARCH)
	output += fmt.Sprintf("  CPUs:         %d\n", runtime.NumCPU())
	output += "\n"

	// Features
	output += color.New(color.FgMagenta, color.Bold).Sprint("🌟 Features\n")
	features := []string{
		"✅ Enhanced Command Line Interface",
		"✅ Live Feedback & Progress Indicators",
		"✅ Rich Colored Output",
		"✅ Advanced Networking Tools",
		"✅ Comprehensive File Operations",
		"✅ Cross-Platform Compatibility",
		"✅ Tab Completion & Smart Suggestions",
		"✅ Security Validation & Sanitization",
		"✅ Performance Monitoring",
		"✅ Modular Architecture",
	}

	for _, feature := range features {
		output += fmt.Sprintf("  %s\n", feature)
	}
	output += "\n"

	// Copyright
	output += color.New(color.FgYellow, color.Bold).Sprint("📄 Information\n")
	output += "  Description:  Advanced command-line shell with enhanced features\n"
	output += "  License:      Open Source\n"
	output += fmt.Sprintf("  Build Date:   %s\n", time.Now().Format("2006-01-02"))

	output += "\n═══════════════════════════════════════════════════════════════\n"
	output += color.New(color.FgHiBlack).Sprintf("Generated at %s\n", time.Now().Format("2006-01-02 15:04:05"))

	return &commands.Result{
		Output:   output,
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}
