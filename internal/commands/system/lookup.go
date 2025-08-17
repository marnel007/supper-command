package system

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"suppercommand/internal/commands"

	"github.com/fatih/color"
)

// LookupCommand provides intelligent command discovery and suggestions
type LookupCommand struct {
	*commands.BaseCommand
	registry *commands.Registry
}

// NewLookupCommand creates a new lookup command
func NewLookupCommand(registry *commands.Registry) *LookupCommand {
	return &LookupCommand{
		BaseCommand: commands.NewBaseCommand(
			"lookup",
			"Intelligent command discovery and suggestions",
			"lookup [-m] [-s] [-c] [-t <task>] [query]",
			[]string{"windows", "linux", "darwin"},
			false,
		),
		registry: registry,
	}
}

// Execute performs intelligent command lookup
func (l *LookupCommand) Execute(ctx context.Context, args *commands.Arguments) (*commands.Result, error) {
	startTime := time.Now()

	// Parse arguments
	showSimilar := false
	showCategories := false
	showMenu := false
	taskBased := ""
	query := ""
	menuSelection := 0

	for i, arg := range args.Raw {
		switch arg {
		case "-s", "--similar":
			showSimilar = true
		case "-c", "--categories":
			showCategories = true
		case "-m", "--menu":
			showMenu = true
		case "-t", "--task":
			if i+1 < len(args.Raw) {
				taskBased = args.Raw[i+1]
			}
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			// Handle numeric menu selections
			fmt.Sscanf(arg, "%d", &menuSelection)
		case "menu-tasks", "menu-popular", "menu-search":
			// Handle special menu commands
			query = arg
		default:
			if !strings.HasPrefix(arg, "-") && arg != taskBased {
				query = arg
			}
		}
	}

	var output strings.Builder
	output.WriteString(color.New(color.FgCyan, color.Bold).Sprint("🔍 INTELLIGENT COMMAND LOOKUP\n"))
	output.WriteString("═══════════════════════════════════════════════════════════════\n")

	// Handle menu selections
	if menuSelection > 0 {
		return l.handleMenuSelection(menuSelection, startTime)
	}

	// Handle special menu commands
	if query == "menu-tasks" {
		return l.showTaskSubmenu(startTime)
	}
	if query == "menu-popular" {
		return l.showPopularCommands(startTime)
	}
	if query == "menu-search" {
		return l.showSearchMenu(startTime)
	}

	if showMenu {
		return l.showInteractiveMenu(startTime)
	}

	if taskBased != "" {
		return l.taskBasedLookup(taskBased, startTime)
	}

	if showCategories {
		return l.showCategories(startTime)
	}

	if query == "" {
		return l.showUsage(startTime)
	}

	// Perform intelligent lookup
	results := l.performLookup(query, showSimilar)

	// Show all matching commands in a comprehensive list
	allMatches := []CommandInfo{}
	allMatches = append(allMatches, results.ExactMatches...)
	allMatches = append(allMatches, results.PartialMatches...)
	allMatches = append(allMatches, results.SimilarCommands...)

	if len(allMatches) > 0 {
		output.WriteString(color.New(color.FgGreen, color.Bold).Sprintf("📋 COMMANDS MATCHING '%s' (%d found)\n", query, len(allMatches)))
		output.WriteString("───────────────────────────────────────────────────────────────\n")

		// Sort matches by relevance (exact first, then partial, then similar)
		exactCount := len(results.ExactMatches)
		partialCount := len(results.PartialMatches)

		for i, match := range allMatches {
			var prefix string
			var nameColor *color.Color

			if i < exactCount {
				prefix = "🎯"
				nameColor = color.New(color.FgGreen, color.Bold)
			} else if i < exactCount+partialCount {
				prefix = "📝"
				nameColor = color.New(color.FgYellow, color.Bold)
			} else {
				prefix = "🔗"
				nameColor = color.New(color.FgBlue, color.Bold)
			}

			output.WriteString(fmt.Sprintf("  %s %-12s - %s\n",
				prefix,
				nameColor.Sprint(match.Name),
				match.Description))

			// Show usage for exact matches
			if i < exactCount && match.Usage != "" {
				output.WriteString(fmt.Sprintf("     Usage: %s\n",
					color.New(color.FgCyan).Sprint(match.Usage)))
			}
		}
		output.WriteString("\n")

		// Show legend
		output.WriteString(color.New(color.FgHiBlack).Sprint("Legend: 🎯 Exact match  📝 Partial match  🔗 Similar command\n"))
		output.WriteString("\n")
	}

	if len(results.Suggestions) > 0 {
		output.WriteString(color.New(color.FgMagenta, color.Bold).Sprint("💡 SMART SUGGESTIONS\n"))
		output.WriteString("───────────────────────────────────────────────────────────────\n")
		for _, suggestion := range results.Suggestions {
			output.WriteString(fmt.Sprintf("  %s\n", color.New(color.FgMagenta).Sprint(suggestion)))
		}
		output.WriteString("\n")
	}

	if len(results.ExactMatches) == 0 && len(results.PartialMatches) == 0 && len(results.SimilarCommands) == 0 {
		output.WriteString(color.New(color.FgRed).Sprint("❌ No matches found for: ") + query + "\n\n")
		output.WriteString("💡 Try:\n")
		output.WriteString("  - lookup -m (interactive menu)\n")
		output.WriteString("  - lookup -c (show categories)\n")
		output.WriteString("  - lookup -t <task> (task-based lookup)\n")
		output.WriteString("  - lookup -s <partial_name> (similar commands)\n")
	}

	// Add interactive suggestion
	if len(results.ExactMatches) > 0 || len(results.PartialMatches) > 0 || len(results.SimilarCommands) > 0 {
		output.WriteString("💡 Use 'lookup -m' for an interactive menu to explore commands\n")
	}

	output.WriteString("═══════════════════════════════════════════════════════════════\n")

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// LookupResult contains the results of a command lookup
type LookupResult struct {
	ExactMatches    []CommandInfo
	PartialMatches  []CommandInfo
	SimilarCommands []CommandInfo
	Suggestions     []string
}

// CommandInfo contains information about a command
type CommandInfo struct {
	Name        string
	Description string
	Category    string
	Usage       string
}

// performLookup performs the actual intelligent lookup
func (l *LookupCommand) performLookup(query string, includeSimilar bool) LookupResult {
	result := LookupResult{}
	query = strings.ToLower(query)

	allCommands := l.registry.GetAllCommands()

	for _, cmd := range allCommands {
		cmdInfo := CommandInfo{
			Name:        cmd.Name(),
			Description: cmd.Description(),
			Category:    l.getCommandCategory(cmd.Name()),
			Usage:       cmd.Usage(),
		}

		cmdName := strings.ToLower(cmd.Name())
		cmdDesc := strings.ToLower(cmd.Description())

		// Exact match
		if cmdName == query {
			result.ExactMatches = append(result.ExactMatches, cmdInfo)
			continue
		}

		// Partial name match
		if strings.Contains(cmdName, query) {
			result.PartialMatches = append(result.PartialMatches, cmdInfo)
			continue
		}

		// Description match
		if strings.Contains(cmdDesc, query) {
			result.PartialMatches = append(result.PartialMatches, cmdInfo)
			continue
		}

		// Similar commands (fuzzy matching) - always check for better user experience
		if l.isSimilar(cmdName, query) {
			result.SimilarCommands = append(result.SimilarCommands, cmdInfo)
		}
	}

	// Generate smart suggestions
	result.Suggestions = l.generateSuggestions(query)

	return result
}

// taskBasedLookup provides task-based command suggestions
func (l *LookupCommand) taskBasedLookup(task string, startTime time.Time) (*commands.Result, error) {
	var output strings.Builder
	output.WriteString(color.New(color.FgCyan, color.Bold).Sprint("🎯 TASK-BASED COMMAND LOOKUP\n"))
	output.WriteString("═══════════════════════════════════════════════════════════════\n")
	output.WriteString(fmt.Sprintf("Task: %s\n\n", color.New(color.FgYellow, color.Bold).Sprint(task)))

	suggestions := l.getTaskSuggestions(strings.ToLower(task))

	if len(suggestions) == 0 {
		output.WriteString(color.New(color.FgRed).Sprint("❌ No specific suggestions for this task.\n"))
		output.WriteString("💡 Try common tasks: network, file, system, security, monitoring\n")
	} else {
		output.WriteString(color.New(color.FgGreen, color.Bold).Sprint("📋 RECOMMENDED COMMANDS\n"))
		output.WriteString("───────────────────────────────────────────────────────────────\n")

		for _, suggestion := range suggestions {
			output.WriteString(fmt.Sprintf("  %s - %s\n",
				color.New(color.FgGreen, color.Bold).Sprint(suggestion.Command),
				suggestion.Description))
			if suggestion.Example != "" {
				output.WriteString(fmt.Sprintf("    Example: %s\n",
					color.New(color.FgBlue).Sprint(suggestion.Example)))
			}
			output.WriteString("\n")
		}
	}

	output.WriteString("═══════════════════════════════════════════════════════════════\n")

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// showCategories displays all command categories
func (l *LookupCommand) showCategories(startTime time.Time) (*commands.Result, error) {
	var output strings.Builder
	output.WriteString(color.New(color.FgCyan, color.Bold).Sprint("📂 COMMAND CATEGORIES\n"))
	output.WriteString("═══════════════════════════════════════════════════════════════\n")

	categories := map[string][]string{
		"System Commands":     {},
		"Filesystem Commands": {},
		"Networking Commands": {},
		"Advanced Tools":      {},
	}

	allCommands := l.registry.GetAllCommands()
	for _, cmd := range allCommands {
		category := l.getCommandCategory(cmd.Name())
		categories[category] = append(categories[category], cmd.Name())
	}

	for category, commands := range categories {
		if len(commands) == 0 {
			continue
		}

		sort.Strings(commands)
		output.WriteString(color.New(color.FgGreen, color.Bold).Sprintf("📁 %s (%d commands)\n", category, len(commands)))
		output.WriteString("───────────────────────────────────────────────────────────────\n")

		for i, cmd := range commands {
			if i > 0 && i%6 == 0 {
				output.WriteString("\n")
			}
			output.WriteString(fmt.Sprintf("%-12s ", cmd))
		}
		output.WriteString("\n\n")
	}

	output.WriteString("💡 Use 'lookup <command_name>' to get details about specific commands\n")
	output.WriteString("═══════════════════════════════════════════════════════════════\n")

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// showUsage displays usage information
func (l *LookupCommand) showUsage(startTime time.Time) (*commands.Result, error) {
	var output strings.Builder
	output.WriteString(color.New(color.FgCyan, color.Bold).Sprint("🔍 INTELLIGENT LOOKUP USAGE\n"))
	output.WriteString("═══════════════════════════════════════════════════════════════\n")

	examples := []struct {
		command     string
		description string
	}{
		{"lookup -m", "Interactive menu with dropdown-style navigation"},
		{"lookup ping", "Find commands related to 'ping'"},
		{"lookup network", "Find all network-related commands"},
		{"lookup -c", "Show all command categories"},
		{"lookup -t network", "Get commands for network tasks"},
		{"lookup -t file", "Get commands for file operations"},
		{"lookup -s net", "Find similar commands to 'net'"},
		{"lookup copy", "Find commands related to copying"},
		{"lookup -t security", "Get security-related commands"},
	}

	output.WriteString(color.New(color.FgGreen, color.Bold).Sprint("💡 USAGE EXAMPLES\n"))
	output.WriteString("───────────────────────────────────────────────────────────────\n")

	for _, example := range examples {
		output.WriteString(fmt.Sprintf("  %s\n", color.New(color.FgBlue).Sprint(example.command)))
		output.WriteString(fmt.Sprintf("    %s\n\n", example.description))
	}

	output.WriteString("═══════════════════════════════════════════════════════════════\n")

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// TaskSuggestion represents a task-based command suggestion
type TaskSuggestion struct {
	Command     string
	Description string
	Example     string
}

// getTaskSuggestions returns command suggestions based on task
func (l *LookupCommand) getTaskSuggestions(task string) []TaskSuggestion {
	taskMap := map[string][]TaskSuggestion{
		"network": {
			{"ping", "Test network connectivity", "ping google.com"},
			{"nslookup", "DNS lookup", "nslookup example.com"},
			{"tracert", "Trace network route", "tracert google.com"},
			{"netstat", "Show network connections", "netstat -an"},
			{"portscan", "Scan network ports", "portscan 192.168.1.1"},
			{"speedtest", "Test internet speed", "speedtest"},
			{"wget", "Download files", "wget https://example.com/file.zip"},
		},
		"file": {
			{"ls", "List files", "ls -la"},
			{"cp", "Copy files", "cp file.txt backup.txt"},
			{"mv", "Move/rename files", "mv old.txt new.txt"},
			{"rm", "Remove files", "rm unwanted.txt"},
			{"cat", "View file contents", "cat readme.txt"},
			{"mkdir", "Create directories", "mkdir new_folder"},
		},
		"system": {
			{"sysinfo", "System information", "sysinfo -v"},
			{"killtask", "Terminate processes", "killtask notepad"},
			{"whoami", "Current user", "whoami"},
			{"hostname", "System hostname", "hostname"},
			{"ver", "Version information", "ver -v"},
		},
		"security": {
			{"sniff", "Packet capture", "sniff -c 10 -p HTTP"},
			{"portscan", "Port scanning", "portscan target.com"},
			{"netstat", "Network monitoring", "netstat -an"},
			{"killtask", "Process management", "killtask suspicious_process"},
		},
		"monitoring": {
			{"sysinfo", "System monitoring", "sysinfo --cpu"},
			{"netstat", "Network monitoring", "netstat"},
			{"speedtest", "Performance testing", "speedtest"},
			{"ping", "Connectivity monitoring", "ping -c 10 server.com"},
		},
	}

	if suggestions, exists := taskMap[task]; exists {
		return suggestions
	}

	return []TaskSuggestion{}
}

// getCommandCategory returns the category of a command
func (l *LookupCommand) getCommandCategory(name string) string {
	systemCommands := []string{"help", "clear", "sysinfo", "whoami", "hostname", "exit", "ver", "helphtml", "winupdate", "killtask", "lookup"}
	fsCommands := []string{"pwd", "ls", "dir", "echo", "cd", "cat", "mkdir", "rm", "rmdir", "cp", "mv"}
	advancedTools := []string{"fastcp-send", "fastcp-recv", "fastcp-backup", "fastcp-restore", "fastcp-dedup", "netdiscover", "sniff"}

	for _, cmd := range systemCommands {
		if cmd == name {
			return "System Commands"
		}
	}
	for _, cmd := range fsCommands {
		if cmd == name {
			return "Filesystem Commands"
		}
	}
	for _, cmd := range advancedTools {
		if cmd == name {
			return "Advanced Tools"
		}
	}
	return "Networking Commands"
}

// isSimilar checks if two strings are similar using enhanced fuzzy matching
func (l *LookupCommand) isSimilar(s1, s2 string) bool {
	if len(s1) == 0 || len(s2) == 0 {
		return false
	}

	// Check if one string is contained in the other (for cases like "pin" in "ping")
	if strings.Contains(s1, s2) || strings.Contains(s2, s1) {
		return true
	}

	// Check for common prefixes (for cases like "pin" and "ping")
	minLen := len(s1)
	if len(s2) < minLen {
		minLen = len(s2)
	}

	if minLen >= 2 {
		// Check if they share a common prefix of at least 2 characters
		if s1[:minLen] == s2[:minLen] {
			return true
		}
	}

	// Original substring matching for longer strings
	if minLen >= 3 {
		// Check for common substrings of length 3 or more
		for i := 0; i <= len(s1)-3; i++ {
			substr := s1[i : i+3]
			if strings.Contains(s2, substr) {
				return true
			}
		}
	}

	// Levenshtein-like distance check for short strings
	if len(s1) <= 5 && len(s2) <= 5 {
		return l.isCloseMatch(s1, s2)
	}

	return false
}

// isCloseMatch checks if two short strings are close matches (1-2 character difference)
func (l *LookupCommand) isCloseMatch(s1, s2 string) bool {
	if abs(len(s1)-len(s2)) > 2 {
		return false
	}

	// Simple edit distance check
	differences := 0
	maxLen := len(s1)
	if len(s2) > maxLen {
		maxLen = len(s2)
	}

	for i := 0; i < maxLen; i++ {
		if i >= len(s1) || i >= len(s2) {
			differences++
		} else if s1[i] != s2[i] {
			differences++
		}

		if differences > 2 {
			return false
		}
	}

	return differences <= 2
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// generateSuggestions generates smart suggestions based on query
func (l *LookupCommand) generateSuggestions(query string) []string {
	suggestions := []string{}

	// Common typos and alternatives
	alternatives := map[string][]string{
		"list":    {"ls", "dir"},
		"copy":    {"cp"},
		"move":    {"mv"},
		"delete":  {"rm"},
		"remove":  {"rm", "rmdir"},
		"network": {"ping", "netstat", "nslookup"},
		"info":    {"sysinfo", "whoami", "hostname"},
		"kill":    {"killtask"},
		"process": {"killtask", "sysinfo"},
		"file":    {"ls", "cat", "cp", "mv"},
		"test":    {"ping", "speedtest", "portscan"},
	}

	if alts, exists := alternatives[query]; exists {
		for _, alt := range alts {
			suggestions = append(suggestions, fmt.Sprintf("Try '%s' for %s operations", alt, query))
		}
	}

	// Add general suggestions
	if len(suggestions) == 0 {
		suggestions = append(suggestions, "Use 'lookup -c' to see all command categories")
		suggestions = append(suggestions, "Use 'lookup -t <task>' for task-based suggestions")
		suggestions = append(suggestions, "Use 'help' to see all available commands")
	}

	return suggestions
}

// showInteractiveMenu displays an interactive dropdown-style menu
func (l *LookupCommand) showInteractiveMenu(startTime time.Time) (*commands.Result, error) {
	var output strings.Builder
	output.WriteString(color.New(color.FgCyan, color.Bold).Sprint("📋 INTERACTIVE COMMAND MENU\n"))
	output.WriteString("═══════════════════════════════════════════════════════════════\n")

	// Main menu options
	menuOptions := []MenuOption{
		{
			ID:          1,
			Title:       "Browse by Category",
			Description: "Explore commands organized by functionality",
			Icon:        "📂",
		},
		{
			ID:          2,
			Title:       "Task-Based Lookup",
			Description: "Find commands for specific tasks",
			Icon:        "🎯",
		},
		{
			ID:          3,
			Title:       "All Commands List",
			Description: "View all available commands",
			Icon:        "📜",
		},
		{
			ID:          4,
			Title:       "Popular Commands",
			Description: "Most commonly used commands",
			Icon:        "⭐",
		},
		{
			ID:          5,
			Title:       "Search Commands",
			Description: "Search by name or description",
			Icon:        "🔍",
		},
	}

	output.WriteString(color.New(color.FgGreen, color.Bold).Sprint("🎮 INTERACTIVE DROPDOWN MENU - Select an option:\n"))
	output.WriteString("───────────────────────────────────────────────────────────────\n")

	for _, option := range menuOptions {
		output.WriteString(fmt.Sprintf("  %s [%d] %s\n",
			option.Icon,
			option.ID,
			color.New(color.FgWhite, color.Bold).Sprint(option.Title)))
		output.WriteString(fmt.Sprintf("      %s\n",
			color.New(color.FgHiBlack).Sprint(option.Description)))
		output.WriteString(fmt.Sprintf("      %s\n\n",
			color.New(color.FgGreen).Sprint("Command: lookup "+fmt.Sprintf("%d", option.ID))))
	}

	output.WriteString("───────────────────────────────────────────────────────────────\n")
	output.WriteString(color.New(color.FgYellow, color.Bold).Sprint("🎯 HOW TO SELECT FROM DROPDOWN:\n"))
	output.WriteString("• Type: lookup 1  (Browse by Category)\n")
	output.WriteString("• Type: lookup 2  (Task-Based Lookup)\n")
	output.WriteString("• Type: lookup 3  (All Commands List)\n")
	output.WriteString("• Type: lookup 4  (Popular Commands)\n")
	output.WriteString("• Type: lookup 5  (Search Commands)\n\n")

	// Show popular commands (option 4 content)
	output.WriteString(color.New(color.FgMagenta, color.Bold).Sprint("⭐ POPULAR COMMANDS (Option 4):\n"))
	output.WriteString("───────────────────────────────────────────────────────────────\n")

	popularCommands := l.getPopularCommands()
	for i, cmd := range popularCommands {
		output.WriteString(fmt.Sprintf("  [%d] %s - %s\n",
			i+1,
			color.New(color.FgCyan, color.Bold).Sprint(cmd.Name),
			cmd.Description))
		if cmd.Example != "" {
			output.WriteString(fmt.Sprintf("      Example: %s\n",
				color.New(color.FgBlue).Sprint(cmd.Example)))
		}
		output.WriteString("\n")
	}

	// Show quick task menu
	output.WriteString(color.New(color.FgBlue, color.Bold).Sprint("🎯 QUICK TASK MENU (Option 2):\n"))
	output.WriteString("───────────────────────────────────────────────────────────────\n")

	quickTasks := []struct {
		task        string
		description string
		command     string
	}{
		{"network", "Network operations and troubleshooting", "lookup -t network"},
		{"file", "File and directory operations", "lookup -t file"},
		{"system", "System information and management", "lookup -t system"},
		{"security", "Security and monitoring tools", "lookup -t security"},
		{"monitoring", "Performance and system monitoring", "lookup -t monitoring"},
	}

	for i, task := range quickTasks {
		output.WriteString(fmt.Sprintf("  [%d] %s - %s\n",
			i+1,
			color.New(color.FgYellow, color.Bold).Sprint(strings.Title(task.task)),
			task.description))
		output.WriteString(fmt.Sprintf("      Command: %s\n\n",
			color.New(color.FgGreen).Sprint(task.command)))
	}

	output.WriteString("═══════════════════════════════════════════════════════════════\n")
	output.WriteString(color.New(color.FgHiGreen, color.Bold).Sprint("🚀 Next Steps:\n"))
	output.WriteString("• Copy and run any of the commands shown above\n")
	output.WriteString("• Use 'help <command>' for detailed information about specific commands\n")
	output.WriteString("• Use 'lookup <search_term>' to search for specific functionality\n")
	output.WriteString("═══════════════════════════════════════════════════════════════\n")

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// MenuOption represents a menu option in the interactive menu
type MenuOption struct {
	ID          int
	Title       string
	Description string
	Icon        string
	Command     string
}

// PopularCommand represents a popular command with example
type PopularCommand struct {
	Name        string
	Description string
	Example     string
}

// getPopularCommands returns a list of the most popular/useful commands
func (l *LookupCommand) getPopularCommands() []PopularCommand {
	return []PopularCommand{
		{
			Name:        "ping",
			Description: "Test network connectivity",
			Example:     "ping google.com",
		},
		{
			Name:        "ls",
			Description: "List files and directories",
			Example:     "ls -la",
		},
		{
			Name:        "sysinfo",
			Description: "Display system information",
			Example:     "sysinfo -v",
		},
		{
			Name:        "help",
			Description: "Get help for any command",
			Example:     "help ping",
		},
		{
			Name:        "nslookup",
			Description: "DNS lookup and resolution",
			Example:     "nslookup google.com",
		},
		{
			Name:        "netstat",
			Description: "Show network connections",
			Example:     "netstat -an",
		},
		{
			Name:        "speedtest",
			Description: "Test internet connection speed",
			Example:     "speedtest",
		},
		{
			Name:        "portscan",
			Description: "Scan network ports",
			Example:     "portscan 192.168.1.1",
		},
		{
			Name:        "wget",
			Description: "Download files from URLs",
			Example:     "wget https://example.com/file.zip",
		},
		{
			Name:        "killtask",
			Description: "Terminate processes",
			Example:     "killtask notepad",
		},
	}
}

// handleMenuSelection handles numeric menu selections (1-5)
func (l *LookupCommand) handleMenuSelection(selection int, startTime time.Time) (*commands.Result, error) {
	switch selection {
	case 1:
		// Browse by Category
		return l.showCategories(startTime)
	case 2:
		// Task-Based Lookup
		return l.showTaskSubmenu(startTime)
	case 3:
		// All Commands List
		return l.showAllCommandsList(startTime)
	case 4:
		// Popular Commands
		return l.showPopularCommands(startTime)
	case 5:
		// Search Commands
		return l.showSearchMenu(startTime)
	default:
		var output strings.Builder
		output.WriteString(color.New(color.FgRed).Sprint("❌ Invalid menu selection: ") + fmt.Sprintf("%d", selection) + "\n\n")
		output.WriteString("Valid options are 1-5. Use 'lookup -m' to see the menu.\n")
		return &commands.Result{
			Output:   output.String(),
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}
}

// showTaskSubmenu shows the task-based submenu
func (l *LookupCommand) showTaskSubmenu(startTime time.Time) (*commands.Result, error) {
	var output strings.Builder
	output.WriteString(color.New(color.FgCyan, color.Bold).Sprint("🎯 TASK-BASED COMMAND LOOKUP MENU\n"))
	output.WriteString("═══════════════════════════════════════════════════════════════\n")

	tasks := []struct {
		id          int
		name        string
		description string
		icon        string
		command     string
	}{
		{1, "Network", "Network operations and troubleshooting", "🌐", "lookup -t network"},
		{2, "File", "File and directory operations", "📁", "lookup -t file"},
		{3, "System", "System information and management", "🖥️", "lookup -t system"},
		{4, "Security", "Security and monitoring tools", "🔒", "lookup -t security"},
		{5, "Monitoring", "Performance and system monitoring", "📊", "lookup -t monitoring"},
	}

	output.WriteString(color.New(color.FgGreen, color.Bold).Sprint("📋 SELECT A TASK CATEGORY:\n"))
	output.WriteString("───────────────────────────────────────────────────────────────\n")

	for _, task := range tasks {
		output.WriteString(fmt.Sprintf("  %s [%d] %s\n",
			task.icon,
			task.id,
			color.New(color.FgWhite, color.Bold).Sprint(task.name)))
		output.WriteString(fmt.Sprintf("      %s\n",
			color.New(color.FgHiBlack).Sprint(task.description)))
		output.WriteString(fmt.Sprintf("      %s\n\n",
			color.New(color.FgGreen).Sprint("Command: "+task.command)))
	}

	output.WriteString("───────────────────────────────────────────────────────────────\n")
	output.WriteString(color.New(color.FgYellow, color.Bold).Sprint("💡 HOW TO SELECT:\n"))
	output.WriteString("• Copy and paste any command above\n")
	output.WriteString("• Or use shortcuts: lookup -t network, lookup -t file, etc.\n")
	output.WriteString("• Back to main menu: lookup -m\n")
	output.WriteString("═══════════════════════════════════════════════════════════════\n")

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// showPopularCommands shows the popular commands submenu
func (l *LookupCommand) showPopularCommands(startTime time.Time) (*commands.Result, error) {
	var output strings.Builder
	output.WriteString(color.New(color.FgCyan, color.Bold).Sprint("⭐ POPULAR COMMANDS MENU\n"))
	output.WriteString("═══════════════════════════════════════════════════════════════\n")

	popularCommands := l.getPopularCommands()

	output.WriteString(color.New(color.FgGreen, color.Bold).Sprint("🎯 MOST COMMONLY USED COMMANDS:\n"))
	output.WriteString("───────────────────────────────────────────────────────────────\n")

	for i, cmd := range popularCommands {
		output.WriteString(fmt.Sprintf("  [%d] %s - %s\n",
			i+1,
			color.New(color.FgCyan, color.Bold).Sprint(cmd.Name),
			cmd.Description))
		if cmd.Example != "" {
			output.WriteString(fmt.Sprintf("      %s\n",
				color.New(color.FgBlue).Sprint("Example: "+cmd.Example)))
		}
		output.WriteString(fmt.Sprintf("      %s\n\n",
			color.New(color.FgGreen).Sprint("Try: "+cmd.Example)))
	}

	output.WriteString("───────────────────────────────────────────────────────────────\n")
	output.WriteString(color.New(color.FgYellow, color.Bold).Sprint("💡 HOW TO USE:\n"))
	output.WriteString("• Copy and paste any example command above\n")
	output.WriteString("• Use 'help <command>' for detailed information\n")
	output.WriteString("• Back to main menu: lookup -m\n")
	output.WriteString("═══════════════════════════════════════════════════════════════\n")

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// showSearchMenu shows the search submenu
func (l *LookupCommand) showSearchMenu(startTime time.Time) (*commands.Result, error) {
	var output strings.Builder
	output.WriteString(color.New(color.FgCyan, color.Bold).Sprint("🔍 SEARCH COMMANDS MENU\n"))
	output.WriteString("═══════════════════════════════════════════════════════════════\n")

	searchOptions := []struct {
		example     string
		description string
		result      string
	}{
		{"lookup ping", "Search for 'ping' command", "Find exact matches and related commands"},
		{"lookup network", "Search for network-related commands", "Find all commands with 'network' in name/description"},
		{"lookup copy", "Search for copy/file operations", "Find cp, mv, and related file commands"},
		{"lookup security", "Search for security tools", "Find sniff, portscan, and security commands"},
		{"lookup monitor", "Search for monitoring tools", "Find sysinfo, netstat, and monitoring commands"},
	}

	output.WriteString(color.New(color.FgGreen, color.Bold).Sprint("🎯 SEARCH EXAMPLES:\n"))
	output.WriteString("───────────────────────────────────────────────────────────────\n")

	for i, option := range searchOptions {
		output.WriteString(fmt.Sprintf("  [%d] %s\n",
			i+1,
			color.New(color.FgCyan, color.Bold).Sprint(option.example)))
		output.WriteString(fmt.Sprintf("      %s\n",
			color.New(color.FgHiBlack).Sprint(option.description)))
		output.WriteString(fmt.Sprintf("      %s\n\n",
			color.New(color.FgGreen).Sprint("Result: "+option.result)))
	}

	output.WriteString("───────────────────────────────────────────────────────────────\n")
	output.WriteString(color.New(color.FgYellow, color.Bold).Sprint("💡 SEARCH TIPS:\n"))
	output.WriteString("• Use single words for broader results: lookup network\n")
	output.WriteString("• Use specific terms for exact matches: lookup ping\n")
	output.WriteString("• Try common tasks: lookup copy, lookup delete, lookup list\n")
	output.WriteString("• Back to main menu: lookup -m\n")
	output.WriteString("═══════════════════════════════════════════════════════════════\n")

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// showAllCommandsList shows all available commands
func (l *LookupCommand) showAllCommandsList(startTime time.Time) (*commands.Result, error) {
	var output strings.Builder
	output.WriteString(color.New(color.FgCyan, color.Bold).Sprint("📜 ALL AVAILABLE COMMANDS\n"))
	output.WriteString("═══════════════════════════════════════════════════════════════\n")

	allCommands := l.registry.GetAllCommands()
	sort.Slice(allCommands, func(i, j int) bool {
		return allCommands[i].Name() < allCommands[j].Name()
	})

	output.WriteString(color.New(color.FgGreen, color.Bold).Sprintf("📋 COMPLETE COMMAND LIST (%d commands):\n", len(allCommands)))
	output.WriteString("───────────────────────────────────────────────────────────────\n")

	for i, cmd := range allCommands {
		output.WriteString(fmt.Sprintf("  [%d] %s - %s\n",
			i+1,
			color.New(color.FgCyan, color.Bold).Sprint(cmd.Name()),
			cmd.Description()))

		// Add usage for important commands
		if i < 10 || cmd.Name() == "ping" || cmd.Name() == "help" || cmd.Name() == "lookup" {
			output.WriteString(fmt.Sprintf("      %s\n",
				color.New(color.FgBlue).Sprint("Usage: "+cmd.Usage())))
		}
		output.WriteString("\n")
	}

	output.WriteString("───────────────────────────────────────────────────────────────\n")
	output.WriteString(color.New(color.FgYellow, color.Bold).Sprint("💡 NEXT STEPS:\n"))
	output.WriteString("• Use 'help <command>' for detailed information about any command\n")
	output.WriteString("• Try popular commands: ping, ls, sysinfo, netstat\n")
	output.WriteString("• Back to main menu: lookup -m\n")
	output.WriteString("═══════════════════════════════════════════════════════════════\n")

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}
