package system

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"suppercommand/internal/commands"

	"github.com/fatih/color"
)

// SysInfoCommand shows system information
type SysInfoCommand struct {
	*commands.BaseCommand
}

// NewSysInfoCommand creates a new sysinfo command
func NewSysInfoCommand() *SysInfoCommand {
	return &SysInfoCommand{
		BaseCommand: commands.NewBaseCommand(
			"sysinfo",
			"Display comprehensive system information",
			"sysinfo [-v|--verbose]",
			[]string{"windows", "linux", "darwin"},
			false,
		),
	}
}

// Execute shows system information
func (s *SysInfoCommand) Execute(ctx context.Context, args *commands.Arguments) (*commands.Result, error) {
	startTime := time.Now()

	verbose := false
	for _, arg := range args.Raw {
		if arg == "-v" || arg == "--verbose" {
			verbose = true
		}
	}

	var output strings.Builder

	// Header
	output.WriteString(color.New(color.FgCyan, color.Bold).Sprint("🖥️  SYSTEM INFORMATION") + "\n")
	output.WriteString("═══════════════════════════════════════════════════════════════\n\n")

	// Operating System
	output.WriteString(color.New(color.FgGreen, color.Bold).Sprint("🌐 Operating System\n"))
	output.WriteString(fmt.Sprintf("  OS:           %s\n", runtime.GOOS))
	output.WriteString(fmt.Sprintf("  Architecture: %s\n", runtime.GOARCH))

	if hostname, err := os.Hostname(); err == nil {
		output.WriteString(fmt.Sprintf("  Hostname:     %s\n", hostname))
	}

	output.WriteString("\n")

	// Runtime Information
	output.WriteString(color.New(color.FgBlue, color.Bold).Sprint("⚡ Runtime Information\n"))
	output.WriteString(fmt.Sprintf("  Go Version:   %s\n", runtime.Version()))
	output.WriteString(fmt.Sprintf("  CPUs:         %d\n", runtime.NumCPU()))
	output.WriteString(fmt.Sprintf("  Goroutines:   %d\n", runtime.NumGoroutine()))

	// Memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	output.WriteString(fmt.Sprintf("  Memory Alloc: %s\n", formatBytes(m.Alloc)))
	output.WriteString(fmt.Sprintf("  Total Alloc:  %s\n", formatBytes(m.TotalAlloc)))
	output.WriteString(fmt.Sprintf("  Sys Memory:   %s\n", formatBytes(m.Sys)))
	output.WriteString("\n")

	// Environment
	output.WriteString(color.New(color.FgYellow, color.Bold).Sprint("🌍 Environment\n"))

	// Key environment variables
	envVars := []string{"PATH", "HOME", "USER", "USERNAME", "USERPROFILE", "TEMP", "TMP"}
	for _, envVar := range envVars {
		if value := os.Getenv(envVar); value != "" {
			if envVar == "PATH" && !verbose {
				// Truncate PATH for readability
				if len(value) > 80 {
					value = value[:77] + "..."
				}
			}
			output.WriteString(fmt.Sprintf("  %-12s %s\n", envVar+":", value))
		}
	}
	output.WriteString("\n")

	// Process Information
	output.WriteString(color.New(color.FgMagenta, color.Bold).Sprint("🔧 Process Information\n"))
	output.WriteString(fmt.Sprintf("  PID:          %d\n", os.Getpid()))
	output.WriteString(fmt.Sprintf("  PPID:         %d\n", os.Getppid()))

	if wd, err := os.Getwd(); err == nil {
		output.WriteString(fmt.Sprintf("  Working Dir:  %s\n", wd))
	}

	if exe, err := os.Executable(); err == nil {
		output.WriteString(fmt.Sprintf("  Executable:   %s\n", exe))
	}

	output.WriteString("\n")

	// Verbose information
	if verbose {
		output.WriteString(color.New(color.FgRed, color.Bold).Sprint("🔍 Detailed Information\n"))

		// All environment variables
		output.WriteString("Environment Variables:\n")
		for _, env := range os.Environ() {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				output.WriteString(fmt.Sprintf("  %s = %s\n", parts[0], parts[1]))
			}
		}
		output.WriteString("\n")

		// Detailed memory stats
		output.WriteString("Detailed Memory Statistics:\n")
		output.WriteString(fmt.Sprintf("  HeapAlloc:    %s\n", formatBytes(m.HeapAlloc)))
		output.WriteString(fmt.Sprintf("  HeapSys:      %s\n", formatBytes(m.HeapSys)))
		output.WriteString(fmt.Sprintf("  HeapIdle:     %s\n", formatBytes(m.HeapIdle)))
		output.WriteString(fmt.Sprintf("  HeapInuse:    %s\n", formatBytes(m.HeapInuse)))
		output.WriteString(fmt.Sprintf("  HeapReleased: %s\n", formatBytes(m.HeapReleased)))
		output.WriteString(fmt.Sprintf("  HeapObjects:  %d\n", m.HeapObjects))
		output.WriteString(fmt.Sprintf("  GC Cycles:    %d\n", m.NumGC))
		output.WriteString(fmt.Sprintf("  Last GC:      %s ago\n", time.Since(time.Unix(0, int64(m.LastGC))).Round(time.Millisecond)))
	}

	// Footer
	output.WriteString("═══════════════════════════════════════════════════════════════\n")
	output.WriteString(color.New(color.FgHiBlack).Sprintf("Generated at %s\n", time.Now().Format("2006-01-02 15:04:05")))

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// formatBytes formats bytes in human readable format
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
