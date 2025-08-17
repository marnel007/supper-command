package shell

import (
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

	// Create the interactive shell using go-prompt
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

	if result.Output != "" {
		fmt.Println(result.Output)
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
