package system

import (
	"io/ioutil"

	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"suppercommand/internal/commands"

	"github.com/fatih/color"
)

// SmartHistoryCommand provides AI-powered command history features
type SmartHistoryCommand struct {
	*commands.BaseCommand
	historyFile string
	registry    *commands.Registry
}

// HistoryEntry represents a command history entry with metadata
type HistoryEntry struct {
	ID        int       `json:"id"`
	Command   string    `json:"command"`
	Timestamp time.Time `json:"timestamp"`
	Directory string    `json:"directory"`
	ExitCode  int       `json:"exit_code"`
	Duration  int64     `json:"duration_ms"`
	Tags      []string  `json:"tags"`
	Category  string    `json:"category"`
}

// CommandPattern represents a detected usage pattern
type CommandPattern struct {
	Pattern     string    `json:"pattern"`
	Commands    []string  `json:"commands"`
	Frequency   int       `json:"frequency"`
	LastUsed    time.Time `json:"last_used"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
}

// NewSmartHistoryCommand creates a new smart history command
func NewSmartHistoryCommand(registry *commands.Registry) *SmartHistoryCommand {
	homeDir, _ := os.UserHomeDir()
	historyFile := filepath.Join(homeDir, ".supershell_history.json")

	return &SmartHistoryCommand{
		BaseCommand: commands.NewBaseCommand(
			"history",
			"Smart command history with AI-powered search and analysis",
			"history [smart|patterns|suggest|timeline|export|stats] [query]",
			[]string{"windows", "linux", "darwin"},
			false,
		),
		historyFile: historyFile,
		registry:    registry,
	}
}

// Execute handles smart history operations
func (h *SmartHistoryCommand) Execute(ctx context.Context, args *commands.Arguments) (*commands.Result, error) {
	startTime := time.Now()

	if len(args.Raw) == 0 {
		return h.showBasicHistory(startTime)
	}

	subcommand := args.Raw[0]
	switch subcommand {
	case "smart":
		query := ""
		if len(args.Raw) > 1 {
			query = strings.Join(args.Raw[1:], " ")
		}
		return h.smartSearch(query, startTime)
	case "patterns":
		return h.showPatterns(startTime)
	case "suggest":
		return h.showSuggestions(startTime)
	case "timeline":
		return h.showTimeline(startTime)
	case "export":
		format := "json"
		if len(args.Raw) > 1 {
			format = args.Raw[1]
		}
		return h.exportHistory(format, startTime)
	case "stats":
		return h.showStatistics(startTime)
	case "clear":
		return h.clearHistory(startTime)
	case "add":
		if len(args.Raw) > 1 {
			command := strings.Join(args.Raw[1:], " ")
			return h.addToHistory(command, startTime)
		}
		return &commands.Result{
			Output:   "Usage: history add <command>",
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	default:
		return h.showBasicHistory(startTime)
	}
}

// showBasicHistory displays recent command history
func (h *SmartHistoryCommand) showBasicHistory(startTime time.Time) (*commands.Result, error) {
	entries, err := h.loadHistory()
	if err != nil {
		return &commands.Result{
			Output:   fmt.Sprintf("Error loading history: %v", err),
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	var output strings.Builder
	headerColor := color.New(color.FgHiCyan, color.Bold)
	idColor := color.New(color.FgHiBlack)
	commandColor := color.New(color.FgWhite)
	timeColor := color.New(color.FgHiBlack)

	output.WriteString(headerColor.Sprint("📚 Command History\n"))
	output.WriteString("═══════════════════════════════════════════════════════════════\n")

	// Show last 20 commands
	start := len(entries) - 20
	if start < 0 {
		start = 0
	}

	for i := start; i < len(entries); i++ {
		entry := entries[i]
		timeStr := entry.Timestamp.Format("15:04:05")

		output.WriteString(fmt.Sprintf("%s %s %s %s\n",
			idColor.Sprintf("%4d", entry.ID),
			timeColor.Sprint(timeStr),
			commandColor.Sprint(entry.Command),
			h.getStatusIcon(entry.ExitCode)))
	}

	output.WriteString("\n")
	output.WriteString(color.New(color.FgHiBlack).Sprint("💡 Try: history smart \"backup files\" | history patterns | history suggest\n"))

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// smartSearch performs natural language search on command history
func (h *SmartHistoryCommand) smartSearch(query string, startTime time.Time) (*commands.Result, error) {
	entries, err := h.loadHistory()
	if err != nil {
		return &commands.Result{
			Output:   fmt.Sprintf("Error loading history: %v", err),
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	if query == "" {
		return &commands.Result{
			Output:   "Usage: history smart <search query>\nExample: history smart \"backup files\"",
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	matches := h.performSmartSearch(entries, query)

	var output strings.Builder
	headerColor := color.New(color.FgHiYellow, color.Bold)
	queryColor := color.New(color.FgYellow, color.Underline)

	output.WriteString(headerColor.Sprint("🔍 Smart History Search\n"))
	output.WriteString("═══════════════════════════════════════════════════════════════\n")
	output.WriteString(fmt.Sprintf("🎯 Query: %s\n", queryColor.Sprint(query)))
	output.WriteString(fmt.Sprintf("📊 Found %d matches\n\n", len(matches)))

	if len(matches) == 0 {
		output.WriteString(color.New(color.FgHiBlack).Sprint("No matches found. Try different keywords or use 'history patterns' to see common patterns.\n"))
	} else {
		for _, match := range matches {
			h.formatHistoryEntry(&output, match, query)
		}
	}

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// showPatterns displays detected command patterns
func (h *SmartHistoryCommand) showPatterns(startTime time.Time) (*commands.Result, error) {
	entries, err := h.loadHistory()
	if err != nil {
		return &commands.Result{
			Output:   fmt.Sprintf("Error loading history: %v", err),
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	patterns := h.detectPatterns(entries)

	var output strings.Builder
	headerColor := color.New(color.FgHiMagenta, color.Bold)

	output.WriteString(headerColor.Sprint("🧠 Command Patterns\n"))
	output.WriteString("═══════════════════════════════════════════════════════════════\n")

	if len(patterns) == 0 {
		output.WriteString(color.New(color.FgHiBlack).Sprint("No patterns detected yet. Use more commands to build pattern recognition.\n"))
	} else {
		for i, pattern := range patterns {
			if i >= 10 { // Show top 10 patterns
				break
			}
			h.formatPattern(&output, pattern, i+1)
		}
	}

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// showSuggestions provides context-aware command suggestions
func (h *SmartHistoryCommand) showSuggestions(startTime time.Time) (*commands.Result, error) {
	entries, err := h.loadHistory()
	if err != nil {
		return &commands.Result{
			Output:   fmt.Sprintf("Error loading history: %v", err),
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	suggestions := h.generateSuggestions(entries)

	var output strings.Builder
	headerColor := color.New(color.FgHiGreen, color.Bold)

	output.WriteString(headerColor.Sprint("💡 Smart Suggestions\n"))
	output.WriteString("═══════════════════════════════════════════════════════════════\n")

	if len(suggestions) == 0 {
		output.WriteString(color.New(color.FgHiBlack).Sprint("No suggestions available. Build more command history for better suggestions.\n"))
	} else {
		for i, suggestion := range suggestions {
			if i >= 5 { // Show top 5 suggestions
				break
			}
			h.formatSuggestion(&output, suggestion, i+1)
		}
	}

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// showTimeline displays command history in timeline format
func (h *SmartHistoryCommand) showTimeline(startTime time.Time) (*commands.Result, error) {
	entries, err := h.loadHistory()
	if err != nil {
		return &commands.Result{
			Output:   fmt.Sprintf("Error loading history: %v", err),
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	var output strings.Builder
	headerColor := color.New(color.FgHiCyan, color.Bold)

	output.WriteString(headerColor.Sprint("📅 Command Timeline\n"))
	output.WriteString("═══════════════════════════════════════════════════════════════\n")

	// Group by date
	dateGroups := make(map[string][]HistoryEntry)
	for _, entry := range entries {
		dateKey := entry.Timestamp.Format("2006-01-02")
		dateGroups[dateKey] = append(dateGroups[dateKey], entry)
	}

	// Sort dates
	var dates []string
	for date := range dateGroups {
		dates = append(dates, date)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(dates)))

	// Show last 7 days
	for i, date := range dates {
		if i >= 7 {
			break
		}
		h.formatTimelineDate(&output, date, dateGroups[date])
	}

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// showStatistics displays command usage statistics
func (h *SmartHistoryCommand) showStatistics(startTime time.Time) (*commands.Result, error) {
	entries, err := h.loadHistory()
	if err != nil {
		return &commands.Result{
			Output:   fmt.Sprintf("Error loading history: %v", err),
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	stats := h.calculateStatistics(entries)

	var output strings.Builder
	headerColor := color.New(color.FgHiYellow, color.Bold)

	output.WriteString(headerColor.Sprint("📊 Command Statistics\n"))
	output.WriteString("═══════════════════════════════════════════════════════════════\n")

	h.formatStatistics(&output, stats, entries)

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// exportHistory exports command history in specified format
func (h *SmartHistoryCommand) exportHistory(format string, startTime time.Time) (*commands.Result, error) {
	entries, err := h.loadHistory()
	if err != nil {
		return &commands.Result{
			Output:   fmt.Sprintf("Error loading history: %v", err),
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	filename := fmt.Sprintf("supershell_history_%s.%s", time.Now().Format("20060102_150405"), format)

	switch format {
	case "json":
		return h.exportJSON(entries, filename, startTime)
	case "csv":
		return h.exportCSV(entries, filename, startTime)
	case "txt":
		return h.exportText(entries, filename, startTime)
	default:
		return &commands.Result{
			Output:   "Supported formats: json, csv, txt",
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}
}

// Helper methods for smart search and pattern detection
func (h *SmartHistoryCommand) performSmartSearch(entries []HistoryEntry, query string) []HistoryEntry {
	var matches []HistoryEntry
	queryLower := strings.ToLower(query)
	keywords := strings.Fields(queryLower)

	for _, entry := range entries {
		commandLower := strings.ToLower(entry.Command)
		score := 0

		// Exact match gets highest score
		if strings.Contains(commandLower, queryLower) {
			score += 10
		}

		// Keyword matching
		for _, keyword := range keywords {
			if strings.Contains(commandLower, keyword) {
				score += 3
			}
		}

		// Category matching
		if strings.Contains(strings.ToLower(entry.Category), queryLower) {
			score += 5
		}

		// Tag matching
		for _, tag := range entry.Tags {
			if strings.Contains(strings.ToLower(tag), queryLower) {
				score += 4
			}
		}

		if score > 0 {
			matches = append(matches, entry)
		}
	}

	// Sort by relevance (most recent first for same score)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Timestamp.After(matches[j].Timestamp)
	})

	return matches
}

// Additional helper methods would go here...
// (detectPatterns, generateSuggestions, formatters, etc.)

// loadHistory loads command history from file
func (h *SmartHistoryCommand) loadHistory() ([]HistoryEntry, error) {
	var entries []HistoryEntry

	if _, err := os.Stat(h.historyFile); os.IsNotExist(err) {
		return entries, nil // Return empty history if file doesn't exist
	}

	data, err := ioutil.ReadFile(h.historyFile)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return entries, nil
	}

	err = json.Unmarshal(data, &entries)
	return entries, err
}

// saveHistory saves command history to file
func (h *SmartHistoryCommand) saveHistory(entries []HistoryEntry) error {
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(h.historyFile, data, 0644)
}

// Helper methods for formatting and utilities
func (h *SmartHistoryCommand) getStatusIcon(exitCode int) string {
	if exitCode == 0 {
		return color.New(color.FgGreen).Sprint("✓")
	}
	return color.New(color.FgRed).Sprint("✗")
}

func (h *SmartHistoryCommand) formatHistoryEntry(output *strings.Builder, entry HistoryEntry, query string) {
	idColor := color.New(color.FgHiBlack)
	timeColor := color.New(color.FgHiBlack)

	// Highlight query terms in command
	command := entry.Command
	if query != "" {
		command = h.highlightQuery(command, query)
	}

	timeStr := entry.Timestamp.Format("01/02 15:04")
	output.WriteString(fmt.Sprintf("%s %s %s %s\n",
		idColor.Sprintf("%4d", entry.ID),
		timeColor.Sprint(timeStr),
		command,
		h.getStatusIcon(entry.ExitCode)))
}

func (h *SmartHistoryCommand) highlightQuery(text, query string) string {
	highlightColor := color.New(color.BgYellow, color.FgBlack)
	queryLower := strings.ToLower(query)
	textLower := strings.ToLower(text)

	if strings.Contains(textLower, queryLower) {
		// Simple highlighting - in a real implementation, you'd want more sophisticated highlighting
		return strings.ReplaceAll(text, query, highlightColor.Sprint(query))
	}
	return text
}

// detectPatterns analyzes command history to find usage patterns
func (h *SmartHistoryCommand) detectPatterns(entries []HistoryEntry) []CommandPattern {
	patterns := []CommandPattern{}

	// Group commands by category
	categoryGroups := make(map[string][]HistoryEntry)
	for _, entry := range entries {
		categoryGroups[entry.Category] = append(categoryGroups[entry.Category], entry)
	}

	// Detect sequential patterns
	patterns = append(patterns, h.detectSequentialPatterns(entries)...)

	// Detect frequency patterns
	patterns = append(patterns, h.detectFrequencyPatterns(entries)...)

	// Detect time-based patterns
	patterns = append(patterns, h.detectTimePatterns(entries)...)

	// Sort patterns by frequency
	sort.Slice(patterns, func(i, j int) bool {
		return patterns[i].Frequency > patterns[j].Frequency
	})

	return patterns
}

func (h *SmartHistoryCommand) detectSequentialPatterns(entries []HistoryEntry) []CommandPattern {
	patterns := []CommandPattern{}
	sequences := make(map[string]int)

	// Look for command sequences (2-3 commands in a row)
	for i := 0; i < len(entries)-1; i++ {
		if entries[i+1].Timestamp.Sub(entries[i].Timestamp) < 5*time.Minute {
			seq := fmt.Sprintf("%s → %s",
				strings.Fields(entries[i].Command)[0],
				strings.Fields(entries[i+1].Command)[0])
			sequences[seq]++
		}
	}

	for seq, freq := range sequences {
		if freq >= 3 { // Only patterns that occur 3+ times
			commands := strings.Split(seq, " → ")
			patterns = append(patterns, CommandPattern{
				Pattern:     seq,
				Commands:    commands,
				Frequency:   freq,
				LastUsed:    time.Now(),
				Category:    "sequential",
				Description: fmt.Sprintf("Common sequence: %s", seq),
			})
		}
	}

	return patterns
}

func (h *SmartHistoryCommand) detectFrequencyPatterns(entries []HistoryEntry) []CommandPattern {
	patterns := []CommandPattern{}
	cmdFreq := make(map[string]int)
	cmdLast := make(map[string]time.Time)

	for _, entry := range entries {
		cmd := strings.Fields(entry.Command)[0]
		cmdFreq[cmd]++
		if entry.Timestamp.After(cmdLast[cmd]) {
			cmdLast[cmd] = entry.Timestamp
		}
	}

	for cmd, freq := range cmdFreq {
		if freq >= 5 { // Commands used 5+ times
			patterns = append(patterns, CommandPattern{
				Pattern:     cmd,
				Commands:    []string{cmd},
				Frequency:   freq,
				LastUsed:    cmdLast[cmd],
				Category:    "frequent",
				Description: fmt.Sprintf("Frequently used command: %s", cmd),
			})
		}
	}

	return patterns
}

func (h *SmartHistoryCommand) detectTimePatterns(entries []HistoryEntry) []CommandPattern {
	patterns := []CommandPattern{}
	hourlyUsage := make(map[int]map[string]int)

	for _, entry := range entries {
		hour := entry.Timestamp.Hour()
		if hourlyUsage[hour] == nil {
			hourlyUsage[hour] = make(map[string]int)
		}
		cmd := strings.Fields(entry.Command)[0]
		hourlyUsage[hour][cmd]++
	}

	// Find commands that are predominantly used at specific times
	for hour, commands := range hourlyUsage {
		for cmd, freq := range commands {
			if freq >= 3 {
				patterns = append(patterns, CommandPattern{
					Pattern:     fmt.Sprintf("%s@%02d:00", cmd, hour),
					Commands:    []string{cmd},
					Frequency:   freq,
					LastUsed:    time.Now(),
					Category:    "temporal",
					Description: fmt.Sprintf("Often used at %02d:00: %s", hour, cmd),
				})
			}
		}
	}

	return patterns
}

func (h *SmartHistoryCommand) generateSuggestions(entries []HistoryEntry) []string {
	suggestions := []string{}

	if len(entries) == 0 {
		return suggestions
	}

	// Get current working directory
	cwd, _ := os.Getwd()

	// Recent command analysis
	recentCommands := h.getRecentCommands(entries, 10)

	// Context-based suggestions
	suggestions = append(suggestions, h.getContextSuggestions(cwd)...)

	// Pattern-based suggestions
	suggestions = append(suggestions, h.getPatternSuggestions(recentCommands)...)

	// Time-based suggestions
	suggestions = append(suggestions, h.getTimeSuggestions()...)

	// Remove duplicates and limit to 5
	suggestions = h.removeDuplicates(suggestions)
	if len(suggestions) > 5 {
		suggestions = suggestions[:5]
	}

	return suggestions
}

func (h *SmartHistoryCommand) getRecentCommands(entries []HistoryEntry, count int) []HistoryEntry {
	if len(entries) <= count {
		return entries
	}
	return entries[len(entries)-count:]
}

func (h *SmartHistoryCommand) getContextSuggestions(cwd string) []string {
	suggestions := []string{}

	if strings.Contains(strings.ToLower(cwd), "project") ||
		strings.Contains(strings.ToLower(cwd), "src") ||
		strings.Contains(strings.ToLower(cwd), "code") {
		suggestions = append(suggestions, "ls -la", "git status", "cat README.md")
	}

	if strings.Contains(strings.ToLower(cwd), "log") {
		suggestions = append(suggestions, "ls -lt | head -10", "cat *.log | tail -50")
	}

	if strings.Contains(strings.ToLower(cwd), "config") {
		suggestions = append(suggestions, "ls -la", "cat *.conf", "backup config")
	}

	return suggestions
}

func (h *SmartHistoryCommand) getPatternSuggestions(recent []HistoryEntry) []string {
	suggestions := []string{}

	// Check for common patterns in recent commands
	hasGit := false
	hasNetwork := false
	hasSystem := false

	for _, entry := range recent {
		cmd := strings.Fields(entry.Command)[0]
		switch cmd {
		case "git":
			hasGit = true
		case "ping", "tracert", "netstat", "nslookup":
			hasNetwork = true
		case "ps", "top", "df", "free":
			hasSystem = true
		}
	}

	if hasGit {
		suggestions = append(suggestions, "git log --oneline -10", "git diff", "git push")
	}

	if hasNetwork {
		suggestions = append(suggestions, "netstat -tulpn", "ping google.com", "speedtest")
	}

	if hasSystem {
		suggestions = append(suggestions, "htop", "df -h", "free -h")
	}

	return suggestions
}

func (h *SmartHistoryCommand) getTimeSuggestions() []string {
	suggestions := []string{}
	hour := time.Now().Hour()

	// Morning suggestions (6-12)
	if hour >= 6 && hour < 12 {
		suggestions = append(suggestions, "sysinfo", "server health", "git status")
	}

	// Afternoon suggestions (12-18)
	if hour >= 12 && hour < 18 {
		suggestions = append(suggestions, "perf analyze", "server services", "backup")
	}

	// Evening suggestions (18-22)
	if hour >= 18 && hour < 22 {
		suggestions = append(suggestions, "history stats", "cleanup", "git commit")
	}

	return suggestions
}

func (h *SmartHistoryCommand) removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}

func (h *SmartHistoryCommand) formatPattern(output *strings.Builder, pattern CommandPattern, index int) {
	patternColor := color.New(color.FgHiCyan)
	freqColor := color.New(color.FgHiYellow)
	descColor := color.New(color.FgHiBlack)

	output.WriteString(fmt.Sprintf("%s %s\n",
		patternColor.Sprintf("🔄 Pattern %d:", index),
		color.New(color.FgWhite, color.Bold).Sprint(pattern.Pattern)))

	output.WriteString(fmt.Sprintf("   %s %s\n",
		descColor.Sprint("Description:"),
		pattern.Description))

	output.WriteString(fmt.Sprintf("   %s %s times, last used %s\n\n",
		descColor.Sprint("Usage:"),
		freqColor.Sprint(pattern.Frequency),
		pattern.LastUsed.Format("01/02 15:04")))
}

func (h *SmartHistoryCommand) formatSuggestion(output *strings.Builder, suggestion string, index int) {
	suggestionColor := color.New(color.FgHiGreen)
	commandColor := color.New(color.FgHiCyan)

	output.WriteString(fmt.Sprintf("%s %s\n",
		suggestionColor.Sprintf("💡 Suggestion %d:", index),
		commandColor.Sprint(suggestion)))
}

func (h *SmartHistoryCommand) formatTimelineDate(output *strings.Builder, date string, entries []HistoryEntry) {
	dateColor := color.New(color.FgHiCyan, color.Bold)
	timeColor := color.New(color.FgHiBlack)
	commandColor := color.New(color.FgWhite)

	// Parse and format date
	parsedDate, _ := time.Parse("2006-01-02", date)
	formattedDate := parsedDate.Format("Monday, January 2, 2006")

	output.WriteString(fmt.Sprintf("%s %s (%d commands)\n",
		dateColor.Sprint("📅"),
		formattedDate,
		len(entries)))

	// Show timeline for the day
	for i, entry := range entries {
		if i >= 10 { // Limit to 10 commands per day
			output.WriteString(fmt.Sprintf("   ... and %d more commands\n", len(entries)-10))
			break
		}

		timeStr := entry.Timestamp.Format("15:04")
		connector := "├─"
		if i == len(entries)-1 || i == 9 {
			connector = "└─"
		}

		output.WriteString(fmt.Sprintf("  %s %s %s %s\n",
			color.New(color.FgHiBlack).Sprint(connector),
			timeColor.Sprint(timeStr),
			commandColor.Sprint(entry.Command),
			h.getStatusIcon(entry.ExitCode)))
	}
	output.WriteString("\n")
}

func (h *SmartHistoryCommand) calculateStatistics(entries []HistoryEntry) map[string]interface{} {
	stats := make(map[string]interface{})

	if len(entries) == 0 {
		return stats
	}

	// Basic counts
	stats["total_commands"] = len(entries)

	// Command frequency
	cmdFreq := make(map[string]int)
	categoryFreq := make(map[string]int)
	hourlyUsage := make(map[int]int)
	dailyUsage := make(map[string]int)
	successRate := 0

	for _, entry := range entries {
		cmd := strings.Fields(entry.Command)[0]
		cmdFreq[cmd]++
		categoryFreq[entry.Category]++
		hourlyUsage[entry.Timestamp.Hour()]++
		dailyUsage[entry.Timestamp.Format("2006-01-02")]++

		if entry.ExitCode == 0 {
			successRate++
		}
	}

	stats["command_frequency"] = cmdFreq
	stats["category_frequency"] = categoryFreq
	stats["hourly_usage"] = hourlyUsage
	stats["daily_usage"] = dailyUsage
	stats["success_rate"] = float64(successRate) / float64(len(entries)) * 100

	// Find most used command
	maxFreq := 0
	mostUsed := ""
	for cmd, freq := range cmdFreq {
		if freq > maxFreq {
			maxFreq = freq
			mostUsed = cmd
		}
	}
	stats["most_used_command"] = mostUsed
	stats["most_used_frequency"] = maxFreq

	// Find busiest hour
	maxHourUsage := 0
	busiestHour := 0
	for hour, usage := range hourlyUsage {
		if usage > maxHourUsage {
			maxHourUsage = usage
			busiestHour = hour
		}
	}
	stats["busiest_hour"] = busiestHour
	stats["busiest_hour_usage"] = maxHourUsage

	return stats
}

func (h *SmartHistoryCommand) formatStatistics(output *strings.Builder, stats map[string]interface{}, entries []HistoryEntry) {
	headerColor := color.New(color.FgHiYellow, color.Bold)
	labelColor := color.New(color.FgHiBlack)
	valueColor := color.New(color.FgHiCyan)

	// Basic statistics
	output.WriteString(headerColor.Sprint("📊 Overview\n"))
	output.WriteString(fmt.Sprintf("   Total Commands: %s\n",
		valueColor.Sprint(stats["total_commands"])))
	output.WriteString(fmt.Sprintf("   Success Rate: %s%%\n",
		valueColor.Sprintf("%.1f", stats["success_rate"])))
	output.WriteString(fmt.Sprintf("   Most Used: %s (%s times)\n\n",
		valueColor.Sprint(stats["most_used_command"]),
		valueColor.Sprint(stats["most_used_frequency"])))

	// Top commands
	output.WriteString(headerColor.Sprint("🏆 Top Commands\n"))
	cmdFreq := stats["command_frequency"].(map[string]int)
	topCommands := h.getTopItems(cmdFreq, 5)
	for i, item := range topCommands {
		bar := h.createUsageBar(item.Count, stats["total_commands"].(int))
		output.WriteString(fmt.Sprintf("   %d. %s %s %s\n",
			i+1,
			valueColor.Sprintf("%-12s", item.Key),
			bar,
			labelColor.Sprintf("(%d)", item.Count)))
	}
	output.WriteString("\n")

	// Activity patterns
	output.WriteString(headerColor.Sprint("⏰ Activity Patterns\n"))
	output.WriteString(fmt.Sprintf("   Busiest Hour: %s:00 (%s commands)\n",
		valueColor.Sprintf("%02d", stats["busiest_hour"]),
		valueColor.Sprint(stats["busiest_hour_usage"])))

	// Category breakdown
	output.WriteString(fmt.Sprintf("   Categories: "))
	categoryFreq := stats["category_frequency"].(map[string]int)
	categories := []string{}
	for cat, freq := range categoryFreq {
		categories = append(categories, fmt.Sprintf("%s(%d)", cat, freq))
	}
	output.WriteString(strings.Join(categories, ", "))
	output.WriteString("\n")
}

type KeyValue struct {
	Key   string
	Count int
}

func (h *SmartHistoryCommand) getTopItems(freq map[string]int, limit int) []KeyValue {
	var items []KeyValue
	for key, count := range freq {
		items = append(items, KeyValue{Key: key, Count: count})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Count > items[j].Count
	})

	if len(items) > limit {
		items = items[:limit]
	}

	return items
}

func (h *SmartHistoryCommand) createUsageBar(count, total int) string {
	barLength := 15
	percentage := float64(count) / float64(total) * 100
	filledLength := int(percentage * float64(barLength) / 100)

	bar := "["
	for i := 0; i < barLength; i++ {
		if i < filledLength {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	bar += "]"

	if percentage >= 50 {
		return color.New(color.FgHiGreen).Sprint(bar)
	} else if percentage >= 25 {
		return color.New(color.FgHiYellow).Sprint(bar)
	} else {
		return color.New(color.FgHiRed).Sprint(bar)
	}
}

func (h *SmartHistoryCommand) exportJSON(entries []HistoryEntry, filename string, startTime time.Time) (*commands.Result, error) {
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return &commands.Result{
			Output:   fmt.Sprintf("Error marshaling JSON: %v", err),
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		return &commands.Result{
			Output:   fmt.Sprintf("Error writing file: %v", err),
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	return &commands.Result{
		Output:   fmt.Sprintf("✅ History exported to %s (%d entries)", filename, len(entries)),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

func (h *SmartHistoryCommand) exportCSV(entries []HistoryEntry, filename string, startTime time.Time) (*commands.Result, error) {
	var output strings.Builder

	// CSV header
	output.WriteString("ID,Command,Timestamp,Directory,ExitCode,Duration,Category,Tags\n")

	// CSV data
	for _, entry := range entries {
		tags := strings.Join(entry.Tags, ";")
		output.WriteString(fmt.Sprintf("%d,\"%s\",\"%s\",\"%s\",%d,%d,\"%s\",\"%s\"\n",
			entry.ID,
			strings.ReplaceAll(entry.Command, "\"", "\"\""), // Escape quotes
			entry.Timestamp.Format("2006-01-02 15:04:05"),
			entry.Directory,
			entry.ExitCode,
			entry.Duration,
			entry.Category,
			tags))
	}

	err := ioutil.WriteFile(filename, []byte(output.String()), 0644)
	if err != nil {
		return &commands.Result{
			Output:   fmt.Sprintf("Error writing file: %v", err),
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	return &commands.Result{
		Output:   fmt.Sprintf("✅ History exported to %s (%d entries)", filename, len(entries)),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

func (h *SmartHistoryCommand) exportText(entries []HistoryEntry, filename string, startTime time.Time) (*commands.Result, error) {
	var output strings.Builder

	output.WriteString("SuperShell Command History Export\n")
	output.WriteString("==================================\n")
	output.WriteString(fmt.Sprintf("Generated: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	output.WriteString(fmt.Sprintf("Total Commands: %d\n\n", len(entries)))

	for _, entry := range entries {
		output.WriteString(fmt.Sprintf("[%d] %s\n", entry.ID, entry.Timestamp.Format("2006-01-02 15:04:05")))
		output.WriteString(fmt.Sprintf("Command: %s\n", entry.Command))
		output.WriteString(fmt.Sprintf("Directory: %s\n", entry.Directory))
		output.WriteString(fmt.Sprintf("Exit Code: %d\n", entry.ExitCode))
		output.WriteString(fmt.Sprintf("Duration: %dms\n", entry.Duration))
		output.WriteString(fmt.Sprintf("Category: %s\n", entry.Category))
		if len(entry.Tags) > 0 {
			output.WriteString(fmt.Sprintf("Tags: %s\n", strings.Join(entry.Tags, ", ")))
		}
		output.WriteString("\n")
	}

	err := ioutil.WriteFile(filename, []byte(output.String()), 0644)
	if err != nil {
		return &commands.Result{
			Output:   fmt.Sprintf("Error writing file: %v", err),
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	return &commands.Result{
		Output:   fmt.Sprintf("✅ History exported to %s (%d entries)", filename, len(entries)),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

func (h *SmartHistoryCommand) clearHistory(startTime time.Time) (*commands.Result, error) {
	err := os.Remove(h.historyFile)
	if err != nil && !os.IsNotExist(err) {
		return &commands.Result{
			Output:   fmt.Sprintf("Error clearing history: %v", err),
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	return &commands.Result{
		Output:   "Command history cleared",
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

func (h *SmartHistoryCommand) addToHistory(command string, startTime time.Time) (*commands.Result, error) {
	entries, err := h.loadHistory()
	if err != nil {
		return &commands.Result{
			Output:   fmt.Sprintf("Error loading history: %v", err),
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	// Create new entry
	newEntry := HistoryEntry{
		ID:        len(entries) + 1,
		Command:   command,
		Timestamp: time.Now(),
		Directory: ".",
		ExitCode:  0,
		Duration:  0,
		Tags:      []string{},
		Category:  h.categorizeCommand(command),
	}

	entries = append(entries, newEntry)

	err = h.saveHistory(entries)
	if err != nil {
		return &commands.Result{
			Output:   fmt.Sprintf("Error saving history: %v", err),
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	return &commands.Result{
		Output:   fmt.Sprintf("Added command to history: %s", command),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

func (h *SmartHistoryCommand) categorizeCommand(command string) string {
	cmd := strings.Fields(command)[0]

	switch cmd {
	case "ls", "dir", "cat", "cp", "mv", "rm", "mkdir":
		return "filesystem"
	case "ping", "tracert", "nslookup", "netstat", "wget":
		return "network"
	case "firewall", "server", "perf", "remote":
		return "management"
	case "help", "lookup", "history":
		return "help"
	default:
		return "other"
	}
}
