package core

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"os/user"
	"sync"

	"net"
	"strconv"

	"os/signal"
	"syscall"

	prompt "github.com/c-bata/go-prompt"
	"github.com/fatih/color"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/pcapgo"
	"gopkg.in/yaml.v2"
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
		desc := cmd.Description()
		descLines := strings.Split(desc, "\n")
		helpText += "  " + cmd.Name() + "\n"
		for _, line := range descLines {
			if strings.TrimSpace(line) != "" {
				helpText += "    " + line + "\n"
			}
		}
	}
	helpText += `
Alias usage:
  alias                # List all aliases
  alias <name> <cmd>   # Create or update an alias (e.g. alias ll ls -l)
  unalias <name>       # Remove an alias

Type 'help' to see this message again.
`
	if !strings.HasSuffix(helpText, "\n") {
		helpText += "\n"
	}
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
	files, err := os.ReadDir(dir)
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
	data, err := os.ReadFile(args[0])
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
func (h *HostnameCommand) Description() string { return "Show the system hostname" }
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

type NetstatCommand struct{}

func (n *NetstatCommand) Name() string { return "netstat" }
func (n *NetstatCommand) Description() string {
	return "Show open network connections (type 'netstat --help' for options)"
}

func (n *NetstatCommand) Help() string {
	return `Show open network connections

Usage:
  netstat [options]

Options:
  -tcp, --tcp           Show only TCP connections
  -udp, --udp           Show only UDP connections
  -state <STATE>        Filter by connection state (e.g. ESTABLISHED, LISTEN)
  -p, --process <PID>   Filter by process ID
  :<port>               Filter by local port (e.g. :80)
  --sort <column>       Sort by column (proto, local, remote, state, pid)
  --desc                Sort descending
  --group               Group by state
  --csv                 Export as CSV
  --json                Export as JSON
  --user                (Not yet implemented) Show only connections for current user

Interactive filter:
  After output, type to filter live. Type enter on empty line to exit filter.`
}

type NetstatEntry struct {
	Proto, Local, Remote, State, PID string
	RawLine                          string
}

func highlightFilter(line, filter string) string {
	if filter == "" {
		return line
	}
	lowerLine := strings.ToLower(line)
	lowerFilter := strings.ToLower(filter)
	var result strings.Builder
	i := 0
	for i < len(line) {
		if len(lowerLine[i:]) >= len(lowerFilter) && lowerLine[i:i+len(lowerFilter)] == lowerFilter {
			// Highlight the match
			result.WriteString(color.New(color.BgYellow, color.FgBlack, color.Bold).Sprint(line[i : i+len(lowerFilter)]))
			i += len(lowerFilter)
		} else {
			result.WriteByte(line[i])
			i++
		}
	}
	return result.String()
}

func badge(text, colorName string) string {
	var c *color.Color
	switch colorName {
	case "green":
		c = color.New(color.FgGreen, color.Bold)
	case "yellow":
		c = color.New(color.FgYellow, color.Bold)
	case "red":
		c = color.New(color.FgRed, color.Bold)
	case "blue":
		c = color.New(color.FgBlue, color.Bold)
	case "magenta":
		c = color.New(color.FgMagenta, color.Bold)
	default:
		c = color.New(color.Bold)
	}
	return c.Sprintf("[%s]", text)
}

func modernNetstatDisplay(entries []NetstatEntry, filter string) {
	// Print a summary bar
	total, established, listening := 0, 0, 0
	for _, e := range entries {
		if filter == "" || strings.Contains(strings.ToLower(e.RawLine), strings.ToLower(filter)) {
			total++
			if strings.Contains(strings.ToLower(e.State), "established") {
				established++
			}
			if strings.Contains(strings.ToLower(e.State), "listen") {
				listening++
			}
		}
	}
	color.New(color.BgBlue, color.FgWhite, color.Bold).Printf(" Netstat Dashboard ")
	fmt.Printf("  Total: %d  ", total)
	color.New(color.BgGreen, color.FgBlack).Printf(" ESTABLISHED: %d ", established)
	color.New(color.BgYellow, color.FgBlack).Printf(" LISTENING: %d ", listening)
	fmt.Println("\n")

	// Print headers
	color.New(color.FgCyan, color.Bold).Printf("%-8s %-25s %-25s %-15s %-8s\n", "PROTO", "LOCAL", "REMOTE", "STATE", "PID")
	color.New(color.FgHiBlack).Println(strings.Repeat("─", 90))

	// Print entries
	for _, e := range entries {
		if filter == "" || strings.Contains(strings.ToLower(e.RawLine), strings.ToLower(filter)) {
			// Protocol icon
			protoIcon := ""
			switch strings.ToLower(e.Proto) {
			case "tcp":
				protoIcon = "🌐"
			case "udp":
				protoIcon = "📡"
			default:
				protoIcon = ""
			}
			// State badge
			stateBadge := ""
			stateLower := strings.ToLower(e.State)
			switch {
			case strings.Contains(stateLower, "established"):
				stateBadge = badge("ESTABLISHED", "green")
			case strings.Contains(stateLower, "listen"):
				stateBadge = badge("LISTEN", "yellow")
			case strings.Contains(stateLower, "close"):
				stateBadge = badge("CLOSE", "red")
			default:
				stateBadge = badge(e.State, "blue")
			}
			// Highlight filter in addresses
			local := highlightFilter(e.Local, filter)
			remote := highlightFilter(e.Remote, filter)
			pid := highlightFilter(e.PID, filter)
			// Print row
			fmt.Printf("%-2s %-6s %-25s %-25s %-15s %-8s\n",
				protoIcon, e.Proto, local, remote, stateBadge, pid)
		}
	}
	fmt.Println()
}

func filterAndDisplay(entries []NetstatEntry, filter string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for _, e := range entries {
		if filter == "" || strings.Contains(strings.ToLower(e.RawLine), strings.ToLower(filter)) {
			// Highlight the filter in the output
			fmt.Fprintln(w, highlightFilter(e.RawLine, filter))
		}
	}
	w.Flush()
}

func (n *NetstatCommand) Execute(args []string) string {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			fmt.Println(n.Help())
			return ""
		}
	}

	proto := ""
	port := ""
	state := ""
	pid := ""
	exportCSV := false
	exportJSON := false
	sortColumn := ""
	sortAsc := true
	groupByState := false

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-tcp" || arg == "--tcp":
			proto = "tcp"
		case arg == "-udp" || arg == "--udp":
			proto = "udp"
		case arg == "-state" && i+1 < len(args):
			state = strings.ToLower(args[i+1])
			i++
		case (arg == "-p" || arg == "--process") && i+1 < len(args):
			pid = args[i+1]
			i++
		case arg == "--csv":
			exportCSV = true
		case arg == "--json":
			exportJSON = true
		case arg == "--sort":
			if i+1 < len(args) {
				sortColumn = strings.ToLower(args[i+1])
				i++
			}
		case arg == "--desc":
			sortAsc = false
		case arg == "--group":
			groupByState = true
		case arg == "--user":
			// This feature is complex and requires platform-specific logic
			// For now, we'll just print a placeholder message.
			return "User-specific filtering (--user) is not yet implemented."
		case strings.HasPrefix(arg, ":"):
			port = arg[1:]
		}
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("netstat", "-ano")
	} else if _, err := exec.LookPath("ss"); err == nil {
		cmd = exec.Command("ss", "-tunap")
	} else {
		cmd = exec.Command("netstat", "-tunap")
	}
	runningCmd = cmd
	defer func() { runningCmd = nil }()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("netstat failed: %v\n%s", err, string(out))
	}

	lines := strings.Split(string(out), "\n")
	var entries []NetstatEntry
	for _, line := range lines {
		lower := strings.ToLower(line)
		if strings.TrimSpace(line) == "" || strings.Contains(lower, "proto") || strings.Contains(lower, "state") || strings.Contains(lower, "local address") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		entry := NetstatEntry{
			Proto:   fields[0],
			Local:   fields[1],
			Remote:  fields[2],
			State:   "",
			PID:     "",
			RawLine: line,
		}
		if len(fields) >= 5 {
			entry.State = fields[3]
			entry.PID = fields[len(fields)-1]
		}
		entries = append(entries, entry)
	}

	if sortColumn != "" {
		sortNetstatEntries(entries, sortColumn, sortAsc)
	}

	var headerPrinted bool
	var total, tcpCount, udpCount, established, listening int
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for _, line := range lines {
		lower := strings.ToLower(line)
		if strings.TrimSpace(line) == "" {
			continue
		}
		// Print header in bold cyan
		if !headerPrinted && (strings.Contains(lower, "proto") || strings.Contains(lower, "state") || strings.Contains(lower, "local address")) {
			color.New(color.FgCyan, color.Bold).Fprintln(w, line)
			headerPrinted = true
			continue
		}
		// Filtering
		if proto != "" && !strings.Contains(lower, proto) {
			continue
		}
		portPattern := ""
		if port != "" {
			portPattern = fmt.Sprintf(`[:.]%s(\\D|$)`, regexp.QuoteMeta(port))
			re := regexp.MustCompile(portPattern)
			if !re.MatchString(line) {
				continue
			}
		}
		if state != "" && !strings.Contains(lower, state) {
			continue
		}
		if pid != "" && !strings.Contains(line, pid) {
			continue
		}
		// Count stats
		total++
		if strings.Contains(lower, "tcp") {
			tcpCount++
		}
		if strings.Contains(lower, "udp") {
			udpCount++
		}
		if strings.Contains(lower, "established") {
			established++
		}
		if strings.Contains(lower, "listen") {
			listening++
		}
		// Colorize columns
		fields := strings.Fields(line)
		for i, f := range fields {
			switch {
			case i == 0 && (f == "TCP" || f == "tcp"):
				fmt.Fprint(w, color.BlueString(f)+"\t")
			case i == 0 && (f == "UDP" || f == "udp"):
				fmt.Fprint(w, color.MagentaString(f)+"\t")
			case strings.Contains(strings.ToLower(f), "established"):
				fmt.Fprint(w, color.HiGreenString(f)+"\t")
			case strings.Contains(strings.ToLower(f), "listen"):
				fmt.Fprint(w, color.GreenString(f)+"\t")
			case strings.Contains(strings.ToLower(f), "close"):
				fmt.Fprint(w, color.RedString(f)+"\t")
			case strings.Contains(f, ":"):
				fmt.Fprint(w, color.CyanString(f)+"\t")
			case i == len(fields)-1 && len(f) < 8 && f != "-" && f != "0":
				fmt.Fprint(w, color.YellowString(f)+"\t") // PID
			default:
				fmt.Fprint(w, f+"\t")
			}
		}
		fmt.Fprintln(w)
	}
	w.Flush()
	color.New(color.FgHiBlack).Printf("\nTotal: %d | TCP: %d | UDP: %d | ESTABLISHED: %d | LISTENING: %d\n", total, tcpCount, udpCount, established, listening)
	color.New(color.FgHiBlack).Println("Options: -tcp, -udp, -state <STATE>, -p/--process <PID>, :<port> (e.g. netstat -tcp :80 -state established)")

	// Export to CSV
	if exportCSV {
		fmt.Println("proto,local,remote,state,pid")
		for _, e := range entries {
			fmt.Printf("%s,%s,%s,%s,%s\n", e.Proto, e.Local, e.Remote, e.State, e.PID)
		}
		return ""
	}
	// Export to JSON
	if exportJSON {
		data, err := json.MarshalIndent(entries, "", "  ")
		if err != nil {
			return "Failed to marshal JSON: " + err.Error()
		}
		fmt.Println(string(data))
		return ""
	}

	// Interactive filtering loop
	filter := ""
	for {
		if groupByState {
			modernNetstatDisplayGrouped(entries, filter)
		} else {
			modernNetstatDisplay(entries, filter)
		}
		topPorts(entries, 5) // Show top 5 ports
		filter = prompt.Input("Filter (enter to exit): ", func(d prompt.Document) []prompt.Suggest { return nil })
		if filter == "" {
			break
		}
		fmt.Print("\033[H\033[2J") // Clear screen
	}

	return ""
}

func groupByState(entries []NetstatEntry) map[string][]NetstatEntry {
	groups := make(map[string][]NetstatEntry)
	for _, e := range entries {
		state := strings.ToUpper(e.State)
		groups[state] = append(groups[state], e)
	}
	return groups
}

func modernNetstatDisplayGrouped(entries []NetstatEntry, filter string) {
	groups := groupByState(entries)
	for state, group := range groups {
		color.New(color.FgHiMagenta, color.Bold).Printf("\n=== %s ===\n", state)
		for _, e := range group {
			if filter == "" || strings.Contains(strings.ToLower(e.RawLine), strings.ToLower(filter)) {
				// Protocol icon
				protoIcon := ""
				switch strings.ToLower(e.Proto) {
				case "tcp":
					protoIcon = "🌐"
				case "udp":
					protoIcon = "📡"
				default:
					protoIcon = ""
				}
				// State badge
				stateBadge := ""
				stateLower := strings.ToLower(e.State)
				switch {
				case strings.Contains(stateLower, "established"):
					stateBadge = badge("ESTABLISHED", "green")
				case strings.Contains(stateLower, "listen"):
					stateBadge = badge("LISTEN", "yellow")
				case strings.Contains(stateLower, "close"):
					stateBadge = badge("CLOSE", "red")
				default:
					stateBadge = badge(e.State, "blue")
				}
				// Highlight filter in addresses
				local := highlightFilter(e.Local, filter)
				remote := highlightFilter(e.Remote, filter)
				pid := highlightFilter(e.PID, filter)
				// Print row
				fmt.Printf("%-2s %-6s %-25s %-25s %-15s %-8s\n",
					protoIcon, e.Proto, local, remote, stateBadge, pid)
			}
		}
	}
}

func topPorts(entries []NetstatEntry, n int) {
	portCount := make(map[string]int)
	for _, e := range entries {
		parts := strings.Split(e.Local, ":")
		if len(parts) > 1 {
			port := parts[len(parts)-1]
			portCount[port]++
		}
	}
	// Sort and print top N
	// ...
}

func sortNetstatEntries(entries []NetstatEntry, column string, asc bool) {
	sort.Slice(entries, func(i, j int) bool {
		var a, b string
		switch column {
		case "proto":
			a, b = entries[i].Proto, entries[j].Proto
		case "local":
			a, b = entries[i].Local, entries[j].Local
		case "remote":
			a, b = entries[i].Remote, entries[j].Remote
		case "state":
			a, b = entries[i].State, entries[j].State
		case "pid":
			a, b = entries[i].PID, entries[j].PID
		default:
			a, b = entries[i].RawLine, entries[j].RawLine
		}
		if asc {
			return a < b
		}
		return a > b
	})
}

type ArpCommand struct{}

func (a *ArpCommand) Name() string { return "arp" }
func (a *ArpCommand) Description() string {
	return `Show the ARP table

Usage:
  arp

Options:
  (no options yet)

Shows the system ARP table. On Windows, uses 'arp -a'. On Unix, uses 'ip neigh' or 'arp -a'.`
}
func (a *ArpCommand) Execute(args []string) string {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("arp", "-a")
	} else if _, err := exec.LookPath("ip"); err == nil {
		cmd = exec.Command("ip", "neigh")
	} else {
		cmd = exec.Command("arp", "-a")
	}
	runningCmd = cmd
	defer func() { runningCmd = nil }()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("arp failed: %v\n%s", err, string(out))
	}
	// Colorize output
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		lower := strings.ToLower(line)
		switch {
		case strings.Contains(lower, "dynamic"):
			color.New(color.FgGreen).Println(line)
		case strings.Contains(lower, "static"):
			color.New(color.FgCyan).Println(line)
		case strings.Contains(lower, "incomplete"):
			color.New(color.FgRed).Println(line)
		default:
			fmt.Println(line)
		}
	}
	return ""
}

type RouteCommand struct{}

func (r *RouteCommand) Name() string { return "route" }
func (r *RouteCommand) Description() string {
	return `route - Show the routing table

  Usage:
    route

  Options:
    (no options yet)

  Notes:
    - Shows the system routing table
    - On Windows, uses 'route print'
    - On Unix, uses 'ip route' or 'netstat -rn'
`
}
func (r *RouteCommand) Execute(args []string) string {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("route", "print")
	} else if _, err := exec.LookPath("ip"); err == nil {
		cmd = exec.Command("ip", "route")
	} else {
		cmd = exec.Command("netstat", "-rn")
	}
	runningCmd = cmd
	defer func() { runningCmd = nil }()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("route failed: %v\n%s", err, string(out))
	}
	// Colorize output
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		lower := strings.ToLower(line)
		switch {
		case strings.Contains(lower, "default") || strings.Contains(lower, "gateway"):
			color.New(color.FgGreen, color.Bold).Println(line)
		case strings.Contains(lower, "metric"):
			color.New(color.FgCyan).Println(line)
		case strings.Contains(lower, "interface"):
			color.New(color.FgYellow).Println(line)
		default:
			fmt.Println(line)
		}
	}
	return ""
}

type SpeedtestCommand struct{}

func (s *SpeedtestCommand) Name() string { return "speedtest" }
func (s *SpeedtestCommand) Description() string {
	return "Run a Go-native speed test (usage: speedtest)"
}
func (s *SpeedtestCommand) Execute(args []string) string {
	// Try to run fast if available
	if _, err := exec.LookPath("fast"); err == nil {
		cmd := exec.Command("fast")
		runningCmd = cmd
		defer func() { runningCmd = nil }()
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return "fast CLI failed: " + err.Error()
		}
		return ""
	}
	// Try to install fast-cli automatically
	fmt.Println("fast CLI not found. Attempting to install fast-cli globally with npm...")
	installCmd := exec.Command("npm", "install", "--global", "fast-cli")
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	if err := installCmd.Run(); err != nil {
		return "Failed to install fast-cli with npm. Please install Node.js and npm, then run: npm install --global fast-cli"
	}
	// Try again to run fast
	if _, err := exec.LookPath("fast"); err == nil {
		cmd := exec.Command("fast")
		runningCmd = cmd
		defer func() { runningCmd = nil }()
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return "fast CLI failed: " + err.Error()
		}
		return ""
	}
	return `fast CLI could not be found or installed.\nPlease ensure Node.js and npm are installed, then run:\n  npm install --global fast-cli\nOr visit https://github.com/ddo/fast for more info.`
}

type HelpEntry struct {
	Name        string
	Description string
}

func GenerateHTMLHelp(filename string) error {
	var entries []HelpEntry
	for name, cmd := range commandRegistry {
		desc := cmd.Description()
		if h, ok := cmd.(interface{ Help() string }); ok {
			desc = h.Help()
		}
		entries = append(entries, HelpEntry{Name: name, Description: desc})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })

	htmlTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>SuperShell Command Reference</title>
    <style>
        body { font-family: 'Segoe UI', Arial, sans-serif; background: #181c20; color: #e0e0e0; margin: 0; padding: 2em; }
        h1 { color: #4ec9b0; }
        .index { margin-bottom: 2em; }
        .index a { color: #4ec9b0; text-decoration: none; margin-right: 1em; font-weight: bold; }
        .command { background: #23272e; border-radius: 8px; margin: 2em 0; padding: 1.5em; box-shadow: 0 2px 8px #0003; }
        .command-name { font-size: 1.3em; color: #569cd6; font-weight: bold; margin-bottom: 0.5em; }
        .command-desc { margin-top: 0.5em; white-space: pre-line; color: #e0e0e0; }
        .footer { margin-top: 3em; color: #888; font-size: 0.9em; text-align: center; }
        .timestamp { color: #4ec9b0; font-size: 0.95em; }
        a.anchor { display: block; position: relative; top: -80px; visibility: hidden; }
    </style>
</head>
<body>
    <h1>SuperShell Command Reference</h1>
    <div class="timestamp">Generated: {{.Timestamp}}</div>
    <div class="index">
        {{range .Entries}}<a href="#{{.Name}}">{{.Name}}</a>{{end}}
    </div>
    {{range .Entries}}
    <a class="anchor" id="{{.Name}}"></a>
    <div class="command">
        <div class="command-name">{{.Name}}</div>
        <div class="command-desc">{{.Description}}</div>
    </div>
    {{end}}
    <div class="footer">SuperShell &copy; {{.Year}} &mdash; <a href="https://github.com/yourrepo" style="color:#4ec9b0;">Project Home</a></div>
</body>
</html>
`
	type htmlData struct {
		Entries   []HelpEntry
		Timestamp string
		Year      int
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	tmpl := template.Must(template.New("help").Parse(htmlTemplate))
	now := time.Now()
	data := htmlData{
		Entries:   entries,
		Timestamp: now.Format("2006-01-02 15:04:05 MST"),
		Year:      now.Year(),
	}
	return tmpl.Execute(f, data)
}

type HelpHTMLCommand struct{}

func (h *HelpHTMLCommand) Name() string        { return "helphtml" }
func (h *HelpHTMLCommand) Description() string { return "Generate an HTML help file for all commands" }
func (h *HelpHTMLCommand) Execute(args []string) string {
	filename := "help.html"
	if len(args) > 0 {
		filename = args[0]
	}
	err := GenerateHTMLHelp(filename)
	if err != nil {
		return "Failed to generate HTML help: " + err.Error()
	}
	return "HTML help file generated: " + filename
}

// PortscanCommand: Fast TCP port scanner
type PortscanCommand struct{}

func (p *PortscanCommand) Name() string { return "portscan" }
func (p *PortscanCommand) Description() string {
	return "Scan TCP ports on a host (usage: portscan <host> [ports])"
}
func (p *PortscanCommand) Execute(args []string) string {
	if len(args) == 0 {
		return "Usage: portscan <host> [ports]"
	}
	host := args[0]
	ports := []int{}
	if len(args) > 1 {
		// Parse ports: single, comma, or range
		for _, part := range strings.Split(args[1], ",") {
			if strings.Contains(part, "-") {
				rangeParts := strings.SplitN(part, "-", 2)
				start, _ := strconv.Atoi(rangeParts[0])
				end, _ := strconv.Atoi(rangeParts[1])
				for i := start; i <= end; i++ {
					ports = append(ports, i)
				}
			} else {
				p, _ := strconv.Atoi(part)
				ports = append(ports, p)
			}
		}
	} else {
		for i := 1; i <= 1024; i++ {
			ports = append(ports, i)
		}
	}
	results := make([]string, len(ports))
	var wg sync.WaitGroup
	wg.Add(len(ports))
	for i, port := range ports {
		go func(i, port int) {
			defer wg.Done()
			address := fmt.Sprintf("%s:%d", host, port)
			conn, err := net.DialTimeout("tcp", address, 500*time.Millisecond)
			if err == nil {
				conn.Close()
				results[i] = color.New(color.FgGreen).Sprintf("%5d OPEN", port)
			} else {
				results[i] = color.New(color.FgRed).Sprintf("%5d closed", port)
			}
		}(i, port)
	}
	wg.Wait()
	openCount := 0
	for _, r := range results {
		if strings.Contains(r, "OPEN") {
			openCount++
		}
	}
	return fmt.Sprintf("Port scan results for %s (open ports in green):\n%s\n%d open, %d closed", host, strings.Join(results, "\n"), openCount, len(ports)-openCount)
}

type NetdiscoverCommand struct{}

func (n *NetdiscoverCommand) Name() string { return "netdiscover" }
func (n *NetdiscoverCommand) Description() string {
	return "Discover live hosts on a subnet (usage: netdiscover <CIDR>)"
}
func (n *NetdiscoverCommand) Execute(args []string) string {
	if len(args) == 0 {
		return "Usage: netdiscover <CIDR> (e.g., netdiscover 192.168.1.0/24)"
	}
	subnet := args[0]
	ip, ipnet, err := net.ParseCIDR(subnet)
	if err != nil {
		return "Invalid CIDR: " + err.Error()
	}
	var hosts []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); incIP(ip) {
		hosts = append(hosts, ip.String())
	}
	// Remove network and broadcast addresses
	if len(hosts) > 2 {
		hosts = hosts[1 : len(hosts)-1]
	}
	results := make([]string, len(hosts))
	var wg sync.WaitGroup
	wg.Add(len(hosts))
	for i, host := range hosts {
		go func(i int, host string) {
			defer wg.Done()
			alive := false
			// Try ICMP ping (using system ping command)
			var pingCmd *exec.Cmd
			if runtime.GOOS == "windows" {
				pingCmd = exec.Command("ping", "-n", "1", "-w", "500", host)
			} else {
				pingCmd = exec.Command("ping", "-c", "1", "-W", "1", host)
			}
			err := pingCmd.Run()
			if err == nil {
				alive = true
			} else {
				// Fallback: try TCP connect to port 80
				conn, err := net.DialTimeout("tcp", host+":80", 500*time.Millisecond)
				if err == nil {
					alive = true
					conn.Close()
				}
			}
			if alive {
				results[i] = color.New(color.FgGreen).Sprintf("%s alive", host)
			} else {
				results[i] = color.New(color.FgRed).Sprintf("%s unreachable", host)
			}
		}(i, host)
	}
	wg.Wait()
	aliveCount := 0
	for _, r := range results {
		if strings.Contains(r, "alive") {
			aliveCount++
		}
	}
	return fmt.Sprintf("Network discovery results for %s (alive hosts in green):\n%s\n%d alive, %d unreachable", subnet, strings.Join(results, "\n"), aliveCount, len(hosts)-aliveCount)
}

// Helper to increment an IP address
func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

type SniffCommand struct{}

func (s *SniffCommand) Name() string { return "sniff" }
func (s *SniffCommand) Description() string {
	return `sniff - Packet sniffer

  Usage:
    sniff <iface|index> [file.pcap] [max_packets] [bpf_filter]

  Options:
    <iface|index>    Interface name or index to capture from (required)
    [file.pcap]      Optional file to save packets (Wireshark-compatible)
    [max_packets]    Optional max packets to capture (default: 50)
    [bpf_filter]     Optional BPF filter (e.g. "tcp port 443")

  Examples:
    sniff 2
    sniff 2 capture.pcap 200
    sniff 2 "" 100 "tcp"
    sniff 2 capture.pcap 100 "tcp port 443 or port 80"
    sniff 2 "" 50 "tcp port 22"

  Filter examples:
    "tcp"                      (all TCP traffic)
    "port 80"                  (HTTP)
    "tcp port 443 or port 80"  (HTTPS or HTTP)
    "tcp and port 22"          (SSH)

  Notes:
    - Saves to .pcap if file is specified
    - Default max_packets is 50
    - BPF filter is optional
    - Use sniff with no arguments to list interfaces and see this help
`
}
func (s *SniffCommand) Execute(args []string) string {
	ifs, err := pcap.FindAllDevs()
	if err != nil {
		return "Error finding interfaces: " + err.Error() + "\nMake sure Npcap is installed (https://nmap.org/npcap/) and you have permission."
	}
	if len(ifs) == 0 {
		return "No network interfaces found."
	}
	iface := ""
	var pcapFile *os.File
	var pcapWriter *pcapgo.Writer
	if len(args) > 0 {
		arg := args[0]
		// Try as index
		if idx, err := strconv.Atoi(arg); err == nil && idx > 0 && idx <= len(ifs) {
			iface = ifs[idx-1].Name
		} else {
			// Try as name
			for _, dev := range ifs {
				if dev.Name == arg {
					iface = dev.Name
					break
				}
			}
		}
		if iface == "" {
			return "Interface not found: " + arg
		}
	} else {
		// No argument: print all interfaces and usage
		var b strings.Builder
		b.WriteString("Available interfaces:\n")
		for i, dev := range ifs {
			b.WriteString(fmt.Sprintf("  %d: %s (%s)\n", i+1, dev.Name, dev.Description))
		}
		b.WriteString("\nUsage: sniff <iface|index> [file.pcap] [max_packets] [bpf_filter]\n")
		b.WriteString("Example: sniff 2 capture.pcap 200 'tcp port 443'\n")
		b.WriteString("Filter examples: 'tcp', 'port 80', 'tcp port 443 or port 80', 'tcp and port 22'\n")
		return b.String()
	}
	// Check for optional pcap file argument
	if len(args) > 1 && args[1] != "" {
		filename := args[1]
		f, err := os.Create(filename)
		if err != nil {
			return "Failed to create pcap file: " + err.Error()
		}
		pcapFile = f
		defer pcapFile.Close()
	}

	maxPackets := 50
	if len(args) > 2 {
		if n, err := strconv.Atoi(args[2]); err == nil && n > 0 {
			maxPackets = n
		}
	}

	bpfFilter := ""
	if len(args) > 3 {
		bpfFilter = args[3]
	}

	fmt.Printf("Sniffing on interface: %s\nPress Ctrl+C to stop.\n", iface)
	handle, err := pcap.OpenLive(iface, 1600, true, pcap.BlockForever)
	if err != nil {
		return "Error opening interface: " + err.Error()
	}
	defer handle.Close()

	if bpfFilter != "" {
		err := handle.SetBPFFilter(bpfFilter)
		if err != nil {
			return fmt.Sprintf("Failed to set BPF filter '%s': %v", bpfFilter, err)
		}
		fmt.Printf("BPF filter applied: %s\n", bpfFilter)
	}

	if pcapFile != nil {
		pcapWriter = pcapgo.NewWriter(pcapFile)
		err := pcapWriter.WriteFileHeader(1600, handle.LinkType())
		if err != nil {
			return "Failed to write pcap file header: " + err.Error()
		}
		fmt.Printf("Saving packets to: %s\n", pcapFile.Name())
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	count := 0

	// Signal handling for Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT)
	defer signal.Stop(sigChan)

	packetChan := packetSource.Packets()
	stopped := false

	// Spinner for live feedback
	spinnerDone := make(chan struct{})
	var lastSrc, lastDst, lastSport, lastDport string
	var lastProto string
	var packetCount int
	go func() {
		spinner := []string{"|", "/", "-", "\\"}
		i := 0
		for {
			select {
			case <-spinnerDone:
				fmt.Print("\r") // Clear spinner line
				return
			default:
				msg := fmt.Sprintf("\rCapturing packets... %s  [count: %d]  ", spinner[i%len(spinner)], packetCount)
				if lastSrc != "" && lastDst != "" {
					msg += fmt.Sprintf("last: %s:%s → %s:%s (%s)", lastSrc, lastSport, lastDst, lastDport, lastProto)
				}
				fmt.Print(msg)
				time.Sleep(150 * time.Millisecond)
				i++
			}
		}
	}()

	for !stopped {
		select {
		case <-sigChan:
			fmt.Println("\n(Ctrl+C detected. Stopping sniff.)")
			stopped = true
		case packet, ok := <-packetChan:
			if !ok {
				stopped = true
				break
			}
			// Clear spinner line before printing packet info
			fmt.Print("\r\033[K")
			count++
			packetCount = count
			netLayer := packet.NetworkLayer()
			transLayer := packet.TransportLayer()
			if netLayer == nil || transLayer == nil {
				continue
			}
			var proto, src, dst, sport, dport string
			proto = transLayer.LayerType().String()
			src = netLayer.NetworkFlow().Src().String()
			dst = netLayer.NetworkFlow().Dst().String()
			sport, dport = "", ""
			switch t := transLayer.(type) {
			case *layers.TCP:
				sport = fmt.Sprintf("%d", t.SrcPort)
				dport = fmt.Sprintf("%d", t.DstPort)
			case *layers.UDP:
				sport = fmt.Sprintf("%d", t.SrcPort)
				dport = fmt.Sprintf("%d", t.DstPort)
			}
			lastSrc, lastDst, lastSport, lastDport, lastProto = src, dst, sport, dport, proto
			fmt.Printf("%4d %-6s %15s:%-5s -> %15s:%-5s\n", count, proto, src, sport, dst, dport)
			// Write to pcap file if enabled
			if pcapWriter != nil {
				ci := packet.Metadata().CaptureInfo
				pcapWriter.WritePacket(ci, packet.Data())
			}
			if count >= maxPackets {
				fmt.Printf("(Limit reached: %d packets. Stopping sniff.)\n", maxPackets)
				stopped = true
			}
		}
	}
	close(spinnerDone)
	fmt.Println("") // Ensure prompt is on a new line after capture ends
	return ""
}
