package system

import (
	"context"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"suppercommand/internal/commands"

	"github.com/fatih/color"
)

// KillTaskCommand terminates processes by PID or name
type KillTaskCommand struct {
	*commands.BaseCommand
}

// NewKillTaskCommand creates a new killtask command
func NewKillTaskCommand() *KillTaskCommand {
	return &KillTaskCommand{
		BaseCommand: commands.NewBaseCommand(
			"killtask",
			"Terminate processes by PID or process name",
			"killtask [-f] [-t] <pid|process_name> [pid2] [process_name2] ...",
			[]string{"windows", "linux", "darwin"},
			true, // May require elevation for some processes
		),
	}
}

// Execute terminates the specified processes
func (k *KillTaskCommand) Execute(ctx context.Context, args *commands.Arguments) (*commands.Result, error) {
	startTime := time.Now()

	if len(args.Raw) == 0 {
		return &commands.Result{
			Output:   "Usage: killtask [-f] [-t] <pid|process_name> [pid2] [process_name2] ...\n",
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	// Parse arguments
	force := false
	tree := false
	var targets []string

	for _, arg := range args.Raw {
		switch arg {
		case "-f", "--force":
			force = true
		case "-t", "--tree":
			tree = true
		default:
			if !strings.HasPrefix(arg, "-") {
				targets = append(targets, arg)
			}
		}
	}

	if len(targets) == 0 {
		return &commands.Result{
			Output:   "Error: No process ID or name specified\n",
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	var output strings.Builder
	output.WriteString(color.New(color.FgRed, color.Bold).Sprint("💀 PROCESS TERMINATION\n"))
	output.WriteString("═══════════════════════════════════════════════════════════════\n")

	if force {
		output.WriteString(color.New(color.FgYellow).Sprint("⚠️  Force mode enabled - processes will be terminated immediately\n"))
	}
	if tree {
		output.WriteString(color.New(color.FgBlue).Sprint("🌳 Tree mode enabled - child processes will also be terminated\n"))
	}
	output.WriteString("───────────────────────────────────────────────────────────────\n")

	successCount := 0
	errorCount := 0

	for _, target := range targets {
		result := k.killProcess(target, force, tree)
		output.WriteString(result.message)
		if result.success {
			successCount++
		} else {
			errorCount++
		}
	}

	// Summary
	output.WriteString("───────────────────────────────────────────────────────────────\n")
	output.WriteString(color.New(color.FgGreen, color.Bold).Sprintf("✅ Successfully terminated: %d process(es)\n", successCount))
	if errorCount > 0 {
		output.WriteString(color.New(color.FgRed, color.Bold).Sprintf("❌ Failed to terminate: %d process(es)\n", errorCount))
	}
	output.WriteString("═══════════════════════════════════════════════════════════════\n")

	exitCode := 0
	if errorCount > 0 {
		exitCode = 1
	}

	return &commands.Result{
		Output:   output.String(),
		ExitCode: exitCode,
		Duration: time.Since(startTime),
	}, nil
}

// KillResult represents the result of a kill operation
type KillResult struct {
	success bool
	message string
}

// killProcess kills a single process by PID or name
func (k *KillTaskCommand) killProcess(target string, force, tree bool) KillResult {
	// Check if target is a PID (numeric)
	if pid, err := strconv.Atoi(target); err == nil {
		return k.killByPID(pid, force, tree)
	}

	// Target is a process name
	return k.killByName(target, force, tree)
}

// killByPID kills a process by its PID
func (k *KillTaskCommand) killByPID(pid int, force, tree bool) KillResult {
	var cmd *exec.Cmd
	var cmdArgs []string

	switch runtime.GOOS {
	case "windows":
		cmdArgs = []string{"taskkill", "/PID", strconv.Itoa(pid)}
		if force {
			cmdArgs = append(cmdArgs, "/F")
		}
		if tree {
			cmdArgs = append(cmdArgs, "/T")
		}
		cmd = exec.Command(cmdArgs[0], cmdArgs[1:]...)
	case "linux", "darwin":
		signal := "TERM"
		if force {
			signal = "KILL"
		}
		cmd = exec.Command("kill", "-"+signal, strconv.Itoa(pid))
	default:
		return KillResult{
			success: false,
			message: color.New(color.FgRed).Sprintf("❌ PID %d: Unsupported operating system\n", pid),
		}
	}

	err := cmd.Run()
	if err != nil {
		return KillResult{
			success: false,
			message: color.New(color.FgRed).Sprintf("❌ PID %d: Failed to terminate (%v)\n", pid, err),
		}
	}

	return KillResult{
		success: true,
		message: color.New(color.FgGreen).Sprintf("✅ PID %d: Process terminated successfully\n", pid),
	}
}

// killByName kills processes by name
func (k *KillTaskCommand) killByName(name string, force, tree bool) KillResult {
	var cmd *exec.Cmd
	var cmdArgs []string

	switch runtime.GOOS {
	case "windows":
		cmdArgs = []string{"taskkill", "/IM", name}
		if force {
			cmdArgs = append(cmdArgs, "/F")
		}
		if tree {
			cmdArgs = append(cmdArgs, "/T")
		}
		cmd = exec.Command(cmdArgs[0], cmdArgs[1:]...)
	case "linux", "darwin":
		signal := "TERM"
		if force {
			signal = "KILL"
		}
		cmd = exec.Command("pkill", "-"+signal, name)
	default:
		return KillResult{
			success: false,
			message: color.New(color.FgRed).Sprintf("❌ %s: Unsupported operating system\n", name),
		}
	}

	err := cmd.Run()
	if err != nil {
		return KillResult{
			success: false,
			message: color.New(color.FgRed).Sprintf("❌ %s: Failed to terminate (%v)\n", name, err),
		}
	}

	return KillResult{
		success: true,
		message: color.New(color.FgGreen).Sprintf("✅ %s: Process(es) terminated successfully\n", name),
	}
}
