package filesystem

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"suppercommand/internal/commands"

	"github.com/fatih/color"
)

// DirCommand lists directory contents (Windows-style)
type DirCommand struct {
	*commands.BaseCommand
}

// NewDirCommand creates a new dir command
func NewDirCommand() *DirCommand {
	return &DirCommand{
		BaseCommand: commands.NewBaseCommand(
			"dir",
			"List directory contents (Windows-style)",
			"dir [directory]",
			[]string{"windows", "linux", "darwin"},
			false,
		),
	}
}

// Execute lists directory contents in enhanced Windows style
func (d *DirCommand) Execute(ctx context.Context, args *commands.Arguments) (*commands.Result, error) {
	startTime := time.Now()

	// Parse arguments
	dir := "."
	pattern := "*"
	showAll := false
	sortByName := false
	sortBySize := false
	sortByDate := false

	for _, arg := range args.Raw {
		if strings.HasPrefix(arg, "/") || strings.HasPrefix(arg, "-") {
			// Handle flags
			switch strings.ToLower(arg) {
			case "/a", "-a", "--all":
				showAll = true
			case "/on", "-n", "--name":
				sortByName = true
			case "/os", "-s", "--size":
				sortBySize = true
			case "/od", "-d", "--date":
				sortByDate = true
			}
		} else if strings.Contains(arg, "*") || strings.Contains(arg, "?") {
			// It's a pattern
			pattern = arg
			dir = filepath.Dir(arg)
			if dir == "." {
				dir = "."
			}
		} else {
			// It's a directory
			dir = arg
		}
	}

	// Get absolute path
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return &commands.Result{
			Output:   "",
			Error:    err,
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, err
	}

	// Read directory
	entries, err := os.ReadDir(dir)
	if err != nil {
		return &commands.Result{
			Output:   "",
			Error:    err,
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, err
	}

	// Filter entries by pattern
	var filteredEntries []os.DirEntry
	for _, entry := range entries {
		// Skip hidden files unless -a flag is used
		if !showAll && strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		// Match pattern
		matched, err := filepath.Match(pattern, entry.Name())
		if err == nil && matched {
			filteredEntries = append(filteredEntries, entry)
		}
	}

	entries = filteredEntries

	// Sort entries
	d.sortEntries(entries, sortByName, sortBySize, sortByDate)

	var output strings.Builder

	// Enhanced header with orange styling
	headerColor := color.New(color.FgHiYellow, color.Bold)
	pathColor := color.New(color.FgYellow, color.Underline)

	output.WriteString(headerColor.Sprint("📁 Directory Listing\n"))
	output.WriteString("═══════════════════════════════════════════════════════════════\n")
	output.WriteString(fmt.Sprintf(" 📂 Path: %s\n", pathColor.Sprint(absDir)))
	output.WriteString("───────────────────────────────────────────────────────────────\n\n")

	// Enhanced color scheme with orange tones
	dirColor := color.New(color.FgHiYellow, color.Bold) // Bright orange for directories
	dirIconColor := color.New(color.FgYellow)           // Orange for directory icons
	fileColor := color.New(color.FgWhite)               // White for regular files
	exeColor := color.New(color.FgHiGreen, color.Bold)  // Green for executables
	docColor := color.New(color.FgHiBlue)               // Blue for documents
	imageColor := color.New(color.FgHiMagenta)          // Magenta for images
	archiveColor := color.New(color.FgHiCyan)           // Cyan for archives
	sizeColor := color.New(color.FgHiBlack)             // Gray for sizes
	dateColor := color.New(color.FgHiBlack)             // Gray for dates

	totalFiles := 0
	totalDirs := 0
	totalSize := int64(0)

	// Separate directories and files for better organization
	var directories []os.DirEntry
	var files []os.DirEntry

	for _, entry := range entries {
		if entry.IsDir() {
			directories = append(directories, entry)
		} else {
			files = append(files, entry)
		}
	}

	// Display directories first with enhanced formatting
	if len(directories) > 0 {
		output.WriteString(dirIconColor.Sprint("📁 DIRECTORIES\n"))
		output.WriteString("───────────────────────────────────────────────────────────────\n")

		for _, entry := range directories {
			info, err := entry.Info()
			if err != nil {
				continue
			}

			modTime := info.ModTime().Format("01/02/2006  03:04 PM")

			output.WriteString(fmt.Sprintf("%s    %s  %s\n",
				dateColor.Sprint(modTime),
				dirColor.Sprint("📁 <DIR>    "),
				dirColor.Sprint(entry.Name())))
			totalDirs++
		}
		output.WriteString("\n")
	}

	// Display files with enhanced formatting and file type colors
	if len(files) > 0 {
		output.WriteString(dirIconColor.Sprint("📄 FILES\n"))
		output.WriteString("───────────────────────────────────────────────────────────────\n")

		for _, entry := range files {
			info, err := entry.Info()
			if err != nil {
				continue
			}

			modTime := info.ModTime().Format("01/02/2006  03:04 PM")
			size := info.Size()

			// Determine file type and color
			fileName := entry.Name()
			fileExt := strings.ToLower(filepath.Ext(fileName))
			var coloredName string
			var icon string

			switch {
			case d.isExecutable(fileExt):
				icon = "⚡"
				coloredName = exeColor.Sprint(fileName)
			case d.isDocument(fileExt):
				icon = "📄"
				coloredName = docColor.Sprint(fileName)
			case d.isImage(fileExt):
				icon = "🖼️"
				coloredName = imageColor.Sprint(fileName)
			case d.isArchive(fileExt):
				icon = "📦"
				coloredName = archiveColor.Sprint(fileName)
			case d.isCode(fileExt):
				icon = "💻"
				coloredName = color.New(color.FgHiGreen).Sprint(fileName)
			case d.isConfig(fileExt):
				icon = "⚙️"
				coloredName = color.New(color.FgHiYellow).Sprint(fileName)
			default:
				icon = "📄"
				coloredName = fileColor.Sprint(fileName)
			}

			// Format size with appropriate units
			sizeStr := d.formatFileSize(size)

			output.WriteString(fmt.Sprintf("%s %s %s %s %s\n",
				dateColor.Sprint(modTime),
				sizeColor.Sprintf("%12s", sizeStr),
				icon,
				coloredName,
				d.getFileTypeDescription(fileExt)))

			totalFiles++
			totalSize += size
		}
	}

	// Enhanced summary with orange styling
	output.WriteString("\n")
	output.WriteString("═══════════════════════════════════════════════════════════════\n")
	summaryColor := color.New(color.FgHiYellow, color.Bold)
	statsColor := color.New(color.FgYellow)

	output.WriteString(summaryColor.Sprint("📊 SUMMARY\n"))
	output.WriteString("───────────────────────────────────────────────────────────────\n")
	output.WriteString(fmt.Sprintf("%s %s\n",
		statsColor.Sprint("📁 Directories:"),
		color.New(color.FgHiWhite, color.Bold).Sprintf("%d", totalDirs)))
	output.WriteString(fmt.Sprintf("%s %s (%s)\n",
		statsColor.Sprint("📄 Files:      "),
		color.New(color.FgHiWhite, color.Bold).Sprintf("%d", totalFiles),
		color.New(color.FgHiCyan).Sprint(d.formatFileSize(totalSize))))

	// Show available space if possible
	if stat, err := os.Stat(dir); err == nil {
		if statT, ok := stat.Sys().(*os.FileInfo); ok {
			_ = statT // Use statT if needed for platform-specific info
		}
	}

	output.WriteString("═══════════════════════════════════════════════════════════════\n")

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// sortEntries sorts directory entries based on specified criteria
func (d *DirCommand) sortEntries(entries []os.DirEntry, byName, bySize, byDate bool) {
	if bySize {
		// Sort by size (largest first)
		for i := 0; i < len(entries)-1; i++ {
			for j := i + 1; j < len(entries); j++ {
				info1, _ := entries[i].Info()
				info2, _ := entries[j].Info()
				if info1 != nil && info2 != nil && info1.Size() < info2.Size() {
					entries[i], entries[j] = entries[j], entries[i]
				}
			}
		}
	} else if byDate {
		// Sort by modification date (newest first)
		for i := 0; i < len(entries)-1; i++ {
			for j := i + 1; j < len(entries); j++ {
				info1, _ := entries[i].Info()
				info2, _ := entries[j].Info()
				if info1 != nil && info2 != nil && info1.ModTime().Before(info2.ModTime()) {
					entries[i], entries[j] = entries[j], entries[i]
				}
			}
		}
	} else if byName {
		// Sort by name (alphabetical)
		for i := 0; i < len(entries)-1; i++ {
			for j := i + 1; j < len(entries); j++ {
				if strings.ToLower(entries[i].Name()) > strings.ToLower(entries[j].Name()) {
					entries[i], entries[j] = entries[j], entries[i]
				}
			}
		}
	}
}

// isExecutable checks if file extension indicates an executable
func (d *DirCommand) isExecutable(ext string) bool {
	executables := []string{".exe", ".bat", ".cmd", ".com", ".scr", ".msi", ".ps1", ".sh", ".bin", ".run"}
	for _, e := range executables {
		if ext == e {
			return true
		}
	}
	return false
}

// isDocument checks if file extension indicates a document
func (d *DirCommand) isDocument(ext string) bool {
	documents := []string{".txt", ".doc", ".docx", ".pdf", ".rtf", ".odt", ".xls", ".xlsx", ".ppt", ".pptx", ".md", ".readme"}
	for _, e := range documents {
		if ext == e {
			return true
		}
	}
	return false
}

// isImage checks if file extension indicates an image
func (d *DirCommand) isImage(ext string) bool {
	images := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".svg", ".ico", ".webp", ".raw"}
	for _, e := range images {
		if ext == e {
			return true
		}
	}
	return false
}

// isArchive checks if file extension indicates an archive
func (d *DirCommand) isArchive(ext string) bool {
	archives := []string{".zip", ".rar", ".7z", ".tar", ".gz", ".bz2", ".xz", ".cab", ".iso", ".dmg"}
	for _, e := range archives {
		if ext == e {
			return true
		}
	}
	return false
}

// isCode checks if file extension indicates source code
func (d *DirCommand) isCode(ext string) bool {
	code := []string{".go", ".py", ".js", ".ts", ".java", ".c", ".cpp", ".h", ".cs", ".php", ".rb", ".rs", ".swift", ".kt"}
	for _, e := range code {
		if ext == e {
			return true
		}
	}
	return false
}

// isConfig checks if file extension indicates a configuration file
func (d *DirCommand) isConfig(ext string) bool {
	config := []string{".json", ".xml", ".yaml", ".yml", ".toml", ".ini", ".cfg", ".conf", ".config", ".env"}
	for _, e := range config {
		if ext == e {
			return true
		}
	}
	return false
}

// formatFileSize formats file size in human-readable format
func (d *DirCommand) formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// getFileTypeDescription returns a brief description of the file type
func (d *DirCommand) getFileTypeDescription(ext string) string {
	switch ext {
	case ".exe", ".com":
		return color.New(color.FgHiBlack).Sprint("(executable)")
	case ".bat", ".cmd":
		return color.New(color.FgHiBlack).Sprint("(batch file)")
	case ".txt":
		return color.New(color.FgHiBlack).Sprint("(text file)")
	case ".pdf":
		return color.New(color.FgHiBlack).Sprint("(PDF document)")
	case ".doc", ".docx":
		return color.New(color.FgHiBlack).Sprint("(Word document)")
	case ".jpg", ".jpeg", ".png", ".gif":
		return color.New(color.FgHiBlack).Sprint("(image)")
	case ".zip", ".rar", ".7z":
		return color.New(color.FgHiBlack).Sprint("(archive)")
	case ".mp3", ".wav", ".flac":
		return color.New(color.FgHiBlack).Sprint("(audio)")
	case ".mp4", ".avi", ".mkv":
		return color.New(color.FgHiBlack).Sprint("(video)")
	case ".go":
		return color.New(color.FgHiBlack).Sprint("(Go source)")
	case ".py":
		return color.New(color.FgHiBlack).Sprint("(Python)")
	case ".js":
		return color.New(color.FgHiBlack).Sprint("(JavaScript)")
	case ".json":
		return color.New(color.FgHiBlack).Sprint("(JSON data)")
	case ".xml":
		return color.New(color.FgHiBlack).Sprint("(XML data)")
	default:
		return ""
	}
}
