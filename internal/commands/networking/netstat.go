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

// NetstatCommand shows network connections and statistics
type NetstatCommand struct {
	*commands.BaseCommand
}

// NewNetstatCommand creates a new netstat command
func NewNetstatCommand() *NetstatCommand {
	return &NetstatCommand{
		BaseCommand: commands.NewBaseCommand(
			"netstat",
			"Display network connections, routing tables, and network statistics",
			"netstat [-a] [-n] [-p] [-r] [-s]",
			[]string{"windows", "linux", "darwin"},
			false,
		),
	}
}

// Execute shows network information with enhanced formatting
func (n *NetstatCommand) Execute(ctx context.Context, args *commands.Arguments) (*commands.Result, error) {
	startTime := time.Now()

	// Parse flags
	showAll := false
	showNumeric := false
	showProcesses := false
	showRouting := false
	showStatistics := false

	for _, arg := range args.Raw {
		switch arg {
		case "-a", "--all":
			showAll = true
		case "-n", "--numeric":
			showNumeric = true
		case "-p", "--processes":
			showProcesses = true
		case "-r", "--route":
			showRouting = true
		case "-s", "--statistics":
			showStatistics = true
		}
	}

	var output strings.Builder

	// Header
	output.WriteString(color.New(color.FgCyan, color.Bold).Sprint("🌐 NETWORK STATUS\n"))
	output.WriteString("═══════════════════════════════════════════════════════════════\n\n")

	// Show progress
	fmt.Print("📊 Gathering network information")
	done := make(chan bool)
	go n.showProgress(done)

	// Build netstat command based on OS and flags
	var cmd *exec.Cmd
	var cmdArgs []string

	if runtime.GOOS == "windows" {
		cmdArgs = []string{}
		if showAll {
			cmdArgs = append(cmdArgs, "-a")
		}
		if showNumeric {
			cmdArgs = append(cmdArgs, "-n")
		}
		if showProcesses {
			cmdArgs = append(cmdArgs, "-o")
		}
		if showRouting {
			cmdArgs = append(cmdArgs, "-r")
		}
		if showStatistics {
			cmdArgs = append(cmdArgs, "-s")
		}
		if len(cmdArgs) == 0 {
			cmdArgs = []string{"-an"}
		}
	} else {
		cmdArgs = []string{}
		if showAll {
			cmdArgs = append(cmdArgs, "-a")
		}
		if showNumeric {
			cmdArgs = append(cmdArgs, "-n")
		}
		if showProcesses {
			cmdArgs = append(cmdArgs, "-p")
		}
		if showRouting {
			cmdArgs = append(cmdArgs, "-r")
		}
		if showStatistics {
			cmdArgs = append(cmdArgs, "-s")
		}
		if len(cmdArgs) == 0 {
			cmdArgs = []string{"-an"}
		}
	}

	cmd = exec.CommandContext(ctx, "netstat", cmdArgs...)

	// Execute command
	cmdOutput, err := cmd.Output()
	done <- true
	fmt.Print("\r\033[K") // Clear progress line

	if err != nil {
		output.WriteString(color.New(color.FgRed, color.Bold).Sprintf("❌ Failed to execute netstat: %v\n", err))
		return &commands.Result{
			Output:   output.String(),
			Error:    err,
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, err
	}

	// Parse and format output
	scanner := bufio.NewScanner(strings.NewReader(string(cmdOutput)))

	// Count different connection types
	connectionCounts := make(map[string]int)
	lineCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineCount++

		// Format the line with colors
		formattedLine := n.formatNetstatLine(line)
		if formattedLine != "" {
			output.WriteString(formattedLine + "\n")
		}

		// Count connection states
		if strings.Contains(line, "ESTABLISHED") {
			connectionCounts["ESTABLISHED"]++
		} else if strings.Contains(line, "LISTENING") || strings.Contains(line, "LISTEN") {
			connectionCounts["LISTENING"]++
		} else if strings.Contains(line, "TIME_WAIT") {
			connectionCounts["TIME_WAIT"]++
		} else if strings.Contains(line, "CLOSE_WAIT") {
			connectionCounts["CLOSE_WAIT"]++
		}
	}

	// Summary
	if len(connectionCounts) > 0 {
		output.WriteString("\n" + color.New(color.FgYellow, color.Bold).Sprint("📊 CONNECTION SUMMARY\n"))
		output.WriteString("═══════════════════════════════════════════════════════════════\n")

		for state, count := range connectionCounts {
			var stateColor *color.Color
			switch state {
			case "ESTABLISHED":
				stateColor = color.New(color.FgGreen, color.Bold)
			case "LISTENING":
				stateColor = color.New(color.FgBlue, color.Bold)
			case "TIME_WAIT":
				stateColor = color.New(color.FgYellow)
			case "CLOSE_WAIT":
				stateColor = color.New(color.FgRed)
			default:
				stateColor = color.New(color.FgWhite)
			}

			output.WriteString(fmt.Sprintf("  %-12s %s\n",
				state+":",
				stateColor.Sprintf("%d connections", count)))
		}
	}

	output.WriteString("\n═══════════════════════════════════════════════════════════════\n")
	output.WriteString(color.New(color.FgHiBlack).Sprintf("Completed in %v (%d lines processed)\n",
		time.Since(startTime).Round(time.Millisecond), lineCount))

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// showProgress shows a spinner during netstat execution
func (n *NetstatCommand) showProgress(done chan bool) {
	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	i := 0

	for {
		select {
		case <-done:
			return
		default:
			fmt.Printf("\r📊 Gathering network information %s",
				color.New(color.FgYellow).Sprint(spinner[i%len(spinner)]))
			time.Sleep(100 * time.Millisecond)
			i++
		}
	}
}

// formatNetstatLine formats a netstat output line with colors
func (n *NetstatCommand) formatNetstatLine(line string) string {
	line = strings.TrimSpace(line)
	if line == "" {
		return ""
	}

	lower := strings.ToLower(line)

	// Skip header lines but format them
	if strings.Contains(lower, "proto") && strings.Contains(lower, "local address") {
		return color.New(color.FgCyan, color.Bold).Sprint(line)
	}

	// Skip separator lines
	if strings.Contains(line, "---") || strings.Contains(line, "===") {
		return ""
	}

	// Color code based on connection state
	switch {
	case strings.Contains(lower, "established"):
		return color.New(color.FgGreen).Sprint("🟢 ") + line
	case strings.Contains(lower, "listening") || strings.Contains(lower, "listen"):
		return color.New(color.FgBlue).Sprint("🔵 ") + line
	case strings.Contains(lower, "time_wait"):
		return color.New(color.FgYellow).Sprint("🟡 ") + line
	case strings.Contains(lower, "close_wait"):
		return color.New(color.FgRed).Sprint("🔴 ") + line
	case strings.Contains(lower, "syn_sent") || strings.Contains(lower, "syn_recv"):
		return color.New(color.FgMagenta).Sprint("🟣 ") + line
	default:
		// Check if it's a data line (contains port numbers)
		if strings.Contains(line, ":") && (strings.Contains(line, "TCP") || strings.Contains(line, "UDP")) {
			return "⚪ " + line
		}
		return line
	}
}
