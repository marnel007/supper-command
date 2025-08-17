package networking

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"suppercommand/internal/commands"

	"github.com/fatih/color"
)

// PingCommand pings a host with live feedback
type PingCommand struct {
	*commands.BaseCommand
}

// NewPingCommand creates a new ping command
func NewPingCommand() *PingCommand {
	return &PingCommand{
		BaseCommand: commands.NewBaseCommand(
			"ping",
			"Send ICMP echo requests to a host with live feedback",
			"ping [-c count] [-t timeout] <host>",
			[]string{"windows", "linux", "darwin"},
			false,
		),
	}
}

// Execute pings a host with live feedback
func (p *PingCommand) Execute(ctx context.Context, args *commands.Arguments) (*commands.Result, error) {
	startTime := time.Now()

	if len(args.Raw) == 0 {
		return &commands.Result{
			Output:   "Usage: ping [-c count] [-t timeout] <host>\n",
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	// Parse arguments
	host := ""
	count := "4"
	timeout := "4000" // milliseconds

	for i, arg := range args.Raw {
		switch arg {
		case "-c":
			if i+1 < len(args.Raw) {
				count = args.Raw[i+1]
			}
		case "-t":
			if i+1 < len(args.Raw) {
				timeout = args.Raw[i+1]
			}
		default:
			if !strings.HasPrefix(arg, "-") && host == "" {
				// Skip if it's a value for a flag
				if i > 0 && (args.Raw[i-1] == "-c" || args.Raw[i-1] == "-t") {
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

	// Show initial message
	fmt.Printf("🌐 PING %s with %s packets of data:\n",
		color.New(color.FgCyan).Sprint(host), count)
	fmt.Println()

	// Create context with timeout to prevent hanging
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	// Build ping command based on OS
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(pingCtx, "ping", host, "-n", count, "-w", timeout)
	} else {
		cmd = exec.CommandContext(pingCtx, "ping", host, "-c", count, "-W", timeout)
	}

	// Get stdout pipe for live output
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return &commands.Result{
			Output:   fmt.Sprintf("Failed to create stdout pipe: %v\n", err),
			Error:    err,
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, err
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return &commands.Result{
			Output:   fmt.Sprintf("Failed to start ping: %v\n", err),
			Error:    err,
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, err
	}

	// Read output line by line for live feedback
	scanner := bufio.NewScanner(stdout)
	var outputLines []string

	for scanner.Scan() {
		line := scanner.Text()
		outputLines = append(outputLines, line)

		// Color code the output based on content
		p.printColoredPingLine(line)
	}

	// Wait for command to complete
	err = cmd.Wait()

	// Handle different exit scenarios
	var summary string
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			summary = color.New(color.FgYellow, color.Bold).Sprint("⏰ Ping timed out")
		} else if ctx.Err() == context.Canceled {
			summary = color.New(color.FgYellow, color.Bold).Sprint("🛑 Ping cancelled by user")
		} else {
			summary = color.New(color.FgRed, color.Bold).Sprintf("❌ Ping failed: %v", err)
		}
	} else {
		summary = color.New(color.FgGreen, color.Bold).Sprint("✅ Ping completed successfully")
	}

	// Print summary
	fmt.Println()
	fmt.Println(summary)

	return &commands.Result{
		Output:   strings.Join(outputLines, "\n") + "\n",
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// printColoredPingLine prints a ping output line with appropriate colors
func (p *PingCommand) printColoredPingLine(line string) {
	lower := strings.ToLower(line)

	switch {
	case strings.Contains(lower, "reply from") || strings.Contains(lower, "bytes from"):
		// Successful ping - green
		parts := strings.Fields(line)
		fmt.Print("📡 ")
		for i, part := range parts {
			if strings.Contains(part, "time=") || strings.Contains(part, "time<") {
				color.New(color.FgGreen, color.Bold).Print(part)
			} else if strings.Contains(part, "bytes") {
				color.New(color.FgCyan).Print(part)
			} else if i == 0 {
				color.New(color.FgWhite).Print(part)
			} else {
				fmt.Print(part)
			}
			if i < len(parts)-1 {
				fmt.Print(" ")
			}
		}
		fmt.Println()

	case strings.Contains(lower, "request timed out") || strings.Contains(lower, "unreachable") ||
		strings.Contains(lower, "timed out") || strings.Contains(lower, "no route"):
		// Failed ping - red
		color.New(color.FgRed, color.Bold).Printf("❌ %s\n", line)

	case strings.Contains(lower, "packets:") || strings.Contains(lower, "statistics") ||
		strings.Contains(lower, "round trip") || strings.Contains(lower, "min/avg/max"):
		// Statistics - cyan
		color.New(color.FgCyan, color.Bold).Printf("📊 %s\n", line)

	case strings.Contains(lower, "pinging") || strings.Contains(lower, "ping statistics"):
		// Header information - yellow
		color.New(color.FgYellow, color.Bold).Printf("🎯 %s\n", line)

	default:
		// Default output
		fmt.Printf("   %s\n", line)
	}
}
