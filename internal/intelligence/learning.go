package intelligence

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type LearningSystem struct {
	mu               sync.RWMutex
	dataDir          string
	commandHistory   []CommandRecord
	workflowPatterns map[string]*WorkflowPattern
	commandFrequency map[string]int
	maxHistorySize   int
}

type CommandRecord struct {
	Command   string    `json:"command"`
	Timestamp time.Time `json:"timestamp"`
	Success   bool      `json:"success"`
}

type WorkflowPattern struct {
	FromCommand string    `json:"from_command"`
	ToCommand   string    `json:"to_command"`
	Occurrences int       `json:"occurrences"`
	Confidence  float64   `json:"confidence"`
	LastSeen    time.Time `json:"last_seen"`
}

type SmartSuggestion struct {
	Command    string  `json:"command"`
	Reason     string  `json:"reason"`
	Confidence float64 `json:"confidence"`
	Frequency  int     `json:"frequency"`
}

type SmartCompletion struct {
	Command    string    `json:"command"`
	Frequency  int       `json:"frequency"`
	Confidence float64   `json:"confidence"`
	LastUsed   time.Time `json:"last_used"`
}

func NewLearningSystem(dataDir string) *LearningSystem {
	return &LearningSystem{
		dataDir:          dataDir,
		commandHistory:   make([]CommandRecord, 0),
		workflowPatterns: make(map[string]*WorkflowPattern),
		commandFrequency: make(map[string]int),
		maxHistorySize:   1000,
	}
}

func (ls *LearningSystem) Load() error {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	historyPath := filepath.Join(ls.dataDir, "command_history.json")
	if data, err := ioutil.ReadFile(historyPath); err == nil {
		json.Unmarshal(data, &ls.commandHistory)
	}

	patternsPath := filepath.Join(ls.dataDir, "workflow_patterns.json")
	if data, err := ioutil.ReadFile(patternsPath); err == nil {
		json.Unmarshal(data, &ls.workflowPatterns)
	}

	ls.rebuildFrequencyMap()
	return nil
}

func (ls *LearningSystem) Save() error {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	historyPath := filepath.Join(ls.dataDir, "command_history.json")
	if data, err := json.MarshalIndent(ls.commandHistory, "", "  "); err == nil {
		ioutil.WriteFile(historyPath, data, 0644)
	}

	patternsPath := filepath.Join(ls.dataDir, "workflow_patterns.json")
	if data, err := json.MarshalIndent(ls.workflowPatterns, "", "  "); err == nil {
		ioutil.WriteFile(patternsPath, data, 0644)
	}

	return nil
}

func (ls *LearningSystem) RecordCommand(command string) {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	command = strings.TrimSpace(command)
	if command == "" {
		return
	}

	record := CommandRecord{
		Command:   command,
		Timestamp: time.Now(),
		Success:   true,
	}

	ls.commandHistory = append(ls.commandHistory, record)

	if len(ls.commandHistory) > ls.maxHistorySize {
		ls.commandHistory = ls.commandHistory[len(ls.commandHistory)-ls.maxHistorySize:]
	}

	ls.commandFrequency[command]++
	ls.learnWorkflowPatterns()
}

func (ls *LearningSystem) GetSmartSuggestion(lastCommand string, minOccurrences int) *SmartSuggestion {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	if lastCommand == "" {
		return nil
	}

	if pattern, exists := ls.workflowPatterns[lastCommand]; exists {
		if pattern.Occurrences >= minOccurrences && pattern.Confidence > 0.6 {
			return &SmartSuggestion{
				Command:    pattern.ToCommand,
				Reason:     "You usually run this command next",
				Confidence: pattern.Confidence,
				Frequency:  pattern.Occurrences,
			}
		}
	}

	return nil
}

func (ls *LearningSystem) GetSmartCompletions(partial string, maxResults int) []SmartCompletion {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	completions := make([]SmartCompletion, 0)
	partial = strings.ToLower(partial)

	for command, frequency := range ls.commandFrequency {
		if strings.HasPrefix(strings.ToLower(command), partial) {
			completion := SmartCompletion{
				Command:    command,
				Frequency:  frequency,
				Confidence: float64(frequency) / 10.0,
				LastUsed:   ls.getLastUsed(command),
			}
			completions = append(completions, completion)
		}
	}

	sort.Slice(completions, func(i, j int) bool {
		return completions[i].Frequency > completions[j].Frequency
	})

	if len(completions) > maxResults {
		completions = completions[:maxResults]
	}

	return completions
}

func (ls *LearningSystem) learnWorkflowPatterns() {
	if len(ls.commandHistory) < 2 {
		return
	}

	historyLen := len(ls.commandHistory)
	lastIdx := historyLen - 1
	prevIdx := historyLen - 2

	fromCmd := ls.commandHistory[prevIdx].Command
	toCmd := ls.commandHistory[lastIdx].Command

	if fromCmd == toCmd {
		return
	}

	if pattern, exists := ls.workflowPatterns[fromCmd]; exists {
		if pattern.ToCommand == toCmd {
			pattern.Occurrences++
			pattern.LastSeen = time.Now()
			pattern.Confidence = float64(pattern.Occurrences) / 10.0
		}
	} else {
		ls.workflowPatterns[fromCmd] = &WorkflowPattern{
			FromCommand: fromCmd,
			ToCommand:   toCmd,
			Occurrences: 1,
			Confidence:  0.1,
			LastSeen:    time.Now(),
		}
	}
}

func (ls *LearningSystem) rebuildFrequencyMap() {
	ls.commandFrequency = make(map[string]int)
	for _, record := range ls.commandHistory {
		ls.commandFrequency[record.Command]++
	}
}

func (ls *LearningSystem) getLastUsed(command string) time.Time {
	for i := len(ls.commandHistory) - 1; i >= 0; i-- {
		if ls.commandHistory[i].Command == command {
			return ls.commandHistory[i].Timestamp
		}
	}
	return time.Time{}
}

func (ls *LearningSystem) GetStats() map[string]interface{} {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	return map[string]interface{}{
		"total_commands":    len(ls.commandHistory),
		"unique_commands":   len(ls.commandFrequency),
		"workflow_patterns": len(ls.workflowPatterns),
		"learning_active":   len(ls.commandHistory) > 0,
	}
}
