package shell

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"suppercommand/internal/commands"
	"suppercommand/internal/config"
	"suppercommand/internal/monitoring"
	"suppercommand/pkg/errors"

	prompt "github.com/c-bata/go-prompt"
)

// Shell interface defines the core shell functionality
type Shell interface {
	Initialize(ctx context.Context) error
	Run(ctx context.Context) error
	ExecuteCommand(ctx context.Context, input string) (*ExecutionResult, error)
	Shutdown(ctx context.Context) error
}

// ExecutionResult contains the result of command execution
type ExecutionResult struct {
	Output     string
	Error      error
	ExitCode   int
	Duration   time.Duration
	MemoryUsed int64
	Warnings   []string
}

// ExecutionContext contains context information for command execution
type ExecutionContext struct {
	Command     string
	Arguments   []string
	Environment map[string]string
	WorkingDir  string
	User        *UserInfo
	Privileges  *PrivilegeInfo
	Timeout     time.Duration
}

// UserInfo contains information about the current user
type UserInfo struct {
	Username string
	UID      string
	GID      string
	HomeDir  string
}

// PrivilegeInfo contains privilege information
type PrivilegeInfo struct {
	IsElevated    bool
	CanElevate    bool
	RequiredLevel PrivilegeLevel
	Platform      string
	Capabilities  []string
}

// PrivilegeLevel represents different privilege levels
type PrivilegeLevel int

const (
	PrivilegeLevelUser PrivilegeLevel = iota
	PrivilegeLevelAdmin
	PrivilegeLevelRoot
)

// Arguments contains parsed command arguments
type Arguments struct {
	Raw     []string
	Parsed  map[string]interface{}
	Flags   map[string]bool
	Options map[string]string
}

// BasicShell implements the Shell interface
type BasicShell struct {
	config    config.ShellConfig
	registry  *commands.Registry
	monitor   monitoring.Monitor
	logger    monitoring.Logger
	executor  *Executor
	completer *Completer
	prompter  *Prompter
}

// NewShell creates a new shell instance
func NewShell(
	config config.ShellConfig,
	registry *commands.Registry,
	monitor monitoring.Monitor,
	logger monitoring.Logger,
) Shell {
	return &BasicShell{
		config:   config,
		registry: registry,
		monitor:  monitor,
		logger:   logger,
	}
}

// Initialize initializes the shell components
func (s *BasicShell) Initialize(ctx context.Context) error {
	s.logger.Info("Initializing shell components")

	// Initialize executor
	s.executor = NewExecutor(s.registry, s.monitor, s.logger)
	if err := s.executor.Initialize(ctx); err != nil {
		return errors.Wrap(err, "failed to initialize executor")
	}

	// Initialize completer
	s.completer = NewCompleter(s.registry, s.logger)
	if err := s.completer.Initialize(ctx); err != nil {
		return errors.Wrap(err, "failed to initialize completer")
	}

	// Initialize prompter
	s.prompter = NewPrompter(s.config, s.logger)
	if err := s.prompter.Initialize(ctx); err != nil {
		return errors.Wrap(err, "failed to initialize prompter")
	}

	s.logger.Info("Shell components initialized successfully")
	return nil
}

// Run starts the shell main loop
func (s *BasicShell) Run(ctx context.Context) error {
	s.logger.Info("Starting shell main loop")

	// Use go-prompt by default, simple shell can be enabled with SUPERSHELL_SIMPLE=1
	// Use stable terminal mode with SUPERSHELL_STABLE=1 for better resize handling
	useSimpleShell := os.Getenv("SUPERSHELL_SIMPLE") == "1"
	useStableMode := os.Getenv("SUPERSHELL_STABLE") == "1"

	if useSimpleShell {
		return s.runSimpleShell(ctx)
	}

	if useStableMode {
		return s.runStableShell(ctx)
	}

	// Create the interactive shell using go-prompt with better terminal handling
	p := prompt.New(
		s.promptExecutor,
		s.promptCompleter,
		prompt.OptionLivePrefix(s.prompter.GetLivePrefix),
		prompt.OptionTitle("SuperShell"),
		prompt.OptionInputTextColor(prompt.White),
		prompt.OptionSuggestionBGColor(prompt.Black),
		prompt.OptionSuggestionTextColor(prompt.White),
		prompt.OptionSelectedSuggestionBGColor(prompt.Blue),
		prompt.OptionSelectedSuggestionTextColor(prompt.White),
		prompt.OptionPreviewSuggestionTextColor(prompt.Yellow),
		// Fix terminal display issues
		prompt.OptionMaxSuggestion(6), // Reduce to prevent skewing
		prompt.OptionShowCompletionAtStart(),
		prompt.OptionCompletionWordSeparator(" "),
	)

	// Run the prompt in a goroutine so we can handle context cancellation
	done := make(chan struct{})
	go func() {
		defer close(done)
		p.Run()
	}()

	// Wait for context cancellation or prompt completion
	select {
	case <-ctx.Done():
		s.logger.Info("Shell context cancelled, shutting down")
		return ctx.Err()
	case <-done:
		s.logger.Info("Shell prompt completed")
		return nil
	}
}

// promptExecutor adapts our executor to go-prompt's expected signature
func (s *BasicShell) promptExecutor(input string) {
	input = strings.TrimSpace(input)
	if input == "" {
		return
	}

	// Handle exit commands
	if input == "exit" || input == "quit" {
		fmt.Println("👋 Goodbye!")
		os.Exit(0)
	}

	// Execute command
	ctx := context.Background()
	result, err := s.executor.Execute(ctx, input)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	// Display result with proper newline handling
	if result.Output != "" {
		fmt.Print(result.Output)
		// Ensure we end with a newline
		if !strings.HasSuffix(result.Output, "\n") {
			fmt.Println()
		}
	}
}

// promptCompleter adapts our completer to go-prompt's expected signature
func (s *BasicShell) promptCompleter(d prompt.Document) []prompt.Suggest {
	completions := s.completer.GetCompletions(d.Text, d.CursorPositionCol())

	var suggestions []prompt.Suggest
	for _, comp := range completions {
		suggestions = append(suggestions, prompt.Suggest{
			Text:        comp.Text,
			Description: comp.Description,
		})
	}

	return suggestions
}

// ExecuteCommand executes a single command
func (s *BasicShell) ExecuteCommand(ctx context.Context, input string) (*ExecutionResult, error) {
	return s.executor.Execute(ctx, input)
}

// Shutdown gracefully shuts down the shell
func (s *BasicShell) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down shell")

	if s.executor != nil {
		if err := s.executor.Shutdown(ctx); err != nil {
			s.logger.Error("Failed to shutdown executor", err)
		}
	}

	if s.completer != nil {
		if err := s.completer.Shutdown(ctx); err != nil {
			s.logger.Error("Failed to shutdown completer", err)
		}
	}

	if s.prompter != nil {
		if err := s.prompter.Shutdown(ctx); err != nil {
			s.logger.Error("Failed to shutdown prompter", err)
		}
	}

	s.logger.Info("Shell shutdown complete")
	return nil
}

// runSimpleShell runs a simple shell without go-prompt for better terminal compatibility
func (s *BasicShell) runSimpleShell(ctx context.Context) error {
	fmt.Println("SuperShell v0.03 - Smart Command Line Interface")
	fmt.Println("Type 'help' for available commands or 'exit' to quit.")
	fmt.Println()

	// Create a channel to handle graceful shutdown
	done := make(chan bool)

	// Handle context cancellation
	go func() {
		<-ctx.Done()
		done <- true
	}()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		select {
		case <-done:
			return ctx.Err()
		default:
			// Display prompt
			fmt.Print(s.prompter.GetPrompt())

			// Read input with proper error handling
			if !scanner.Scan() {
				if err := scanner.Err(); err != nil {
					fmt.Printf("❌ Input error: %v\n", err)
				}
				break
			}

			input := strings.TrimSpace(scanner.Text())
			if input == "" {
				continue
			}

			// Handle exit
			if input == "exit" || input == "quit" {
				fmt.Println("👋 Goodbye!")
				return nil
			}

			// Execute command
			result, err := s.executor.Execute(ctx, input)
			if err != nil {
				fmt.Printf("❌ Error: %v\n", err)
				continue
			}

			// Display result with proper newline handling
			if result.Output != "" {
				fmt.Print(result.Output)
				if !strings.HasSuffix(result.Output, "\n") {
					fmt.Println()
				}
			}

			// Display error if any
			if result.Error != nil {
				fmt.Printf("❌ Error: %v\n", result.Error)
			}

			// Display warnings if any
			for _, warning := range result.Warnings {
				fmt.Printf("⚠️  Warning: %s\n", warning)
			}
		}
	}

	return scanner.Err()
}

// runStableShell runs a stable shell with better terminal resize handling
func (s *BasicShell) runStableShell(ctx context.Context) error {
	fmt.Println("SuperShell v0.03 - Smart Command Line Interface (Stable Mode)")
	fmt.Println("Type 'help' for available commands or 'exit' to quit.")
	fmt.Println("Press TAB for command suggestions.")
	fmt.Println()

	// Create a channel to handle graceful shutdown
	done := make(chan bool)

	// Handle context cancellation
	go func() {
		<-ctx.Done()
		done <- true
	}()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		select {
		case <-done:
			return ctx.Err()
		default:
			// Display prompt
			fmt.Print(s.prompter.GetPrompt())

			// Read input with proper error handling
			if !scanner.Scan() {
				if err := scanner.Err(); err != nil {
					fmt.Printf("❌ Input error: %v\n", err)
				}
				break
			}

			input := strings.TrimSpace(scanner.Text())
			if input == "" {
				continue
			}

			// Handle exit
			if input == "exit" || input == "quit" {
				fmt.Println("👋 Goodbye!")
				return nil
			}

			// Execute command
			result, err := s.executor.Execute(ctx, input)
			if err != nil {
				fmt.Printf("❌ Error: %v\n", err)
				continue
			}

			// Display result
			if result.Output != "" {
				fmt.Print(result.Output)
				// Ensure we end with a newline
				if !strings.HasSuffix(result.Output, "\n") {
					fmt.Println()
				}
			}
		}
	}
}

// clearTerminalState clears any problematic terminal state
func (s *BasicShell) clearTerminalState() {
	// Send escape sequence to reset terminal state
	fmt.Print("\033[0m")   // Reset all attributes
	fmt.Print("\033[?25h") // Show cursor
	fmt.Print("\r")        // Return to beginning of line
}
