package networking

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"suppercommand/internal/commands"

	"github.com/fatih/color"
)

// IpconfigCommand shows network interface configuration
type IpconfigCommand struct {
	*commands.BaseCommand
}

// NewIpconfigCommand creates a new ipconfig command
func NewIpconfigCommand() *IpconfigCommand {
	return &IpconfigCommand{
		BaseCommand: commands.NewBaseCommand(
			"ipconfig",
			"Display network interface configuration",
			"ipconfig [/all] [/release] [/renew] [/flushdns]",
			[]string{"windows", "linux", "darwin"},
			false,
		),
	}
}

// Execute shows network interface configuration
func (i *IpconfigCommand) Execute(ctx context.Context, args *commands.Arguments) (*commands.Result, error) {
	startTime := time.Now()

	// Parse flags
	showAll := false
	release := false
	renew := false
	flushDNS := false

	for _, arg := range args.Raw {
		switch strings.ToLower(arg) {
		case "/all", "-all", "--all":
			showAll = true
		case "/release", "-release", "--release":
			release = true
		case "/renew", "-renew", "--renew":
			renew = true
		case "/flushdns", "-flushdns", "--flushdns":
			flushDNS = true
		}
	}

	var output strings.Builder

	// Header
	output.WriteString(color.New(color.FgCyan, color.Bold).Sprint("🌐 NETWORK CONFIGURATION\n"))
	output.WriteString("═══════════════════════════════════════════════════════════════\n\n")

	// Handle special operations first
	if release || renew || flushDNS {
		return i.handleSpecialOperations(ctx, release, renew, flushDNS, startTime)
	}

	// Show progress
	fmt.Print("📊 Gathering network interface information")
	done := make(chan bool)
	go i.showProgress(done)

	// Get network interfaces using Go's net package
	interfaces, err := net.Interfaces()
	done <- true
	fmt.Print("\r\033[K") // Clear progress line

	if err != nil {
		output.WriteString(color.New(color.FgRed).Sprintf("❌ Failed to get network interfaces: %v\n", err))
		return &commands.Result{
			Output:   output.String(),
			Error:    err,
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, err
	}

	// Display interfaces
	activeInterfaces := 0
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 && !showAll {
			continue // Skip down interfaces unless showing all
		}

		activeInterfaces++
		i.displayInterface(&output, iface, showAll)
	}

	// System ipconfig output (if available)
	if showAll {
		output.WriteString("\n" + color.New(color.FgYellow, color.Bold).Sprint("📋 System Configuration Details:\n"))
		output.WriteString("═══════════════════════════════════════════════════════════════\n")

		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.CommandContext(ctx, "ipconfig", "/all")
		} else {
			cmd = exec.CommandContext(ctx, "ifconfig", "-a")
		}

		if cmdOutput, err := cmd.Output(); err == nil {
			scanner := bufio.NewScanner(strings.NewReader(string(cmdOutput)))
			for scanner.Scan() {
				line := scanner.Text()
				i.formatSystemLine(&output, line)
			}
		}
	}

	// Summary
	output.WriteString("\n" + color.New(color.FgGreen, color.Bold).Sprint("📊 SUMMARY\n"))
	output.WriteString("═══════════════════════════════════════════════════════════════\n")
	output.WriteString(fmt.Sprintf("  Total Interfaces: %d\n", len(interfaces)))
	output.WriteString(fmt.Sprintf("  Active Interfaces: %d\n", activeInterfaces))

	output.WriteString("\n═══════════════════════════════════════════════════════════════\n")
	output.WriteString(color.New(color.FgHiBlack).Sprintf("Completed in %v\n",
		time.Since(startTime).Round(time.Millisecond)))

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// displayInterface displays information for a single network interface
func (i *IpconfigCommand) displayInterface(output *strings.Builder, iface net.Interface, showAll bool) {
	// Interface header
	status := "DOWN"
	statusColor := color.New(color.FgRed)
	if iface.Flags&net.FlagUp != 0 {
		status = "UP"
		statusColor = color.New(color.FgGreen)
	}

	output.WriteString(color.New(color.FgBlue, color.Bold).Sprintf("🔌 %s\n", iface.Name))
	output.WriteString(fmt.Sprintf("   Status:      %s\n", statusColor.Sprint(status)))

	if showAll {
		output.WriteString(fmt.Sprintf("   Index:       %d\n", iface.Index))
		output.WriteString(fmt.Sprintf("   MTU:         %d\n", iface.MTU))
		if iface.HardwareAddr != nil {
			output.WriteString(fmt.Sprintf("   MAC Address: %s\n",
				color.New(color.FgMagenta).Sprint(iface.HardwareAddr.String())))
		}
		output.WriteString(fmt.Sprintf("   Flags:       %s\n", i.formatFlags(iface.Flags)))
	}

	// Get addresses
	addrs, err := iface.Addrs()
	if err != nil {
		output.WriteString(fmt.Sprintf("   Error getting addresses: %v\n", err))
	} else {
		ipv4Count := 0
		ipv6Count := 0

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				ip := ipnet.IP
				if ip.To4() != nil {
					// IPv4
					ipv4Count++
					output.WriteString(fmt.Sprintf("   IPv4:        %s\n",
						color.New(color.FgGreen).Sprint(ip.String())))
					if showAll {
						output.WriteString(fmt.Sprintf("   Subnet:      %s\n", ipnet.Mask.String()))
					}
				} else if ip.To16() != nil {
					// IPv6
					ipv6Count++
					if showAll || !ip.IsLinkLocalUnicast() {
						output.WriteString(fmt.Sprintf("   IPv6:        %s\n",
							color.New(color.FgCyan).Sprint(ip.String())))
					}
				}
			}
		}

		if showAll {
			output.WriteString(fmt.Sprintf("   IPv4 Count:  %d\n", ipv4Count))
			output.WriteString(fmt.Sprintf("   IPv6 Count:  %d\n", ipv6Count))
		}
	}

	output.WriteString("\n")
}

// formatFlags formats interface flags
func (i *IpconfigCommand) formatFlags(flags net.Flags) string {
	var flagStrings []string

	if flags&net.FlagUp != 0 {
		flagStrings = append(flagStrings, color.New(color.FgGreen).Sprint("UP"))
	}
	if flags&net.FlagBroadcast != 0 {
		flagStrings = append(flagStrings, "BROADCAST")
	}
	if flags&net.FlagLoopback != 0 {
		flagStrings = append(flagStrings, "LOOPBACK")
	}
	if flags&net.FlagPointToPoint != 0 {
		flagStrings = append(flagStrings, "POINT-TO-POINT")
	}
	if flags&net.FlagMulticast != 0 {
		flagStrings = append(flagStrings, "MULTICAST")
	}

	return strings.Join(flagStrings, " ")
}

// formatSystemLine formats a line from system ipconfig/ifconfig output
func (i *IpconfigCommand) formatSystemLine(output *strings.Builder, line string) {
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}

	lower := strings.ToLower(line)

	switch {
	case strings.Contains(lower, "adapter") || strings.Contains(lower, "ethernet"):
		output.WriteString(color.New(color.FgBlue, color.Bold).Sprintf("   %s\n", line))
	case strings.Contains(lower, "ip address") || strings.Contains(lower, "inet "):
		output.WriteString(color.New(color.FgGreen).Sprintf("   %s\n", line))
	case strings.Contains(lower, "subnet mask") || strings.Contains(lower, "netmask"):
		output.WriteString(color.New(color.FgYellow).Sprintf("   %s\n", line))
	case strings.Contains(lower, "default gateway") || strings.Contains(lower, "gateway"):
		output.WriteString(color.New(color.FgCyan).Sprintf("   %s\n", line))
	case strings.Contains(lower, "dns") || strings.Contains(lower, "nameserver"):
		output.WriteString(color.New(color.FgMagenta).Sprintf("   %s\n", line))
	default:
		output.WriteString(fmt.Sprintf("   %s\n", line))
	}
}

// handleSpecialOperations handles release, renew, and flushdns operations
func (i *IpconfigCommand) handleSpecialOperations(ctx context.Context, release, renew, flushDNS bool, startTime time.Time) (*commands.Result, error) {
	var output strings.Builder

	output.WriteString(color.New(color.FgYellow, color.Bold).Sprint("⚙️  NETWORK OPERATIONS\n"))
	output.WriteString("═══════════════════════════════════════════════════════════════\n\n")

	if flushDNS {
		output.WriteString("🔄 Flushing DNS cache...\n")

		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.CommandContext(ctx, "ipconfig", "/flushdns")
		} else {
			// On Linux/macOS, try different methods
			cmd = exec.CommandContext(ctx, "sudo", "systemctl", "restart", "systemd-resolved")
		}

		if err := cmd.Run(); err != nil {
			output.WriteString(color.New(color.FgRed).Sprintf("❌ Failed to flush DNS: %v\n", err))
		} else {
			output.WriteString(color.New(color.FgGreen).Sprint("✅ DNS cache flushed successfully\n"))
		}
		output.WriteString("\n")
	}

	if release {
		output.WriteString("📤 Releasing IP configuration...\n")

		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.CommandContext(ctx, "ipconfig", "/release")
		} else {
			output.WriteString(color.New(color.FgYellow).Sprint("⚠️  Release operation not supported on this platform\n"))
		}

		if cmd != nil {
			if err := cmd.Run(); err != nil {
				output.WriteString(color.New(color.FgRed).Sprintf("❌ Failed to release IP: %v\n", err))
			} else {
				output.WriteString(color.New(color.FgGreen).Sprint("✅ IP configuration released\n"))
			}
		}
		output.WriteString("\n")
	}

	if renew {
		output.WriteString("📥 Renewing IP configuration...\n")

		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.CommandContext(ctx, "ipconfig", "/renew")
		} else {
			output.WriteString(color.New(color.FgYellow).Sprint("⚠️  Renew operation not supported on this platform\n"))
		}

		if cmd != nil {
			if err := cmd.Run(); err != nil {
				output.WriteString(color.New(color.FgRed).Sprintf("❌ Failed to renew IP: %v\n", err))
			} else {
				output.WriteString(color.New(color.FgGreen).Sprint("✅ IP configuration renewed\n"))
			}
		}
	}

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// showProgress shows a spinner during interface gathering
func (i *IpconfigCommand) showProgress(done chan bool) {
	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	idx := 0

	for {
		select {
		case <-done:
			return
		default:
			fmt.Printf("\r📊 Gathering network interface information %s",
				color.New(color.FgYellow).Sprint(spinner[idx%len(spinner)]))
			time.Sleep(100 * time.Millisecond)
			idx++
		}
	}
}
