# Save this as setup.ps1 and run it
Write-Host "🔧 Setting up SuperShell..." -ForegroundColor Cyan

# Clean current directory
Remove-Item -Path "*.go", "go.mod", "go.sum" -ErrorAction SilentlyContinue

# Create structure
New-Item -ItemType Directory -Path "cmd\supershell" -Force | Out-Null

# Create go.mod
@"
module github.com/supershell/supershell

go 1.21

require (
	github.com/fatih/color v1.16.0
	github.com/spf13/cobra v1.8.0
)
"@ | Out-File -FilePath "go.mod" -Encoding UTF8

# Create main.go
@"
package main

import (
	"fmt"
	"os"
	"strings"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "supershell",
		Short: "SuperShell - Next-generation CLI",
		Run:   runShell,
	}
	rootCmd.Execute()
}

func runShell(cmd *cobra.Command, args []string) {
	green := color.New(color.FgGreen)
	cyan := color.New(color.FgCyan)
	
	green.Println("🚀 SuperShell v1.0.0 - Ready!")
	fmt.Println("Type 'help' for commands, 'exit' to quit")
	
	for {
		cyan.Print("supershell> ")
		var input string
		fmt.Scanln(&input)
		
		switch strings.ToLower(strings.TrimSpace(input)) {
		case "exit", "quit":
			green.Println("👋 Goodbye!")
			return
		case "help":
			fmt.Println("Commands: help, version, clear, echo <text>, exit")
		case "version":
			green.Println("SuperShell v1.0.0")
		case "clear":
			fmt.Print("\033[2J\033[H")
		default:
			if strings.HasPrefix(input, "echo ") {
				fmt.Println(strings.TrimPrefix(input, "echo "))
			} else if input != "" {
				fmt.Printf("✅ Executed: %s\n", input)
			}
		}
	}
}
"@ | Out-File -FilePath "cmd\supershell\main.go" -Encoding UTF8

# Build
go mod download
go build -o supershell.exe .\cmd\supershell

Write-Host "✅ SuperShell built successfully!" -ForegroundColor Green
Write-Host "🚀 Run with: .\supershell.exe" -ForegroundColor Cyan