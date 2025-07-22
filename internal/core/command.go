package core

import (
	"bufio"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"
)

type Command interface {
	Name() string
	Execute(args []string) string
	Description() string // Add a description method for help output
}

var commandRegistry = make(map[string]Command)

func Register(cmd Command) {
	commandRegistry[cmd.Name()] = cmd
}

func Dispatch(input string, depth ...int) string {
	d := 0
	if len(depth) > 0 {
		d = depth[0]
	}
	if d > 10 { // Changed from maxAliasDepth to 10
		return "Alias expansion too deep (possible recursion)."
	}

	fmt.Printf("DEBUG: Dispatch called with input=%q, depth=%d\n", input, d)

	parts := strings.Fields(input)
	if len(parts) == 0 {
		return ""
	}
	cmdName := parts[0]

	cmd, ok := commandRegistry[cmdName]
	if !ok {
		return "Unknown command: " + cmdName
	}
	return cmd.Execute(parts[1:])
}

type PingCommand struct{}

func (p *PingCommand) Name() string        { return "ping" }
func (p *PingCommand) Description() string { return "Ping a host to test connectivity" }
func (p *PingCommand) Execute(args []string) string {
	if len(args) == 0 {
		return "Usage: ping <host> [count]"
	}
	host := args[0]
	count := "4"
	if len(args) > 1 {
		count = args[1]
	}
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("ping", host, "-n", count)
	} else {
		cmd = exec.Command("ping", host, "-c", count)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "Failed to start ping: " + err.Error()
	}
	if err := cmd.Start(); err != nil {
		return "Failed to start ping: " + err.Error()
	}
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		lower := strings.ToLower(line)
		switch {
		case strings.Contains(lower, "reply from") || strings.Contains(lower, "bytes from"):
			color.New(color.FgGreen).Println(line)
		case strings.Contains(lower, "request timed out") || strings.Contains(lower, "unreachable") || strings.Contains(lower, "timed out"):
			color.New(color.FgRed).Println(line)
		case strings.Contains(lower, "packets:") || strings.Contains(lower, "statistics") || strings.Contains(lower, "round trip"):
			color.New(color.FgCyan).Println(line)
		default:
			fmt.Println(line)
		}
	}
	cmd.Wait()
	return ""
}

type NslookupCommand struct{}

func (n *NslookupCommand) Name() string        { return "nslookup" }
func (n *NslookupCommand) Description() string { return "Query DNS records for a domain" }
func (n *NslookupCommand) Execute(args []string) string {
	if len(args) == 0 {
		return "Usage: nslookup <domain>"
	}
	domain := args[0]
	out, err := exec.Command("nslookup", domain).CombinedOutput()
	if err != nil {
		return fmt.Sprintf("nslookup failed: %v\n%s", err, string(out))
	}
	return string(out)
}

type TracertCommand struct{}

func (t *TracertCommand) Name() string        { return "tracert" }
func (t *TracertCommand) Description() string { return "Trace the route to a host" }
func (t *TracertCommand) Execute(args []string) string {
	if len(args) == 0 {
		return "Usage: tracert <host>"
	}
	host := args[0]
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("tracert", host)
	} else {
		cmd = exec.Command("traceroute", host)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "Failed to start tracert: " + err.Error()
	}

	// Spinner for live feedback
	done := make(chan struct{})
	go func() {
		spinner := []string{"|", "/", "-", "\\"}
		i := 0
		for {
			select {
			case <-done:
				fmt.Print("\r") // Clear spinner line
				return
			default:
				fmt.Printf("\rTracing route, please wait... %s", spinner[i%len(spinner)])
				time.Sleep(200 * time.Millisecond)
				i++
			}
		}
	}()

	if err := cmd.Start(); err != nil {
		close(done)
		fmt.Println()
		return "Failed to start tracert: " + err.Error()
	}
	scanner := bufio.NewScanner(stdout)
	firstLine := true
	for scanner.Scan() {
		if firstLine {
			close(done)
			fmt.Print("\r") // Clear spinner line
			firstLine = false
		}
		line := scanner.Text()
		lower := strings.ToLower(line)
		switch {
		case strings.HasPrefix(lower, "tracing route") || strings.HasPrefix(lower, "over a maximum") || strings.HasPrefix(lower, "trace complete"):
			color.New(color.FgCyan, color.Bold).Println(line)
		case strings.Contains(line, "*"):
			color.New(color.FgRed).Println(line)
		case strings.Contains(line, "ms") && strings.Contains(line, "."):
			color.New(color.FgGreen).Println(line)
		case strings.Contains(line, "[") && strings.Contains(line, "]"):
			color.New(color.FgMagenta).Println(line)
		case strings.Contains(line, "timed out") || strings.Contains(line, "unreachable"):
			color.New(color.FgRed, color.Bold).Println(line)
		default:
			fmt.Println(line)
		}
	}
	cmd.Wait()
	return ""
}
