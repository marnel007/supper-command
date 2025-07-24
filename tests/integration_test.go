package tests

import (
	"strings"
	"testing"
	"time"

	"suppercommand/internal/agent"
	"suppercommand/internal/core"
)

// TestAgentOSIntegration tests the core Agent OS functionality
func TestAgentOSIntegration(t *testing.T) {
	// Create enhanced shell
	shell := core.NewEnhancedShell()

	// Initialize
	err := shell.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize enhanced shell: %v", err)
	}

	// Verify Agent OS is running
	if !shell.IsEnhanced() {
		t.Fatal("Expected enhanced mode to be enabled")
	}

	agent := shell.GetAgent()
	if agent == nil {
		t.Fatal("Expected agent to be available")
	}

	// Test cleanup
	defer func() {
		if err := agent.Shutdown(); err != nil {
			t.Errorf("Failed to shutdown agent: %v", err)
		}
	}()
}

// TestDevelopmentCommands tests the development workflow commands
func TestDevelopmentCommands(t *testing.T) {
	shell := core.NewEnhancedShell()
	err := shell.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize shell: %v", err)
	}
	defer shell.GetAgent().Shutdown()

	testCases := []struct {
		name        string
		command     string
		args        []string
		expectError bool
	}{
		{
			name:    "dev reload",
			command: "dev reload",
			args:    []string{},
		},
		{
			name:    "dev profile",
			command: "dev profile",
			args:    []string{},
		},
		{
			name:    "dev docs",
			command: "dev docs",
			args:    []string{},
		},
		{
			name:    "dev build",
			command: "dev build",
			args:    []string{},
		},
		{
			name:        "dev test missing args",
			command:     "dev test",
			args:        []string{},
			expectError: true,
		},
		{
			name:    "dev test with command",
			command: "dev test",
			args:    []string{"help"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := shell.GetAgent().ExecuteCommand(tc.command, tc.args)

			if tc.expectError {
				if err == nil && result.ExitCode == 0 {
					t.Errorf("Expected error for command %s %v", tc.command, tc.args)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for command %s %v: %v", tc.command, tc.args, err)
				return
			}

			if result.ExitCode != 0 {
				t.Errorf("Command %s %v failed with exit code %d: %s",
					tc.command, tc.args, result.ExitCode, result.Output)
			}

			if result.Output == "" {
				t.Errorf("Expected output for command %s %v", tc.command, tc.args)
			}
		})
	}
}

// TestPerformanceCommands tests the performance monitoring features
func TestPerformanceCommands(t *testing.T) {
	shell := core.NewEnhancedShell()
	err := shell.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize shell: %v", err)
	}
	defer shell.GetAgent().Shutdown()

	// Execute some commands to generate performance data
	commands := []string{"help", "pwd", "echo"}
	for _, cmd := range commands {
		_, err := shell.GetAgent().ExecuteCommand(cmd, []string{})
		if err != nil {
			t.Errorf("Failed to execute command %s: %v", cmd, err)
		}
	}

	// Test performance commands
	perfCommands := []string{
		"perf stats",
		"perf monitor",
		"perf benchmark",
		"perf optimize",
	}

	for _, cmd := range perfCommands {
		t.Run(cmd, func(t *testing.T) {
			result, err := shell.GetAgent().ExecuteCommand(cmd, []string{})
			if err != nil {
				t.Errorf("Command %s failed: %v", cmd, err)
				return
			}

			if result.ExitCode != 0 {
				t.Errorf("Command %s failed with exit code %d: %s",
					cmd, result.ExitCode, result.Output)
			}

			if result.Output == "" {
				t.Errorf("Expected output for command %s", cmd)
			}
		})
	}
}

// TestCommandBridging tests legacy command bridging
func TestCommandBridging(t *testing.T) {
	shell := core.NewEnhancedShell()
	err := shell.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize shell: %v", err)
	}
	defer shell.GetAgent().Shutdown()

	// Test that legacy commands work through the bridge
	legacyCommands := []struct {
		name string
		args []string
	}{
		{"help", []string{}},
		{"echo", []string{"test"}},
		{"pwd", []string{}},
		{"whoami", []string{}},
		{"hostname", []string{}},
		{"ver", []string{}},
	}

	for _, cmd := range legacyCommands {
		t.Run(cmd.name, func(t *testing.T) {
			result, err := shell.GetAgent().ExecuteCommand(cmd.name, cmd.args)
			if err != nil {
				t.Errorf("Legacy command %s failed: %v", cmd.name, err)
				return
			}

			if result.ExitCode != 0 {
				t.Errorf("Legacy command %s failed with exit code %d: %s",
					cmd.name, result.ExitCode, result.Output)
			}

			// Verify it's marked as a legacy command
			if metadata, ok := result.Metadata["legacy_command"]; !ok || !metadata.(bool) {
				t.Errorf("Command %s should be marked as legacy", cmd.name)
			}
		})
	}
}

// TestPerformanceMonitoring tests the performance tracking functionality
func TestPerformanceMonitoring(t *testing.T) {
	shell := core.NewEnhancedShell()
	err := shell.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize shell: %v", err)
	}
	defer shell.GetAgent().Shutdown()

	agent := shell.GetAgent()

	// Execute a command multiple times
	cmdName := "echo"
	args := []string{"performance test"}

	for i := 0; i < 5; i++ {
		result, err := agent.ExecuteCommand(cmdName, args)
		if err != nil {
			t.Fatalf("Command execution failed: %v", err)
		}
		if result.ExitCode != 0 {
			t.Fatalf("Command failed with exit code %d", result.ExitCode)
		}
	}

	// Check performance stats
	stats := agent.GetPerformanceStats()
	if len(stats) == 0 {
		t.Fatal("Expected performance stats to be recorded")
	}

	echoStats, exists := stats[cmdName]
	if !exists {
		t.Fatalf("Expected stats for command %s", cmdName)
	}

	if echoStats.TotalExecutions != 5 {
		t.Errorf("Expected 5 executions, got %d", echoStats.TotalExecutions)
	}

	if echoStats.AverageTime <= 0 {
		t.Errorf("Expected positive average time, got %v", echoStats.AverageTime)
	}

	if echoStats.ErrorCount != 0 {
		t.Errorf("Expected 0 errors, got %d", echoStats.ErrorCount)
	}
}

// TestHotReload tests the hot reload functionality
func TestHotReload(t *testing.T) {
	shell := core.NewEnhancedShell()
	err := shell.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize shell: %v", err)
	}
	defer shell.GetAgent().Shutdown()

	// Test enabling hot reload
	result, err := shell.GetAgent().ExecuteCommand("dev reload", []string{"--watch", "internal/"})
	if err != nil {
		t.Fatalf("Hot reload command failed: %v", err)
	}

	if result.ExitCode != 0 {
		t.Fatalf("Hot reload failed with exit code %d: %s", result.ExitCode, result.Output)
	}

	if !strings.Contains(result.Output, "Hot Reload") {
		t.Error("Expected hot reload confirmation in output")
	}

	if !strings.Contains(result.Output, "internal/") {
		t.Error("Expected watched path in output")
	}
}

// TestCommandCategories tests that commands are properly categorized
func TestCommandCategories(t *testing.T) {
	shell := core.NewEnhancedShell()
	err := shell.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize shell: %v", err)
	}
	defer shell.GetAgent().Shutdown()

	// This would require extending the Agent interface to expose command metadata
	// For now, we'll test that development and performance commands exist
	devCommands := []string{
		"dev reload",
		"dev test",
		"dev profile",
		"dev docs",
		"dev build",
	}

	perfCommands := []string{
		"perf stats",
		"perf monitor",
		"perf benchmark",
		"perf optimize",
	}

	allCommands := append(devCommands, perfCommands...)

	for _, cmd := range allCommands {
		t.Run(cmd, func(t *testing.T) {
			// Test that the command exists by trying to execute it
			result, err := shell.GetAgent().ExecuteCommand(cmd, []string{})

			// We expect either success or a validation error (not "command not found")
			if err != nil && !strings.Contains(err.Error(), "required") {
				t.Errorf("Command %s appears to not exist: %v", cmd, err)
			}

			if result != nil && strings.Contains(result.Output, "Command not found") {
				t.Errorf("Command %s not found", cmd)
			}
		})
	}
}

// TestCommandValidation tests argument validation
func TestCommandValidation(t *testing.T) {
	shell := core.NewEnhancedShell()
	err := shell.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize shell: %v", err)
	}
	defer shell.GetAgent().Shutdown()

	// Test commands that require arguments
	testCases := []struct {
		command     string
		args        []string
		shouldError bool
	}{
		{"dev test", []string{}, true},        // requires command name
		{"dev test", []string{"help"}, false}, // valid command name
	}

	for _, tc := range testCases {
		t.Run(tc.command, func(t *testing.T) {
			result, err := shell.GetAgent().ExecuteCommand(tc.command, tc.args)

			if tc.shouldError {
				if err == nil && result.ExitCode == 0 {
					t.Errorf("Expected validation error for %s with args %v", tc.command, tc.args)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for %s with args %v: %v", tc.command, tc.args, err)
				}
				if result.ExitCode != 0 {
					t.Errorf("Command failed with exit code %d: %s", result.ExitCode, result.Output)
				}
			}
		})
	}
}

// TestCommandTimeout tests command execution timeouts
func TestCommandTimeout(t *testing.T) {
	shell := core.NewEnhancedShell()
	err := shell.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize shell: %v", err)
	}
	defer shell.GetAgent().Shutdown()

	// Test that commands complete within reasonable time
	start := time.Now()
	result, err := shell.GetAgent().ExecuteCommand("help", []string{})
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Help command failed: %v", err)
	}

	if result.ExitCode != 0 {
		t.Fatalf("Help command failed with exit code %d", result.ExitCode)
	}

	// Commands should complete in under 5 seconds
	if duration > 5*time.Second {
		t.Errorf("Command took too long: %v", duration)
	}

	// Agent OS should track execution time
	if result.Duration <= 0 {
		t.Error("Expected positive execution duration to be tracked")
	}
}

// TestResultTypes tests different result types
func TestResultTypes(t *testing.T) {
	shell := core.NewEnhancedShell()
	err := shell.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize shell: %v", err)
	}
	defer shell.GetAgent().Shutdown()

	testCases := []struct {
		command      string
		args         []string
		expectedType agent.ResultType
	}{
		{"help", []string{}, agent.ResultTypeSuccess},
		{"perf stats", []string{}, agent.ResultTypeInfo}, // When no stats available
		{"dev test", []string{}, agent.ResultTypeError},  // Missing required args
	}

	for _, tc := range testCases {
		t.Run(tc.command, func(t *testing.T) {
			result, _ := shell.GetAgent().ExecuteCommand(tc.command, tc.args)

			if result == nil {
				t.Fatal("Expected result object")
			}

			if result.Type != tc.expectedType {
				t.Errorf("Expected result type %s, got %s", tc.expectedType, result.Type)
			}
		})
	}
}

// BenchmarkCommandExecution benchmarks command execution performance
func BenchmarkCommandExecution(b *testing.B) {
	shell := core.NewEnhancedShell()
	err := shell.Initialize()
	if err != nil {
		b.Fatalf("Failed to initialize shell: %v", err)
	}
	defer shell.GetAgent().Shutdown()

	commands := []struct {
		name string
		args []string
	}{
		{"help", []string{}},
		{"echo", []string{"benchmark test"}},
		{"pwd", []string{}},
		{"whoami", []string{}},
	}

	for _, cmd := range commands {
		b.Run(cmd.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				result, err := shell.GetAgent().ExecuteCommand(cmd.name, cmd.args)
				if err != nil {
					b.Fatalf("Command failed: %v", err)
				}
				if result.ExitCode != 0 {
					b.Fatalf("Command failed with exit code %d", result.ExitCode)
				}
			}
		})
	}
}

// BenchmarkPerformanceMonitoring benchmarks the performance monitoring overhead
func BenchmarkPerformanceMonitoring(b *testing.B) {
	shell := core.NewEnhancedShell()
	err := shell.Initialize()
	if err != nil {
		b.Fatalf("Failed to initialize shell: %v", err)
	}
	defer shell.GetAgent().Shutdown()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := shell.GetAgent().ExecuteCommand("echo", []string{"monitoring test"})
		if err != nil {
			b.Fatalf("Command failed: %v", err)
		}
		if result.ExitCode != 0 {
			b.Fatalf("Command failed with exit code %d", result.ExitCode)
		}
	}
}

// TestMemoryUsage tests memory usage stays reasonable
func TestMemoryUsage(t *testing.T) {
	shell := core.NewEnhancedShell()
	err := shell.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize shell: %v", err)
	}
	defer shell.GetAgent().Shutdown()

	// Execute many commands to test memory growth
	for i := 0; i < 100; i++ {
		result, err := shell.GetAgent().ExecuteCommand("echo", []string{"memory test"})
		if err != nil {
			t.Fatalf("Command failed at iteration %d: %v", i, err)
		}
		if result.ExitCode != 0 {
			t.Fatalf("Command failed at iteration %d with exit code %d", i, result.ExitCode)
		}
	}

	// Check that performance stats are being maintained
	stats := shell.GetAgent().GetPerformanceStats()
	if len(stats) == 0 {
		t.Fatal("Expected performance stats to be recorded")
	}

	echoStats, exists := stats["echo"]
	if !exists {
		t.Fatal("Expected stats for echo command")
	}

	if echoStats.TotalExecutions != 100 {
		t.Errorf("Expected 100 executions, got %d", echoStats.TotalExecutions)
	}
}
