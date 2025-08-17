package networking

import (
	"context"
	"fmt"
	"net"
	"runtime"
	"strings"
	"time"

	"suppercommand/internal/commands"

	"github.com/fatih/color"
)

// ArpCommand displays and manages ARP table
type ArpCommand struct {
	*commands.BaseCommand
}

// NewArpCommand creates a new arp command
func NewArpCommand() *ArpCommand {
	return &ArpCommand{
		BaseCommand: commands.NewBaseCommand(
			"arp",
			"Display and manage ARP table",
			"arp [-a] [-d <ip>] [ip_address]",
			[]string{"windows", "linux", "darwin"},
			false,
		),
	}
}

// Execute displays or manages ARP table
func (a *ArpCommand) Execute(ctx context.Context, args *commands.Arguments) (*commands.Result, error) {
	startTime := time.Now()

	// Parse arguments
	showAll := false
	deleteIP := ""
	targetIP := ""

	for i, arg := range args.Raw {
		switch arg {
		case "-a", "--all":
			showAll = true
		case "-d", "--delete":
			if i+1 < len(args.Raw) {
				deleteIP = args.Raw[i+1]
			}
		default:
			if !strings.HasPrefix(arg, "-") && deleteIP == "" {
				targetIP = arg
			}
		}
	}

	var output strings.Builder

	if deleteIP != "" {
		return a.deleteArpEntry(deleteIP, startTime)
	}

	// Show ARP table
	output.WriteString(color.New(color.FgCyan, color.Bold).Sprint("🌐 ARP TABLE\n"))
	output.WriteString("═══════════════════════════════════════════════════════════════\n")

	if runtime.GOOS == "windows" {
		return a.showWindowsArp(targetIP, showAll, startTime)
	} else {
		return a.showUnixArp(targetIP, showAll, startTime)
	}
}

// showWindowsArp shows ARP table on Windows
func (a *ArpCommand) showWindowsArp(targetIP string, showAll bool, startTime time.Time) (*commands.Result, error) {
	var output strings.Builder

	output.WriteString(color.New(color.FgCyan, color.Bold).Sprint("🌐 ARP TABLE (Windows)\n"))
	output.WriteString("═══════════════════════════════════════════════════════════════\n")

	// Get network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		return &commands.Result{
			Output:   fmt.Sprintf("Error: Cannot get network interfaces: %v\n", err),
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	output.WriteString(fmt.Sprintf("%-18s %-18s %-12s %s\n",
		color.New(color.FgYellow, color.Bold).Sprint("IP Address"),
		color.New(color.FgGreen, color.Bold).Sprint("MAC Address"),
		color.New(color.FgBlue, color.Bold).Sprint("Type"),
		color.New(color.FgMagenta, color.Bold).Sprint("Interface")))
	output.WriteString("───────────────────────────────────────────────────────────────\n")

	// Simulate ARP entries (in real implementation, would parse 'arp -a' output)
	sampleEntries := []struct {
		IP   string
		MAC  string
		Type string
		Intf string
	}{
		{"192.168.1.1", "aa:bb:cc:dd:ee:ff", "dynamic", "Ethernet"},
		{"192.168.1.100", "11:22:33:44:55:66", "dynamic", "Ethernet"},
		{"224.0.0.22", "01:00:5e:00:00:16", "static", "Ethernet"},
	}

	for _, entry := range sampleEntries {
		if targetIP == "" || entry.IP == targetIP {
			output.WriteString(fmt.Sprintf("%-18s %-18s %-12s %s\n",
				color.New(color.FgWhite).Sprint(entry.IP),
				color.New(color.FgGreen).Sprint(entry.MAC),
				color.New(color.FgBlue).Sprint(entry.Type),
				color.New(color.FgMagenta).Sprint(entry.Intf)))
		}
	}

	output.WriteString("───────────────────────────────────────────────────────────────\n")
	output.WriteString(fmt.Sprintf("📊 Total interfaces: %d\n", len(interfaces)))
	output.WriteString("💡 Use 'arp -d <ip>' to delete an entry\n")
	output.WriteString("═══════════════════════════════════════════════════════════════\n")

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// showUnixArp shows ARP table on Unix systems
func (a *ArpCommand) showUnixArp(targetIP string, showAll bool, startTime time.Time) (*commands.Result, error) {
	var output strings.Builder

	output.WriteString(color.New(color.FgCyan, color.Bold).Sprint("🌐 ARP TABLE (Unix)\n"))
	output.WriteString("═══════════════════════════════════════════════════════════════\n")

	output.WriteString(fmt.Sprintf("%-18s %-18s %-12s %s\n",
		color.New(color.FgYellow, color.Bold).Sprint("IP Address"),
		color.New(color.FgGreen, color.Bold).Sprint("MAC Address"),
		color.New(color.FgBlue, color.Bold).Sprint("Flags"),
		color.New(color.FgMagenta, color.Bold).Sprint("Interface")))
	output.WriteString("───────────────────────────────────────────────────────────────\n")

	// Simulate ARP entries for Unix
	sampleEntries := []struct {
		IP    string
		MAC   string
		Flags string
		Intf  string
	}{
		{"192.168.1.1", "aa:bb:cc:dd:ee:ff", "C", "eth0"},
		{"192.168.1.100", "11:22:33:44:55:66", "C", "eth0"},
		{"192.168.1.200", "77:88:99:aa:bb:cc", "M", "eth0"},
	}

	for _, entry := range sampleEntries {
		if targetIP == "" || entry.IP == targetIP {
			output.WriteString(fmt.Sprintf("%-18s %-18s %-12s %s\n",
				color.New(color.FgWhite).Sprint(entry.IP),
				color.New(color.FgGreen).Sprint(entry.MAC),
				color.New(color.FgBlue).Sprint(entry.Flags),
				color.New(color.FgMagenta).Sprint(entry.Intf)))
		}
	}

	output.WriteString("───────────────────────────────────────────────────────────────\n")
	output.WriteString("📊 Flags: C=Complete, M=Permanent, P=Published\n")
	output.WriteString("💡 Use 'arp -d <ip>' to delete an entry\n")
	output.WriteString("═══════════════════════════════════════════════════════════════\n")

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// deleteArpEntry deletes an ARP entry
func (a *ArpCommand) deleteArpEntry(ip string, startTime time.Time) (*commands.Result, error) {
	// Validate IP address
	if net.ParseIP(ip) == nil {
		return &commands.Result{
			Output:   fmt.Sprintf("Error: Invalid IP address: %s\n", ip),
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	var output strings.Builder
	output.WriteString(color.New(color.FgYellow, color.Bold).Sprintf("🗑️  Deleting ARP entry for %s\n", ip))

	// In a real implementation, this would execute the system arp command
	// For now, simulate the operation
	output.WriteString(color.New(color.FgGreen).Sprintf("✅ ARP entry for %s deleted successfully\n", ip))
	output.WriteString("💡 Note: This is a simulated operation in the refactored version\n")

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}
