package commands

import (
	"sort"
	"strings"
)

// CompletionEngine provides command auto-completion functionality
type CompletionEngine struct {
	registry    *Registry
	completions map[string][]string
}

// NewCompletionEngine creates a new completion engine
func NewCompletionEngine(registry *Registry) *CompletionEngine {
	return &CompletionEngine{
		registry:    registry,
		completions: GetAutoCompletions(),
	}
}

// Complete returns completion suggestions for a given input
func (c *CompletionEngine) Complete(input string) []string {
	if input == "" {
		return c.getAllCommands()
	}

	parts := strings.Fields(input)
	if len(parts) == 0 {
		return c.getAllCommands()
	}

	// Complete command names
	if len(parts) == 1 {
		return c.completeCommand(parts[0])
	}

	// Complete subcommands and options
	return c.completeSubcommand(parts)
}

// getAllCommands returns all available command names
func (c *CompletionEngine) getAllCommands() []string {
	commands := c.registry.List()
	sort.Strings(commands)
	return commands
}

// completeCommand completes command names
func (c *CompletionEngine) completeCommand(prefix string) []string {
	allCommands := c.getAllCommands()
	matches := make([]string, 0)

	for _, cmd := range allCommands {
		if strings.HasPrefix(cmd, prefix) {
			matches = append(matches, cmd)
		}
	}

	return matches
}

// completeSubcommand completes subcommands and options
func (c *CompletionEngine) completeSubcommand(parts []string) []string {
	if len(parts) < 2 {
		return []string{}
	}

	command := parts[0]
	subcommand := parts[1]

	// Look for exact matches first
	key := command + " " + subcommand
	if completions, exists := c.completions[key]; exists {
		return c.filterCompletions(completions, parts[2:])
	}

	// Look for command-level completions
	if completions, exists := c.completions[command]; exists {
		matches := make([]string, 0)
		for _, completion := range completions {
			if strings.HasPrefix(completion, subcommand) {
				matches = append(matches, completion)
			}
		}
		return matches
	}

	return []string{}
}

// filterCompletions filters completions based on already provided arguments
func (c *CompletionEngine) filterCompletions(completions []string, provided []string) []string {
	if len(provided) == 0 {
		return completions
	}

	// Remove already provided options
	providedSet := make(map[string]bool)
	for _, arg := range provided {
		if strings.HasPrefix(arg, "--") {
			providedSet[arg] = true
		}
	}

	filtered := make([]string, 0)
	for _, completion := range completions {
		if !providedSet[completion] {
			filtered = append(filtered, completion)
		}
	}

	return filtered
}

// GetCommandSuggestions returns suggestions for a partial command
func (c *CompletionEngine) GetCommandSuggestions(partial string) []string {
	suggestions := make([]string, 0)
	commands := c.getAllCommands()

	for _, cmd := range commands {
		if strings.Contains(cmd, partial) {
			suggestions = append(suggestions, cmd)
		}
	}

	// Add fuzzy matching for better UX
	if len(suggestions) == 0 {
		suggestions = c.fuzzyMatch(partial, commands)
	}

	return suggestions
}

// fuzzyMatch performs fuzzy matching on command names
func (c *CompletionEngine) fuzzyMatch(pattern string, candidates []string) []string {
	matches := make([]string, 0)
	pattern = strings.ToLower(pattern)

	for _, candidate := range candidates {
		if c.fuzzyScore(pattern, strings.ToLower(candidate)) > 0 {
			matches = append(matches, candidate)
		}
	}

	return matches
}

// fuzzyScore calculates a fuzzy matching score
func (c *CompletionEngine) fuzzyScore(pattern, text string) int {
	if pattern == "" {
		return 1
	}

	score := 0
	patternIndex := 0

	for i, char := range text {
		if patternIndex < len(pattern) && rune(pattern[patternIndex]) == char {
			score++
			patternIndex++

			// Bonus for consecutive matches
			if i > 0 && patternIndex > 1 {
				score++
			}
		}
	}

	// Must match all pattern characters
	if patternIndex != len(pattern) {
		return 0
	}

	return score
}

// GetContextualHelp returns contextual help for a command
func (c *CompletionEngine) GetContextualHelp(command string) string {
	helpMap := GetCommandHelp()
	if help, exists := helpMap[command]; exists {
		return strings.TrimSpace(help)
	}

	// Try to get help from the command itself
	if cmd, err := c.registry.Get(command); err == nil {
		return cmd.Usage()
	}

	return "No help available for command: " + command
}

// GetCommandsByCategory returns commands organized by category
func (c *CompletionEngine) GetCommandsByCategory() map[string][]string {
	categories := GetCommandCategories()

	// Filter to only include registered commands
	registeredCommands := make(map[string]bool)
	for _, cmd := range c.getAllCommands() {
		registeredCommands[cmd] = true
	}

	filtered := make(map[string][]string)
	for category, commands := range categories {
		availableCommands := make([]string, 0)
		for _, cmd := range commands {
			if registeredCommands[cmd] {
				availableCommands = append(availableCommands, cmd)
			}
		}
		if len(availableCommands) > 0 {
			filtered[category] = availableCommands
		}
	}

	return filtered
}

// ValidateCommand checks if a command exists and is valid
func (c *CompletionEngine) ValidateCommand(command string) bool {
	_, err := c.registry.Get(command)
	return err == nil
}

// GetSimilarCommands returns commands similar to the given input
func (c *CompletionEngine) GetSimilarCommands(input string, maxResults int) []string {
	if maxResults <= 0 {
		maxResults = 5
	}

	type scoredCommand struct {
		command string
		score   int
	}

	commands := c.getAllCommands()
	scored := make([]scoredCommand, 0)

	for _, cmd := range commands {
		score := c.fuzzyScore(strings.ToLower(input), strings.ToLower(cmd))
		if score > 0 {
			scored = append(scored, scoredCommand{cmd, score})
		}
	}

	// Sort by score (descending)
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Return top results
	results := make([]string, 0, maxResults)
	for i, sc := range scored {
		if i >= maxResults {
			break
		}
		results = append(results, sc.command)
	}

	return results
}

// GetAutoCompletions returns predefined auto-completion mappings
func GetAutoCompletions() map[string][]string {
	return map[string][]string{
		"firewall":        {"status", "enable", "disable", "rules", "help"},
		"firewall rules":  {"list", "add", "remove"},
		"perf":            {"analyze", "monitor", "report", "baseline", "help"},
		"perf baseline":   {"create", "list", "delete"},
		"server":          {"health", "services", "users", "alerts", "backup", "session", "help"},
		"server services": {"list", "start", "stop", "restart"},
		"server session":  {"list", "kill"},
		"remote":          {"list", "add", "remove", "exec", "cluster", "sync", "help"},
		"remote cluster":  {"list", "create", "delete"},
		"remote sync":     {"list", "create", "execute"},
		"ping":            {"-c", "--count", "-t", "--timeout", "-i", "--interval"},
		"tracert":         {"-m", "--max-hops", "-t", "--timeout"},
		"nslookup":        {"-s", "--server"},
		"netstat":         {"-tcp", "--tcp", "-udp", "--udp", "-state", "-p", "--process", "--csv", "--json", "--sort", "--desc", "--group"},
		"portscan":        {"-p", "--ports", "-t", "--timeout", "--top-ports"},
		"sniff":           {"-i", "--interface", "-c", "--count", "-p", "--protocol", "-s", "--source", "-d", "--dest", "--port", "-v", "--verbose", "--hex", "--save", "--continuous", "-t", "--timeout"},
		"wget":            {"-v", "--verbose"},
		"arp":             {"-a", "--all", "-d", "--delete"},
		"route":           {"print", "show", "add", "delete", "-4", "--ipv4", "-6", "--ipv6"},
		"speedtest":       {"-s", "--simple", "-q", "--quiet", "--download-only", "--upload-only"},
		"sysinfo":         {"-v", "--verbose", "--cpu", "--memory", "--disk", "--network"},
		"killtask":        {"-f", "--force", "-t", "--tree"},
		"lookup":          {"-m", "--menu", "-s", "--similar", "-c", "--categories", "-t", "--task"},
		"ver":             {"-v", "--verbose"},
	}
}

// GetCommandHelp returns help text for commands
func GetCommandHelp() map[string]string {
	return map[string]string{
		// Management Commands
		"firewall": "Manage system firewall settings and rules. Control Windows Defender Firewall, view status, and manage security policies.",
		"perf":     "Performance monitoring and system analysis. Monitor CPU, memory, disk, and network usage with baseline comparison capabilities.",
		"server":   "Server management and system administration. Monitor health, manage services, track users, and maintain system components.",
		"remote":   "Remote server management and SSH operations. Add servers, execute commands remotely, and manage distributed systems.",

		// Network Commands
		"ping":        "Send ICMP echo requests to test network connectivity and measure response times to remote hosts.",
		"tracert":     "Trace the network route packets take to reach a destination, showing each hop along the path.",
		"nslookup":    "Query DNS servers for domain name information, IP addresses, and various DNS record types.",
		"netstat":     "Display active network connections, listening ports, and network statistics with filtering options.",
		"portscan":    "Scan remote hosts for open ports and services, useful for network security assessment.",
		"sniff":       "Capture and analyze network packets in real-time with protocol filtering and detailed inspection.",
		"wget":        "Download files from web servers using HTTP/HTTPS with progress monitoring and resume capability.",
		"arp":         "Display and modify the ARP (Address Resolution Protocol) table showing IP to MAC address mappings.",
		"route":       "Display and modify the system routing table to control network packet forwarding.",
		"speedtest":   "Test internet connection speed by measuring download/upload bandwidth and latency.",
		"ipconfig":    "Display network interface configuration including IP addresses, subnet masks, and gateways.",
		"netdiscover": "Discover active devices on the local network using ARP requests and network scanning.",

		// File System Commands
		"ls":    "List directory contents with various formatting options and file information display.",
		"dir":   "Windows-style directory listing showing files and folders with detailed information.",
		"cat":   "Display the contents of text files to the console with optional line numbering.",
		"cp":    "Copy files and directories from source to destination with preservation of attributes.",
		"mv":    "Move or rename files and directories, supporting both local and cross-directory operations.",
		"rm":    "Remove files and directories with support for wildcards and recursive deletion.",
		"mkdir": "Create new directories with optional parent directory creation.",
		"rmdir": "Remove empty directories or recursively delete directory trees.",
		"pwd":   "Print the current working directory path to show your current location.",
		"cd":    "Change the current working directory to navigate the file system.",

		// System Commands
		"sysinfo":   "Display comprehensive system information including hardware, OS, and performance metrics.",
		"killtask":  "Terminate running processes by name or PID with force termination options.",
		"whoami":    "Display the current user account name and authentication context.",
		"hostname":  "Show the system hostname and network identification information.",
		"ver":       "Display SuperShell version information and build details.",
		"clear":     "Clear the terminal screen and reset the display for better readability.",
		"echo":      "Print text to the console, useful for displaying messages and variables.",
		"winupdate": "Manage Windows Update operations including checking for and installing updates.",

		// Help and Utility Commands
		"help":   "Display comprehensive help information for all commands with detailed usage examples.",
		"lookup": "Interactive command discovery system with search, categorization, and suggestion features.",
		"exit":   "Exit the SuperShell application and return to the system command prompt.",

		// FastCP Commands
		"fastcp-send":    "Ultra-fast file transfer sender with encryption, compression, and resume capability.",
		"fastcp-recv":    "Ultra-fast file transfer receiver with automatic decompression and verification.",
		"fastcp-backup":  "Create encrypted, compressed backups with deduplication and cloud storage support.",
		"fastcp-restore": "Restore files from FastCP backups with integrity verification and selective recovery.",
		"fastcp-dedup":   "Manage file deduplication to optimize storage usage and backup efficiency.",
	}
}

// GetCommandCategories returns commands organized by category
func GetCommandCategories() map[string][]string {
	return map[string][]string{
		"🔥 Security & Firewall":    {"firewall"},
		"⚡ Performance Monitoring": {"perf"},
		"🖥️ Server Management":     {"server", "sysinfo", "killtask", "winupdate"},
		"🌐 Remote Administration":  {"remote"},
		"🌐 Network Tools":          {"ping", "tracert", "nslookup", "netstat", "portscan", "sniff", "wget", "arp", "route", "speedtest", "ipconfig", "netdiscover"},
		"📁 File Operations":        {"ls", "dir", "cat", "cp", "mv", "rm", "mkdir", "rmdir", "pwd", "cd"},
		"⚙️ System Information":    {"whoami", "hostname", "sysinfo", "ver", "clear", "echo"},
		"🔍 Help & Discovery":       {"help", "lookup", "exit"},
		"🚀 FastCP File Transfer":   {"fastcp-send", "fastcp-recv", "fastcp-backup", "fastcp-restore", "fastcp-dedup"},
	}
}
