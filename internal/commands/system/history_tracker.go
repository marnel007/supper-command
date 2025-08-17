package system

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// HistoryTracker automatically tracks command execution
type HistoryTracker struct {
	historyFile string
	maxEntries  int
}

// NewHistoryTracker creates a new history tracker
func NewHistoryTracker() *HistoryTracker {
	homeDir, _ := os.UserHomeDir()
	historyFile := filepath.Join(homeDir, ".supershell_history.json")

	return &HistoryTracker{
		historyFile: historyFile,
		maxEntries:  1000, // Keep last 1000 commands
	}
}

// TrackCommand adds a command to the history
func (ht *HistoryTracker) TrackCommand(command string, directory string, exitCode int, duration time.Duration) error {
	entries, err := ht.loadHistory()
	if err != nil {
		// If we can't load history, start with empty slice
		entries = []HistoryEntry{}
	}

	// Create new entry
	newEntry := HistoryEntry{
		ID:        len(entries) + 1,
		Command:   command,
		Timestamp: time.Now(),
		Directory: directory,
		ExitCode:  exitCode,
		Duration:  duration.Milliseconds(),
		Tags:      ht.generateTags(command),
		Category:  ht.categorizeCommand(command),
	}

	// Add to entries
	entries = append(entries, newEntry)

	// Trim to max entries if needed
	if len(entries) > ht.maxEntries {
		entries = entries[len(entries)-ht.maxEntries:]
		// Renumber IDs
		for i := range entries {
			entries[i].ID = i + 1
		}
	}

	// Save back to file
	return ht.saveHistory(entries)
}

// loadHistory loads command history from file
func (ht *HistoryTracker) loadHistory() ([]HistoryEntry, error) {
	var entries []HistoryEntry

	if _, err := os.Stat(ht.historyFile); os.IsNotExist(err) {
		return entries, nil
	}

	data, err := os.ReadFile(ht.historyFile)
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
func (ht *HistoryTracker) saveHistory(entries []HistoryEntry) error {
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(ht.historyFile, data, 0644)
}

// generateTags generates relevant tags for a command
func (ht *HistoryTracker) generateTags(command string) []string {
	tags := []string{}
	cmdLower := strings.ToLower(command)

	// File operations
	if strings.Contains(cmdLower, "ls") || strings.Contains(cmdLower, "dir") {
		tags = append(tags, "file-listing")
	}
	if strings.Contains(cmdLower, "cd") {
		tags = append(tags, "navigation")
	}
	if strings.Contains(cmdLower, "cp") || strings.Contains(cmdLower, "mv") || strings.Contains(cmdLower, "rm") {
		tags = append(tags, "file-operations")
	}
	if strings.Contains(cmdLower, "cat") || strings.Contains(cmdLower, "less") || strings.Contains(cmdLower, "more") {
		tags = append(tags, "file-viewing")
	}

	// Network operations
	if strings.Contains(cmdLower, "ping") || strings.Contains(cmdLower, "tracert") {
		tags = append(tags, "network-diagnostics")
	}
	if strings.Contains(cmdLower, "wget") || strings.Contains(cmdLower, "curl") {
		tags = append(tags, "download")
	}
	if strings.Contains(cmdLower, "netstat") || strings.Contains(cmdLower, "ss") {
		tags = append(tags, "network-monitoring")
	}

	// System monitoring
	if strings.Contains(cmdLower, "ps") || strings.Contains(cmdLower, "top") || strings.Contains(cmdLower, "htop") {
		tags = append(tags, "process-monitoring")
	}
	if strings.Contains(cmdLower, "df") || strings.Contains(cmdLower, "du") {
		tags = append(tags, "disk-usage")
	}
	if strings.Contains(cmdLower, "free") || strings.Contains(cmdLower, "vmstat") {
		tags = append(tags, "memory-monitoring")
	}

	// Development
	if strings.Contains(cmdLower, "git") {
		tags = append(tags, "version-control")
		if strings.Contains(cmdLower, "commit") {
			tags = append(tags, "git-commit")
		}
		if strings.Contains(cmdLower, "push") || strings.Contains(cmdLower, "pull") {
			tags = append(tags, "git-sync")
		}
	}
	if strings.Contains(cmdLower, "docker") {
		tags = append(tags, "containerization")
	}
	if strings.Contains(cmdLower, "npm") || strings.Contains(cmdLower, "yarn") {
		tags = append(tags, "package-management")
	}

	// System administration
	if strings.Contains(cmdLower, "systemctl") || strings.Contains(cmdLower, "service") {
		tags = append(tags, "service-management")
	}
	if strings.Contains(cmdLower, "firewall") {
		tags = append(tags, "security")
	}
	if strings.Contains(cmdLower, "backup") || strings.Contains(cmdLower, "restore") {
		tags = append(tags, "backup")
	}

	// Search and analysis
	if strings.Contains(cmdLower, "grep") || strings.Contains(cmdLower, "find") || strings.Contains(cmdLower, "locate") {
		tags = append(tags, "search")
	}
	if strings.Contains(cmdLower, "tail") || strings.Contains(cmdLower, "head") {
		tags = append(tags, "log-analysis")
	}

	// SuperShell specific
	if strings.Contains(cmdLower, "perf") {
		tags = append(tags, "performance")
	}
	if strings.Contains(cmdLower, "server") {
		tags = append(tags, "server-management")
	}
	if strings.Contains(cmdLower, "remote") {
		tags = append(tags, "remote-management")
	}

	return tags
}

// categorizeCommand categorizes a command into a broad category
func (ht *HistoryTracker) categorizeCommand(command string) string {
	cmd := strings.ToLower(strings.Fields(command)[0])

	switch cmd {
	case "ls", "dir", "cat", "cp", "mv", "rm", "mkdir", "rmdir", "chmod", "chown":
		return "filesystem"
	case "ping", "tracert", "nslookup", "netstat", "wget", "curl", "ssh", "scp":
		return "network"
	case "firewall", "server", "perf", "remote", "cluster":
		return "management"
	case "git", "svn", "hg":
		return "version-control"
	case "docker", "kubectl", "helm":
		return "containers"
	case "ps", "top", "htop", "kill", "killall", "jobs":
		return "processes"
	case "df", "du", "mount", "umount", "fdisk":
		return "storage"
	case "systemctl", "service", "crontab":
		return "system-services"
	case "grep", "find", "locate", "which", "whereis":
		return "search"
	case "help", "man", "info", "lookup", "history":
		return "help"
	case "cd", "pwd", "pushd", "popd":
		return "navigation"
	default:
		return "other"
	}
}

// GetHistoryFile returns the path to the history file
func (ht *HistoryTracker) GetHistoryFile() string {
	return ht.historyFile
}

// GetRecentCommands returns the most recent commands
func (ht *HistoryTracker) GetRecentCommands(count int) ([]HistoryEntry, error) {
	entries, err := ht.loadHistory()
	if err != nil {
		return nil, err
	}

	if len(entries) <= count {
		return entries, nil
	}

	return entries[len(entries)-count:], nil
}

// SearchHistory searches through command history
func (ht *HistoryTracker) SearchHistory(query string) ([]HistoryEntry, error) {
	entries, err := ht.loadHistory()
	if err != nil {
		return nil, err
	}

	var matches []HistoryEntry
	queryLower := strings.ToLower(query)

	for _, entry := range entries {
		if strings.Contains(strings.ToLower(entry.Command), queryLower) ||
			strings.Contains(strings.ToLower(entry.Category), queryLower) {
			matches = append(matches, entry)
		}

		// Check tags
		for _, tag := range entry.Tags {
			if strings.Contains(strings.ToLower(tag), queryLower) {
				matches = append(matches, entry)
				break
			}
		}
	}

	return matches, nil
}
