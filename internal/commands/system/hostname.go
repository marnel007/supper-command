package system

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"suppercommand/internal/commands"

	"github.com/fatih/color"
)

// HostnameCommand shows or sets the system hostname
type HostnameCommand struct {
	*commands.BaseCommand
}

// NewHostnameCommand creates a new hostname command
func NewHostnameCommand() *HostnameCommand {
	return &HostnameCommand{
		BaseCommand: commands.NewBaseCommand(
			"hostname",
			"Display or set the system hostname",
			"hostname [-v|--verbose] [-i|--ip]",
			[]string{"windows", "linux", "darwin"},
			false,
		),
	}
}

// Execute shows hostname information
func (h *HostnameCommand) Execute(ctx context.Context, args *commands.Arguments) (*commands.Result, error) {
	startTime := time.Now()

	verbose := false
	showIP := false

	for _, arg := range args.Raw {
		switch arg {
		case "-v", "--verbose":
			verbose = true
		case "-i", "--ip":
			showIP = true
		}
	}

	hostname, err := os.Hostname()
	if err != nil {
		return &commands.Result{
			Output:   fmt.Sprintf("Error getting hostname: %v\n", err),
			Error:    err,
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, err
	}

	if !verbose && !showIP {
		// Simple output - just hostname
		return &commands.Result{
			Output:   hostname + "\n",
			ExitCode: 0,
			Duration: time.Since(startTime),
		}, nil
	}

	var output strings.Builder

	if verbose {
		// Verbose output with detailed information
		output.WriteString(color.New(color.FgCyan, color.Bold).Sprint("🖥️  HOSTNAME INFORMATION\n"))
		output.WriteString("═══════════════════════════════════════════════════════════════\n\n")

		// Basic hostname info
		output.WriteString(color.New(color.FgGreen, color.Bold).Sprint("📋 System Identity\n"))
		output.WriteString(fmt.Sprintf("  Hostname:     %s\n", color.New(color.FgWhite, color.Bold).Sprint(hostname)))

		// Try to get FQDN
		if addrs, err := net.LookupAddr("127.0.0.1"); err == nil && len(addrs) > 0 {
			for _, addr := range addrs {
				if strings.Contains(addr, ".") {
					output.WriteString(fmt.Sprintf("  FQDN:         %s\n", addr))
					break
				}
			}
		}
		output.WriteString("\n")
	}

	if showIP || verbose {
		// Network information
		if verbose {
			output.WriteString(color.New(color.FgBlue, color.Bold).Sprint("🌐 Network Interfaces\n"))
		}

		interfaces, err := net.Interfaces()
		if err != nil {
			output.WriteString(color.New(color.FgRed).Sprintf("Error getting network interfaces: %v\n", err))
		} else {
			for _, iface := range interfaces {
				// Skip loopback and down interfaces for simple IP display
				if !showIP && !verbose && (iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0) {
					continue
				}

				addrs, err := iface.Addrs()
				if err != nil {
					continue
				}

				if verbose {
					output.WriteString(fmt.Sprintf("  Interface: %s\n", color.New(color.FgCyan).Sprint(iface.Name)))
					output.WriteString(fmt.Sprintf("    Status:  %s\n", h.getInterfaceStatus(iface.Flags)))
				}

				for _, addr := range addrs {
					if ipnet, ok := addr.(*net.IPNet); ok {
						ip := ipnet.IP
						if ip.To4() != nil {
							// IPv4
							if verbose {
								output.WriteString(fmt.Sprintf("    IPv4:    %s\n", color.New(color.FgGreen).Sprint(ip.String())))
							} else if showIP {
								output.WriteString(fmt.Sprintf("%s\n", ip.String()))
							}
						} else if ip.To16() != nil && !ip.IsLoopback() {
							// IPv6 (non-loopback)
							if verbose {
								output.WriteString(fmt.Sprintf("    IPv6:    %s\n", color.New(color.FgMagenta).Sprint(ip.String())))
							}
						}
					}
				}
				if verbose {
					output.WriteString("\n")
				}
			}
		}
	}

	if verbose {
		// DNS information
		output.WriteString(color.New(color.FgYellow, color.Bold).Sprint("🔍 DNS Resolution\n"))

		// Try to resolve our own hostname
		if ips, err := net.LookupIP(hostname); err == nil && len(ips) > 0 {
			output.WriteString("  Resolved IPs:\n")
			for _, ip := range ips {
				if ip.To4() != nil {
					output.WriteString(fmt.Sprintf("    %s (IPv4)\n", color.New(color.FgGreen).Sprint(ip.String())))
				} else {
					output.WriteString(fmt.Sprintf("    %s (IPv6)\n", color.New(color.FgMagenta).Sprint(ip.String())))
				}
			}
		} else {
			output.WriteString(fmt.Sprintf("  Resolution: %s\n", color.New(color.FgRed).Sprint("Failed")))
		}

		output.WriteString("\n═══════════════════════════════════════════════════════════════\n")
		output.WriteString(color.New(color.FgHiBlack).Sprintf("Generated at %s\n", time.Now().Format("2006-01-02 15:04:05")))
	}

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// getInterfaceStatus returns a human-readable status for network interface flags
func (h *HostnameCommand) getInterfaceStatus(flags net.Flags) string {
	var status []string

	if flags&net.FlagUp != 0 {
		status = append(status, color.New(color.FgGreen).Sprint("UP"))
	} else {
		status = append(status, color.New(color.FgRed).Sprint("DOWN"))
	}

	if flags&net.FlagLoopback != 0 {
		status = append(status, "LOOPBACK")
	}

	if flags&net.FlagBroadcast != 0 {
		status = append(status, "BROADCAST")
	}

	if flags&net.FlagMulticast != 0 {
		status = append(status, "MULTICAST")
	}

	return strings.Join(status, " ")
}
