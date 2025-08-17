package shell

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"suppercommand/internal/config"
	"suppercommand/internal/monitoring"
)

// Prompter handles prompt rendering and management
type Prompter struct {
	config config.ShellConfig
	logger monitoring.Logger
}

// NewPrompter creates a new prompter
func NewPrompter(
	config config.ShellConfig,
	logger monitoring.Logger,
) *Prompter {
	return &Prompter{
		config: config,
		logger: logger,
	}
}

// Initialize initializes the prompter
func (p *Prompter) Initialize(ctx context.Context) error {
	p.logger.Info("Prompt renderer initialized")
	return nil
}

// GetPrompt returns the current prompt string
func (p *Prompter) GetPrompt() string {
	if p.config.Colors.Enabled {
		return p.getColoredPrompt()
	}
	return p.getPlainPrompt()
}

// getColoredPrompt returns a colored prompt
func (p *Prompter) getColoredPrompt() string {
	cwd, _ := os.Getwd()
	shortPath := p.getShortenedPath(cwd)

	var prompt strings.Builder

	// SuperShell branding with colors
	prompt.WriteString("\033[38;5;51m🚀 \033[1;36mSuperShell\033[0m") // Cyan rocket + text
	prompt.WriteString("\033[38;5;46m●\033[0m")                      // Green dot

	// Directory path
	prompt.WriteString(fmt.Sprintf("\033[90m[\033[33m%s\033[90m]\033[0m", shortPath))

	// Prompt arrows
	prompt.WriteString(" \033[38;5;51m❯\033[38;5;45m❯\033[38;5;39m❯\033[0m ")

	return prompt.String()
}

// getPlainPrompt returns a plain text prompt
func (p *Prompter) getPlainPrompt() string {
	cwd, _ := os.Getwd()
	shortPath := p.getShortenedPath(cwd)

	return fmt.Sprintf("SuperShell [%s] >>> ", shortPath)
}

// getShortenedPath shortens long paths for display
func (p *Prompter) getShortenedPath(path string) string {
	// Convert to OS-appropriate path separators
	path = filepath.Clean(path)

	// If path is too long, show beginning + ... + end
	if len(path) > 40 {
		parts := strings.Split(path, string(os.PathSeparator))
		if len(parts) > 3 {
			return parts[0] + string(os.PathSeparator) + "..." + string(os.PathSeparator) + parts[len(parts)-1]
		}
	}

	return path
}

// GetLivePrefix returns a live prefix for go-prompt (without ANSI codes)
func (p *Prompter) GetLivePrefix() (string, bool) {
	cwd, _ := os.Getwd()
	shortPath := p.getShortenedPath(cwd)

	// Clean prompt without ANSI codes for go-prompt compatibility
	return fmt.Sprintf("🚀 SuperShell●[%s] ❯❯❯ ", shortPath), true
}

// Shutdown gracefully shuts down the prompter
func (p *Prompter) Shutdown(ctx context.Context) error {
	p.logger.Info("Prompt renderer shutdown")
	return nil
}
