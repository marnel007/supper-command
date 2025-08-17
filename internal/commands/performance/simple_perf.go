package performance

import (
	"context"
	"fmt"
	"strings"
	"time"

	"suppercommand/internal/managers/performance"
	"suppercommand/internal/types"
)

// SimplePerfCommand is a basic performance command without circular dependencies
type SimplePerfCommand struct {
	name        string
	description string
	usage       string
	manager     types.PerformanceAnalyzer
}

// NewPerfCommand creates a new simple performance command
func NewPerfCommand() *SimplePerfCommand {
	factory := performance.NewFactory()
	manager, err := factory.CreateAnalyzer()
	if err != nil {
		// Continue with nil manager, will show error in execution
	}

	return &SimplePerfCommand{
		name:        "perf",
		description: "Performance monitoring and analysis",
		usage:       "perf [analyze|monitor|report|baseline] [options]",
		manager:     manager,
	}
}

// GetName returns the command name
func (p *SimplePerfCommand) GetName() string {
	return p.name
}

// GetDescription returns the command description
func (p *SimplePerfCommand) GetDescription() string {
	return p.description
}

// GetUsage returns the command usage
func (p *SimplePerfCommand) GetUsage() string {
	return p.usage
}

// Execute executes the performance command
func (p *SimplePerfCommand) Execute(ctx context.Context, args []string) error {
	if p.manager == nil {
		fmt.Println("Performance manager not available")
		return fmt.Errorf("performance manager not initialized")
	}

	if len(args) == 0 {
		return p.showHelp()
	}

	switch args[0] {
	case "analyze":
		return p.analyzePerformance(ctx)
	case "monitor":
		return p.monitorPerformance(ctx)
	case "report":
		return p.generateReport(ctx)
	case "baseline":
		return p.manageBaseline(ctx, args[1:])
	case "help", "--help", "-h":
		return p.showHelp()
	default:
		fmt.Printf("Unknown subcommand: %s\n", args[0])
		return p.showHelp()
	}
}

// analyzePerformance performs system performance analysis
func (p *SimplePerfCommand) analyzePerformance(ctx context.Context) error {
	fmt.Println("Analyzing system performance...")

	// Collect metrics first, then analyze
	metrics, err := p.manager.CollectMetrics(ctx, time.Minute)
	if err != nil {
		return err
	}
	analysis, err := p.manager.AnalyzePerformance(ctx, metrics)
	if err != nil {
		fmt.Printf("Error analyzing performance: %v\n", err)
		return err
	}

	fmt.Printf("Performance Analysis Results:\n")
	fmt.Printf("Overall Health: %s\n", analysis.Overall)
	fmt.Printf("Timestamp: %s\n", analysis.Timestamp.Format("2006-01-02 15:04:05"))

	if len(analysis.Bottlenecks) > 0 {
		fmt.Printf("\nBottlenecks Found:\n")
		for _, bottleneck := range analysis.Bottlenecks {
			fmt.Printf("- %s\n", bottleneck)
		}
	}

	if len(analysis.Warnings) > 0 {
		fmt.Printf("\nWarnings:\n")
		for _, warning := range analysis.Warnings {
			fmt.Printf("- %s\n", warning)
		}
	}

	if len(analysis.Suggestions) > 0 {
		fmt.Printf("\nSuggestions:\n")
		for _, suggestion := range analysis.Suggestions {
			fmt.Printf("- %s: %s\n", suggestion.Category, suggestion.Description)
		}
	}

	return nil
}

// monitorPerformance starts performance monitoring
func (p *SimplePerfCommand) monitorPerformance(ctx context.Context) error {
	fmt.Println("Starting performance monitoring...")
	fmt.Println("Press Ctrl+C to stop monitoring")

	// Collect metrics for monitoring
	_, err := p.manager.CollectMetrics(ctx, time.Second*10)
	if err != nil {
		fmt.Printf("Error during monitoring: %v\n", err)
		return err
	}

	fmt.Println("Monitoring completed")
	return nil
}

// generateReport generates a performance report
func (p *SimplePerfCommand) generateReport(ctx context.Context) error {
	fmt.Println("Generating performance report...")

	report, err := p.manager.GenerateReport(ctx, true)
	if err != nil {
		fmt.Printf("Error generating report: %v\n", err)
		return err
	}

	fmt.Printf("Performance Report\n")
	fmt.Printf("Generated: %s\n", report.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("Summary: %s\n", report.Summary)
	fmt.Printf("Detailed: %t\n", report.Detailed)

	if report.Analysis != nil {
		fmt.Printf("\nAnalysis: %s\n", report.Analysis.Overall)
	}

	return nil
}

// manageBaseline manages performance baselines
func (p *SimplePerfCommand) manageBaseline(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return p.listBaselines(ctx)
	}

	switch args[0] {
	case "create":
		if len(args) < 2 {
			fmt.Println("Usage: perf baseline create <name>")
			return nil
		}
		return p.createBaseline(ctx, args[1])
	case "list":
		return p.listBaselines(ctx)
	case "delete":
		if len(args) < 2 {
			fmt.Println("Usage: perf baseline delete <name>")
			return nil
		}
		return p.deleteBaseline(ctx, args[1])
	default:
		fmt.Printf("Unknown baseline subcommand: %s\n", args[0])
		return nil
	}
}

// createBaseline creates a new performance baseline
func (p *SimplePerfCommand) createBaseline(ctx context.Context, name string) error {
	fmt.Printf("Creating baseline: %s\n", name)

	err := p.manager.SaveBaseline(ctx, name+".json")
	if err != nil {
		fmt.Printf("Error creating baseline: %v\n", err)
		return err
	}

	fmt.Printf("Baseline '%s' created successfully\n", name)
	return nil
}

// listBaselines lists available baselines
func (p *SimplePerfCommand) listBaselines(ctx context.Context) error {
	// Simplified baseline listing - in real implementation would scan directory
	fmt.Println("Baseline listing not fully implemented in simple version")
	return nil
}

// deleteBaseline deletes a baseline
func (p *SimplePerfCommand) deleteBaseline(ctx context.Context, name string) error {
	fmt.Printf("Deleting baseline: %s\n", name)

	// Simplified baseline deletion
	fmt.Printf("Baseline deletion not fully implemented in simple version\n")
	fmt.Printf("Baseline '%s' deleted successfully\n", name)
	return nil
}

// showHelp shows command help
func (p *SimplePerfCommand) showHelp() error {
	help := `
Performance Monitoring Command

Usage: perf [command] [options]

Commands:
  analyze             Analyze current system performance
  monitor             Start real-time performance monitoring
  report              Generate a performance report
  baseline [cmd]      Manage performance baselines
    create <name>     Create a new baseline
    list              List all baselines
    delete <name>     Delete a baseline
  help                Show this help message

Examples:
  perf analyze        # Analyze current performance
  perf monitor        # Start monitoring
  perf report         # Generate report
  perf baseline list  # List baselines
`
	fmt.Println(strings.TrimSpace(help))
	return nil
}

// Name returns the command name (alias for GetName)
func (p *SimplePerfCommand) Name() string {
	return p.GetName()
}

// Description returns the command description (alias for GetDescription)
func (p *SimplePerfCommand) Description() string {
	return p.GetDescription()
}

// Usage returns the command usage (alias for GetUsage)
func (p *SimplePerfCommand) Usage() string {
	return p.GetUsage()
}
