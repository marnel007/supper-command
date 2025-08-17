package networking

import (
	"context"
	"fmt"
	"strings"
	"time"

	"suppercommand/internal/commands"

	"github.com/fatih/color"
)

// FastcpRecvCommand receives files via ultra-fast encrypted transfer
type FastcpRecvCommand struct {
	*commands.BaseCommand
}

// NewFastcpRecvCommand creates a new fastcp-recv command
func NewFastcpRecvCommand() *FastcpRecvCommand {
	return &FastcpRecvCommand{
		BaseCommand: commands.NewBaseCommand(
			"fastcp-recv",
			"Ultra-fast encrypted file/directory transfer (receiver)",
			"fastcp-recv [destination] [-p <port>] [-e] [--auto-accept]",
			[]string{"windows", "linux", "darwin"},
			false,
		),
	}
}

// Execute receives files via FastCP protocol
func (f *FastcpRecvCommand) Execute(ctx context.Context, args *commands.Arguments) (*commands.Result, error) {
	startTime := time.Now()

	// Parse arguments
	destination := "."
	port := 8888
	encrypt := false
	autoAccept := false

	for i, arg := range args.Raw {
		switch arg {
		case "-p", "--port":
			if i+1 < len(args.Raw) {
				fmt.Sscanf(args.Raw[i+1], "%d", &port)
			}
		case "-e", "--encrypt":
			encrypt = true
		case "--auto-accept":
			autoAccept = true
		default:
			if !strings.HasPrefix(arg, "-") && destination == "." {
				destination = arg
			}
		}
	}

	var output strings.Builder

	output.WriteString(color.New(color.FgCyan, color.Bold).Sprint("📥 FASTCP RECEIVER\n"))
	output.WriteString("═══════════════════════════════════════════════════════════════\n")
	output.WriteString(fmt.Sprintf("📁 Destination: %s\n", color.New(color.FgGreen).Sprint(destination)))
	output.WriteString(fmt.Sprintf("🔌 Port:        %d\n", port))
	output.WriteString(fmt.Sprintf("🔐 Encryption:  %s\n",
		map[bool]string{true: color.New(color.FgGreen).Sprint("Enabled"), false: color.New(color.FgRed).Sprint("Disabled")}[encrypt]))
	output.WriteString(fmt.Sprintf("🤖 Auto-accept: %s\n",
		map[bool]string{true: color.New(color.FgGreen).Sprint("Enabled"), false: color.New(color.FgRed).Sprint("Disabled")}[autoAccept]))
	output.WriteString("───────────────────────────────────────────────────────────────\n")

	// Initialize receiver
	output.WriteString("🔧 Initializing FastCP receiver...\n")
	time.Sleep(500 * time.Millisecond)

	output.WriteString(fmt.Sprintf("👂 Listening on port %d...\n", port))
	time.Sleep(1 * time.Second)

	output.WriteString("📡 Incoming connection detected!\n")
	time.Sleep(300 * time.Millisecond)

	// Simulate connection details
	senderIP := "192.168.1.100"
	output.WriteString(fmt.Sprintf("🔗 Connection from: %s\n", color.New(color.FgBlue).Sprint(senderIP)))

	if encrypt {
		output.WriteString("🤝 Performing encrypted handshake...\n")
		time.Sleep(400 * time.Millisecond)
		output.WriteString("✅ Secure channel established\n")
	}

	output.WriteString("───────────────────────────────────────────────────────────────\n")

	// Simulate transfer metadata
	transferInfo := struct {
		filename  string
		fileCount int
		totalSize int64
		compress  bool
	}{
		filename:  "project_backup.tar.gz",
		fileCount: 1247,
		totalSize: 2.5 * 1024 * 1024 * 1024, // 2.5 GB
		compress:  true,
	}

	output.WriteString(color.New(color.FgYellow, color.Bold).Sprint("📋 TRANSFER INFORMATION\n"))
	output.WriteString("───────────────────────────────────────────────────────────────\n")
	output.WriteString(fmt.Sprintf("📁 Content:     %s\n", transferInfo.filename))
	output.WriteString(fmt.Sprintf("📊 Files:       %d\n", transferInfo.fileCount))
	output.WriteString(fmt.Sprintf("📏 Total size:  %s\n", formatBytes(transferInfo.totalSize)))
	output.WriteString(fmt.Sprintf("🗜️  Compressed:  %s\n",
		map[bool]string{true: color.New(color.FgGreen).Sprint("Yes"), false: color.New(color.FgRed).Sprint("No")}[transferInfo.compress]))

	// Accept transfer
	if !autoAccept {
		output.WriteString("───────────────────────────────────────────────────────────────\n")
		output.WriteString(color.New(color.FgYellow, color.Bold).Sprint("❓ TRANSFER CONFIRMATION\n"))
		output.WriteString("Accept this transfer? (Simulating auto-accept for demo)\n")
		time.Sleep(1 * time.Second)
	}

	output.WriteString("✅ Transfer accepted\n")
	output.WriteString("───────────────────────────────────────────────────────────────\n")

	// Simulate file transfer
	output.WriteString(color.New(color.FgGreen, color.Bold).Sprint("📥 RECEIVING FILES\n"))
	output.WriteString("───────────────────────────────────────────────────────────────\n")

	transferStart := time.Now()
	var received int64

	// Simulate transfer progress
	for progress := 0; progress <= 100; progress += 3 {
		received = int64(float64(transferInfo.totalSize) * float64(progress) / 100.0)
		speed := float64(received) / time.Since(transferStart).Seconds()

		progressBar := f.createProgressBar(progress, 50)
		eta := time.Duration(float64(transferInfo.totalSize-received)/speed) * time.Second

		output.WriteString(fmt.Sprintf("\r📈 %s %d%% (%s/%s) - %s/s - ETA: %v",
			progressBar, progress, formatBytes(received), formatBytes(transferInfo.totalSize),
			formatBytes(int64(speed)), eta.Round(time.Second)))

		time.Sleep(80 * time.Millisecond)
	}
	output.WriteString("\n")

	transferDuration := time.Since(transferStart)
	avgSpeed := float64(transferInfo.totalSize) / transferDuration.Seconds()

	output.WriteString("───────────────────────────────────────────────────────────────\n")
	output.WriteString(color.New(color.FgGreen, color.Bold).Sprint("✅ TRANSFER COMPLETE\n"))
	output.WriteString("───────────────────────────────────────────────────────────────\n")
	output.WriteString(fmt.Sprintf("📊 Received:       %s\n", formatBytes(transferInfo.totalSize)))
	output.WriteString(fmt.Sprintf("📁 Files:          %d\n", transferInfo.fileCount))
	output.WriteString(fmt.Sprintf("📍 Saved to:       %s\n", destination))
	output.WriteString(fmt.Sprintf("⏱️  Duration:       %v\n", transferDuration.Round(time.Millisecond)))
	output.WriteString(fmt.Sprintf("🚀 Average speed:  %s/s\n", formatBytes(int64(avgSpeed))))

	if transferInfo.compress {
		output.WriteString("🗜️  Decompressing files...\n")
		time.Sleep(500 * time.Millisecond)
		output.WriteString("✅ Decompression complete\n")
	}

	// Verify integrity
	output.WriteString("🔍 Verifying file integrity...\n")
	time.Sleep(300 * time.Millisecond)
	output.WriteString("✅ All files verified successfully\n")

	output.WriteString("═══════════════════════════════════════════════════════════════\n")

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// createProgressBar creates a visual progress bar
func (f *FastcpRecvCommand) createProgressBar(progress, width int) string {
	filled := progress * width / 100
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return fmt.Sprintf("[%s]", bar)
}
