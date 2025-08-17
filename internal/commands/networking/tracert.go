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

// TracertCommand traces the route to a host with live feedback
type TracertCommand struct {
	*commands.BaseCommand
}

// NewTracertCommand creates a new tracert command
func NewTracertCommand() *TracertCommand {
	return &TracertCommand{
		BaseCommand: commands.NewBaseCommand(
			"tracert",
			"Trace the route to a host with live feedback",
			"tracert [-h maxhops] <host>",
			[]string{"windows", "linux", "darwin"},
			false,
		),
	}
}

// Execute traces the route to a host with live feedback
func (t *TracertCommand) Execute(ctx context.Context, args *commands.Arguments) (*commands.Result, error) {
	startTime := time.Now()

	if len(args.Raw) == 0 {
		return &commands.Result{
			Output:   "Usage: tracert [-h maxhops] <host>\n",
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	// Parse arguments
	host := ""
	maxHops := "30"

	for i, arg := range args.Raw {
		switch arg {
		case "-h":
			if i+1 < len(args.Raw) {
				maxHops = args.Raw[i+1]
			}
		default:
			if !strings.HasPrefix(arg, "-") && host == "" {
				// Skip if it's a value for a flag
				if i > 0 && args.Raw[i-1] == "-h" {
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

	// Show initial message with spinner
	fmt.Printf("🛣️  TRACING ROUTE to %s with maximum %s hops:\n",
		color.New(color.FgCyan, color.Bold).Sprint(host), maxHops)
	fmt.Println()

	// Create context with timeout to prevent hanging
	traceCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// Build tracert command based on OS
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(traceCtx, "tracert", "-h", maxHops, host)
	} else {
		cmd = exec.CommandContext(traceCtx, "traceroute", "-m", maxHops, host)
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
			Output:   fmt.Sprintf("Failed to start tracert: %v\n", err),
			Error:    err,
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, err
	}

	// Progress tracking
	var outputLines []string
	hopCount := 0
	showingSpinner := false
	var spinnerDone chan bool

	// Read output line by line for live feedback
	scanner := bufio.NewScanner(stdout)

	for scanner.Scan() {
		line := scanner.Text()
		outputLines = append(outputLines, line)

		// Stop any active spinner
		if showingSpinner {
			close(spinnerDone)
			showingSpinner = false
			fmt.Print("\r\033[K") // Clear spinner line
		}

		// Color code the output based on content
		needsSpinner := t.printColoredTracertLine(line, &hopCount)

		// Start spinner for next hop if needed
		if needsSpinner {
			spinnerDone = make(chan bool)
			showingSpinner = true
			go t.showProgress(spinnerDone)
		}
	}

	// Ensure spinner is stopped
	if showingSpinner {
		close(spinnerDone)
		fmt.Print("\r\033[K") // Clear spinner line
	}

	// Wait for command to complete with proper error handling
	err = cmd.Wait()

	// Handle different exit scenarios
	var exitCode int
	var summary string

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			summary = color.New(color.FgYellow, color.Bold).Sprint("⏰ Tracert timed out after 5 minutes")
			exitCode = 1
		} else if ctx.Err() == context.Canceled {
			summary = color.New(color.FgYellow, color.Bold).Sprint("🛑 Tracert cancelled by user")
			exitCode = 1
		} else {
			summary = color.New(color.FgRed, color.Bold).Sprintf("❌ Tracert failed: %v", err)
			exitCode = 1
		}
	} else {
		summary = color.New(color.FgGreen, color.Bold).Sprintf("✅ Trace completed with %d hops", hopCount)
		exitCode = 0
	}

	// Print summary
	fmt.Println()
	fmt.Println(summary)

	return &commands.Result{
		Output:   strings.Join(outputLines, "\n") + "\n",
		ExitCode: exitCode,
		Duration: time.Since(startTime),
	}, nil
}

// showProgress shows a spinner while tracing
func (t *TracertCommand) showProgress(done chan bool) {
	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	i := 0
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			fmt.Printf("\r%s Tracing...", color.New(color.FgYellow).Sprint(spinner[i%len(spinner)]))
			i++
		}
	}
}

// printColoredTracertLine prints a tracert output line with appropriate colors
func (t *TracertCommand) printColoredTracertLine(line string, hopCount *int) bool {
	lower := strings.ToLower(line)

	// Clear any existing spinner
	fmt.Print("\r\033[K")

	switch {
	case strings.HasPrefix(strings.TrimSpace(line), "Tracing route") ||
		strings.Contains(lower, "traceroute to"):
		// Header information - cyan
		color.New(color.FgCyan, color.Bold).Printf("🎯 %s\n", line)
		return false

	case strings.Contains(line, "ms") && (strings.Contains(line, "*") ||
		strings.Contains(line, ".")):
		// Hop with timing - parse and colorize
		*hopCount++
		parts := strings.Fields(line)
		if len(parts) > 0 {
			// Hop number
			fmt.Printf("%s ", color.New(color.FgWhite, color.Bold).Sprintf("%2s", parts[0]))

			// Process the rest
			for i := 1; i < len(parts); i++ {
				part := parts[i]
				if strings.Contains(part, "ms") {
					// Timing - color based on speed
					if strings.Contains(part, "*") {
						color.New(color.FgRed).Print("    * ")
					} else {
						// Extract timing and color accordingly
						color.New(color.FgGreen).Printf("%8s ", part)
					}
				} else if strings.Contains(part, ".") && len(strings.Split(part, ".")) == 4 {
					// IP address
					color.New(color.FgCyan).Printf("%s ", part)
				} else if part != "" && !strings.Contains(part, "[") {
					// Hostname
					color.New(color.FgYellow).Printf("%s ", part)
				} else {
					fmt.Printf("%s ", part)
				}
			}
			fmt.Println()
		}
		return true

	case strings.Contains(lower, "request timed out") || strings.Contains(lower, "* * *"):
		// Timeout - red
		*hopCount++
		color.New(color.FgRed, color.Bold).Printf("%2d ❌ Request timed out\n", *hopCount)
		return true

	case strings.Contains(lower, "trace complete") || strings.Contains(lower, "reached"):
		// Completion - green
		color.New(color.FgGreen, color.Bold).Printf("🏁 %s\n", line)
		return false

	case strings.Contains(lower, "unable to resolve") || strings.Contains(lower, "unknown host"):
		// Error - red
		color.New(color.FgRed, color.Bold).Printf("❌ %s\n", line)
		return false

	default:
		// Default output
		if strings.TrimSpace(line) != "" {
			fmt.Printf("   %s\n", line)
		}
		return false
	}
}
