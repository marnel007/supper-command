package main

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"suppercommand/internal/app"

	"github.com/fatih/color"
)

func main() {
	// Create application context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create and initialize application
	application := app.NewApplication()
	if err := application.Initialize(ctx); err != nil {
		color.New(color.FgRed).Printf("❌ Failed to initialize SuperShell: %v\n", err)
		os.Exit(1)
	}

	// Check for command-line execution (-c flag)
	if len(os.Args) >= 3 && os.Args[1] == "-c" {
		// Execute single command and exit
		command := strings.Join(os.Args[2:], " ")
		result, err := application.ExecuteCommand(ctx, command)
		if err != nil {
			color.New(color.FgRed).Printf("❌ Command failed: %v\n", err)
			os.Exit(1)
		}
		if result.Output != "" {
			color.New(color.FgWhite).Println(result.Output)
		}
		return
	}

	// Start application in background
	go func() {
		if err := application.Run(ctx); err != nil {
			color.New(color.FgRed).Printf("❌ Application error: %v\n", err)
			cancel()
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	color.New(color.FgYellow).Println("\n🛑 Shutdown signal received...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := application.Shutdown(shutdownCtx); err != nil {
		color.New(color.FgRed).Printf("❌ Shutdown error: %v\n", err)
		os.Exit(1)
	}

	color.New(color.FgGreen).Println("👋 SuperShell shutdown complete")
}
