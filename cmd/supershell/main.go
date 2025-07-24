package main

import (
	"os"
	"strings"
	"suppercommand/internal/core"

	"github.com/fatih/color"
)

func main() {
	// Create enhanced shell with Agent OS
	shell := core.NewEnhancedShell()

	// Initialize with error handling
	if err := shell.Initialize(); err != nil {
		color.New(color.FgRed).Printf("❌ Failed to initialize SuperShell: %v\n", err)
		os.Exit(1)
	}

	// Check for command-line execution (-c flag)
	if len(os.Args) >= 3 && os.Args[1] == "-c" {
		// Execute single command and exit
		command := strings.Join(os.Args[2:], " ")
		shell.ExecuteEnhanced(command)
		return
	}

	// Run intelligent interactive shell
	shell.RunIntelligent() // ← Use intelligent shell instead of Run()
}
