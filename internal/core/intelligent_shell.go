package core

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"suppercommand/internal/intelligence"

	prompt "github.com/c-bata/go-prompt"
	"github.com/fatih/color"
)

// IntelligentShell wraps the regular shell with intelligence features
type IntelligentShell struct {
	*Shell
	intelligence          *intelligence.IntelligenceEngine
	isIntelligenceEnabled bool
	lastCommand           string
	inputHistory          []string
}

// NewIntelligentShell creates a new intelligent shell
func NewIntelligentShell() *IntelligentShell {
	shell := NewShell()

	intelligenceEngine, err := intelligence.NewIntelligenceEngine()
	if err != nil {
		color.New(color.FgRed).Printf("❌ Failed to initialize intelligence: %v\n", err)
		return &IntelligentShell{
			Shell:                 shell,
			isIntelligenceEnabled: false,
		}
	}

	return &IntelligentShell{
		Shell:                 shell,
		intelligence:          intelligenceEngine,
		isIntelligenceEnabled: true,
		inputHistory:          make([]string, 0),
	}
}

// Initialize initializes the intelligent shell
func (is *IntelligentShell) Initialize() error {
	if is.isIntelligenceEnabled {
		ctx := context.Background()
		if err := is.intelligence.Initialize(ctx); err != nil {
			color.New(color.FgYellow).Printf("⚠️ Intelligence disabled: %v\n", err)
			is.isIntelligenceEnabled = false
		}
	}
	return nil
}

// getCleanPrompt returns a prompt without ANSI codes for go-prompt
func (is *IntelligentShell) getCleanPrompt() (string, bool) {
	cwd, _ := os.Getwd()
	shortPath := getShortenedPath(cwd)

	// Clean prompt without ANSI codes
	return fmt.Sprintf("🚀 SuperShell●[%s] ❯❯❯ ", shortPath), true
}

// RunWithIntelligence runs the shell with intelligent features using go-prompt
func (is *IntelligentShell) RunWithIntelligence() {
	if !is.isIntelligenceEnabled {
		is.Shell.Run()
		return
	}

	color.New(color.FgGreen).Println("🧠 SuperShell with Intelligence Engine activated!")
	color.New(color.FgHiBlack).Println("   Live completions as you type! TAB to complete, Ctrl+C to exit")
	fmt.Println()

	// Create intelligent prompt with clean prompt function
	p := prompt.New(
		is.intelligentExecutor,                     // Command executor
		is.intelligentCompleter,                    // Live completer
		prompt.OptionLivePrefix(is.getCleanPrompt), // Use clean prompt
		prompt.OptionTitle("SuperShell Intelligence"),
		prompt.OptionInputTextColor(prompt.White),
		prompt.OptionSuggestionBGColor(prompt.DarkBlue),        // Dark blue background
		prompt.OptionSuggestionTextColor(prompt.White),         // White text
		prompt.OptionSelectedSuggestionBGColor(prompt.Blue),    // Bright blue for selected
		prompt.OptionSelectedSuggestionTextColor(prompt.White), // White text for selected
		prompt.OptionPreviewSuggestionTextColor(prompt.Yellow), // Yellow for preview
		prompt.OptionDescriptionBGColor(prompt.Black),          // Black background for descriptions
		prompt.OptionDescriptionTextColor(prompt.Green),        // Green text for descriptions
		prompt.OptionScrollbarThumbColor(prompt.Cyan),          // Cyan scrollbar
		prompt.OptionScrollbarBGColor(prompt.DarkGray),         // Dark gray scrollbar background
		prompt.OptionMaxSuggestion(10),                         // Show up to 10 suggestions
		prompt.OptionShowCompletionAtStart(),                   // Show completions immediately
		prompt.OptionCompletionWordSeparator(" "),              // Use space as word separator
	)

	p.Run()
}

// intelligentExecutor handles command execution with intelligence
func (is *IntelligentShell) intelligentExecutor(input string) {
	input = strings.TrimSpace(input)
	if input == "" {
		return
	}

	// Handle exit commands
	if input == "exit" || input == "quit" {
		fmt.Println("👋 Goodbye!")
		os.Exit(0)
	}

	// Show smart suggestion for next command
	is.showSmartSuggestion()

	// Record command for learning
	is.recordCommand(input)

	// Handle special intelligence commands
	if is.handleSpecialCommands(input) {
		return
	}

	// Execute command
	startTime := time.Now()
	output := Dispatch(input)
	if output != "" {
		fmt.Println(output)
	}

	duration := time.Since(startTime)
	if duration > 3*time.Second {
		color.New(color.FgYellow).Printf("⏱️  Command took %v to complete\n", duration)
	}
}

// intelligentCompleter provides live completions as you type
func (is *IntelligentShell) intelligentCompleter(d prompt.Document) []prompt.Suggest {
	if !is.isIntelligenceEnabled {
		return []prompt.Suggest{}
	}

	// Get current input
	input := d.TextBeforeCursor()
	if len(input) < 1 {
		return is.getStartupSuggestions()
	}

	// Get intelligent completions
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	result, err := is.intelligence.GetCompletions(ctx, input, len(input))
	if err != nil || len(result.Completions) == 0 {
		return []prompt.Suggest{}
	}

	// Convert to prompt.Suggest format
	suggestions := make([]prompt.Suggest, 0)
	for _, comp := range result.Completions {
		if comp.Score > 30 { // Lower threshold for live feedback

			// Create description with category and score info
			description := comp.Description
			if comp.Category != "" {
				description += fmt.Sprintf(" (%s)", comp.Category)
			}

			// Add icon to the description instead of the text to avoid rendering issues
			if comp.Icon != "" {
				description = comp.Icon + " " + description
			}

			suggestions = append(suggestions, prompt.Suggest{
				Text:        comp.Text,
				Description: description,
			})
		}
	}

	// Limit to top 10 for performance
	if len(suggestions) > 10 {
		suggestions = suggestions[:10]
	}

	return suggestions
}

// getStartupSuggestions provides suggestions when starting to type
func (is *IntelligentShell) getStartupSuggestions() []prompt.Suggest {
	return []prompt.Suggest{
		{Text: "help", Description: "📖 Show all available commands"},
		{Text: "marketplace search", Description: "🏪 Search community plugins"},
		{Text: "dev reload", Description: "🔧 Hot reload development tools"},
		{Text: "smart", Description: "🤖 Show smart suggestions"},
		{Text: "ls", Description: "📁 List directory contents"},
		{Text: "cd", Description: "📂 Change directory"},
		{Text: "pwd", Description: "📍 Show current directory"},
		{Text: "clear", Description: "🧹 Clear the screen"},
	}
}

// handleSpecialCommands handles special intelligence commands
func (is *IntelligentShell) handleSpecialCommands(input string) bool {
	switch {
	case input == "smart" || input == "suggestions":
		is.showAllSmartSuggestions()
		return true
	case input == "intel" || input == "intelligence":
		is.showIntelligenceStats()
		return true
	case strings.HasPrefix(input, "learn "):
		command := strings.TrimPrefix(input, "learn ")
		is.recordCommand(command)
		color.New(color.FgGreen).Printf("✅ Command '%s' recorded for learning\n", command)
		return true
	case input == "test-intel":
		is.testIntelligenceSystem()
		return true
	}
	return false
}

// showSmartSuggestion shows workflow suggestions
func (is *IntelligentShell) showSmartSuggestion() {
	if !is.isIntelligenceEnabled || is.lastCommand == "" {
		return
	}

	suggestion := is.intelligence.GetSmartSuggestion(is.lastCommand)
	if suggestion != nil && suggestion.Confidence > 0.7 {
		fmt.Println()
		color.New(color.FgCyan).Printf("🤖 Smart Suggestion: %s\n", suggestion.Reason)
		color.New(color.FgYellow).Printf("   💡 Try: %s\n", suggestion.Command)
		fmt.Println()
	}
}

// showAllSmartSuggestions shows comprehensive smart suggestions
func (is *IntelligentShell) showAllSmartSuggestions() {
	fmt.Println()
	color.New(color.FgCyan, color.Bold).Println("🤖 SuperShell Smart Suggestions")
	fmt.Println(strings.Repeat("─", 50))

	// Context-aware suggestions
	contextDetector := is.intelligence.GetContextDetector()
	context := contextDetector.GetContext()

	// Git suggestions
	if context.GitRepository != nil && context.GitRepository.IsRepository {
		color.New(color.FgGreen).Println("📂 Git Repository Detected:")
		if context.GitRepository.HasUncommitted {
			fmt.Println("   🔄 git status - Check uncommitted changes")
			fmt.Println("   📝 git add . - Stage all changes")
			fmt.Println("   💾 git commit -m \"message\" - Commit changes")
		} else {
			fmt.Println("   ⬇️  git pull - Pull latest changes")
			fmt.Println("   🌿 git branch - List branches")
		}
		fmt.Println()
	}

	// Project type suggestions
	if context.ProjectType != "unknown" {
		color.New(color.FgBlue).Printf("🛠️  %s Project Detected:\n", strings.Title(context.ProjectType))
		switch context.ProjectType {
		case "go":
			fmt.Println("   🔨 go build - Build the project")
			fmt.Println("   🧪 go test - Run tests")
			fmt.Println("   📦 go mod tidy - Clean dependencies")
		case "node.js":
			fmt.Println("   📦 npm install - Install dependencies")
			fmt.Println("   🚀 npm start - Start the application")
			fmt.Println("   🧪 npm test - Run tests")
		case "python":
			fmt.Println("   📦 pip install -r requirements.txt")
			fmt.Println("   🐍 python main.py - Run the application")
		}
		fmt.Println()
	}

	// Tool suggestions
	if len(context.AvailableTools) > 0 {
		color.New(color.FgMagenta).Println("🔧 Available Tools:")
		for _, tool := range context.AvailableTools[:min(len(context.AvailableTools), 5)] {
			fmt.Printf("   ⚡ %s - Available for use\n", tool)
		}
		fmt.Println()
	}

	// Recent commands
	if len(is.inputHistory) > 0 {
		color.New(color.FgYellow).Println("🕒 Recent Commands:")
		seen := make(map[string]bool)
		count := 0
		for i := len(is.inputHistory) - 1; i >= 0 && count < 5; i-- {
			cmd := is.inputHistory[i]
			if !seen[cmd] && cmd != "" {
				fmt.Printf("   📝 %s\n", cmd)
				seen[cmd] = true
				count++
			}
		}
		fmt.Println()
	}

	color.New(color.FgHiBlack).Println("💡 Start typing any command to see live completions!")
	fmt.Println()
}

// showIntelligenceStats shows intelligence system statistics
func (is *IntelligentShell) showIntelligenceStats() {
	fmt.Println()
	color.New(color.FgCyan, color.Bold).Println("📊 SuperShell Intelligence Statistics")
	fmt.Println(strings.Repeat("─", 45))

	stats := is.GetIntelligenceStats()

	fmt.Printf("🧠 Intelligence: %s\n", color.New(color.FgGreen).Sprint("Enabled"))
	fmt.Printf("📝 Commands this session: %v\n", stats["commands_in_session"])
	fmt.Printf("🔄 Last command: %v\n", stats["last_command"])

	if learningStats := is.intelligence.GetLearningSystem().GetStats(); learningStats != nil {
		fmt.Printf("📚 Total commands learned: %v\n", learningStats["total_commands"])
		fmt.Printf("🔍 Unique commands: %v\n", learningStats["unique_commands"])
		fmt.Printf("🧩 Workflow patterns: %v\n", learningStats["workflow_patterns"])
	}

	context := is.intelligence.GetContextDetector().GetContext()
	fmt.Printf("📂 Current directory: %s\n", context.CurrentDirectory)
	fmt.Printf("🛠️  Project type: %s\n", context.ProjectType)
	fmt.Printf("🔧 Available tools: %d\n", len(context.AvailableTools))

	fmt.Println()
}

// testIntelligenceSystem tests the intelligence features
func (is *IntelligentShell) testIntelligenceSystem() {
	fmt.Println()
	color.New(color.FgYellow, color.Bold).Println("🧪 Testing Intelligence System")
	fmt.Println(strings.Repeat("─", 40))

	testCases := []string{"cd", "ls", "help", "ma", "dev", "doc", "git"}

	for _, testCase := range testCases {
		fmt.Printf("🔍 Testing completions for: '%s'\n", testCase)

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		result, err := is.intelligence.GetCompletions(ctx, testCase, len(testCase))
		cancel()

		if err != nil {
			color.New(color.FgRed).Printf("   ❌ Error: %v\n", err)
			continue
		}

		if len(result.Completions) == 0 {
			color.New(color.FgYellow).Println("   ⚠️ No completions found")
			continue
		}

		color.New(color.FgGreen).Printf("   ✅ Found %d completions:\n", len(result.Completions))
		for i, comp := range result.Completions[:min(3, len(result.Completions))] {
			fmt.Printf("      %d. %s %s (%.0f%%)\n", i+1, comp.Icon, comp.Display, comp.Score)
		}
		fmt.Println()
	}
}

// recordCommand records a command for learning
func (is *IntelligentShell) recordCommand(command string) {
	if !is.isIntelligenceEnabled {
		return
	}

	is.intelligence.RecordCommand(command)
	is.lastCommand = command
	is.inputHistory = append(is.inputHistory, command)

	if len(is.inputHistory) > 100 {
		is.inputHistory = is.inputHistory[len(is.inputHistory)-100:]
	}
}

// GetIntelligenceStats returns intelligence system statistics
func (is *IntelligentShell) GetIntelligenceStats() map[string]interface{} {
	if !is.isIntelligenceEnabled {
		return map[string]interface{}{"enabled": false}
	}

	stats := map[string]interface{}{
		"enabled":             true,
		"commands_in_session": len(is.inputHistory),
		"last_command":        is.lastCommand,
	}

	if learningStats := is.intelligence.GetLearningSystem().GetStats(); learningStats != nil {
		for k, v := range learningStats {
			stats[k] = v
		}
	}

	return stats
}

// Shutdown gracefully shuts down the intelligent shell
func (is *IntelligentShell) Shutdown() error {
	if is.isIntelligenceEnabled {
		return is.intelligence.Shutdown()
	}
	return nil
}
