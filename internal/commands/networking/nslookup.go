package networking

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"

	"suppercommand/internal/commands"

	"github.com/fatih/color"
)

// NslookupCommand performs DNS lookups
type NslookupCommand struct {
	*commands.BaseCommand
}

// NewNslookupCommand creates a new nslookup command
func NewNslookupCommand() *NslookupCommand {
	return &NslookupCommand{
		BaseCommand: commands.NewBaseCommand(
			"nslookup",
			"Query DNS records for a domain with enhanced output",
			"nslookup <domain> [server]",
			[]string{"windows", "linux", "darwin"},
			false,
		),
	}
}

// Execute performs DNS lookup with enhanced formatting
func (n *NslookupCommand) Execute(ctx context.Context, args *commands.Arguments) (*commands.Result, error) {
	startTime := time.Now()

	if len(args.Raw) == 0 {
		return &commands.Result{
			Output:   "Usage: nslookup <domain> [server]\n",
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	domain := args.Raw[0]
	var server string
	if len(args.Raw) > 1 {
		server = args.Raw[1]
	}

	var output strings.Builder

	// Header
	output.WriteString(color.New(color.FgCyan, color.Bold).Sprintf("🔍 DNS LOOKUP for %s\n", domain))
	output.WriteString("═══════════════════════════════════════════════════════════════\n\n")

	// Show progress
	fmt.Print("🌐 Resolving DNS records")
	done := make(chan bool)
	go n.showProgress(done)

	// Perform Go-native DNS lookup first
	ips, err := net.LookupIP(domain)
	done <- true
	fmt.Print("\r\033[K") // Clear progress line

	if err != nil {
		output.WriteString(color.New(color.FgRed, color.Bold).Sprintf("❌ DNS Resolution Failed: %v\n", err))
	} else {
		output.WriteString(color.New(color.FgGreen, color.Bold).Sprint("✅ DNS Resolution Successful\n\n"))

		// Categorize IPs
		var ipv4s, ipv6s []net.IP
		for _, ip := range ips {
			if ip.To4() != nil {
				ipv4s = append(ipv4s, ip)
			} else {
				ipv6s = append(ipv6s, ip)
			}
		}

		// Display IPv4 addresses
		if len(ipv4s) > 0 {
			output.WriteString(color.New(color.FgBlue, color.Bold).Sprint("🌐 IPv4 Addresses:\n"))
			for _, ip := range ipv4s {
				output.WriteString(fmt.Sprintf("  %s\n", color.New(color.FgGreen).Sprint(ip.String())))
			}
			output.WriteString("\n")
		}

		// Display IPv6 addresses
		if len(ipv6s) > 0 {
			output.WriteString(color.New(color.FgMagenta, color.Bold).Sprint("🌐 IPv6 Addresses:\n"))
			for _, ip := range ipv6s {
				output.WriteString(fmt.Sprintf("  %s\n", color.New(color.FgCyan).Sprint(ip.String())))
			}
			output.WriteString("\n")
		}
	}

	// Additional DNS information using system nslookup
	output.WriteString(color.New(color.FgYellow, color.Bold).Sprint("📋 Detailed DNS Information:\n"))

	var cmd *exec.Cmd
	if server != "" {
		cmd = exec.CommandContext(ctx, "nslookup", domain, server)
	} else {
		cmd = exec.CommandContext(ctx, "nslookup", domain)
	}

	cmdOutput, err := cmd.Output()
	if err != nil {
		output.WriteString(color.New(color.FgRed).Sprintf("System nslookup failed: %v\n", err))
	} else {
		// Parse and colorize nslookup output
		scanner := bufio.NewScanner(strings.NewReader(string(cmdOutput)))
		for scanner.Scan() {
			line := scanner.Text()
			n.formatNslookupLine(&output, line)
		}
	}

	// Additional lookups
	output.WriteString("\n" + color.New(color.FgCyan, color.Bold).Sprint("🔍 Additional Information:\n"))

	// CNAME lookup
	if cname, err := net.LookupCNAME(domain); err == nil && cname != domain+"." {
		output.WriteString(fmt.Sprintf("  CNAME: %s\n", color.New(color.FgYellow).Sprint(cname)))
	}

	// MX records
	if mxRecords, err := net.LookupMX(domain); err == nil && len(mxRecords) > 0 {
		output.WriteString(color.New(color.FgBlue).Sprint("  MX Records:\n"))
		for _, mx := range mxRecords {
			output.WriteString(fmt.Sprintf("    %d %s\n", mx.Pref, mx.Host))
		}
	}

	// TXT records
	if txtRecords, err := net.LookupTXT(domain); err == nil && len(txtRecords) > 0 {
		output.WriteString(color.New(color.FgGreen).Sprint("  TXT Records:\n"))
		for _, txt := range txtRecords {
			if len(txt) > 80 {
				output.WriteString(fmt.Sprintf("    %s...\n", txt[:77]))
			} else {
				output.WriteString(fmt.Sprintf("    %s\n", txt))
			}
		}
	}

	output.WriteString("\n═══════════════════════════════════════════════════════════════\n")
	output.WriteString(color.New(color.FgHiBlack).Sprintf("Lookup completed in %v\n", time.Since(startTime).Round(time.Millisecond)))

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// showProgress shows a spinner during DNS lookup
func (n *NslookupCommand) showProgress(done chan bool) {
	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	i := 0

	for {
		select {
		case <-done:
			return
		default:
			fmt.Printf("\r🌐 Resolving DNS records %s", color.New(color.FgYellow).Sprint(spinner[i%len(spinner)]))
			time.Sleep(100 * time.Millisecond)
			i++
		}
	}
}

// formatNslookupLine formats a line from nslookup output
func (n *NslookupCommand) formatNslookupLine(output *strings.Builder, line string) {
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}

	lower := strings.ToLower(line)

	switch {
	case strings.Contains(lower, "server:") || strings.Contains(lower, "address:"):
		output.WriteString(fmt.Sprintf("  %s\n", color.New(color.FgCyan).Sprint(line)))
	case strings.Contains(lower, "name:"):
		output.WriteString(fmt.Sprintf("  %s\n", color.New(color.FgGreen, color.Bold).Sprint(line)))
	case strings.Contains(lower, "canonical name") || strings.Contains(lower, "alias"):
		output.WriteString(fmt.Sprintf("  %s\n", color.New(color.FgYellow).Sprint(line)))
	case strings.Contains(lower, "non-authoritative") || strings.Contains(lower, "authoritative"):
		output.WriteString(fmt.Sprintf("  %s\n", color.New(color.FgBlue).Sprint(line)))
	default:
		output.WriteString(fmt.Sprintf("  %s\n", line))
	}
}
