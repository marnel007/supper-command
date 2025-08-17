package commands

import (
	"context"
	"sync"

	"suppercommand/internal/monitoring"
	"suppercommand/internal/security"
	"suppercommand/pkg/errors"
)

// Registry manages command registration and execution with dependency injection
type Registry struct {
	mu       sync.RWMutex
	commands map[string]Command
	security security.Validator
	logger   monitoring.Logger
}

// NewRegistry creates a new command registry
func NewRegistry(logger monitoring.Logger) *Registry {
	return &Registry{
		commands: make(map[string]Command),
		logger:   logger,
	}
}

// Initialize initializes the registry and registers built-in commands
func (r *Registry) Initialize(ctx context.Context, validator security.Validator) error {
	r.security = validator

	// Register built-in commands
	if err := r.registerBuiltinCommands(); err != nil {
		return errors.Wrap(err, "failed to register builtin commands")
	}

	r.logger.Info("Command registry initialized",
		monitoring.Field{Key: "command_count", Value: len(r.commands)})

	return nil
}

// Register adds a command to the registry
func (r *Registry) Register(cmd Command) error {
	if cmd == nil {
		return errors.NewValidationError("command cannot be nil")
	}

	name := cmd.Name()
	if name == "" {
		return errors.NewValidationError("command name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.commands[name]; exists {
		return errors.NewValidationError("command '%s' already registered", name)
	}

	r.commands[name] = cmd
	r.logger.Debug("Command registered",
		monitoring.Field{Key: "command", Value: name})

	return nil
}

// Unregister removes a command from the registry
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.commands[name]; !exists {
		return errors.NewValidationError("command '%s' not found", name)
	}

	delete(r.commands, name)
	r.logger.Debug("Command unregistered",
		monitoring.Field{Key: "command", Value: name})

	return nil
}

// Get retrieves a command by name
func (r *Registry) Get(name string) (Command, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cmd, exists := r.commands[name]
	if !exists {
		return nil, errors.NewValidationError("command '%s' not found", name)
	}

	return cmd, nil
}

// List returns all registered command names
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.commands))
	for name := range r.commands {
		names = append(names, name)
	}

	return names
}

// GetAllCommands returns all registered commands
func (r *Registry) GetAllCommands() []Command {
	r.mu.RLock()
	defer r.mu.RUnlock()

	commands := make([]Command, 0, len(r.commands))
	for _, cmd := range r.commands {
		commands = append(commands, cmd)
	}

	return commands
}

// Execute executes a command with validation
func (r *Registry) Execute(ctx context.Context, name string, args *Arguments) (*Result, error) {
	// Get command
	cmd, err := r.Get(name)
	if err != nil {
		return nil, err
	}

	// Validate arguments
	if err := cmd.Validate(args); err != nil {
		return nil, errors.Wrap(err, "argument validation failed")
	}

	// Security validation
	if r.security != nil {
		if err := r.security.ValidateCommand(name, args.Raw); err != nil {
			return nil, errors.Wrap(err, "security validation failed")
		}
	}

	// Execute command
	result, err := cmd.Execute(ctx, args)
	if err != nil {
		r.logger.Error("Command execution failed", err,
			monitoring.Field{Key: "command", Value: name})
		return nil, err
	}

	r.logger.Debug("Command executed successfully",
		monitoring.Field{Key: "command", Value: name},
		monitoring.Field{Key: "duration", Value: result.Duration})

	return result, nil
}

// registerBuiltinCommands registers all built-in commands
func (r *Registry) registerBuiltinCommands() error {
	// Register all new commands using the adapter
	if err := RegisterAllCommands(r); err != nil {
		return errors.Wrap(err, "failed to register new commands")
	}

	r.logger.Info("All commands registered successfully",
		monitoring.Field{Key: "total_commands", Value: len(r.commands)})

	return nil
}

// RegisterAllCommands registers all available commands with the registry
func RegisterAllCommands(registry *Registry) error {
	// This function is called from the app package to register commands
	// The actual registration is handled by the app package
	return nil
}
