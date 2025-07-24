package intelligence

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

type IntelligenceEngine struct {
	mu                  sync.RWMutex
	completionProviders []CompletionProvider
	learningSystem      *LearningSystem
	contextDetector     *ContextDetector
	fuzzyMatcher        *FuzzyMatcher
	uiRenderer          *UIRenderer
	settings            *Settings
	dataDir             string
	isInitialized       bool
}

type Settings struct {
	Enabled                 bool `json:"enabled"`
	FuzzyMatchingEnabled    bool `json:"fuzzy_matching_enabled"`
	SmartSuggestionsEnabled bool `json:"smart_suggestions_enabled"`
	ExternalToolsEnabled    bool `json:"external_tools_enabled"`
	LearningEnabled         bool `json:"learning_enabled"`
	MaxSuggestions          int  `json:"max_suggestions"`
	MinPatternOccurrences   int  `json:"min_pattern_occurrences"`
	LearningThreshold       int  `json:"learning_threshold"`
	ResponseTimeoutMs       int  `json:"response_timeout_ms"`
}

func DefaultSettings() *Settings {
	return &Settings{
		Enabled:                 true,
		FuzzyMatchingEnabled:    true,
		SmartSuggestionsEnabled: true,
		ExternalToolsEnabled:    true,
		LearningEnabled:         true,
		MaxSuggestions:          10,
		MinPatternOccurrences:   3,
		LearningThreshold:       5,
		ResponseTimeoutMs:       100,
	}
}

func NewIntelligenceEngine() (*IntelligenceEngine, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	dataDir := filepath.Join(homeDir, ".supershell", "intelligence")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create intelligence data directory: %w", err)
	}

	engine := &IntelligenceEngine{
		dataDir: dataDir,
	}

	if err := engine.loadSettings(); err != nil {
		engine.settings = DefaultSettings()
	}

	return engine, nil
}

func (ie *IntelligenceEngine) Initialize(ctx context.Context) error {
	ie.mu.Lock()
	defer ie.mu.Unlock()

	if ie.isInitialized {
		return nil
	}

	color.New(color.FgCyan).Println("🧠 Initializing SuperShell Intelligence Engine...")

	ie.learningSystem = NewLearningSystem(ie.dataDir)
	ie.contextDetector = NewContextDetector()
	ie.fuzzyMatcher = NewFuzzyMatcher()
	ie.uiRenderer = NewUIRenderer()

	ie.completionProviders = []CompletionProvider{
		NewInternalCommandProvider(),
		NewFileSystemProvider(),
		NewExternalToolProvider(),
		NewSmartSuggestionProvider(ie.learningSystem),
	}

	if err := ie.learningSystem.Load(); err != nil {
		color.New(color.FgYellow).Printf("⚠️  Could not load learning data: %v\n", err)
	}

	if ie.settings.ExternalToolsEnabled {
		go ie.detectExternalTools()
	}

	ie.isInitialized = true
	color.New(color.FgGreen).Println("✅ Intelligence Engine initialized")
	return nil
}

func (ie *IntelligenceEngine) GetCompletions(ctx context.Context, input string, cursorPos int) (*CompletionResult, error) {
	if !ie.isInitialized || !ie.settings.Enabled {
		return &CompletionResult{}, nil
	}

	ie.mu.RLock()
	defer ie.mu.RUnlock()

	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(ie.settings.ResponseTimeoutMs)*time.Millisecond)
	defer cancel()

	contextInfo := ie.contextDetector.GetContext()
	inputInfo := ParseInput(input, cursorPos)

	allCompletions := make([]Completion, 0)

	for _, provider := range ie.completionProviders {
		completions, err := provider.GetCompletions(timeoutCtx, inputInfo, contextInfo)
		if err != nil {
			continue
		}
		allCompletions = append(allCompletions, completions...)
	}

	if ie.settings.FuzzyMatchingEnabled {
		allCompletions = ie.fuzzyMatcher.FilterAndRank(inputInfo.CurrentWord, allCompletions)
	}

	if len(allCompletions) > ie.settings.MaxSuggestions {
		allCompletions = allCompletions[:ie.settings.MaxSuggestions]
	}

	return &CompletionResult{
		Completions: allCompletions,
		InputInfo:   inputInfo,
		Context:     contextInfo,
	}, nil
}

func (ie *IntelligenceEngine) RecordCommand(command string) {
	if !ie.isInitialized || !ie.settings.LearningEnabled {
		return
	}
	ie.learningSystem.RecordCommand(command)
}

func (ie *IntelligenceEngine) GetSmartSuggestion(lastCommand string) *SmartSuggestion {
	if !ie.isInitialized || !ie.settings.SmartSuggestionsEnabled {
		return nil
	}
	return ie.learningSystem.GetSmartSuggestion(lastCommand, ie.settings.MinPatternOccurrences)
}

func (ie *IntelligenceEngine) GetUIRenderer() *UIRenderer {
	return ie.uiRenderer
}

func (ie *IntelligenceEngine) GetLearningSystem() *LearningSystem {
	return ie.learningSystem
}

func (ie *IntelligenceEngine) GetContextDetector() *ContextDetector {
	return ie.contextDetector
}

func (ie *IntelligenceEngine) Shutdown() error {
	ie.mu.Lock()
	defer ie.mu.Unlock()

	if !ie.isInitialized {
		return nil
	}

	if err := ie.learningSystem.Save(); err != nil {
		return fmt.Errorf("failed to save learning data: %w", err)
	}

	if err := ie.saveSettings(); err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	return nil
}

func (ie *IntelligenceEngine) detectExternalTools() {
	tools := []string{"git", "docker", "npm", "pip", "go", "python", "node"}
	detected := make([]string, 0)

	for _, tool := range tools {
		if ie.isToolAvailable(tool) {
			detected = append(detected, tool)
		}
	}

	if len(detected) > 0 {
		color.New(color.FgGreen).Printf("🔧 Detected external tools: %s\n", strings.Join(detected, ", "))
	}
}

func (ie *IntelligenceEngine) isToolAvailable(tool string) bool {
	_, err := exec.LookPath(tool)
	return err == nil
}

func (ie *IntelligenceEngine) loadSettings() error {
	settingsPath := filepath.Join(ie.dataDir, "settings.json")

	data, err := ioutil.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			ie.settings = DefaultSettings()
			return ie.saveSettings()
		}
		return err
	}

	ie.settings = &Settings{}
	return json.Unmarshal(data, ie.settings)
}

func (ie *IntelligenceEngine) saveSettings() error {
	settingsPath := filepath.Join(ie.dataDir, "settings.json")

	data, err := json.MarshalIndent(ie.settings, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(settingsPath, data, 0644)
}
