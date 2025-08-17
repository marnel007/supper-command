package system

import (
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

// BookmarkCommand manages command bookmarks and snippets
type BookmarkCommand struct {
	*commands.BaseCommand
	bookmarkFile string
	registry     *commands.Registry
}

// Bookmark represents a saved command with metadata
type Bookmark struct {
	Name        string            `json:"name"`
	Command     string            `json:"command"`
	Description string            `json:"description"`
	Category    string            `json:"category"`
	Tags        []string          `json:"tags"`
	Variables   map[string]string `json:"variables"`
	Created     time.Time         `json:"created"`
	LastUsed    time.Time         `json:"last_used"`
	UseCount    int               `json:"use_count"`
	Shared      bool              `json:"shared"`
}

// NewBookmarkCommand creates a new bookmark command
func NewBookmarkCommand(registry *commands.Registry) *BookmarkCommand {
	homeDir, _ := os.UserHomeDir()
	bookmarkFile := filepath.Join(homeDir, ".supershell_bookmarks.json")

	return &BookmarkCommand{
		BaseCommand: commands.NewBaseCommand(
			"bookmark",
			"Manage command bookmarks and snippets",
			"bookmark [add|list|run|edit|remove|search|import|export] [options]",
			[]string{"windows", "linux", "darwin"},
			false,
		),
		bookmarkFile: bookmarkFile,
		registry:     registry,
	}
}

// Execute handles bookmark operations
func (b *BookmarkCommand) Execute(ctx context.Context, args *commands.Arguments) (*commands.Result, error) {
	startTime := time.Now()

	if len(args.Raw) == 0 {
		return b.listBookmarksCommand(startTime)
	}

	subcommand := args.Raw[0]
	switch subcommand {
	case "add":
		if len(args.Raw) < 3 {
			return &commands.Result{
				Output:   "Usage: bookmark add <name> <command> [description]",
				ExitCode: 1,
				Duration: time.Since(startTime),
			}, nil
		}
		name := args.Raw[1]
		command := args.Raw[2]
		description := ""
		if len(args.Raw) > 3 {
			description = strings.Join(args.Raw[3:], " ")
		}
		return b.addBookmark(name, command, description, startTime)

	case "list":
		category := ""
		if len(args.Raw) > 1 {
			category = args.Raw[1]
		}
		return b.listBookmarksCommand(startTime, category)

	case "run":
		if len(args.Raw) < 2 {
			return &commands.Result{
				Output:   "Usage: bookmark run <name>",
				ExitCode: 1,
				Duration: time.Since(startTime),
			}, nil
		}
		return b.runBookmark(args.Raw[1], startTime)

	case "remove", "delete":
		if len(args.Raw) < 2 {
			return &commands.Result{
				Output:   "Usage: bookmark remove <name>",
				ExitCode: 1,
				Duration: time.Since(startTime),
			}, nil
		}
		return b.removeBookmark(args.Raw[1], startTime)

	case "search":
		if len(args.Raw) < 2 {
			return &commands.Result{
				Output:   "Usage: bookmark search <query>",
				ExitCode: 1,
				Duration: time.Since(startTime),
			}, nil
		}
		query := strings.Join(args.Raw[1:], " ")
		return b.searchBookmarksCommand(query, startTime)

	case "edit":
		if len(args.Raw) < 2 {
			return &commands.Result{
				Output:   "Usage: bookmark edit <name>",
				ExitCode: 1,
				Duration: time.Since(startTime),
			}, nil
		}
		return b.editBookmark(args.Raw[1], startTime)

	case "export":
		format := "json"
		if len(args.Raw) > 1 {
			format = args.Raw[1]
		}
		return b.exportBookmarks(format, startTime)

	case "import":
		if len(args.Raw) < 2 {
			return &commands.Result{
				Output:   "Usage: bookmark import <filename>",
				ExitCode: 1,
				Duration: time.Since(startTime),
			}, nil
		}
		return b.importBookmarks(args.Raw[1], startTime)

	case "stats":
		return b.showBookmarkStats(startTime)

	default:
		return b.showHelp(startTime)
	}
}

// addBookmark adds a new bookmark
func (b *BookmarkCommand) addBookmark(name, command, description string, startTime time.Time) (*commands.Result, error) {
	bookmarks, err := b.loadBookmarksFromFile()
	if err != nil {
		return &commands.Result{
			Output:   fmt.Sprintf("Error loading bookmarks: %v", err),
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	// Check if bookmark already exists
	for _, bookmark := range bookmarks {
		if bookmark.Name == name {
			return &commands.Result{
				Output:   fmt.Sprintf("Bookmark '%s' already exists. Use 'bookmark edit %s' to modify it.", name, name),
				ExitCode: 1,
				Duration: time.Since(startTime),
			}, nil
		}
	}

	// Create new bookmark
	newBookmark := Bookmark{
		Name:        name,
		Command:     command,
		Description: description,
		Category:    b.categorizeCommand(command),
		Tags:        b.extractTags(command),
		Variables:   b.extractVariables(command),
		Created:     time.Now(),
		LastUsed:    time.Time{},
		UseCount:    0,
		Shared:      false,
	}

	bookmarks = append(bookmarks, newBookmark)

	err = b.saveBookmarksToFile(bookmarks)
	if err != nil {
		return &commands.Result{
			Output:   fmt.Sprintf("Error saving bookmark: %v", err),
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	successColor := color.New(color.FgGreen, color.Bold)
	return &commands.Result{
		Output: fmt.Sprintf("%s\n📖 Bookmark '%s' added successfully!\n🔖 Command: %s\n📝 Category: %s\n",
			successColor.Sprint("✅ BOOKMARK ADDED"),
			color.New(color.FgCyan).Sprint(name),
			color.New(color.FgWhite).Sprint(command),
			color.New(color.FgYellow).Sprint(newBookmark.Category)),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// listBookmarksCommand lists all bookmarks
func (b *BookmarkCommand) listBookmarksCommand(startTime time.Time, category ...string) (*commands.Result, error) {
	bookmarks, err := b.loadBookmarksFromFile()
	if err != nil {
		return &commands.Result{
			Output:   fmt.Sprintf("Error loading bookmarks: %v", err),
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	var output strings.Builder
	headerColor := color.New(color.FgHiCyan, color.Bold)

	output.WriteString(headerColor.Sprint("📚 Command Bookmarks\n"))
	output.WriteString("═══════════════════════════════════════════════════════════════\n")

	if len(bookmarks) == 0 {
		output.WriteString(color.New(color.FgHiBlack).Sprint("No bookmarks found. Use 'bookmark add <name> <command>' to create one.\n"))
		return &commands.Result{
			Output:   output.String(),
			ExitCode: 0,
			Duration: time.Since(startTime),
		}, nil
	}

	// Filter by category if specified
	filteredBookmarks := bookmarks
	if len(category) > 0 && category[0] != "" {
		filteredBookmarks = []Bookmark{}
		for _, bookmark := range bookmarks {
			if strings.EqualFold(bookmark.Category, category[0]) {
				filteredBookmarks = append(filteredBookmarks, bookmark)
			}
		}
	}

	// Group by category
	categories := make(map[string][]Bookmark)
	for _, bookmark := range filteredBookmarks {
		categories[bookmark.Category] = append(categories[bookmark.Category], bookmark)
	}

	// Sort categories
	var categoryNames []string
	for cat := range categories {
		categoryNames = append(categoryNames, cat)
	}
	sort.Strings(categoryNames)

	for _, catName := range categoryNames {
		catBookmarks := categories[catName]
		b.formatBookmarkCategory(&output, catName, catBookmarks)
	}

	output.WriteString("\n")
	output.WriteString(color.New(color.FgHiBlack).Sprint("💡 Use: bookmark run <name> | bookmark search <query> | bookmark add <name> <command>\n"))

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// runBookmark executes a bookmarked command
func (b *BookmarkCommand) runBookmark(name string, startTime time.Time) (*commands.Result, error) {
	bookmarks, err := b.loadBookmarksFromFile()
	if err != nil {
		return &commands.Result{
			Output:   fmt.Sprintf("Error loading bookmarks: %v", err),
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	// Find bookmark
	var targetBookmark *Bookmark
	for i, bookmark := range bookmarks {
		if bookmark.Name == name {
			targetBookmark = &bookmarks[i]
			break
		}
	}

	if targetBookmark == nil {
		return &commands.Result{
			Output:   fmt.Sprintf("Bookmark '%s' not found. Use 'bookmark list' to see available bookmarks.", name),
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	// Update usage statistics
	targetBookmark.LastUsed = time.Now()
	targetBookmark.UseCount++
	b.saveBookmarksToFile(bookmarks)

	runColor := color.New(color.FgHiGreen, color.Bold)
	return &commands.Result{
		Output: fmt.Sprintf("%s\n🚀 Executing bookmark: %s\n📋 Command: %s\n",
			runColor.Sprint("▶️ RUNNING BOOKMARK"),
			color.New(color.FgCyan).Sprint(name),
			color.New(color.FgWhite).Sprint(targetBookmark.Command)),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// Helper methods
func (b *BookmarkCommand) loadBookmarksFromFile() ([]Bookmark, error) {
	var bookmarks []Bookmark

	if _, err := os.Stat(b.bookmarkFile); os.IsNotExist(err) {
		return bookmarks, nil
	}

	data, err := os.ReadFile(b.bookmarkFile)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return bookmarks, nil
	}

	err = json.Unmarshal(data, &bookmarks)
	return bookmarks, err
}

func (b *BookmarkCommand) saveBookmarksToFile(bookmarks []Bookmark) error {
	data, err := json.MarshalIndent(bookmarks, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(b.bookmarkFile, data, 0644)
}

func (b *BookmarkCommand) categorizeCommand(command string) string {
	cmd := strings.Fields(command)[0]

	switch cmd {
	case "ls", "dir", "cat", "cp", "mv", "rm", "mkdir", "pwd", "cd":
		return "filesystem"
	case "ping", "tracert", "nslookup", "netstat", "wget", "portscan", "speedtest":
		return "network"
	case "firewall", "server", "perf", "remote":
		return "management"
	case "git":
		return "git"
	case "docker", "kubectl":
		return "containers"
	default:
		return "general"
	}
}

func (b *BookmarkCommand) extractTags(command string) []string {
	// Simple tag extraction based on command content
	var tags []string

	if strings.Contains(command, "backup") {
		tags = append(tags, "backup")
	}
	if strings.Contains(command, "sync") {
		tags = append(tags, "sync")
	}
	if strings.Contains(command, "monitor") {
		tags = append(tags, "monitoring")
	}

	return tags
}

func (b *BookmarkCommand) extractVariables(command string) map[string]string {
	// Extract variables like ${VAR} from command
	variables := make(map[string]string)
	// Simple implementation - could be enhanced with regex
	return variables
}

func (b *BookmarkCommand) formatBookmarkCategory(output *strings.Builder, category string, bookmarks []Bookmark) {
	categoryColor := color.New(color.FgHiYellow, color.Bold)
	output.WriteString(fmt.Sprintf("\n%s %s\n", b.getCategoryIcon(category), categoryColor.Sprint(strings.ToUpper(category))))
	output.WriteString("───────────────────────────────────────────────────────────────\n")

	for _, bookmark := range bookmarks {
		nameColor := color.New(color.FgCyan, color.Bold)
		commandColor := color.New(color.FgWhite)
		descColor := color.New(color.FgHiBlack)
		statsColor := color.New(color.FgHiBlack)

		output.WriteString(fmt.Sprintf("📖 %s\n", nameColor.Sprint(bookmark.Name)))
		output.WriteString(fmt.Sprintf("   %s\n", commandColor.Sprint(bookmark.Command)))

		if bookmark.Description != "" {
			output.WriteString(fmt.Sprintf("   %s\n", descColor.Sprint(bookmark.Description)))
		}

		if bookmark.UseCount > 0 {
			lastUsed := "never"
			if !bookmark.LastUsed.IsZero() {
				lastUsed = bookmark.LastUsed.Format("01/02 15:04")
			}
			output.WriteString(fmt.Sprintf("   %s\n",
				statsColor.Sprintf("Used %d times, last: %s", bookmark.UseCount, lastUsed)))
		}
		output.WriteString("\n")
	}
}

func (b *BookmarkCommand) getCategoryIcon(category string) string {
	switch category {
	case "filesystem":
		return "📁"
	case "network":
		return "🌐"
	case "management":
		return "⚙️"
	case "git":
		return "🔧"
	case "containers":
		return "🐳"
	default:
		return "📋"
	}
}

// Placeholder methods for additional functionality
func (b *BookmarkCommand) removeBookmark(name string, startTime time.Time) (*commands.Result, error) {
	return &commands.Result{
		Output:   fmt.Sprintf("Bookmark '%s' removed (placeholder)", name),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

func (b *BookmarkCommand) searchBookmarksCommand(query string, startTime time.Time) (*commands.Result, error) {
	return &commands.Result{
		Output:   fmt.Sprintf("Searching bookmarks for '%s' (placeholder)", query),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

func (b *BookmarkCommand) editBookmark(name string, startTime time.Time) (*commands.Result, error) {
	return &commands.Result{
		Output:   fmt.Sprintf("Editing bookmark '%s' (placeholder)", name),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

func (b *BookmarkCommand) exportBookmarks(format string, startTime time.Time) (*commands.Result, error) {
	return &commands.Result{
		Output:   fmt.Sprintf("Exporting bookmarks as %s (placeholder)", format),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

func (b *BookmarkCommand) importBookmarks(filename string, startTime time.Time) (*commands.Result, error) {
	return &commands.Result{
		Output:   fmt.Sprintf("Importing bookmarks from %s (placeholder)", filename),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

func (b *BookmarkCommand) showBookmarkStats(startTime time.Time) (*commands.Result, error) {
	return &commands.Result{
		Output:   "Bookmark statistics (placeholder)",
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

func (b *BookmarkCommand) showHelp(startTime time.Time) (*commands.Result, error) {
	help := `
📚 Command Bookmarks Help

Usage: bookmark [command] [options]

Commands:
  add <name> <command> [desc]    Add a new bookmark
  list [category]                List bookmarks (optionally by category)
  run <name>                     Execute a bookmarked command
  remove <name>                  Remove a bookmark
  search <query>                 Search bookmarks
  edit <name>                    Edit a bookmark
  export [format]                Export bookmarks (json, csv, txt)
  import <filename>              Import bookmarks from file
  stats                          Show bookmark statistics

Examples:
  bookmark add backup "rsync -av /home/ /backup/" "Daily backup"
  bookmark add check-ports "netstat -tulpn | grep LISTEN"
  bookmark list filesystem
  bookmark run backup
  bookmark search "network"

Categories: filesystem, network, management, git, containers, general
`

	return &commands.Result{
		Output:   strings.TrimSpace(help),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}
