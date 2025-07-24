package intelligence

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type ContextInfo struct {
	CurrentDirectory string    `json:"current_directory"`
	GitRepository    *GitInfo  `json:"git_repository,omitempty"`
	ProjectType      string    `json:"project_type"`
	AvailableTools   []string  `json:"available_tools"`
	LastCommandTime  time.Time `json:"last_command_time"`
}

type GitInfo struct {
	IsRepository   bool     `json:"is_repository"`
	CurrentBranch  string   `json:"current_branch"`
	HasUncommitted bool     `json:"has_uncommitted"`
	ModifiedFiles  []string `json:"modified_files"`
}

type ContextDetector struct {
	cachedContext         *ContextInfo
	lastContextTime       time.Time
	cacheValidityDuration time.Duration
}

func NewContextDetector() *ContextDetector {
	return &ContextDetector{
		cacheValidityDuration: 30 * time.Second,
	}
}

func (cd *ContextDetector) GetContext() *ContextInfo {
	now := time.Now()

	if cd.cachedContext != nil && now.Sub(cd.lastContextTime) < cd.cacheValidityDuration {
		return cd.cachedContext
	}

	context := &ContextInfo{
		LastCommandTime: now,
	}

	cd.detectCurrentDirectory(context)
	cd.detectGitRepository(context)
	cd.detectProjectType(context)
	cd.detectAvailableTools(context)

	cd.cachedContext = context
	cd.lastContextTime = now

	return context
}

func (cd *ContextDetector) detectCurrentDirectory(context *ContextInfo) {
	if cwd, err := os.Getwd(); err == nil {
		context.CurrentDirectory = cwd
	}
}

func (cd *ContextDetector) detectGitRepository(context *ContextInfo) {
	gitInfo := &GitInfo{}

	if _, err := exec.LookPath("git"); err == nil {
		if output, err := exec.Command("git", "rev-parse", "--is-inside-work-tree").Output(); err == nil {
			if strings.TrimSpace(string(output)) == "true" {
				gitInfo.IsRepository = true

				if output, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output(); err == nil {
					gitInfo.CurrentBranch = strings.TrimSpace(string(output))
				}

				if output, err := exec.Command("git", "status", "--porcelain").Output(); err == nil {
					gitInfo.HasUncommitted = len(strings.TrimSpace(string(output))) > 0
				}
			}
		}
	}

	if gitInfo.IsRepository {
		context.GitRepository = gitInfo
	}
}

func (cd *ContextDetector) detectProjectType(context *ContextInfo) {
	cwd := context.CurrentDirectory
	if cwd == "" {
		return
	}

	projectIndicators := map[string][]string{
		"node.js": {"package.json", "yarn.lock"},
		"python":  {"requirements.txt", "setup.py"},
		"go":      {"go.mod", "go.sum"},
	}

	for projectType, indicators := range projectIndicators {
		for _, indicator := range indicators {
			fullPath := filepath.Join(cwd, indicator)
			if _, err := os.Stat(fullPath); err == nil {
				context.ProjectType = projectType
				return
			}
		}
	}

	context.ProjectType = "unknown"
}

func (cd *ContextDetector) detectAvailableTools(context *ContextInfo) {
	tools := []string{"git", "docker", "npm", "pip", "go", "python", "node"}
	availableTools := make([]string, 0)

	for _, tool := range tools {
		if _, err := exec.LookPath(tool); err == nil {
			availableTools = append(availableTools, tool)
		}
	}

	context.AvailableTools = availableTools
}

func (cd *ContextDetector) IsInGitRepository() bool {
	context := cd.GetContext()
	return context.GitRepository != nil && context.GitRepository.IsRepository
}

func (cd *ContextDetector) GetProjectType() string {
	context := cd.GetContext()
	return context.ProjectType
}

func (cd *ContextDetector) IsToolAvailable(tool string) bool {
	context := cd.GetContext()
	for _, availableTool := range context.AvailableTools {
		if availableTool == tool {
			return true
		}
	}
	return false
}

func (cd *ContextDetector) GetGitInfo() *GitInfo {
	context := cd.GetContext()
	return context.GitRepository
}
