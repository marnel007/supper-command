package monitoring

import (
	"context"
	"sync"
	"time"

	"suppercommand/internal/config"
)

// Monitor interface defines performance monitoring methods
type Monitor interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	RecordCommandExecution(cmd string, duration time.Duration, success bool)
	RecordMemoryUsage(bytes int64)
	RecordError(err error, context map[string]interface{})
	GetMetrics() *Metrics
}

// Metrics contains collected performance metrics
type Metrics struct {
	CommandExecutions map[string]*CommandMetrics
	MemoryUsage       *MemoryMetrics
	ErrorCounts       map[string]int64
	StartTime         time.Time
	LastUpdate        time.Time
}

// CommandMetrics contains metrics for a specific command
type CommandMetrics struct {
	Count         int64
	TotalDuration time.Duration
	AvgDuration   time.Duration
	MinDuration   time.Duration
	MaxDuration   time.Duration
	SuccessCount  int64
	ErrorCount    int64
}

// MemoryMetrics contains memory usage metrics
type MemoryMetrics struct {
	Current int64
	Peak    int64
	Average int64
	Samples int64
}

// BasicMonitor implements the Monitor interface
type BasicMonitor struct {
	mu      sync.RWMutex
	config  config.MonitoringConfig
	logger  Logger
	metrics *Metrics
	running bool
}

// NewMonitor creates a new monitor instance
func NewMonitor(config config.MonitoringConfig, logger Logger) Monitor {
	monitor := &BasicMonitor{
		config: config,
		logger: logger,
		metrics: &Metrics{
			CommandExecutions: make(map[string]*CommandMetrics),
			MemoryUsage:       &MemoryMetrics{},
			ErrorCounts:       make(map[string]int64),
			StartTime:         time.Now(),
			LastUpdate:        time.Now(),
		},
	}

	return monitor
}

// Start begins monitoring
func (m *BasicMonitor) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return nil
	}

	m.running = true
	m.logger.Info("Monitor started")

	// Start background metrics collection
	go m.collectMetrics(ctx)

	return nil
}

// Stop stops monitoring
func (m *BasicMonitor) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return nil
	}

	m.running = false
	m.logger.Info("Monitor stopped")

	return nil
}

// RecordCommandExecution records command execution metrics
func (m *BasicMonitor) RecordCommandExecution(cmd string, duration time.Duration, success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	metrics, exists := m.metrics.CommandExecutions[cmd]
	if !exists {
		metrics = &CommandMetrics{
			MinDuration: duration,
			MaxDuration: duration,
		}
		m.metrics.CommandExecutions[cmd] = metrics
	}

	metrics.Count++
	metrics.TotalDuration += duration

	if duration < metrics.MinDuration {
		metrics.MinDuration = duration
	}
	if duration > metrics.MaxDuration {
		metrics.MaxDuration = duration
	}

	metrics.AvgDuration = metrics.TotalDuration / time.Duration(metrics.Count)

	if success {
		metrics.SuccessCount++
	} else {
		metrics.ErrorCount++
	}

	m.metrics.LastUpdate = time.Now()
}

// RecordMemoryUsage records memory usage
func (m *BasicMonitor) RecordMemoryUsage(bytes int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	mem := m.metrics.MemoryUsage
	mem.Current = bytes
	mem.Samples++

	if bytes > mem.Peak {
		mem.Peak = bytes
	}

	// Calculate running average
	if mem.Samples == 1 {
		mem.Average = bytes
	} else {
		mem.Average = (mem.Average*(mem.Samples-1) + bytes) / mem.Samples
	}

	m.metrics.LastUpdate = time.Now()
}

// RecordError records error occurrences
func (m *BasicMonitor) RecordError(err error, context map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	errorType := err.Error()
	m.metrics.ErrorCounts[errorType]++
	m.metrics.LastUpdate = time.Now()

	m.logger.Error("Error recorded", err,
		Field{Key: "context", Value: context})
}

// GetMetrics returns current metrics
func (m *BasicMonitor) GetMetrics() *Metrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to avoid race conditions
	metricsCopy := &Metrics{
		CommandExecutions: make(map[string]*CommandMetrics),
		MemoryUsage: &MemoryMetrics{
			Current: m.metrics.MemoryUsage.Current,
			Peak:    m.metrics.MemoryUsage.Peak,
			Average: m.metrics.MemoryUsage.Average,
			Samples: m.metrics.MemoryUsage.Samples,
		},
		ErrorCounts: make(map[string]int64),
		StartTime:   m.metrics.StartTime,
		LastUpdate:  m.metrics.LastUpdate,
	}

	for cmd, metrics := range m.metrics.CommandExecutions {
		metricsCopy.CommandExecutions[cmd] = &CommandMetrics{
			Count:         metrics.Count,
			TotalDuration: metrics.TotalDuration,
			AvgDuration:   metrics.AvgDuration,
			MinDuration:   metrics.MinDuration,
			MaxDuration:   metrics.MaxDuration,
			SuccessCount:  metrics.SuccessCount,
			ErrorCount:    metrics.ErrorCount,
		}
	}

	for errorType, count := range m.metrics.ErrorCounts {
		metricsCopy.ErrorCounts[errorType] = count
	}

	return metricsCopy
}

// collectMetrics runs background metrics collection
func (m *BasicMonitor) collectMetrics(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !m.running {
				return
			}
			// Collect system metrics here
			// For now, this is a placeholder
		}
	}
}
