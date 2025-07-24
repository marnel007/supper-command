package intelligence

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

type UIRenderer struct {
	maxWidth     int
	maxHeight    int
	colorEnabled bool
}

func NewUIRenderer() *UIRenderer {
	return &UIRenderer{
		maxWidth:     80,
		maxHeight:    10,
		colorEnabled: true,
	}
}

type CompletionDisplay struct {
	Lines        []string `json:"lines"`
	SelectedLine int      `json:"selected_line"`
	TotalHeight  int      `json:"total_height"`
}

func (ur *UIRenderer) RenderCompletions(result *CompletionResult, selectedIndex int) *CompletionDisplay {
	if len(result.Completions) == 0 {
		return &CompletionDisplay{Lines: []string{}, TotalHeight: 0}
	}

	lines := make([]string, 0)

	header := "🧠 Intelligent Completions"
	if ur.colorEnabled {
		header = color.New(color.FgCyan, color.Bold).Sprint(header)
	}
	lines = append(lines, header)
	lines = append(lines, strings.Repeat("─", 40))

	maxItems := len(result.Completions)
	if maxItems > ur.maxHeight-3 {
		maxItems = ur.maxHeight - 3
	}

	for i := 0; i < maxItems; i++ {
		completion := result.Completions[i]

		icon := completion.Icon
		if icon == "" {
			icon = "📦"
		}

		line := fmt.Sprintf("%s %s", icon, completion.Display)

		if completion.Description != "" {
			line += " - " + completion.Description
		}

		if i == selectedIndex && ur.colorEnabled {
			line = color.New(color.BgBlue, color.FgWhite).Sprint(line)
		}

		lines = append(lines, line)
	}

	if len(result.Completions) > maxItems {
		footer := fmt.Sprintf("... and %d more", len(result.Completions)-maxItems)
		if ur.colorEnabled {
			footer = color.New(color.FgHiBlack).Sprint(footer)
		}
		lines = append(lines, footer)
	}

	return &CompletionDisplay{
		Lines:        lines,
		SelectedLine: selectedIndex + 2,
		TotalHeight:  len(lines),
	}
}
