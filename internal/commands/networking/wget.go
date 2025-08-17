package networking

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"suppercommand/internal/commands"

	"github.com/fatih/color"
)

// WgetCommand downloads files from URLs
type WgetCommand struct {
	*commands.BaseCommand
}

// NewWgetCommand creates a new wget command
func NewWgetCommand() *WgetCommand {
	return &WgetCommand{
		BaseCommand: commands.NewBaseCommand(
			"wget",
			"Download files from URLs",
			"wget [-v] <url> [filename]",
			[]string{"windows", "linux", "darwin"},
			false,
		),
	}
}

// Execute downloads a file from a URL
func (w *WgetCommand) Execute(ctx context.Context, args *commands.Arguments) (*commands.Result, error) {
	startTime := time.Now()

	if len(args.Raw) == 0 {
		return &commands.Result{
			Output:   "Usage: wget [-v] <url> [filename]\n",
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	// Parse arguments
	verbose := false
	var url, filename string

	for i, arg := range args.Raw {
		if arg == "-v" || arg == "--verbose" {
			verbose = true
		} else if url == "" {
			url = arg
		} else if filename == "" {
			filename = arg
		} else if i > 0 && args.Raw[i-1] != "-v" && args.Raw[i-1] != "--verbose" {
			filename = arg
		}
	}

	if url == "" {
		return &commands.Result{
			Output:   "Error: No URL specified\n",
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	// If no filename specified, extract from URL
	if filename == "" {
		filename = filepath.Base(url)
		if filename == "." || filename == "/" {
			filename = "index.html"
		}
	}

	var output strings.Builder

	if verbose {
		output.WriteString(color.New(color.FgCyan, color.Bold).Sprintf("🌐 WGET - File Download\n"))
		output.WriteString("═══════════════════════════════════════════════════════════════\n")
		output.WriteString(fmt.Sprintf("📡 URL:      %s\n", color.New(color.FgBlue).Sprint(url)))
		output.WriteString(fmt.Sprintf("📁 File:     %s\n", color.New(color.FgGreen).Sprint(filename)))
		output.WriteString("───────────────────────────────────────────────────────────────\n")
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Make the request
	if verbose {
		output.WriteString("🔄 Connecting to server...\n")
	}

	resp, err := client.Get(url)
	if err != nil {
		return &commands.Result{
			Output:   fmt.Sprintf("Error: Failed to connect to %s: %v\n", url, err),
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &commands.Result{
			Output:   fmt.Sprintf("Error: HTTP %d - %s\n", resp.StatusCode, resp.Status),
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	// Get content length for progress
	contentLength := resp.ContentLength
	if verbose {
		if contentLength > 0 {
			output.WriteString(fmt.Sprintf("📊 Size:     %s\n", formatBytes(contentLength)))
		}
		output.WriteString(fmt.Sprintf("✅ Status:   %s\n", resp.Status))
		output.WriteString("───────────────────────────────────────────────────────────────\n")
		output.WriteString("⬇️  Downloading...\n")
	}

	// Create the output file
	file, err := os.Create(filename)
	if err != nil {
		return &commands.Result{
			Output:   fmt.Sprintf("Error: Cannot create file %s: %v\n", filename, err),
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}
	defer file.Close()

	// Copy with progress tracking
	downloadStart := time.Now()
	var written int64

	if verbose && contentLength > 0 {
		// Progress tracking for known size
		buffer := make([]byte, 32*1024) // 32KB buffer
		for {
			n, err := resp.Body.Read(buffer)
			if n > 0 {
				file.Write(buffer[:n])
				written += int64(n)

				// Show progress every 1MB or at end
				if written%1048576 == 0 || err == io.EOF {
					progress := float64(written) / float64(contentLength) * 100
					output.WriteString(fmt.Sprintf("\r📈 Progress: %.1f%% (%s/%s)",
						progress, formatBytes(written), formatBytes(contentLength)))
				}
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				return &commands.Result{
					Output:   fmt.Sprintf("Error: Download failed: %v\n", err),
					ExitCode: 1,
					Duration: time.Since(startTime),
				}, nil
			}
		}
		output.WriteString("\n")
	} else {
		// Simple copy for unknown size or non-verbose
		written, err = io.Copy(file, resp.Body)
		if err != nil {
			return &commands.Result{
				Output:   fmt.Sprintf("Error: Download failed: %v\n", err),
				ExitCode: 1,
				Duration: time.Since(startTime),
			}, nil
		}
	}

	downloadDuration := time.Since(downloadStart)

	if verbose {
		output.WriteString("───────────────────────────────────────────────────────────────\n")
		output.WriteString(color.New(color.FgGreen, color.Bold).Sprint("✅ DOWNLOAD COMPLETE\n"))
		output.WriteString(fmt.Sprintf("📁 File:     %s\n", filename))
		output.WriteString(fmt.Sprintf("📊 Size:     %s\n", formatBytes(written)))
		output.WriteString(fmt.Sprintf("⏱️  Time:     %v\n", downloadDuration.Round(time.Millisecond)))
		if downloadDuration.Seconds() > 0 {
			speed := float64(written) / downloadDuration.Seconds()
			output.WriteString(fmt.Sprintf("🚀 Speed:    %s/s\n", formatBytes(int64(speed))))
		}
		output.WriteString("═══════════════════════════════════════════════════════════════\n")
	} else {
		output.WriteString(fmt.Sprintf("✅ Downloaded: %s (%s)\n", filename, formatBytes(written)))
	}

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// formatBytes formats byte count as human readable string
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
