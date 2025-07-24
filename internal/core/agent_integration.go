package core

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"suppercommand/internal/agent"

	"github.com/fatih/color"
)

// EnhancedShell wraps the original Shell with Agent OS capabilities
type EnhancedShell struct {
	*Shell   // Embedded original shell
	agent    *agent.Agent
	enhanced bool
}

// NewEnhancedShell creates a new shell with Agent OS integration
func NewEnhancedShell() *EnhancedShell {
	originalShell := NewShell()
	agentOS := agent.NewAgent()

	return &EnhancedShell{
		Shell:    originalShell,
		agent:    agentOS,
		enhanced: true,
	}
}

// Initialize sets up the enhanced shell with Agent OS
func (es *EnhancedShell) Initialize() error {
	color.New(color.FgCyan, color.Bold).Println("🚀 SuperShell - Agent OS Edition")
	color.New(color.FgGreen).Println("   Next-generation PowerShell/Bash replacement")
	color.New(color.FgYellow).Println("   🌐 Advanced Networking • 🛡️ Security • ⚡ Performance")
	fmt.Println()

	// Initialize Agent OS
	if err := es.agent.Initialize(); err != nil {
		color.New(color.FgRed).Printf("❌ Failed to initialize Agent OS: %v\n", err)
		color.New(color.FgYellow).Println("🔄 Falling back to standard SuperShell mode...")
		es.enhanced = false
		return nil
	}

	// Bridge existing commands to Agent OS
	if err := es.bridgeCommands(); err != nil {
		color.New(color.FgRed).Printf("❌ Failed to bridge commands: %v\n", err)
		return err
	}

	// Show welcome message with available features
	es.showWelcomeMessage()

	return nil
}

// bridgeCommands converts existing SuperShell commands to Agent OS format
func (es *EnhancedShell) bridgeCommands() error {
	color.New(color.FgBlue).Println("🔗 Bridging legacy commands to Agent OS...")

	bridgedCount := 0
	for name, cmd := range commandRegistry {
		// Create bridge adapter for legacy commands
		bridgeCmd := &LegacyCommandBridge{
			name:        name,
			legacyCmd:   cmd,
			description: cmd.Description(),
		}

		// Register with Agent OS
		es.agent.RegisterCommand(name, bridgeCmd)
		bridgedCount++
	}

	color.New(color.FgGreen).Printf("✅ Bridged %d legacy commands successfully\n", bridgedCount)
	return nil
}

// ExecuteEnhanced runs commands through Agent OS if available
func (es *EnhancedShell) ExecuteEnhanced(input string) {
	input = strings.TrimSpace(input)
	if input == "" {
		return
	}

	// Handle exit commands
	if input == "exit" || input == "quit" {
		es.shutdown()
		return
	}

	// Parse command and arguments
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return
	}

	if es.enhanced {
		// Try multi-word commands first (for Agent OS commands like "dev profile")
		var result *agent.Result
		var err error
		var cmdName string
		var args []string

		// Try 2-word command first
		if len(parts) >= 2 {
			cmdName = strings.Join(parts[:2], " ")
			args = parts[2:]
			result, err = es.agent.ExecuteCommand(cmdName, args)

			// If 2-word command not found, try 1-word command
			if err != nil || (result != nil && strings.Contains(result.Output, "Command not found")) {
				cmdName = parts[0]
				args = parts[1:]
				result, err = es.agent.ExecuteCommand(cmdName, args)
			}
		} else {
			// Single word command
			cmdName = parts[0]
			args = parts[1:]
			result, err = es.agent.ExecuteCommand(cmdName, args)
		}

		if err != nil {
			color.New(color.FgRed).Printf("❌ Command execution error: %v\n", err)
			return
		}

		// Display results with enhanced formatting
		es.displayResult(result)
	} else {
		// Fallback to original execution
		output := Dispatch(input)
		if output != "" {
			fmt.Println(output)
		}
	}
}

// displayResult shows command results with enhanced formatting
func (es *EnhancedShell) displayResult(result *agent.Result) {
	if result == nil {
		return
	}

	// Color code based on result type
	switch result.Type {
	case agent.ResultTypeSuccess:
		// Standard output for successful commands
		if result.Output != "" {
			fmt.Print(result.Output)
		}
	case agent.ResultTypeError:
		color.New(color.FgRed).Print(result.Output)
	case agent.ResultTypeWarning:
		color.New(color.FgYellow).Print(result.Output)
	case agent.ResultTypeInfo:
		color.New(color.FgCyan).Print(result.Output)
	default:
		fmt.Print(result.Output)
	}

	// Show performance info if slow execution
	if result.Duration > 100*time.Millisecond {
		color.New(color.FgHiBlack).Printf("\n⏱️  Execution time: %v\n", result.Duration)
	}
}

// shutdown gracefully stops the enhanced shell
func (es *EnhancedShell) shutdown() {
	color.New(color.FgYellow).Println("🔄 Shutting down SuperShell...")

	if es.enhanced && es.agent != nil {
		if err := es.agent.Shutdown(); err != nil {
			color.New(color.FgRed).Printf("❌ Error during Agent OS shutdown: %v\n", err)
		}
	}

	color.New(color.FgGreen).Println("👋 Thank you for using SuperShell!")
	fmt.Println("   💡 Visit github.com/your-repo/suppercommand for updates")
}

// showWelcomeMessage displays the enhanced welcome screen
func (es *EnhancedShell) showWelcomeMessage() {
	fmt.Println()
	color.New(color.FgCyan, color.Bold).Println("🎯 ENHANCED FEATURES AVAILABLE")
	color.New(color.FgHiBlack).Println("────────────────────────────────────────────────────────────────")

	features := []struct {
		icon string
		name string
		desc string
	}{
		{"🔥", "Hot Reload", "dev reload - Live command updates"},
		{"📊", "Performance", "perf stats - Real-time monitoring"},
		{"🧪", "Testing", "dev test <cmd> - Interactive testing"},
		{"📚", "Documentation", "dev docs - Auto-generated help"},
		{"🔧", "Build Tools", "dev build - Cross-platform builds"},
		{"⚡", "Optimization", "perf optimize - Auto performance tuning"},
	}

	for _, feature := range features {
		color.New(color.FgWhite).Printf("  %s %-15s %s\n",
			feature.icon, feature.name,
			color.New(color.FgHiBlack).Sprint(feature.desc))
	}

	fmt.Println()
	color.New(color.FgGreen).Println("💡 Type 'help' for all commands or 'dev' for development tools")
	color.New(color.FgHiBlack).Println("────────────────────────────────────────────────────────────────")
	fmt.Println()
}

// LegacyCommandBridge adapts old commands to new Agent OS interface
type LegacyCommandBridge struct {
	name        string
	legacyCmd   Command
	description string
}

func (lcb *LegacyCommandBridge) Name() string        { return lcb.name }
func (lcb *LegacyCommandBridge) Description() string { return lcb.description }
func (lcb *LegacyCommandBridge) Category() string {
	// Categorize legacy commands
	switch {
	case strings.Contains(lcb.name, "net") || lcb.name == "ping" || lcb.name == "portscan":
		return "networking"
	case lcb.name == "ls" || lcb.name == "cd" || lcb.name == "pwd":
		return "filesystem"
	case lcb.name == "help" || lcb.name == "clear" || lcb.name == "echo":
		return "core"
	default:
		return "system"
	}
}

func (lcb *LegacyCommandBridge) Examples() []string {
	// Generate basic examples
	return []string{lcb.name}
}

func (lcb *LegacyCommandBridge) ValidateArgs(args []string) error {
	// Legacy commands don't have built-in validation
	return nil
}

func (lcb *LegacyCommandBridge) Execute(ctx context.Context, args []string) (*agent.Result, error) {
	start := time.Now()

	// Execute legacy command
	output := lcb.legacyCmd.Execute(args)

	// Determine result type based on output
	resultType := agent.ResultTypeSuccess
	exitCode := 0

	if strings.Contains(strings.ToLower(output), "error") {
		resultType = agent.ResultTypeError
		exitCode = 1
	} else if strings.Contains(strings.ToLower(output), "warning") {
		resultType = agent.ResultTypeWarning
	}

	return &agent.Result{
		Output:   output,
		ExitCode: exitCode,
		Duration: time.Since(start),
		Type:     resultType,
		Metadata: map[string]any{
			"legacy_command": true,
			"bridge_version": "1.0.0",
		},
	}, nil
}

// GetAgent returns the underlying Agent OS instance
func (es *EnhancedShell) GetAgent() *agent.Agent {
	return es.agent
}

// IsEnhanced returns whether Agent OS features are active
func (es *EnhancedShell) IsEnhanced() bool {
	return es.enhanced
}

// Run starts the enhanced shell with Windows compatibility
func (es *EnhancedShell) Run() {
	if es.enhanced {
		// Use Agent OS enhanced execution for better Windows compatibility
		es.runEnhanced()
	} else {
		// Fallback to original shell
		es.Shell.Run()
	}
}

// runEnhanced provides a Windows-compatible shell loop
func (es *EnhancedShell) runEnhanced() {
	reader := bufio.NewReader(os.Stdin)

	for {
		// Get current directory for prompt
		cwd, _ := os.Getwd()
		prompt := strings.ReplaceAll(cwd, "/", "\\") + "> "

		// Print prompt
		fmt.Print(prompt)

		// Read input line
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			continue
		}

		// Process command
		es.ExecuteEnhanced(strings.TrimSpace(input))
	}
}

// RegisterAgentPlugin allows external plugins to be registered
func (es *EnhancedShell) RegisterAgentPlugin(plugin agent.Plugin) error {
	if !es.enhanced {
		return fmt.Errorf("Agent OS not available")
	}
	return es.agent.RegisterPlugin(plugin)
}

// Add this method to EnhancedShell (replace the existing RunIntelligent method)
func (es *EnhancedShell) RunIntelligent() {
	// Create intelligent shell wrapper
	intelligentShell := NewIntelligentShell()

	// Initialize intelligence
	if err := intelligentShell.Initialize(); err != nil {
		color.New(color.FgRed).Printf("❌ Failed to initialize intelligence: %v\n", err)
		// Fallback to regular enhanced shell
		es.runEnhanced()
		return
	}

	// Run with intelligence (using go-prompt)
	intelligentShell.RunWithIntelligence()
}
