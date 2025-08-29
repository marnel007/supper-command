package shell

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"suppercommand/internal/commands"
	"suppercommand/internal/monitoring"
)

// Completer handles tab completion
type Completer struct {
	registry *commands.Registry
	logger   monitoring.Logger
}

// Completion represents a completion suggestion
type Completion struct {
	Text        string
	Description string
	Type        CompletionType
}

// CompletionType represents the type of completion
type CompletionType int

const (
	CompletionTypeCommand CompletionType = iota
	CompletionTypeFile
	CompletionTypeDirectory
	CompletionTypeOption
)

// NewCompleter creates a new completer
func NewCompleter(
	registry *commands.Registry,
	logger monitoring.Logger,
) *Completer {
	return &Completer{
		registry: registry,
		logger:   logger,
	}
}

// Initialize initializes the completer
func (c *Completer) Initialize(ctx context.Context) error {
	c.logger.Info("Tab completer initialized")
	return nil
}

// GetCompletions returns completion suggestions for the given input
func (c *Completer) GetCompletions(input string, cursorPos int) []Completion {
	if cursorPos > len(input) {
		cursorPos = len(input)
	}

	textBeforeCursor := input[:cursorPos]
	parts := strings.Fields(textBeforeCursor)

	if len(parts) == 0 {
		// No input, return command completions
		return c.getCommandCompletions("")
	}

	if len(parts) == 1 && !strings.HasSuffix(textBeforeCursor, " ") {
		// Completing command name
		return c.getCommandCompletions(parts[0])
	}

	// Completing arguments - for now, just file/directory completion
	lastPart := ""
	if len(parts) > 1 {
		lastPart = parts[len(parts)-1]
	}

	return c.getFileCompletions(lastPart)
}

// getCommandCompletions returns command name completions
func (c *Completer) getCommandCompletions(prefix string) []Completion {
	var completions []Completion

	commandNames := c.registry.List()
	for _, name := range commandNames {
		if strings.HasPrefix(name, prefix) {
			cmd, err := c.registry.Get(name)
			if err != nil {
				continue
			}

			description := cmd.Description()
			// Truncate long descriptions to prevent skewing
			if len(description) > 40 {
				description = description[:37] + "..."
			}

			completions = append(completions, Completion{
				Text:        name,
				Description: description,
				Type:        CompletionTypeCommand,
			})
		}
	}

	// Limit completions to prevent display issues
	if len(completions) > 8 {
		completions = completions[:8]
	}

	return completions
}

// getFileCompletions returns file and directory completions
func (c *Completer) getFileCompletions(prefix string) []Completion {
	var completions []Completion

	dir, filePrefix := filepath.Split(prefix)
	if dir == "" {
		dir = "."
	}

	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		return completions
	}

	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, filePrefix) {
			fullPath := filepath.Join(dir, name)

			if entry.IsDir() {
				completions = append(completions, Completion{
					Text:        fullPath + string(os.PathSeparator),
					Description: "Directory",
					Type:        CompletionTypeDirectory,
				})
			} else {
				completions = append(completions, Completion{
					Text:        fullPath,
					Description: "File",
					Type:        CompletionTypeFile,
				})
			}
		}
	}

	// Limit file completions to prevent display issues
	if len(completions) > 6 {
		completions = completions[:6]
	}

	return completions
}

// Shutdown gracefully shuts down the completer
func (c *Completer) Shutdown(ctx context.Context) error {
	c.logger.Info("Tab completer shutdown")
	return nil
}
