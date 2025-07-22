package core

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/fatih/color"
	"gopkg.in/yaml.v3"
)

var (
	dirColor   = color.New(color.FgCyan).SprintFunc()
	fileColor  = color.New(color.FgWhite).SprintFunc()
	exeColor   = color.New(color.FgGreen).SprintFunc()
	errorColor = color.New(color.FgRed).SprintFunc()
	sumColor   = color.New(color.FgHiBlack).SprintFunc()
)

var runningCmd *exec.Cmd
var cancelWget context.CancelFunc

// Help command (already in command.go, but for clarity, you can move it here)
type HelpCommand struct{}

func (h *HelpCommand) Name() string        { return "help" }
func (h *HelpCommand) Description() string { return "Show this help message" }
func (h *HelpCommand) Execute(args []string) string {
	helpText := "SuperShell - Available Commands:\n"
	for _, cmd := range commandRegistry {
		helpText += "  " + cmd.Name() + " - " + cmd.Description() + "\n"
	}
	helpText += `
Alias usage:
  alias                # List all aliases
  alias <name> <cmd>   # Create or update an alias (e.g. alias ll ls -l)
  unalias <name>       # Remove an alias

Type 'help' to see this message again.
`
	return helpText
}

// Clear/cls command
type ClearCommand struct{}

func (c *ClearCommand) Name() string        { return "clear" }
func (c *ClearCommand) Description() string { return "Clear the screen" }
func (c *ClearCommand) Execute(args []string) string {
	cmd := exec.Command("cmd", "/c", "cls")
	cmd.Stdout = os.Stdout
	cmd.Run()
	return ""
}

// Echo command
type EchoCommand struct{}

func (e *EchoCommand) Name() string        { return "echo" }
func (e *EchoCommand) Description() string { return "Print text to the screen" }
func (e *EchoCommand) Execute(args []string) string {
	return strings.Join(args, " ")
}

// PWD command
type PwdCommand struct{}

func (p *PwdCommand) Name() string        { return "pwd" }
func (p *PwdCommand) Description() string { return "Print working directory" }
func (p *PwdCommand) Execute(args []string) string {
	dir, _ := os.Getwd()
	return dir
}

// LS command
type LsCommand struct{}

func (l *LsCommand) Name() string        { return "ls" }
func (l *LsCommand) Description() string { return "List directory contents" }
func (l *LsCommand) Execute(args []string) string {
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return "Error: " + err.Error()
	}
	var out strings.Builder
	for _, f := range files {
		out.WriteString(f.Name())
		if f.IsDir() {
			out.WriteString("/")
		}
		out.WriteString("\n")
	}
	return out.String()
}

// CD command
type CdCommand struct{}

func (c *CdCommand) Name() string        { return "cd" }
func (c *CdCommand) Description() string { return "Change directory" }
func (c *CdCommand) Execute(args []string) string {
	if len(args) == 0 {
		return "Usage: cd <directory>"
	}
	err := os.Chdir(args[0])
	if err != nil {
		return "Error: " + err.Error()
	}
	return ""
}

// Exit command (for help listing only)
type ExitCommand struct{}

func (e *ExitCommand) Name() string        { return "exit" }
func (e *ExitCommand) Description() string { return "Exit the shell" }
func (e *ExitCommand) Execute(args []string) string {
	// Handled in shell loop
	return ""
}

// --- cat command ---
type CatCommand struct{}

func (c *CatCommand) Name() string        { return "cat" }
func (c *CatCommand) Description() string { return "Show file contents" }
func (c *CatCommand) Execute(args []string) string {
	if len(args) == 0 {
		return "Usage: cat <file>"
	}
	data, err := ioutil.ReadFile(args[0])
	if err != nil {
		return "Error: " + err.Error()
	}
	return string(data)
}

// --- mkdir command ---
type MkdirCommand struct{}

func (m *MkdirCommand) Name() string        { return "mkdir" }
func (m *MkdirCommand) Description() string { return "Create a new directory" }
func (m *MkdirCommand) Execute(args []string) string {
	if len(args) == 0 {
		return "Usage: mkdir <directory>"
	}
	err := os.Mkdir(args[0], 0755)
	if err != nil {
		return "Error: " + err.Error()
	}
	return ""
}

// --- rm command ---
type RmCommand struct{}

func (r *RmCommand) Name() string        { return "rm" }
func (r *RmCommand) Description() string { return "Delete a file" }
func (r *RmCommand) Execute(args []string) string {
	if len(args) == 0 {
		return "Usage: rm <file>"
	}
	err := os.Remove(args[0])
	if err != nil {
		return "Error: " + err.Error()
	}
	return ""
}

// --- rmdir command ---
type RmdirCommand struct{}

func (r *RmdirCommand) Name() string        { return "rmdir" }
func (r *RmdirCommand) Description() string { return "Remove a directory" }
func (r *RmdirCommand) Execute(args []string) string {
	if len(args) == 0 {
		return "Usage: rmdir <directory>"
	}
	err := os.RemoveAll(args[0])
	if err != nil {
		return "Error: " + err.Error()
	}
	return ""
}

// --- cp command ---
type CpCommand struct{}

func (c *CpCommand) Name() string        { return "cp" }
func (c *CpCommand) Description() string { return "Copy a file" }
func (c *CpCommand) Execute(args []string) string {
	if len(args) < 2 {
		return "Usage: cp <source> <destination>"
	}
	src, err := os.Open(args[0])
	if err != nil {
		return "Error: " + err.Error()
	}
	defer src.Close()
	dst, err := os.Create(args[1])
	if err != nil {
		return "Error: " + err.Error()
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	if err != nil {
		return "Error: " + err.Error()
	}
	return ""
}

// --- mv command ---
type MvCommand struct{}

func (m *MvCommand) Name() string        { return "mv" }
func (m *MvCommand) Description() string { return "Move or rename a file" }
func (m *MvCommand) Execute(args []string) string {
	if len(args) < 2 {
		return "Usage: mv <source> <destination>"
	}
	err := os.Rename(args[0], args[1])
	if err != nil {
		return "Error: " + err.Error()
	}
	return ""
}

// --- whoami command ---
type WhoamiCommand struct{}

func (w *WhoamiCommand) Name() string        { return "whoami" }
func (w *WhoamiCommand) Description() string { return "Show current user" }
func (w *WhoamiCommand) Execute(args []string) string {
	u, err := user.Current()
	if err != nil {
		return "Error: " + err.Error()
	}
	return u.Username
}

// --- hostname command ---
type HostnameCommand struct{}

func (h *HostnameCommand) Name() string        { return "hostname" }
func (h *HostnameCommand) Description() string { return "Show computer name" }
func (h *HostnameCommand) Execute(args []string) string {
	name, err := os.Hostname()
	if err != nil {
		return "Error: " + err.Error()
	}
	return name
}

// --- ver command ---
type VerCommand struct{}

func (v *VerCommand) Name() string        { return "ver" }
func (v *VerCommand) Description() string { return "Show shell version" }
func (v *VerCommand) Execute(args []string) string {
	return "SuperShell v1.0.0"
}

type DirCommand struct{}

func (d *DirCommand) Name() string        { return "dir" }
func (d *DirCommand) Description() string { return "List directory contents (Windows style)" }
func (d *DirCommand) Execute(args []string) string {
	pattern := "*"
	if len(args) > 0 {
		pattern = args[0]
	}
	cwd, _ := os.Getwd()
	entries, err := os.ReadDir(cwd)
	if err != nil {
		return color.New(color.FgRed).Sprint("The system cannot read the directory.")
	}
	fmt.Println("DEBUG: All files in directory:")
	for _, entry := range entries {
		fmt.Println(" -", entry.Name())
	}

	var dirs, regularFiles []os.DirEntry
	for _, entry := range entries {
		patternLower := strings.ToLower(pattern)
		matched, _ := path.Match(patternLower, strings.ToLower(entry.Name()))
		if !matched {
			continue
		}
		if entry.IsDir() {
			dirs = append(dirs, entry)
		} else {
			regularFiles = append(regularFiles, entry)
		}
	}
	if len(dirs) == 0 && len(regularFiles) == 0 {
		return color.New(color.FgRed).Sprint("The system cannot find the path specified.")
	}
	sort.Slice(dirs, func(i, j int) bool { return strings.ToLower(dirs[i].Name()) < strings.ToLower(dirs[j].Name()) })
	sort.Slice(regularFiles, func(i, j int) bool {
		return strings.ToLower(regularFiles[i].Name()) < strings.ToLower(regularFiles[j].Name())
	})

	var out strings.Builder
	out.WriteString(fmt.Sprintf("\n Directory of %s\n\n", strings.ReplaceAll(cwd, "/", "\\")))

	var fileCount, dirCount, totalSize int64

	for _, entry := range dirs {
		info, _ := entry.Info()
		modTime := info.ModTime().Format("01/02/2006  03:04 AM")
		out.WriteString(fmt.Sprintf("%s    <DIR>          %s\n", modTime, color.New(color.FgCyan).Sprint(entry.Name())))
		dirCount++
	}
	for _, entry := range regularFiles {
		info, _ := entry.Info()
		modTime := info.ModTime().Format("01/02/2006  03:04 AM")
		name := entry.Name()
		if strings.HasSuffix(strings.ToLower(name), ".exe") {
			name = color.New(color.FgGreen).Sprint(name)
		} else {
			name = color.New(color.FgWhite).Sprint(name)
		}
		out.WriteString(fmt.Sprintf("%s    %12d %s\n", modTime, info.Size(), name))
		fileCount++
		totalSize += info.Size()
	}
	out.WriteString(fmt.Sprintf("    %d File(s) %d bytes\n", fileCount, totalSize))
	out.WriteString(fmt.Sprintf("    %d Dir(s)\n", dirCount))
	return out.String()
}

type Config struct {
	// Add more settings as needed
}

var configFilePath = "supershell.yaml"
var config *Config

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return &Config{}, nil // default if not found
	}
	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func SaveConfig(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		fmt.Println("YAML marshal error:", err)
		return err
	}
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		fmt.Println("File write error:", err)
	}
	return err
}

func RunScript(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // skip empty lines and comments
		}
		output := Dispatch(line)
		if output != "" {
			fmt.Println(output)
		}
	}
}

func LoadScripts(dir string) {
	files, _ := os.ReadDir(dir)
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".ss") {
			RunScript(filepath.Join(dir, f.Name()))
		}
	}
}

func (c *CdCommand) Completer(args []string) []prompt.Suggest {
	if len(args) > 0 && args[0] == "cd" {
		entries, _ := os.ReadDir(".")
		var dirSugg []prompt.Suggest
		for _, entry := range entries {
			if entry.IsDir() {
				dirSugg = append(dirSugg, prompt.Suggest{Text: entry.Name() + string(os.PathSeparator), Description: "Directory"})
			}
		}
		return prompt.FilterHasPrefix(dirSugg, args[len(args)-1], true)
	}
	return nil
}

func (h *HelpCommand) Completer(args []string) []prompt.Suggest {
	return nil
}

func (c *ClearCommand) Completer(args []string) []prompt.Suggest {
	return nil
}

func (e *EchoCommand) Completer(args []string) []prompt.Suggest {
	return nil
}

func (p *PwdCommand) Completer(args []string) []prompt.Suggest {
	return nil
}

func (l *LsCommand) Completer(args []string) []prompt.Suggest {
	return nil
}

func (m *MkdirCommand) Completer(args []string) []prompt.Suggest {
	return nil
}

func (r *RmCommand) Completer(args []string) []prompt.Suggest {
	return nil
}

func (rm *RmdirCommand) Completer(args []string) []prompt.Suggest {
	return nil
}

func (c *CpCommand) Completer(args []string) []prompt.Suggest {
	return nil
}

func (m *MvCommand) Completer(args []string) []prompt.Suggest {
	return nil
}

func (w *WhoamiCommand) Completer(args []string) []prompt.Suggest {
	return nil
}

func (h *HostnameCommand) Completer(args []string) []prompt.Suggest {
	return nil
}

func (v *VerCommand) Completer(args []string) []prompt.Suggest {
	return nil
}

func (d *DirCommand) Completer(args []string) []prompt.Suggest {
	return nil
}

type WgetCommand struct{}

func (w *WgetCommand) Name() string { return "wget" }
func (w *WgetCommand) Description() string {
	return "Download a file from a URL (usage: wget <url> [filename])"
}
func (w *WgetCommand) Execute(args []string) string {
	if len(args) == 0 {
		return "Usage: wget <url> [filename]"
	}
	url := args[0]
	var cmd *exec.Cmd
	if len(args) > 1 {
		filename := args[1]
		cmd = exec.Command("curl", "-L", "-o", filename, url)
	} else {
		cmd = exec.Command("curl", "-L", "-O", url)
	}
	runningCmd = cmd
	defer func() { runningCmd = nil }()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Sprintf("Download failed: %v", err)
	}
	return "Download complete."
}

type IpconfigCommand struct{}

func (i *IpconfigCommand) Name() string        { return "ipconfig" }
func (i *IpconfigCommand) Description() string { return "Show network interfaces and IP addresses" }
func (i *IpconfigCommand) Execute(args []string) string {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("ipconfig", "/all")
	} else {
		// Try 'ifconfig', fallback to 'ip addr'
		if _, err := exec.LookPath("ifconfig"); err == nil {
			cmd = exec.Command("ifconfig")
		} else {
			cmd = exec.Command("ip", "addr")
		}
	}
	runningCmd = cmd
	defer func() { runningCmd = nil }()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("ipconfig failed: %v\n%s", err, string(out))
	}

	// Colorize output line by line
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		lower := strings.ToLower(line)
		switch {
		case strings.Contains(lower, "adapter") || strings.HasPrefix(lower, "interface") || strings.HasPrefix(lower, "en") || strings.HasPrefix(lower, "eth"):
			color.New(color.FgCyan, color.Bold).Println(line)
		case strings.Contains(lower, "ipv4") || strings.Contains(lower, "inet "):
			color.New(color.FgGreen).Println(line)
		case strings.Contains(lower, "ipv6"):
			color.New(color.FgHiGreen).Println(line)
		case strings.Contains(lower, "physical") || strings.Contains(lower, "mac") || strings.Contains(lower, "ether"):
			color.New(color.FgMagenta).Println(line)
		case strings.Contains(lower, "dns") || strings.Contains(lower, "gateway") || strings.Contains(lower, "router"):
			color.New(color.FgBlue).Println(line)
		default:
			fmt.Println(line)
		}
	}
	return ""
}
