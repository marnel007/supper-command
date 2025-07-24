package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fatih/color"
)

// Agent represents the Agent OS core engine
type Agent struct {
	version     string
	plugins     map[string]Plugin
	commands    map[string]Command
	performance *PerformanceMonitor
	hotReload   *HotReloader
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

// Plugin interface for modular functionality
type Plugin interface {
	Name() string
	Version() string
	Commands() []Command
	Initialize(ctx context.Context, agent *Agent) error
	Shutdown() error
}

// Command interface enhanced with Agent OS features
type Command interface {
	Name() string
	Description() string
	Category() string
	Examples() []string
	Execute(ctx context.Context, args []string) (*Result, error)
	ValidateArgs(args []string) error
}

// Result represents command execution results with metadata
type Result struct {
	Output     string         `json:"output"`
	ExitCode   int            `json:"exit_code"`
	Duration   time.Duration  `json:"duration"`
	MemoryUsed int64          `json:"memory_used"`
	Metadata   map[string]any `json:"metadata"`
	Type       ResultType     `json:"type"`
}

type ResultType string

const (
	ResultTypeSuccess ResultType = "success"
	ResultTypeError   ResultType = "error"
	ResultTypeWarning ResultType = "warning"
	ResultTypeInfo    ResultType = "info"
)

// PerformanceMonitor tracks command execution metrics
type PerformanceMonitor struct {
	metrics map[string]*CommandMetrics
	mu      sync.RWMutex
}

type CommandMetrics struct {
	TotalExecutions int64         `json:"total_executions"`
	AverageTime     time.Duration `json:"average_time"`
	MaxTime         time.Duration `json:"max_time"`
	MinTime         time.Duration `json:"min_time"`
	ErrorCount      int64         `json:"error_count"`
	LastExecuted    time.Time     `json:"last_executed"`
}

// HotReloader enables live command reloading
type HotReloader struct {
	watchedPaths []string
	callbacks    []func()
	enabled      bool
}

// NewAgent creates a new Agent OS instance
func NewAgent() *Agent {
	ctx, cancel := context.WithCancel(context.Background())

	return &Agent{
		version:     "1.0.0",
		plugins:     make(map[string]Plugin),
		commands:    make(map[string]Command),
		performance: NewPerformanceMonitor(),
		hotReload:   NewHotReloader(),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Initialize starts the Agent OS engine
func (a *Agent) Initialize() error {
	color.New(color.FgCyan, color.Bold).Println("🤖 Agent OS - SuperShell Edition")
	color.New(color.FgGreen).Printf("   Version: %s\n", a.version)
	color.New(color.FgYellow).Println("   Initializing enhanced shell capabilities...")

	// Initialize core components
	if err := a.performance.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize performance monitor: %w", err)
	}

	if err := a.hotReload.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize hot reloader: %w", err)
	}

	// Load built-in plugins
	if err := a.loadCorePlugins(); err != nil {
		return fmt.Errorf("failed to load core plugins: %w", err)
	}

	color.New(color.FgGreen).Println("✅ Agent OS initialized successfully!")
	return nil
}

// ExecuteCommand runs a command with full Agent OS instrumentation
func (a *Agent) ExecuteCommand(name string, args []string) (*Result, error) {
	start := time.Now()

	a.mu.RLock()
	cmd, exists := a.commands[name]
	a.mu.RUnlock()

	if !exists {
		return &Result{
			Output:   fmt.Sprintf("Command not found: %s", name),
			ExitCode: 1,
			Duration: time.Since(start),
			Type:     ResultTypeError,
		}, nil
	}

	// Validate arguments
	if err := cmd.ValidateArgs(args); err != nil {
		return &Result{
			Output:   fmt.Sprintf("Invalid arguments: %s", err.Error()),
			ExitCode: 1,
			Duration: time.Since(start),
			Type:     ResultTypeError,
		}, nil
	}

	// Execute with context and monitoring
	ctx, cancel := context.WithTimeout(a.ctx, 30*time.Second)
	defer cancel()

	result, err := cmd.Execute(ctx, args)
	if err != nil {
		result = &Result{
			Output:   fmt.Sprintf("Execution error: %s", err.Error()),
			ExitCode: 1,
			Duration: time.Since(start),
			Type:     ResultTypeError,
		}
	}

	// Update performance metrics
	result.Duration = time.Since(start)
	a.performance.RecordExecution(name, result)

	return result, nil
}

// RegisterPlugin adds a new plugin to the Agent
func (a *Agent) RegisterPlugin(plugin Plugin) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if err := plugin.Initialize(a.ctx, a); err != nil {
		return fmt.Errorf("failed to initialize plugin %s: %w", plugin.Name(), err)
	}

	a.plugins[plugin.Name()] = plugin

	// Register plugin commands
	for _, cmd := range plugin.Commands() {
		a.commands[cmd.Name()] = cmd
	}

	color.New(color.FgBlue).Printf("🔌 Plugin loaded: %s v%s\n", plugin.Name(), plugin.Version())
	return nil
}

// GetPerformanceStats returns current performance statistics
func (a *Agent) GetPerformanceStats() map[string]*CommandMetrics {
	return a.performance.GetStats()
}

// RegisterCommand adds a single command to the Agent OS
func (a *Agent) RegisterCommand(name string, cmd Command) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.commands[name] = cmd
}

// EnableHotReload activates hot reloading functionality
func (a *Agent) EnableHotReload(paths []string) error {
	return a.hotReload.Enable(paths)
}

// Shutdown gracefully stops the Agent OS
func (a *Agent) Shutdown() error {
	a.cancel()

	a.mu.Lock()
	defer a.mu.Unlock()

	// Shutdown all plugins
	for name, plugin := range a.plugins {
		if err := plugin.Shutdown(); err != nil {
			color.New(color.FgRed).Printf("Error shutting down plugin %s: %v\n", name, err)
		}
	}

	color.New(color.FgYellow).Println("🤖 Agent OS shutdown complete")
	return nil
}

// loadCorePlugins initializes essential Agent OS plugins
func (a *Agent) loadCorePlugins() error {
	color.New(color.FgBlue).Println("📦 Loading core plugins...")

	plugins := []Plugin{
		&DevelopmentPlugin{},
		&PerformancePlugin{},
		&CloudPlugin{},
		&SecurityPlugin{},
		&MonitoringPlugin{},
		&AutomationPlugin{},
		&MarketplacePlugin{}, // ← Added
		&TestingPlugin{},     // ← Added
	}

	for _, plugin := range plugins {
		if err := a.registerPluginUnsafe(plugin); err != nil {
			return fmt.Errorf("failed to load plugin %s: %w", plugin.Name(), err)
		}
	}

	return nil
}

// registerPluginUnsafe adds a plugin without acquiring locks (for internal use during initialization)
func (a *Agent) registerPluginUnsafe(plugin Plugin) error {
	if err := plugin.Initialize(a.ctx, a); err != nil {
		return fmt.Errorf("failed to initialize plugin %s: %w", plugin.Name(), err)
	}

	a.plugins[plugin.Name()] = plugin

	// Register plugin commands
	for _, cmd := range plugin.Commands() {
		a.commands[cmd.Name()] = cmd
	}

	color.New(color.FgBlue).Printf("🔌 Plugin loaded: %s v%s\n", plugin.Name(), plugin.Version())
	return nil
}

// NewPerformanceMonitor creates a new performance monitoring instance
func NewPerformanceMonitor() *PerformanceMonitor {
	return &PerformanceMonitor{
		metrics: make(map[string]*CommandMetrics),
	}
}

// Initialize sets up the performance monitor
func (pm *PerformanceMonitor) Initialize() error {
	color.New(color.FgYellow).Println("📊 Performance monitoring enabled")
	return nil
}

// RecordExecution records metrics for a command execution
func (pm *PerformanceMonitor) RecordExecution(cmdName string, result *Result) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	metrics, exists := pm.metrics[cmdName]
	if !exists {
		metrics = &CommandMetrics{
			MinTime: result.Duration,
			MaxTime: result.Duration,
		}
		pm.metrics[cmdName] = metrics
	}

	metrics.TotalExecutions++
	metrics.LastExecuted = time.Now()

	// Update timing statistics
	if result.Duration < metrics.MinTime {
		metrics.MinTime = result.Duration
	}
	if result.Duration > metrics.MaxTime {
		metrics.MaxTime = result.Duration
	}

	// Calculate running average
	totalTime := time.Duration(int64(metrics.AverageTime) * (metrics.TotalExecutions - 1))
	metrics.AverageTime = (totalTime + result.Duration) / time.Duration(metrics.TotalExecutions)

	if result.ExitCode != 0 {
		metrics.ErrorCount++
	}
}

// GetStats returns current performance statistics
func (pm *PerformanceMonitor) GetStats() map[string]*CommandMetrics {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Return a copy to avoid race conditions
	stats := make(map[string]*CommandMetrics)
	for name, metrics := range pm.metrics {
		stats[name] = &CommandMetrics{
			TotalExecutions: metrics.TotalExecutions,
			AverageTime:     metrics.AverageTime,
			MaxTime:         metrics.MaxTime,
			MinTime:         metrics.MinTime,
			ErrorCount:      metrics.ErrorCount,
			LastExecuted:    metrics.LastExecuted,
		}
	}
	return stats
}

// NewHotReloader creates a new hot reload instance
func NewHotReloader() *HotReloader {
	return &HotReloader{
		watchedPaths: make([]string, 0),
		callbacks:    make([]func(), 0),
		enabled:      false,
	}
}

// Initialize sets up the hot reloader
func (hr *HotReloader) Initialize() error {
	color.New(color.FgMagenta).Println("🔥 Hot reload system ready")
	return nil
}

// Enable activates hot reloading for specified paths
func (hr *HotReloader) Enable(paths []string) error {
	hr.watchedPaths = append(hr.watchedPaths, paths...)
	hr.enabled = true
	color.New(color.FgMagenta).Printf("🔥 Hot reload enabled for %d paths\n", len(paths))
	return nil
}
