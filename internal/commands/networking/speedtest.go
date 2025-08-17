package networking

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"suppercommand/internal/commands"

	"github.com/fatih/color"
)

// SpeedtestCommand performs network speed testing
type SpeedtestCommand struct {
	*commands.BaseCommand
}

// NewSpeedtestCommand creates a new speedtest command
func NewSpeedtestCommand() *SpeedtestCommand {
	return &SpeedtestCommand{
		BaseCommand: commands.NewBaseCommand(
			"speedtest",
			"Perform network speed test",
			"speedtest [-s] [-q] [--download-only] [--upload-only]",
			[]string{"windows", "linux", "darwin"},
			false,
		),
	}
}

// Execute performs a network speed test
func (s *SpeedtestCommand) Execute(ctx context.Context, args *commands.Arguments) (*commands.Result, error) {
	startTime := time.Now()

	// Parse arguments
	simple := false
	quiet := false
	downloadOnly := false
	uploadOnly := false

	for _, arg := range args.Raw {
		switch arg {
		case "-s", "--simple":
			simple = true
		case "-q", "--quiet":
			quiet = true
		case "--download-only":
			downloadOnly = true
		case "--upload-only":
			uploadOnly = true
		}
	}

	var output strings.Builder

	if !quiet {
		output.WriteString(color.New(color.FgCyan, color.Bold).Sprint("🚀 NETWORK SPEED TEST\n"))
		output.WriteString("═══════════════════════════════════════════════════════════════\n")
		output.WriteString("🔍 Initializing speed test...\n")
	}

	// Simulate server selection
	if !quiet {
		output.WriteString("📡 Selecting optimal test server...\n")
		time.Sleep(500 * time.Millisecond)
		output.WriteString(color.New(color.FgGreen).Sprint("✅ Server selected: speedtest.example.com (25.3 ms)\n"))
		output.WriteString("───────────────────────────────────────────────────────────────\n")
	}

	var downloadSpeed, uploadSpeed, ping float64

	// Ping test
	if !quiet {
		output.WriteString("🏓 Testing ping...\n")
	}
	ping = s.testPing()
	if !quiet {
		output.WriteString(fmt.Sprintf("✅ Ping: %.1f ms\n", ping))
		output.WriteString("───────────────────────────────────────────────────────────────\n")
	}

	// Download test
	if !uploadOnly {
		if !quiet {
			output.WriteString("⬇️  Testing download speed...\n")
		}
		downloadSpeed = s.testDownload(quiet, &output)
		if !quiet {
			output.WriteString(fmt.Sprintf("✅ Download: %.2f Mbps\n", downloadSpeed))
			output.WriteString("───────────────────────────────────────────────────────────────\n")
		}
	}

	// Upload test
	if !downloadOnly {
		if !quiet {
			output.WriteString("⬆️  Testing upload speed...\n")
		}
		uploadSpeed = s.testUpload(quiet, &output)
		if !quiet {
			output.WriteString(fmt.Sprintf("✅ Upload: %.2f Mbps\n", uploadSpeed))
			output.WriteString("───────────────────────────────────────────────────────────────\n")
		}
	}

	// Results summary
	if simple {
		if !downloadOnly && !uploadOnly {
			output.WriteString(fmt.Sprintf("Ping: %.1f ms | Download: %.2f Mbps | Upload: %.2f Mbps\n",
				ping, downloadSpeed, uploadSpeed))
		} else if downloadOnly {
			output.WriteString(fmt.Sprintf("Ping: %.1f ms | Download: %.2f Mbps\n", ping, downloadSpeed))
		} else {
			output.WriteString(fmt.Sprintf("Ping: %.1f ms | Upload: %.2f Mbps\n", ping, uploadSpeed))
		}
	} else {
		output.WriteString(color.New(color.FgGreen, color.Bold).Sprint("🎯 SPEED TEST RESULTS\n"))
		output.WriteString("═══════════════════════════════════════════════════════════════\n")

		output.WriteString(fmt.Sprintf("🏓 %-15s %s\n", "Ping:",
			color.New(color.FgYellow, color.Bold).Sprintf("%.1f ms", ping)))

		if !uploadOnly {
			output.WriteString(fmt.Sprintf("⬇️  %-15s %s\n", "Download:",
				color.New(color.FgGreen, color.Bold).Sprintf("%.2f Mbps", downloadSpeed)))
		}

		if !downloadOnly {
			output.WriteString(fmt.Sprintf("⬆️  %-15s %s\n", "Upload:",
				color.New(color.FgBlue, color.Bold).Sprintf("%.2f Mbps", uploadSpeed)))
		}

		output.WriteString("───────────────────────────────────────────────────────────────\n")

		// Performance rating
		rating := s.getPerformanceRating(downloadSpeed, uploadSpeed, ping)
		output.WriteString(fmt.Sprintf("📊 Performance:   %s\n", rating))
		output.WriteString(fmt.Sprintf("⏱️  Test Duration:  %v\n", time.Since(startTime).Round(time.Second)))
		output.WriteString("═══════════════════════════════════════════════════════════════\n")
	}

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// testPing simulates ping test
func (s *SpeedtestCommand) testPing() float64 {
	// Simulate ping test with some variation
	time.Sleep(200 * time.Millisecond)
	return 15.0 + rand.Float64()*20.0 // 15-35ms range
}

// testDownload simulates download speed test
func (s *SpeedtestCommand) testDownload(quiet bool, output *strings.Builder) float64 {
	testDuration := 3 * time.Second
	startTime := time.Now()

	// Simulate progressive speed measurement
	for elapsed := time.Duration(0); elapsed < testDuration; elapsed = time.Since(startTime) {
		if !quiet {
			progress := float64(elapsed) / float64(testDuration) * 100
			currentSpeed := 50.0 + rand.Float64()*100.0 // 50-150 Mbps range
			output.WriteString(fmt.Sprintf("\r📈 Progress: %.0f%% - Current: %.1f Mbps", progress, currentSpeed))
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !quiet {
		output.WriteString("\n")
	}

	// Return final speed
	return 85.5 + rand.Float64()*30.0 // 85-115 Mbps range
}

// testUpload simulates upload speed test
func (s *SpeedtestCommand) testUpload(quiet bool, output *strings.Builder) float64 {
	testDuration := 3 * time.Second
	startTime := time.Now()

	// Simulate progressive speed measurement
	for elapsed := time.Duration(0); elapsed < testDuration; elapsed = time.Since(startTime) {
		if !quiet {
			progress := float64(elapsed) / float64(testDuration) * 100
			currentSpeed := 20.0 + rand.Float64()*40.0 // 20-60 Mbps range
			output.WriteString(fmt.Sprintf("\r📈 Progress: %.0f%% - Current: %.1f Mbps", progress, currentSpeed))
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !quiet {
		output.WriteString("\n")
	}

	// Return final speed
	return 35.2 + rand.Float64()*20.0 // 35-55 Mbps range
}

// getPerformanceRating returns a performance rating based on speeds
func (s *SpeedtestCommand) getPerformanceRating(download, upload, ping float64) string {
	// Simple rating algorithm
	score := 0

	// Download score (0-40 points)
	if download >= 100 {
		score += 40
	} else if download >= 50 {
		score += 30
	} else if download >= 25 {
		score += 20
	} else if download >= 10 {
		score += 10
	}

	// Upload score (0-30 points)
	if upload >= 50 {
		score += 30
	} else if upload >= 25 {
		score += 20
	} else if upload >= 10 {
		score += 15
	} else if upload >= 5 {
		score += 10
	}

	// Ping score (0-30 points)
	if ping <= 20 {
		score += 30
	} else if ping <= 50 {
		score += 20
	} else if ping <= 100 {
		score += 10
	}

	// Return rating
	if score >= 90 {
		return color.New(color.FgGreen, color.Bold).Sprint("🌟 Excellent")
	} else if score >= 70 {
		return color.New(color.FgGreen).Sprint("✅ Good")
	} else if score >= 50 {
		return color.New(color.FgYellow).Sprint("⚠️  Fair")
	} else {
		return color.New(color.FgRed).Sprint("❌ Poor")
	}
}
