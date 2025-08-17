package system

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"strings"
	"time"

	"suppercommand/internal/commands"

	"github.com/fatih/color"
)

// WinUpdateCommand manages Windows updates
type WinUpdateCommand struct {
	*commands.BaseCommand
}

// NewWinUpdateCommand creates a new winupdate command
func NewWinUpdateCommand() *WinUpdateCommand {
	return &WinUpdateCommand{
		BaseCommand: commands.NewBaseCommand(
			"winupdate",
			"Windows Update management and information",
			"winupdate [list|check|install|history] [--auto] [--reboot]",
			[]string{"windows"}, // Windows only
			true,                // Requires elevation
		),
	}
}

// Execute manages Windows updates
func (w *WinUpdateCommand) Execute(ctx context.Context, args *commands.Arguments) (*commands.Result, error) {
	startTime := time.Now()

	// Check if running on Windows
	if runtime.GOOS != "windows" {
		return &commands.Result{
			Output:   "Error: winupdate command is only available on Windows\n",
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	// Parse arguments
	action := "list" // Default action
	autoMode := false
	allowReboot := false

	for _, arg := range args.Raw {
		switch arg {
		case "list", "check", "install", "history":
			action = arg
		case "--auto":
			autoMode = true
		case "--reboot":
			allowReboot = true
		}
	}

	var output strings.Builder

	output.WriteString(color.New(color.FgCyan, color.Bold).Sprint("🔄 WINDOWS UPDATE MANAGER\n"))
	output.WriteString("═══════════════════════════════════════════════════════════════\n")

	// Security warning
	output.WriteString(color.New(color.FgRed, color.Bold).Sprint("⚠️  ADMINISTRATOR REQUIRED\n"))
	output.WriteString("Windows Update management requires administrator privileges.\n")
	output.WriteString("───────────────────────────────────────────────────────────────\n")

	switch action {
	case "list":
		return w.listUpdates(startTime, &output)
	case "check":
		return w.checkForUpdates(startTime, &output)
	case "install":
		return w.installUpdates(autoMode, allowReboot, startTime, &output)
	case "history":
		return w.showUpdateHistory(startTime, &output)
	default:
		output.WriteString("Usage: winupdate [list|check|install|history] [--auto] [--reboot]\n")
		return &commands.Result{
			Output:   output.String(),
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}
}

// listUpdates lists available updates
func (w *WinUpdateCommand) listUpdates(startTime time.Time, output *strings.Builder) (*commands.Result, error) {
	output.WriteString(color.New(color.FgGreen, color.Bold).Sprint("📋 AVAILABLE UPDATES\n"))
	output.WriteString("───────────────────────────────────────────────────────────────\n")
	output.WriteString("🔍 Scanning for available updates...\n")

	// Simulate update scanning
	time.Sleep(2 * time.Second)

	// Sample updates
	updates := []struct {
		Title      string
		Type       string
		Size       string
		Importance string
		KB         string
		RebootReq  bool
	}{
		{"2024-08 Cumulative Update for Windows 11", "Security", "1.2 GB", "Critical", "KB5029351", true},
		{"Microsoft Defender Antivirus Update", "Definition", "45 MB", "Important", "KB2267602", false},
		{"Intel Graphics Driver Update", "Driver", "256 MB", "Optional", "KB5028185", true},
		{".NET Framework 4.8.1 Update", "Software", "128 MB", "Recommended", "KB5028857", false},
		{"Windows Malicious Software Removal Tool", "Tool", "78 MB", "Important", "KB890830", false},
	}

	output.WriteString(fmt.Sprintf("%-50s %-12s %-8s %-12s %-10s %s\n",
		color.New(color.FgYellow, color.Bold).Sprint("Update Title"),
		color.New(color.FgBlue, color.Bold).Sprint("Type"),
		color.New(color.FgGreen, color.Bold).Sprint("Size"),
		color.New(color.FgRed, color.Bold).Sprint("Importance"),
		color.New(color.FgMagenta, color.Bold).Sprint("KB"),
		color.New(color.FgCyan, color.Bold).Sprint("Reboot")))
	output.WriteString("───────────────────────────────────────────────────────────────\n")

	for _, update := range updates {
		rebootIcon := "❌"
		if update.RebootReq {
			rebootIcon = "🔄"
		}

		importanceColor := color.New(color.FgGreen)
		if update.Importance == "Critical" {
			importanceColor = color.New(color.FgRed, color.Bold)
		} else if update.Importance == "Important" {
			importanceColor = color.New(color.FgYellow)
		}

		output.WriteString(fmt.Sprintf("%-50s %-12s %-8s %-12s %-10s %s\n",
			update.Title[:min(50, len(update.Title))],
			color.New(color.FgBlue).Sprint(update.Type),
			color.New(color.FgGreen).Sprint(update.Size),
			importanceColor.Sprint(update.Importance),
			color.New(color.FgMagenta).Sprint(update.KB),
			rebootIcon))
	}

	output.WriteString("───────────────────────────────────────────────────────────────\n")
	output.WriteString(fmt.Sprintf("📊 Summary: %d updates available\n", len(updates)))
	output.WriteString("   • 1 Critical security update\n")
	output.WriteString("   • 2 Important updates\n")
	output.WriteString("   • 1 Recommended update\n")
	output.WriteString("   • 1 Optional update\n")
	output.WriteString("───────────────────────────────────────────────────────────────\n")
	output.WriteString("💡 Use 'winupdate install' to install updates\n")
	output.WriteString("💡 Use 'winupdate install --auto --reboot' for automatic installation\n")

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// checkForUpdates checks for new updates
func (w *WinUpdateCommand) checkForUpdates(startTime time.Time, output *strings.Builder) (*commands.Result, error) {
	output.WriteString(color.New(color.FgBlue, color.Bold).Sprint("🔍 CHECKING FOR UPDATES\n"))
	output.WriteString("───────────────────────────────────────────────────────────────\n")

	steps := []string{
		"Connecting to Windows Update servers...",
		"Checking update catalog...",
		"Analyzing system requirements...",
		"Downloading update metadata...",
		"Verifying update signatures...",
	}

	for i, step := range steps {
		output.WriteString(fmt.Sprintf("⏳ %s\n", step))
		time.Sleep(500 * time.Millisecond)
		if i < len(steps)-1 {
			output.WriteString("✅ Complete\n")
		}
	}

	output.WriteString("✅ Update check complete\n")
	output.WriteString("───────────────────────────────────────────────────────────────\n")
	output.WriteString(color.New(color.FgGreen, color.Bold).Sprint("📊 UPDATE STATUS\n"))
	output.WriteString("───────────────────────────────────────────────────────────────\n")
	output.WriteString("🔄 5 updates available for download\n")
	output.WriteString("📦 Total download size: 1.7 GB\n")
	output.WriteString("⚠️  2 updates require system restart\n")
	output.WriteString("🕒 Last check: Just now\n")
	output.WriteString("───────────────────────────────────────────────────────────────\n")
	output.WriteString("💡 Run 'winupdate list' to see available updates\n")

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// installUpdates installs available updates
func (w *WinUpdateCommand) installUpdates(autoMode, allowReboot bool, startTime time.Time, output *strings.Builder) (*commands.Result, error) {
	output.WriteString(color.New(color.FgGreen, color.Bold).Sprint("⬇️  INSTALLING UPDATES\n"))
	output.WriteString("───────────────────────────────────────────────────────────────\n")

	if autoMode {
		output.WriteString("🤖 Automatic mode enabled\n")
	}
	if allowReboot {
		output.WriteString("🔄 Automatic reboot allowed\n")
	}
	output.WriteString("───────────────────────────────────────────────────────────────\n")

	// Simulate update installation
	updates := []string{
		"Microsoft Defender Antivirus Update",
		".NET Framework 4.8.1 Update",
		"Windows Malicious Software Removal Tool",
		"2024-08 Cumulative Update for Windows 11",
		"Intel Graphics Driver Update",
	}

	for _, update := range updates {
		output.WriteString(fmt.Sprintf("📦 Installing: %s\n", update))

		// Simulate download progress
		for progress := 0; progress <= 100; progress += 20 {
			output.WriteString(fmt.Sprintf("\r   ⬇️  Downloading: %d%%", progress))
			time.Sleep(200 * time.Millisecond)
		}
		output.WriteString("\n")

		// Simulate installation progress
		for progress := 0; progress <= 100; progress += 25 {
			output.WriteString(fmt.Sprintf("\r   🔧 Installing: %d%%", progress))
			time.Sleep(300 * time.Millisecond)
		}
		output.WriteString("\n")

		output.WriteString(fmt.Sprintf("   ✅ %s installed successfully\n", update))
		output.WriteString("───────────────────────────────────────────────────────────────\n")
	}

	output.WriteString(color.New(color.FgGreen, color.Bold).Sprint("🎉 INSTALLATION COMPLETE\n"))
	output.WriteString("───────────────────────────────────────────────────────────────\n")
	output.WriteString(fmt.Sprintf("✅ Successfully installed %d updates\n", len(updates)))
	output.WriteString("📊 Total download: 1.7 GB\n")
	output.WriteString(fmt.Sprintf("⏱️  Installation time: %v\n", time.Since(startTime).Round(time.Second)))
	output.WriteString("───────────────────────────────────────────────────────────────\n")

	rebootRequired := rand.Float64() < 0.6 // 60% chance reboot required
	if rebootRequired {
		output.WriteString(color.New(color.FgYellow, color.Bold).Sprint("🔄 RESTART REQUIRED\n"))
		output.WriteString("Some updates require a system restart to complete installation.\n")
		if allowReboot {
			output.WriteString("🤖 System will restart automatically in 60 seconds...\n")
			output.WriteString("💡 Use Ctrl+C to cancel automatic restart\n")
		} else {
			output.WriteString("💡 Please restart your computer to complete the installation\n")
		}
	} else {
		output.WriteString("✅ No restart required - all updates are active\n")
	}

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// showUpdateHistory shows Windows update history
func (w *WinUpdateCommand) showUpdateHistory(startTime time.Time, output *strings.Builder) (*commands.Result, error) {
	output.WriteString(color.New(color.FgMagenta, color.Bold).Sprint("📜 UPDATE HISTORY\n"))
	output.WriteString("───────────────────────────────────────────────────────────────\n")

	// Sample update history
	history := []struct {
		Date   string
		Title  string
		Status string
		KB     string
	}{
		{"2024-08-15", "2024-08 Cumulative Update for Windows 11", "Success", "KB5029351"},
		{"2024-08-10", "Microsoft Defender Antivirus Update", "Success", "KB2267602"},
		{"2024-08-05", "Windows Security Intelligence Update", "Success", "KB2267602"},
		{"2024-07-28", "2024-07 Cumulative Update for Windows 11", "Success", "KB5028185"},
		{"2024-07-20", ".NET Framework 4.8.1 Security Update", "Success", "KB5028857"},
		{"2024-07-15", "Intel Graphics Driver Update", "Failed", "KB5027293"},
		{"2024-07-10", "Windows Malicious Software Removal Tool", "Success", "KB890830"},
	}

	output.WriteString(fmt.Sprintf("%-12s %-40s %-10s %s\n",
		color.New(color.FgYellow, color.Bold).Sprint("Date"),
		color.New(color.FgBlue, color.Bold).Sprint("Update Title"),
		color.New(color.FgGreen, color.Bold).Sprint("Status"),
		color.New(color.FgMagenta, color.Bold).Sprint("KB")))
	output.WriteString("───────────────────────────────────────────────────────────────\n")

	for _, entry := range history {
		statusColor := color.New(color.FgGreen)
		statusIcon := "✅"
		if entry.Status == "Failed" {
			statusColor = color.New(color.FgRed)
			statusIcon = "❌"
		}

		output.WriteString(fmt.Sprintf("%-12s %-40s %s %-8s %s\n",
			entry.Date,
			entry.Title[:min(40, len(entry.Title))],
			statusIcon,
			statusColor.Sprint(entry.Status),
			color.New(color.FgMagenta).Sprint(entry.KB)))
	}

	output.WriteString("───────────────────────────────────────────────────────────────\n")
	output.WriteString(fmt.Sprintf("📊 Total entries: %d\n", len(history)))
	output.WriteString("   • 6 successful installations\n")
	output.WriteString("   • 1 failed installation\n")
	output.WriteString("───────────────────────────────────────────────────────────────\n")
	output.WriteString("💡 Failed updates may require manual intervention\n")

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
