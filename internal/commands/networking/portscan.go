package networking

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"suppercommand/internal/commands"

	"github.com/fatih/color"
)

// PortscanCommand performs TCP port scanning
type PortscanCommand struct {
	*commands.BaseCommand
}

// NewPortscanCommand creates a new portscan command
func NewPortscanCommand() *PortscanCommand {
	return &PortscanCommand{
		BaseCommand: commands.NewBaseCommand(
			"portscan",
			"Fast TCP port scanner with live feedback",
			"portscan [-p ports] [-t timeout] [-c concurrency] <host>",
			[]string{"windows", "linux", "darwin"},
			false,
		),
	}
}

// Execute performs port scanning with live feedback
func (p *PortscanCommand) Execute(ctx context.Context, args *commands.Arguments) (*commands.Result, error) {
	startTime := time.Now()

	if len(args.Raw) == 0 {
		return &commands.Result{
			Output:   "Usage: portscan [-p ports] [-t timeout] [-c concurrency] <host>\n",
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	// Parse arguments
	host := ""
	ports := "1-1000"
	timeout := 1 * time.Second
	concurrency := 100

	for i, arg := range args.Raw {
		switch arg {
		case "-p":
			if i+1 < len(args.Raw) {
				ports = args.Raw[i+1]
			}
		case "-t":
			if i+1 < len(args.Raw) {
				if t, err := strconv.Atoi(args.Raw[i+1]); err == nil {
					timeout = time.Duration(t) * time.Second
				}
			}
		case "-c":
			if i+1 < len(args.Raw) {
				if c, err := strconv.Atoi(args.Raw[i+1]); err == nil {
					concurrency = c
				}
			}
		default:
			if !strings.HasPrefix(arg, "-") && host == "" {
				// Skip if it's a value for a flag
				if i > 0 && (args.Raw[i-1] == "-p" || args.Raw[i-1] == "-t" || args.Raw[i-1] == "-c") {
					continue
				}
				host = arg
			}
		}
	}

	if host == "" {
		return &commands.Result{
			Output:   "Error: No host specified\n",
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	var output strings.Builder

	// Header
	output.WriteString(color.New(color.FgCyan, color.Bold).Sprintf("🎯 PORT SCAN: %s\n", host))
	output.WriteString("═══════════════════════════════════════════════════════════════\n\n")

	// Parse port range
	portList, err := p.parsePortRange(ports)
	if err != nil {
		output.WriteString(color.New(color.FgRed).Sprintf("❌ Invalid port range: %v\n", err))
		return &commands.Result{
			Output:   output.String(),
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	// Show scan parameters
	output.WriteString(color.New(color.FgBlue, color.Bold).Sprint("📋 Scan Parameters:\n"))
	output.WriteString(fmt.Sprintf("  Target:       %s\n", color.New(color.FgWhite, color.Bold).Sprint(host)))
	output.WriteString(fmt.Sprintf("  Ports:        %s (%d ports)\n", ports, len(portList)))
	output.WriteString(fmt.Sprintf("  Timeout:      %v\n", timeout))
	output.WriteString(fmt.Sprintf("  Concurrency:  %d\n", concurrency))
	output.WriteString("\n")

	// Resolve host
	fmt.Printf("🔍 Resolving %s...\n", host)
	ips, err := net.LookupIP(host)
	if err != nil {
		output.WriteString(color.New(color.FgRed).Sprintf("❌ Failed to resolve host: %v\n", err))
		return &commands.Result{
			Output:   output.String(),
			Error:    err,
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, err
	}

	// Use first IPv4 address
	var targetIP string
	for _, ip := range ips {
		if ip.To4() != nil {
			targetIP = ip.String()
			break
		}
	}

	if targetIP == "" {
		output.WriteString(color.New(color.FgRed).Sprint("❌ No IPv4 address found\n"))
		return &commands.Result{
			Output:   output.String(),
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	output.WriteString(color.New(color.FgGreen).Sprintf("✅ Resolved to: %s\n\n", targetIP))

	// Start scanning
	fmt.Printf("🚀 Scanning %d ports...\n\n", len(portList))

	openPorts := p.scanPorts(ctx, targetIP, portList, timeout, concurrency)

	// Results
	output.WriteString(color.New(color.FgYellow, color.Bold).Sprint("📊 SCAN RESULTS\n"))
	output.WriteString("═══════════════════════════════════════════════════════════════\n")

	if len(openPorts) == 0 {
		output.WriteString(color.New(color.FgRed).Sprint("❌ No open ports found\n"))
	} else {
		output.WriteString(color.New(color.FgGreen, color.Bold).Sprintf("✅ Found %d open port(s):\n\n", len(openPorts)))

		// Sort ports
		sort.Ints(openPorts)

		for _, port := range openPorts {
			service := p.getServiceName(port)
			output.WriteString(fmt.Sprintf("  %s %d/tcp %s\n",
				color.New(color.FgGreen).Sprint("🟢"),
				port,
				color.New(color.FgCyan).Sprint(service)))
		}
	}

	output.WriteString("\n═══════════════════════════════════════════════════════════════\n")
	output.WriteString(color.New(color.FgHiBlack).Sprintf("Scan completed in %v\n",
		time.Since(startTime).Round(time.Millisecond)))

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// parsePortRange parses port range specification
func (p *PortscanCommand) parsePortRange(portSpec string) ([]int, error) {
	var ports []int

	parts := strings.Split(portSpec, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)

		if strings.Contains(part, "-") {
			// Range
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("invalid range format: %s", part)
			}

			start, err := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
			if err != nil {
				return nil, fmt.Errorf("invalid start port: %s", rangeParts[0])
			}

			end, err := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
			if err != nil {
				return nil, fmt.Errorf("invalid end port: %s", rangeParts[1])
			}

			if start > end {
				return nil, fmt.Errorf("start port greater than end port: %d > %d", start, end)
			}

			for i := start; i <= end; i++ {
				if i >= 1 && i <= 65535 {
					ports = append(ports, i)
				}
			}
		} else {
			// Single port
			port, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid port: %s", part)
			}

			if port >= 1 && port <= 65535 {
				ports = append(ports, port)
			}
		}
	}

	return ports, nil
}

// scanPorts performs concurrent port scanning with live feedback
func (p *PortscanCommand) scanPorts(ctx context.Context, host string, ports []int, timeout time.Duration, concurrency int) []int {
	var openPorts []int
	var mu sync.Mutex

	// Create semaphore for concurrency control
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	// Progress tracking
	completed := 0
	total := len(ports)

	for _, port := range ports {
		wg.Add(1)
		go func(port int) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			// Scan port
			if p.isPortOpen(ctx, host, port, timeout) {
				mu.Lock()
				openPorts = append(openPorts, port)
				fmt.Printf("🟢 Found open port: %d/tcp\n", port)
				mu.Unlock()
			}

			// Update progress
			mu.Lock()
			completed++
			if completed%50 == 0 || completed == total {
				fmt.Printf("📊 Progress: %d/%d ports scanned (%.1f%%)\n",
					completed, total, float64(completed)/float64(total)*100)
			}
			mu.Unlock()
		}(port)
	}

	wg.Wait()
	fmt.Println()

	return openPorts
}

// isPortOpen checks if a port is open
func (p *PortscanCommand) isPortOpen(ctx context.Context, host string, port int, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// getServiceName returns the common service name for a port
func (p *PortscanCommand) getServiceName(port int) string {
	services := map[int]string{
		21:   "ftp",
		22:   "ssh",
		23:   "telnet",
		25:   "smtp",
		53:   "dns",
		80:   "http",
		110:  "pop3",
		143:  "imap",
		443:  "https",
		993:  "imaps",
		995:  "pop3s",
		1433: "mssql",
		3306: "mysql",
		3389: "rdp",
		5432: "postgresql",
		6379: "redis",
		8080: "http-alt",
		8443: "https-alt",
	}

	if service, exists := services[port]; exists {
		return service
	}
	return "unknown"
}
