package commands

import (
	"context"
	"time"

	"suppercommand/pkg/errors"
)

// Command interface defines the enhanced command interface with context and security
type Command interface {
	Name() string
	Description() string
	Usage() string
	Execute(ctx context.Context, args *Arguments) (*Result, error)
	Validate(args *Arguments) error
	RequiresElevation() bool
	SupportedPlatforms() []string
}

// Arguments contains parsed command arguments
type Arguments struct {
	Raw     []string
	Parsed  map[string]interface{}
	Flags   map[string]bool
	Options map[string]string
}

// Result contains the result of command execution
type Result struct {
	Output     string
	Error      error
	ExitCode   int
	Duration   time.Duration
	MemoryUsed int64
	Warnings   []string
}

// BaseCommand provides common functionality for all commands
type BaseCommand struct {
	name        string
	description string
	usage       string
	platforms   []string
	elevation   bool
}

// NewBaseCommand creates a new base command
func NewBaseCommand(name, description, usage string, platforms []string, elevation bool) *BaseCommand {
	return &BaseCommand{
		name:        name,
		description: description,
		usage:       usage,
		platforms:   platforms,
		elevation:   elevation,
	}
}

// Name returns the command name
func (c *BaseCommand) Name() string {
	return c.name
}

// Description returns the command description
func (c *BaseCommand) Description() string {
	return c.description
}

// Usage returns the command usage information
func (c *BaseCommand) Usage() string {
	return c.usage
}

// RequiresElevation returns whether the command requires elevated privileges
func (c *BaseCommand) RequiresElevation() bool {
	return c.elevation
}

// SupportedPlatforms returns the list of supported platforms
func (c *BaseCommand) SupportedPlatforms() []string {
	return c.platforms
}

// Validate provides basic argument validation
func (c *BaseCommand) Validate(args *Arguments) error {
	if args == nil {
		return errors.NewValidationError("arguments cannot be nil")
	}
	return nil
}

// ParseArguments parses raw arguments into structured format
func ParseArguments(raw []string) *Arguments {
	args := &Arguments{
		Raw:     raw,
		Parsed:  make(map[string]interface{}),
		Flags:   make(map[string]bool),
		Options: make(map[string]string),
	}

	for i, arg := range raw {
		if len(arg) > 0 && arg[0] == '-' {
			if len(arg) > 1 && arg[1] == '-' {
				// Long option (--option or --option=value)
				if eq := findEquals(arg); eq != -1 {
					key := arg[2:eq]
					value := arg[eq+1:]
					args.Options[key] = value
					args.Parsed[key] = value
				} else {
					key := arg[2:]
					args.Flags[key] = true
					args.Parsed[key] = true
				}
			} else {
				// Short option (-o or -o value)
				key := arg[1:]
				if i+1 < len(raw) && len(raw[i+1]) > 0 && raw[i+1][0] != '-' {
					// Next argument is the value
					args.Options[key] = raw[i+1]
					args.Parsed[key] = raw[i+1]
				} else {
					// It's a flag
					args.Flags[key] = true
					args.Parsed[key] = true
				}
			}
		}
	}

	return args
}

// findEquals finds the position of '=' in a string
func findEquals(s string) int {
	for i, c := range s {
		if c == '=' {
			return i
		}
	}
	return -1
}
