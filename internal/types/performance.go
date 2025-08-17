package types

import (
	"context"
	"time"
)

// PerformanceAnalyzer defines the interface for performance analysis operations
type PerformanceAnalyzer interface {
	CollectMetrics(ctx context.Context, duration time.Duration) (*PerformanceMetrics, error)
	AnalyzePerformance(ctx context.Context, metrics *PerformanceMetrics) (*AnalysisReport, error)
	GenerateReport(ctx context.Context, detailed bool) (*PerformanceReport, error)
	SaveBaseline(ctx context.Context, filepath string) error
	CompareBaseline(ctx context.Context, baselinePath string) (*ComparisonReport, error)
	GetOptimizationSuggestions(ctx context.Context) ([]*OptimizationSuggestion, error)
}

// PerformanceMetrics contains comprehensive system performance metrics
type PerformanceMetrics struct {
	Timestamp time.Time      `json:"timestamp"`
	CPU       CPUMetrics     `json:"cpu"`
	Memory    MemoryMetrics  `json:"memory"`
	Disk      DiskMetrics    `json:"disk"`
	Network   NetworkMetrics `json:"network"`
	Duration  time.Duration  `json:"duration"`
}

// CPUMetrics contains CPU performance data
type CPUMetrics struct {
	Usage       float64   `json:"usage"`        // Overall CPU usage percentage
	LoadAverage []float64 `json:"load_average"` // 1, 5, 15 minute load averages
	CoreUsage   []float64 `json:"core_usage"`   // Per-core usage percentages
	Processes   int       `json:"processes"`    // Number of running processes
	Threads     int       `json:"threads"`      // Number of threads
}

// MemoryMetrics contains memory performance data
type MemoryMetrics struct {
	Total     uint64  `json:"total"`      // Total memory in bytes
	Used      uint64  `json:"used"`       // Used memory in bytes
	Available uint64  `json:"available"`  // Available memory in bytes
	Usage     float64 `json:"usage"`      // Memory usage percentage
	SwapTotal uint64  `json:"swap_total"` // Total swap in bytes
	SwapUsed  uint64  `json:"swap_used"`  // Used swap in bytes
	SwapUsage float64 `json:"swap_usage"` // Swap usage percentage
	Cached    uint64  `json:"cached"`     // Cached memory in bytes
	Buffers   uint64  `json:"buffers"`    // Buffer memory in bytes
}

// DiskMetrics contains disk performance data
type DiskMetrics struct {
	Usage      []DiskUsage `json:"usage"`       // Per-disk usage
	IOStats    DiskIOStats `json:"io_stats"`    // Disk I/O statistics
	ReadSpeed  uint64      `json:"read_speed"`  // Read speed in bytes/sec
	WriteSpeed uint64      `json:"write_speed"` // Write speed in bytes/sec
}

// DiskUsage contains usage information for a single disk
type DiskUsage struct {
	Device     string  `json:"device"`     // Device name
	MountPoint string  `json:"mountpoint"` // Mount point
	Total      uint64  `json:"total"`      // Total space in bytes
	Used       uint64  `json:"used"`       // Used space in bytes
	Available  uint64  `json:"available"`  // Available space in bytes
	Usage      float64 `json:"usage"`      // Usage percentage
}

// DiskIOStats contains disk I/O statistics
type DiskIOStats struct {
	ReadOps    uint64 `json:"read_ops"`    // Number of read operations
	WriteOps   uint64 `json:"write_ops"`   // Number of write operations
	ReadBytes  uint64 `json:"read_bytes"`  // Bytes read
	WriteBytes uint64 `json:"write_bytes"` // Bytes written
}

// NetworkMetrics contains network performance data
type NetworkMetrics struct {
	Interfaces      []NetworkInterface `json:"interfaces"`       // Per-interface statistics
	Connections     int                `json:"connections"`      // Active connections
	BytesReceived   uint64             `json:"bytes_received"`   // Total bytes received
	BytesSent       uint64             `json:"bytes_sent"`       // Total bytes sent
	PacketsReceived uint64             `json:"packets_received"` // Total packets received
	PacketsSent     uint64             `json:"packets_sent"`     // Total packets sent
}

// NetworkInterface contains statistics for a network interface
type NetworkInterface struct {
	Name            string `json:"name"`             // Interface name
	BytesReceived   uint64 `json:"bytes_received"`   // Bytes received
	BytesSent       uint64 `json:"bytes_sent"`       // Bytes sent
	PacketsReceived uint64 `json:"packets_received"` // Packets received
	PacketsSent     uint64 `json:"packets_sent"`     // Packets sent
	Errors          uint64 `json:"errors"`           // Error count
	Drops           uint64 `json:"drops"`            // Dropped packets
}

// AnalysisReport contains performance analysis results
type AnalysisReport struct {
	Timestamp   time.Time                 `json:"timestamp"`
	Overall     HealthLevel               `json:"overall"`
	Components  map[string]HealthLevel    `json:"components"`
	Bottlenecks []string                  `json:"bottlenecks"`
	Warnings    []string                  `json:"warnings"`
	Suggestions []*OptimizationSuggestion `json:"suggestions"`
}

// PerformanceReport contains a comprehensive performance report
type PerformanceReport struct {
	Timestamp time.Time           `json:"timestamp"`
	Summary   string              `json:"summary"`
	Metrics   *PerformanceMetrics `json:"metrics"`
	Analysis  *AnalysisReport     `json:"analysis"`
	Trends    []TrendData         `json:"trends,omitempty"`
	Detailed  bool                `json:"detailed"`
}

// ComparisonReport contains baseline comparison results
type ComparisonReport struct {
	Timestamp     time.Time           `json:"timestamp"`
	BaselineDate  time.Time           `json:"baseline_date"`
	Current       *PerformanceMetrics `json:"current"`
	Baseline      *PerformanceMetrics `json:"baseline"`
	Improvements  []string            `json:"improvements"`
	Degradations  []string            `json:"degradations"`
	OverallChange string              `json:"overall_change"`
}

// OptimizationSuggestion contains a performance optimization suggestion
type OptimizationSuggestion struct {
	Category    string   `json:"category"`    // CPU, Memory, Disk, Network
	Priority    string   `json:"priority"`    // High, Medium, Low
	Title       string   `json:"title"`       // Short description
	Description string   `json:"description"` // Detailed description
	Impact      string   `json:"impact"`      // Expected impact
	Commands    []string `json:"commands"`    // Suggested commands to run
	Safe        bool     `json:"safe"`        // Whether it's safe to auto-apply
}

// TrendData contains historical trend information
type TrendData struct {
	Timestamp time.Time `json:"timestamp"`
	Metric    string    `json:"metric"`
	Value     float64   `json:"value"`
	Unit      string    `json:"unit"`
}
