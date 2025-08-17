package commands_test

import (
	"context"
	"testing"
	"time"

	"suppercommand/internal/commands"
	"suppercommand/internal/config"
	"suppercommand/internal/monitoring"
)

// MockCommand implements the Command interface for testing
type MockCommand struct {
	name        string
	description string
	usage       string
	platforms   []string
	elevation   bool
	executeFunc func(ctx context.Context, args *commands.Arguments) (*commands.Result, error)
}

func (m *MockCommand) Name() string                            { return m.name }
func (m *MockCommand) Description() string                     { return m.description }
func (m *MockCommand) Usage() string                           { return m.usage }
func (m *MockCommand) RequiresElevation() bool                 { return m.elevation }
func (m *MockCommand) SupportedPlatforms() []string            { return m.platforms }
func (m *MockCommand) Validate(args *commands.Arguments) error { return nil }
func (m *MockCommand) Execute(ctx context.Context, args *commands.Arguments) (*commands.Result, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, args)
	}
	return &commands.Result{
		Output:   "mock output",
		Duration: 10 * time.Millisecond,
	}, nil
}

func TestRegistry_Register(t *testing.T) {
	logger := monitoring.NewLogger(config.MonitoringConfig{})
	registry := commands.NewRegistry(logger)

	mockCmd := &MockCommand{
		name:        "test",
		description: "Test command",
		usage:       "test [options]",
		platforms:   []string{"windows", "linux"},
	}

	// Test successful registration
	err := registry.Register(mockCmd)
	if err != nil {
		t.Errorf("Register() failed: %v", err)
	}

	// Test duplicate registration
	err = registry.Register(mockCmd)
	if err == nil {
		t.Error("Register() should fail for duplicate command")
	}

	// Test nil command
	err = registry.Register(nil)
	if err == nil {
		t.Error("Register() should fail for nil command")
	}
}

func TestRegistry_Get(t *testing.T) {
	logger := monitoring.NewLogger(config.MonitoringConfig{})
	registry := commands.NewRegistry(logger)

	mockCmd := &MockCommand{
		name:        "test",
		description: "Test command",
	}

	// Register command
	registry.Register(mockCmd)

	// Test successful get
	cmd, err := registry.Get("test")
	if err != nil {
		t.Errorf("Get() failed: %v", err)
	}
	if cmd.Name() != "test" {
		t.Errorf("Get() returned wrong command: got %s, want test", cmd.Name())
	}

	// Test non-existent command
	_, err = registry.Get("nonexistent")
	if err == nil {
		t.Error("Get() should fail for non-existent command")
	}
}

func TestRegistry_List(t *testing.T) {
	logger := monitoring.NewLogger(config.MonitoringConfig{})
	registry := commands.NewRegistry(logger)

	// Empty registry
	names := registry.List()
	if len(names) != 0 {
		t.Errorf("List() should return empty slice for empty registry, got %d items", len(names))
	}

	// Add commands
	registry.Register(&MockCommand{name: "cmd1"})
	registry.Register(&MockCommand{name: "cmd2"})

	names = registry.List()
	if len(names) != 2 {
		t.Errorf("List() should return 2 items, got %d", len(names))
	}
}

func TestParseArguments(t *testing.T) {
	tests := []struct {
		name        string
		raw         []string
		wantFlags   map[string]bool
		wantOptions map[string]string
	}{
		{
			name: "flags and options",
			raw:  []string{"--verbose", "-f", "file.txt", "--output=result.txt"},
			wantFlags: map[string]bool{
				"verbose": true,
			},
			wantOptions: map[string]string{
				"f":      "file.txt",
				"output": "result.txt",
			},
		},
		{
			name:        "no arguments",
			raw:         []string{},
			wantFlags:   map[string]bool{},
			wantOptions: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := commands.ParseArguments(tt.raw)

			for flag, expected := range tt.wantFlags {
				if args.Flags[flag] != expected {
					t.Errorf("ParseArguments() flag %s = %v, want %v", flag, args.Flags[flag], expected)
				}
			}

			for option, expected := range tt.wantOptions {
				if args.Options[option] != expected {
					t.Errorf("ParseArguments() option %s = %v, want %v", option, args.Options[option], expected)
				}
			}
		})
	}
}
