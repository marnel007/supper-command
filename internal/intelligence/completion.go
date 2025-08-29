package intelligence

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type CompletionType int

const (
	CompletionTypeCommand CompletionType = iota
	CompletionTypeFile
	CompletionTypeDirectory
	CompletionTypeFlag
	CompletionTypeExternal
	CompletionTypeSmart
)

type Completion struct {
	Text        string         `json:"text"`
	Display     string         `json:"display"`
	Description string         `json:"description"`
	Type        CompletionType `json:"type"`
	Category    string         `json:"category"`
	Icon        string         `json:"icon"`
	Score       float64        `json:"score"`
	InsertText  string         `json:"insert_text"`
	Metadata    map[string]any `json:"metadata"`
}

type CompletionResult struct {
	Completions []Completion `json:"completions"`
	InputInfo   *InputInfo   `json:"input_info"`
	Context     *ContextInfo `json:"context"`
}

type InputInfo struct {
	FullLine      string   `json:"full_line"`
	CurrentWord   string   `json:"current_word"`
	PreviousWords []string `json:"previous_words"`
	CursorPos     int      `json:"cursor_pos"`
	WordStart     int      `json:"word_start"`
	WordEnd       int      `json:"word_end"`
}

func ParseInput(input string, cursorPos int) *InputInfo {
	if cursorPos > len(input) {
		cursorPos = len(input)
	}

	wordStart := cursorPos
	wordEnd := cursorPos

	for wordStart > 0 && input[wordStart-1] != ' ' {
		wordStart--
	}

	for wordEnd < len(input) && input[wordEnd] != ' ' {
		wordEnd++
	}

	currentWord := ""
	if wordStart < len(input) {
		currentWord = input[wordStart:min(wordEnd, len(input))]
	}

	parts := strings.Fields(input[:wordStart])

	return &InputInfo{
		FullLine:      input,
		CurrentWord:   currentWord,
		PreviousWords: parts,
		CursorPos:     cursorPos,
		WordStart:     wordStart,
		WordEnd:       wordEnd,
	}
}

type CompletionProvider interface {
	GetCompletions(ctx context.Context, input *InputInfo, context *ContextInfo) ([]Completion, error)
	Name() string
	Priority() int
}

type InternalCommandProvider struct {
	commands map[string]CommandInfo
}

type CommandInfo struct {
	Name        string
	Description string
	Category    string
	Flags       []string
	Examples    []string
}

func NewInternalCommandProvider() *InternalCommandProvider {
	return &InternalCommandProvider{
		commands: buildInternalCommands(),
	}
}

func (p *InternalCommandProvider) Name() string  { return "internal_commands" }
func (p *InternalCommandProvider) Priority() int { return 100 }

func (p *InternalCommandProvider) GetCompletions(ctx context.Context, input *InputInfo, context *ContextInfo) ([]Completion, error) {
	completions := make([]Completion, 0)

	if len(input.PreviousWords) == 0 {
		for name, cmd := range p.commands {
			if strings.HasPrefix(name, input.CurrentWord) {
				completions = append(completions, Completion{
					Text:        name,
					Display:     name,
					Description: cmd.Description,
					Type:        CompletionTypeCommand,
					Category:    cmd.Category,
					Icon:        getCommandIcon(cmd.Category),
					Score:       calculateScore(name, input.CurrentWord),
					InsertText:  name,
				})
			}
		}
	}

	return completions, nil
}

type FileSystemProvider struct{}

func NewFileSystemProvider() *FileSystemProvider {
	return &FileSystemProvider{}
}

func (p *FileSystemProvider) Name() string  { return "filesystem" }
func (p *FileSystemProvider) Priority() int { return 80 }

func (p *FileSystemProvider) GetCompletions(ctx context.Context, input *InputInfo, context *ContextInfo) ([]Completion, error) {
	completions := make([]Completion, 0)

	// Check if this is a file/path completion scenario
	shouldCompleteFiles := false
	searchDir := "."
	searchPattern := input.CurrentWord

	// If we have previous words, check if the first word is a file-related command
	if len(input.PreviousWords) > 0 {
		firstWord := input.PreviousWords[0]
		fileCommands := []string{"cd", "ls", "cat", "cp", "mv", "rm", "rmdir", "mkdir"}
		for _, cmd := range fileCommands {
			if firstWord == cmd {
				shouldCompleteFiles = true
				break
			}
		}
	}

	// Also complete if the current word looks like a path
	if strings.Contains(input.CurrentWord, "/") || strings.Contains(input.CurrentWord, "\\") || strings.Contains(input.CurrentWord, ":") {
		shouldCompleteFiles = true
	}

	if !shouldCompleteFiles {
		return completions, nil
	}

	// Handle path parsing for Windows and Unix
	if strings.Contains(input.CurrentWord, "\\") || strings.Contains(input.CurrentWord, ":") {
		// Windows path
		if strings.Contains(input.CurrentWord, "\\") {
			dir := filepath.Dir(input.CurrentWord)
			if dir != "." && dir != input.CurrentWord {
				searchDir = dir
				searchPattern = filepath.Base(input.CurrentWord)
			} else {
				searchPattern = input.CurrentWord
			}
		} else if strings.Contains(input.CurrentWord, ":") {
			// Handle drive letters like "e:"
			if len(input.CurrentWord) >= 2 && input.CurrentWord[1] == ':' {
				if len(input.CurrentWord) == 2 {
					// Just "e:" - search root of drive
					searchDir = input.CurrentWord + "\\"
					searchPattern = ""
				} else {
					// "e:\something" - parse normally
					searchDir = filepath.Dir(input.CurrentWord)
					searchPattern = filepath.Base(input.CurrentWord)
				}
			}
		}
	} else if strings.Contains(input.CurrentWord, "/") {
		// Unix path
		dir := filepath.Dir(input.CurrentWord)
		if dir != "." {
			searchDir = dir
		}
		searchPattern = filepath.Base(input.CurrentWord)
	}

	// Read directory
	entries, err := ioutil.ReadDir(searchDir)
	if err != nil {
		return completions, nil
	}

	for _, entry := range entries {
		name := entry.Name()

		// Skip hidden files unless explicitly requested
		if strings.HasPrefix(name, ".") && !strings.HasPrefix(searchPattern, ".") {
			continue
		}

		// Check if name matches pattern (case-insensitive)
		if searchPattern == "" || strings.HasPrefix(strings.ToLower(name), strings.ToLower(searchPattern)) {
			fullPath := name
			if searchDir != "." {
				fullPath = filepath.Join(searchDir, name)
			}

			icon := "📄"
			compType := CompletionTypeFile
			description := "File"

			if entry.IsDir() {
				icon = "📁"
				compType = CompletionTypeDirectory
				description = "Directory"
				// For directories, add separator for easier navigation
				if !strings.HasSuffix(fullPath, string(os.PathSeparator)) {
					fullPath += string(os.PathSeparator)
				}
			}

			score := calculateScore(name, searchPattern)
			// Boost score for exact prefix matches
			if strings.HasPrefix(strings.ToLower(name), strings.ToLower(searchPattern)) {
				score += 20
			}

			completions = append(completions, Completion{
				Text:        fullPath,
				Display:     name,
				Description: description,
				Type:        compType,
				Category:    "filesystem",
				Icon:        icon,
				Score:       score,
				InsertText:  fullPath,
			})
		}
	}

	return completions, nil
}

type ExternalToolProvider struct {
	tools map[string]ToolInfo
}

type ToolInfo struct {
	Name        string
	Description string
	Subcommands []string
	CommonFlags []string
}

func NewExternalToolProvider() *ExternalToolProvider {
	return &ExternalToolProvider{
		tools: buildExternalTools(),
	}
}

func (p *ExternalToolProvider) Name() string  { return "external_tools" }
func (p *ExternalToolProvider) Priority() int { return 70 }

func (p *ExternalToolProvider) GetCompletions(ctx context.Context, input *InputInfo, context *ContextInfo) ([]Completion, error) {
	completions := make([]Completion, 0)

	if len(input.PreviousWords) == 0 {
		for name, tool := range p.tools {
			if strings.HasPrefix(name, input.CurrentWord) {
				completions = append(completions, Completion{
					Text:        name,
					Display:     name,
					Description: tool.Description,
					Type:        CompletionTypeExternal,
					Category:    "external",
					Icon:        "🔧",
					Score:       calculateScore(name, input.CurrentWord),
					InsertText:  name,
				})
			}
		}
	}

	return completions, nil
}

type SmartSuggestionProvider struct {
	learningSystem *LearningSystem
}

func NewSmartSuggestionProvider(learning *LearningSystem) *SmartSuggestionProvider {
	return &SmartSuggestionProvider{
		learningSystem: learning,
	}
}

func (p *SmartSuggestionProvider) Name() string  { return "smart_suggestions" }
func (p *SmartSuggestionProvider) Priority() int { return 90 }

func (p *SmartSuggestionProvider) GetCompletions(ctx context.Context, input *InputInfo, context *ContextInfo) ([]Completion, error) {
	completions := make([]Completion, 0)

	suggestions := p.learningSystem.GetSmartCompletions(input.CurrentWord, 5)

	for _, suggestion := range suggestions {
		completions = append(completions, Completion{
			Text:        suggestion.Command,
			Display:     suggestion.Command,
			Description: fmt.Sprintf("🤖 Smart suggestion (used %d times)", suggestion.Frequency),
			Type:        CompletionTypeSmart,
			Category:    "smart",
			Icon:        "🤖",
			Score:       float64(suggestion.Frequency) * 10,
			InsertText:  suggestion.Command,
			Metadata: map[string]any{
				"frequency":  suggestion.Frequency,
				"confidence": suggestion.Confidence,
			},
		})
	}

	return completions, nil
}

func calculateScore(text, pattern string) float64 {
	if text == pattern {
		return 100.0
	}
	if strings.HasPrefix(strings.ToLower(text), strings.ToLower(pattern)) {
		return 90.0 - float64(len(text)-len(pattern))
	}
	return 50.0
}

func getCommandIcon(category string) string {
	icons := map[string]string{
		"marketplace": "🏪", "development": "🔧", "performance": "⚡",
		"system": "⚙️", "filesystem": "📁",
	}

	if icon, exists := icons[category]; exists {
		return icon
	}
	return "📦"
}

func buildInternalCommands() map[string]CommandInfo {
	return map[string]CommandInfo{
		"help": {
			Name: "help", Description: "Show help information", Category: "system",
		},
		"marketplace": {
			Name: "marketplace", Description: "Community plugin marketplace", Category: "marketplace",
		},
		"dev": {
			Name: "dev", Description: "Development tools", Category: "development",
		},
		"ls": {
			Name: "ls", Description: "List directory contents", Category: "filesystem",
		},
		"cd": {
			Name: "cd", Description: "Change directory", Category: "filesystem",
		},
	}
}

func buildExternalTools() map[string]ToolInfo {
	return map[string]ToolInfo{
		"git": {
			Name: "git", Description: "Git version control",
		},
		"docker": {
			Name: "docker", Description: "Docker container platform",
		},
		"npm": {
			Name: "npm", Description: "Node.js package manager",
		},
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
