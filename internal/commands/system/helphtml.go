package system

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"suppercommand/internal/commands"

	"github.com/fatih/color"
)

// HelpHTMLCommand generates HTML help documentation
type HelpHTMLCommand struct {
	*commands.BaseCommand
	registry *commands.Registry
}

// NewHelpHTMLCommand creates a new helphtml command
func NewHelpHTMLCommand(registry *commands.Registry) *HelpHTMLCommand {
	return &HelpHTMLCommand{
		BaseCommand: commands.NewBaseCommand(
			"helphtml",
			"Generate HTML help documentation for all commands",
			"helphtml [filename]",
			[]string{"windows", "linux", "darwin"},
			false,
		),
		registry: registry,
	}
}

// Execute generates HTML help documentation
func (h *HelpHTMLCommand) Execute(ctx context.Context, args *commands.Arguments) (*commands.Result, error) {
	startTime := time.Now()

	// Default filename
	filename := "supershell-help.html"
	if len(args.Raw) > 0 {
		filename = args.Raw[0]
		if !strings.HasSuffix(filename, ".html") {
			filename += ".html"
		}
	}

	var output strings.Builder
	output.WriteString(color.New(color.FgCyan, color.Bold).Sprint("📄 GENERATING HTML HELP\n"))
	output.WriteString("═══════════════════════════════════════════════════════════════\n")
	output.WriteString(fmt.Sprintf("📁 Output file: %s\n", color.New(color.FgGreen).Sprint(filename)))
	output.WriteString("───────────────────────────────────────────────────────────────\n")

	// Generate HTML content
	htmlContent := h.generateHTML()

	// Write to file
	err := ioutil.WriteFile(filename, []byte(htmlContent), 0644)
	if err != nil {
		return &commands.Result{
			Output:   fmt.Sprintf("Error: Cannot write to file %s: %v\n", filename, err),
			ExitCode: 1,
			Duration: time.Since(startTime),
		}, nil
	}

	// Get file size
	fileInfo, _ := os.Stat(filename)
	fileSize := fileInfo.Size()

	output.WriteString(color.New(color.FgGreen).Sprint("✅ HTML documentation generated successfully\n"))
	output.WriteString(fmt.Sprintf("📊 File size: %d bytes\n", fileSize))
	output.WriteString(fmt.Sprintf("📋 Commands documented: %d\n", len(h.registry.GetAllCommands())))
	output.WriteString("───────────────────────────────────────────────────────────────\n")
	output.WriteString("💡 Open the file in your web browser to view the documentation\n")
	output.WriteString("═══════════════════════════════════════════════════════════════\n")

	return &commands.Result{
		Output:   output.String(),
		ExitCode: 0,
		Duration: time.Since(startTime),
	}, nil
}

// generateHTML creates the HTML documentation
func (h *HelpHTMLCommand) generateHTML() string {
	var html strings.Builder

	// HTML header with modern responsive design and side navigation
	html.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>SuperShell Command Reference</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            line-height: 1.6;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
        }
        
        .app-container {
            display: flex;
            min-height: 100vh;
        }
        
        /* Side Navigation */
        .sidebar {
            width: 300px;
            background: rgba(255, 255, 255, 0.95);
            backdrop-filter: blur(10px);
            border-right: 1px solid rgba(0, 0, 0, 0.1);
            position: fixed;
            height: 100vh;
            overflow-y: auto;
            z-index: 1000;
            transition: transform 0.3s ease;
        }
        
        .sidebar-header {
            padding: 20px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            text-align: center;
        }
        
        .sidebar-header h1 {
            font-size: 1.5em;
            margin-bottom: 5px;
        }
        
        .sidebar-header .version {
            font-size: 0.9em;
            opacity: 0.9;
        }
        
        .search-box {
            padding: 15px;
            border-bottom: 1px solid rgba(0, 0, 0, 0.1);
        }
        
        .search-input {
            width: 100%;
            padding: 10px 15px;
            border: 1px solid #ddd;
            border-radius: 25px;
            font-size: 14px;
            outline: none;
            transition: border-color 0.3s ease;
        }
        
        .search-input:focus {
            border-color: #667eea;
            box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
        }
        
        .nav-section {
            margin-bottom: 10px;
        }
        
        .nav-section-title {
            padding: 15px 20px 10px;
            font-weight: 600;
            color: #333;
            font-size: 0.9em;
            text-transform: uppercase;
            letter-spacing: 0.5px;
            border-bottom: 1px solid rgba(0, 0, 0, 0.05);
        }
        
        .nav-item {
            display: block;
            padding: 12px 20px;
            color: #555;
            text-decoration: none;
            transition: all 0.3s ease;
            border-left: 3px solid transparent;
        }
        
        .nav-item:hover {
            background: rgba(102, 126, 234, 0.1);
            color: #667eea;
            border-left-color: #667eea;
        }
        
        .nav-item.active {
            background: rgba(102, 126, 234, 0.15);
            color: #667eea;
            border-left-color: #667eea;
            font-weight: 500;
        }
        
        .nav-item .emoji {
            margin-right: 8px;
        }
        
        /* Main Content */
        .main-content {
            flex: 1;
            margin-left: 300px;
            padding: 0;
            background: white;
            min-height: 100vh;
        }
        
        .content-header {
            background: white;
            padding: 30px 40px;
            border-bottom: 1px solid rgba(0, 0, 0, 0.1);
            position: sticky;
            top: 0;
            z-index: 100;
            backdrop-filter: blur(10px);
        }
        
        .content-header h1 {
            color: #2c3e50;
            font-size: 2.5em;
            margin-bottom: 10px;
        }
        
        .content-header .subtitle {
            color: #7f8c8d;
            font-size: 1.1em;
        }
        
        .content-body {
            padding: 40px;
        }
        
        .category-section {
            margin-bottom: 60px;
        }
        
        .category-title {
            font-size: 2em;
            color: #2c3e50;
            margin-bottom: 30px;
            padding-bottom: 15px;
            border-bottom: 3px solid #667eea;
            display: flex;
            align-items: center;
        }
        
        .category-title .emoji {
            margin-right: 15px;
            font-size: 1.2em;
        }
        
        .command-card {
            background: white;
            border: 1px solid rgba(0, 0, 0, 0.1);
            border-radius: 12px;
            padding: 25px;
            margin-bottom: 25px;
            box-shadow: 0 2px 10px rgba(0, 0, 0, 0.05);
            transition: all 0.3s ease;
        }
        
        .command-card:hover {
            box-shadow: 0 8px 25px rgba(0, 0, 0, 0.1);
            transform: translateY(-2px);
        }
        
        .command-header {
            display: flex;
            align-items: center;
            justify-content: space-between;
            margin-bottom: 15px;
        }
        
        .command-name {
            font-size: 1.4em;
            font-weight: 600;
            color: #667eea;
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
        }
        
        .command-badges {
            display: flex;
            gap: 8px;
        }
        
        .badge {
            padding: 4px 12px;
            border-radius: 20px;
            font-size: 0.8em;
            font-weight: 500;
        }
        
        .badge-admin {
            background: #e74c3c;
            color: white;
        }
        
        .badge-platform {
            background: #3498db;
            color: white;
        }
        
        .command-description {
            color: #555;
            font-size: 1.1em;
            margin-bottom: 20px;
            line-height: 1.7;
        }
        
        .command-usage {
            background: #2c3e50;
            color: #ecf0f1;
            padding: 15px 20px;
            border-radius: 8px;
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
            font-size: 0.95em;
            margin-bottom: 20px;
            position: relative;
            overflow-x: auto;
        }
        
        .copy-button {
            position: absolute;
            top: 10px;
            right: 10px;
            background: rgba(255, 255, 255, 0.2);
            border: none;
            color: white;
            padding: 5px 10px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 0.8em;
            transition: background 0.3s ease;
        }
        
        .copy-button:hover {
            background: rgba(255, 255, 255, 0.3);
        }
        
        .tabs {
            display: flex;
            border-bottom: 1px solid #ddd;
            margin-bottom: 20px;
        }
        
        .tab {
            padding: 12px 20px;
            background: none;
            border: none;
            cursor: pointer;
            font-size: 0.95em;
            color: #666;
            border-bottom: 2px solid transparent;
            transition: all 0.3s ease;
        }
        
        .tab.active {
            color: #667eea;
            border-bottom-color: #667eea;
            font-weight: 500;
        }
        
        .tab:hover {
            color: #667eea;
            background: rgba(102, 126, 234, 0.05);
        }
        
        .tab-content {
            display: none;
        }
        
        .tab-content.active {
            display: block;
        }
        
        .options-grid {
            display: grid;
            gap: 15px;
            margin-bottom: 20px;
        }
        
        .option-item {
            padding: 15px;
            background: #f8f9fa;
            border-radius: 8px;
            border-left: 4px solid #667eea;
        }
        
        .option-flag {
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
            background: #e9ecef;
            padding: 3px 8px;
            border-radius: 4px;
            color: #2c3e50;
            font-weight: 600;
            font-size: 0.9em;
        }
        
        .option-description {
            margin-top: 8px;
            color: #555;
            line-height: 1.6;
        }
        
        .examples-grid {
            display: grid;
            gap: 20px;
            margin-bottom: 20px;
        }
        
        .example-item {
            padding: 20px;
            background: #f8f9fa;
            border-radius: 8px;
            border-left: 4px solid #27ae60;
        }
        
        .example-command {
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
            background: #2c3e50;
            color: #ecf0f1;
            padding: 12px 16px;
            border-radius: 6px;
            margin-bottom: 12px;
            font-size: 0.9em;
            position: relative;
        }
        
        .example-description {
            color: #555;
            line-height: 1.6;
        }
        
        .use-cases-grid {
            display: grid;
            gap: 15px;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
        }
        
        .use-case-item {
            padding: 20px;
            background: #fff3cd;
            border-radius: 8px;
            border-left: 4px solid #f39c12;
        }
        
        .use-case-title {
            font-weight: 600;
            color: #d68910;
            margin-bottom: 10px;
            font-size: 1.1em;
        }
        
        .use-case-description {
            color: #555;
            line-height: 1.6;
        }
        
        /* Mobile Responsive */
        @media (max-width: 768px) {
            .sidebar {
                transform: translateX(-100%);
                width: 280px;
            }
            
            .sidebar.open {
                transform: translateX(0);
            }
            
            .main-content {
                margin-left: 0;
            }
            
            .mobile-menu-btn {
                display: block;
                position: fixed;
                top: 20px;
                left: 20px;
                z-index: 1001;
                background: #667eea;
                color: white;
                border: none;
                padding: 10px;
                border-radius: 8px;
                cursor: pointer;
            }
            
            .content-header {
                padding: 20px;
                padding-left: 70px;
            }
            
            .content-body {
                padding: 20px;
            }
            
            .use-cases-grid {
                grid-template-columns: 1fr;
            }
        }
        
        .mobile-menu-btn {
            display: none;
        }
        
        /* Smooth scrolling */
        html {
            scroll-behavior: smooth;
        }
        
        /* Custom scrollbar */
        .sidebar::-webkit-scrollbar {
            width: 6px;
        }
        
        .sidebar::-webkit-scrollbar-track {
            background: #f1f1f1;
        }
        
        .sidebar::-webkit-scrollbar-thumb {
            background: #c1c1c1;
            border-radius: 3px;
        }
        
        .sidebar::-webkit-scrollbar-thumb:hover {
            background: #a8a8a8;
        }
    </style>
</head>
<body>
    <div class="app-container">
        <button class="mobile-menu-btn" onclick="toggleSidebar()">☰</button>
        
        <!-- Sidebar Navigation -->
        <nav class="sidebar" id="sidebar">
            <div class="sidebar-header">
                <h1>🚀 SuperShell</h1>
                <div class="version">Command Reference v2.0</div>
            </div>
            
            <div class="search-box">
                <input type="text" class="search-input" placeholder="Search commands..." onkeyup="filterCommands(this.value)">
            </div>
            
            <div class="nav-sections" id="navSections">
                <div class="nav-section">
                    <div class="nav-section-title">🔥 Security & Firewall</div>
                    <a href="#firewall" class="nav-item" data-category="security"><span class="emoji">🛡️</span>firewall</a>
                </div>
                
                <div class="nav-section">
                    <div class="nav-section-title">⚡ Performance</div>
                    <a href="#perf" class="nav-item" data-category="performance"><span class="emoji">📊</span>perf</a>
                </div>
                
                <div class="nav-section">
                    <div class="nav-section-title">🖥️ Server Management</div>
                    <a href="#server" class="nav-item" data-category="server"><span class="emoji">🖥️</span>server</a>
                    <a href="#sysinfo" class="nav-item" data-category="server"><span class="emoji">ℹ️</span>sysinfo</a>
                    <a href="#killtask" class="nav-item" data-category="server"><span class="emoji">⚡</span>killtask</a>
                    <a href="#winupdate" class="nav-item" data-category="server"><span class="emoji">🔄</span>winupdate</a>
                </div>
                
                <div class="nav-section">
                    <div class="nav-section-title">🌐 Remote Administration</div>
                    <a href="#remote" class="nav-item" data-category="remote"><span class="emoji">🌐</span>remote</a>
                </div>
                
                <div class="nav-section">
                    <div class="nav-section-title">🌐 Network Tools</div>
                    <a href="#ping" class="nav-item" data-category="network"><span class="emoji">📡</span>ping</a>
                    <a href="#tracert" class="nav-item" data-category="network"><span class="emoji">🛤️</span>tracert</a>
                    <a href="#nslookup" class="nav-item" data-category="network"><span class="emoji">🔍</span>nslookup</a>
                    <a href="#netstat" class="nav-item" data-category="network"><span class="emoji">📊</span>netstat</a>
                    <a href="#portscan" class="nav-item" data-category="network"><span class="emoji">🔍</span>portscan</a>
                    <a href="#sniff" class="nav-item" data-category="network"><span class="emoji">👁️</span>sniff</a>
                    <a href="#wget" class="nav-item" data-category="network"><span class="emoji">⬇️</span>wget</a>
                    <a href="#arp" class="nav-item" data-category="network"><span class="emoji">🔗</span>arp</a>
                    <a href="#route" class="nav-item" data-category="network"><span class="emoji">🛣️</span>route</a>
                    <a href="#speedtest" class="nav-item" data-category="network"><span class="emoji">⚡</span>speedtest</a>
                    <a href="#ipconfig" class="nav-item" data-category="network"><span class="emoji">🌐</span>ipconfig</a>
                    <a href="#netdiscover" class="nav-item" data-category="network"><span class="emoji">🔍</span>netdiscover</a>
                </div>
                
                <div class="nav-section">
                    <div class="nav-section-title">📁 File Operations</div>
                    <a href="#ls" class="nav-item" data-category="files"><span class="emoji">📋</span>ls</a>
                    <a href="#dir" class="nav-item" data-category="files"><span class="emoji">📁</span>dir</a>
                    <a href="#cat" class="nav-item" data-category="files"><span class="emoji">📄</span>cat</a>
                    <a href="#cp" class="nav-item" data-category="files"><span class="emoji">📋</span>cp</a>
                    <a href="#mv" class="nav-item" data-category="files"><span class="emoji">➡️</span>mv</a>
                    <a href="#rm" class="nav-item" data-category="files"><span class="emoji">🗑️</span>rm</a>
                    <a href="#mkdir" class="nav-item" data-category="files"><span class="emoji">📁</span>mkdir</a>
                    <a href="#rmdir" class="nav-item" data-category="files"><span class="emoji">🗂️</span>rmdir</a>
                    <a href="#pwd" class="nav-item" data-category="files"><span class="emoji">📍</span>pwd</a>
                    <a href="#cd" class="nav-item" data-category="files"><span class="emoji">📂</span>cd</a>
                </div>
                
                <div class="nav-section">
                    <div class="nav-section-title">⚙️ System Information</div>
                    <a href="#whoami" class="nav-item" data-category="system"><span class="emoji">👤</span>whoami</a>
                    <a href="#hostname" class="nav-item" data-category="system"><span class="emoji">🏷️</span>hostname</a>
                    <a href="#ver" class="nav-item" data-category="system"><span class="emoji">ℹ️</span>ver</a>
                    <a href="#clear" class="nav-item" data-category="system"><span class="emoji">🧹</span>clear</a>
                    <a href="#echo" class="nav-item" data-category="system"><span class="emoji">📢</span>echo</a>
                </div>
                
                <div class="nav-section">
                    <div class="nav-section-title">🔍 Help & Discovery</div>
                    <a href="#help" class="nav-item" data-category="help"><span class="emoji">❓</span>help</a>
                    <a href="#lookup" class="nav-item" data-category="help"><span class="emoji">🔍</span>lookup</a>
                    <a href="#history" class="nav-item" data-category="help"><span class="emoji">🧠</span>history</a>
                    <a href="#helphtml" class="nav-item" data-category="help"><span class="emoji">📄</span>helphtml</a>
                    <a href="#exit" class="nav-item" data-category="help"><span class="emoji">🚪</span>exit</a>
                </div>
                
                <div class="nav-section">
                    <div class="nav-section-title">🚀 FastCP Transfer</div>
                    <a href="#fastcp-send" class="nav-item" data-category="fastcp"><span class="emoji">📤</span>fastcp-send</a>
                    <a href="#fastcp-recv" class="nav-item" data-category="fastcp"><span class="emoji">📥</span>fastcp-recv</a>
                    <a href="#fastcp-backup" class="nav-item" data-category="fastcp"><span class="emoji">💾</span>fastcp-backup</a>
                    <a href="#fastcp-restore" class="nav-item" data-category="fastcp"><span class="emoji">♻️</span>fastcp-restore</a>
                    <a href="#fastcp-dedup" class="nav-item" data-category="fastcp"><span class="emoji">🔄</span>fastcp-dedup</a>
                </div>
            </div>
        </nav>
        
        <!-- Main Content -->
        <main class="main-content">
            <div class="content-header">
                <h1>SuperShell Command Reference</h1>
                <div class="subtitle">
                    Comprehensive documentation for all SuperShell commands<br>
                    Generated on ` + time.Now().Format("January 2, 2006 at 3:04 PM") + `
                </div>
            </div>
            
            <div class="content-body" id="contentBody">
`)

	// Get all commands and categorize them
	allCommands := h.registry.GetAllCommands()

	// Categorize commands with emojis
	categories := map[string]struct {
		Commands []commands.Command
		Emoji    string
	}{
		"Security & Firewall":    {Commands: []commands.Command{}, Emoji: "🔥"},
		"Performance Monitoring": {Commands: []commands.Command{}, Emoji: "⚡"},
		"Server Management":      {Commands: []commands.Command{}, Emoji: "🖥️"},
		"Remote Administration":  {Commands: []commands.Command{}, Emoji: "🌐"},
		"Network Tools":          {Commands: []commands.Command{}, Emoji: "🌐"},
		"File Operations":        {Commands: []commands.Command{}, Emoji: "📁"},
		"System Information":     {Commands: []commands.Command{}, Emoji: "⚙️"},
		"Help & Discovery":       {Commands: []commands.Command{}, Emoji: "🔍"},
		"FastCP Transfer":        {Commands: []commands.Command{}, Emoji: "🚀"},
	}

	// Categorize all commands
	for _, cmd := range allCommands {
		name := cmd.Name()
		switch {
		case name == "firewall":
			cat := categories["Security & Firewall"]
			cat.Commands = append(cat.Commands, cmd)
			categories["Security & Firewall"] = cat
		case name == "perf":
			cat := categories["Performance Monitoring"]
			cat.Commands = append(cat.Commands, cmd)
			categories["Performance Monitoring"] = cat
		case name == "server" || name == "sysinfo" || name == "killtask" || name == "winupdate":
			cat := categories["Server Management"]
			cat.Commands = append(cat.Commands, cmd)
			categories["Server Management"] = cat
		case name == "remote":
			cat := categories["Remote Administration"]
			cat.Commands = append(cat.Commands, cmd)
			categories["Remote Administration"] = cat
		case h.isNetworkCommand(name):
			cat := categories["Network Tools"]
			cat.Commands = append(cat.Commands, cmd)
			categories["Network Tools"] = cat
		case h.isFilesystemCommand(name):
			cat := categories["File Operations"]
			cat.Commands = append(cat.Commands, cmd)
			categories["File Operations"] = cat
		case h.isSystemCommand(name):
			cat := categories["System Information"]
			cat.Commands = append(cat.Commands, cmd)
			categories["System Information"] = cat
		case h.isHelpCommand(name):
			cat := categories["Help & Discovery"]
			cat.Commands = append(cat.Commands, cmd)
			categories["Help & Discovery"] = cat
		case h.isFastCPCommand(name):
			cat := categories["FastCP Transfer"]
			cat.Commands = append(cat.Commands, cmd)
			categories["FastCP Transfer"] = cat
		}
	}

	// Generate sections for each category
	categoryOrder := []string{
		"Security & Firewall",
		"Performance Monitoring",
		"Server Management",
		"Remote Administration",
		"Network Tools",
		"File Operations",
		"System Information",
		"Help & Discovery",
		"FastCP Transfer",
	}

	for _, categoryName := range categoryOrder {
		categoryData := categories[categoryName]
		if len(categoryData.Commands) == 0 {
			continue
		}

		html.WriteString(fmt.Sprintf(`
                <section class="category-section" id="category-%s">
                    <h2 class="category-title">
                        <span class="emoji">%s</span>%s
                    </h2>
`, strings.ToLower(strings.ReplaceAll(categoryName, " ", "-")), categoryData.Emoji, categoryName))

		for _, cmd := range categoryData.Commands {
			badges := ""
			if cmd.RequiresElevation() {
				badges += `<span class="badge badge-admin">⚠️ Admin Required</span>`
			}

			platforms := strings.Join(cmd.SupportedPlatforms(), ", ")
			if platforms != "" {
				badges += fmt.Sprintf(`<span class="badge badge-platform">%s</span>`, platforms)
			}

			detailedHelp := h.getEnhancedHTMLHelp(cmd.Name())

			html.WriteString(fmt.Sprintf(`
                    <div class="command-card" id="%s">
                        <div class="command-header">
                            <div class="command-name">%s</div>
                            <div class="command-badges">%s</div>
                        </div>
                        <div class="command-description">%s</div>
                        <div class="command-usage">
                            <button class="copy-button" onclick="copyToClipboard('%s')">Copy</button>
                            %s
                        </div>
                        %s
                    </div>
`, cmd.Name(), cmd.Name(), badges, cmd.Description(), cmd.Usage(), cmd.Usage(), detailedHelp))
		}

		html.WriteString(`                </section>
`)
	}

	// HTML footer with JavaScript
	html.WriteString(`
            </div>
        </main>
    </div>

    <script>
        // Mobile sidebar toggle
        function toggleSidebar() {
            const sidebar = document.getElementById('sidebar');
            sidebar.classList.toggle('open');
        }

        // Search functionality
        function filterCommands(searchTerm) {
            const navItems = document.querySelectorAll('.nav-item');
            const sections = document.querySelectorAll('.nav-section');
            
            searchTerm = searchTerm.toLowerCase();
            
            navItems.forEach(item => {
                const commandName = item.textContent.toLowerCase();
                const isVisible = commandName.includes(searchTerm);
                item.style.display = isVisible ? 'block' : 'none';
            });
            
            // Show/hide sections based on visible items
            sections.forEach(section => {
                const visibleItems = section.querySelectorAll('.nav-item[style="display: block"], .nav-item:not([style])');
                const hasVisibleItems = Array.from(visibleItems).some(item => 
                    !searchTerm || item.textContent.toLowerCase().includes(searchTerm)
                );
                section.style.display = hasVisibleItems ? 'block' : 'none';
            });
        }

        // Copy to clipboard functionality
        function copyToClipboard(text) {
            navigator.clipboard.writeText(text).then(() => {
                // Show feedback
                const button = event.target;
                const originalText = button.textContent;
                button.textContent = 'Copied!';
                button.style.background = 'rgba(46, 204, 113, 0.8)';
                
                setTimeout(() => {
                    button.textContent = originalText;
                    button.style.background = 'rgba(255, 255, 255, 0.2)';
                }, 2000);
            }).catch(err => {
                console.error('Failed to copy text: ', err);
            });
        }

        // Tab functionality
        function showTab(tabName, commandId) {
            // Hide all tab contents for this command
            const command = document.getElementById(commandId);
            const tabContents = command.querySelectorAll('.tab-content');
            const tabs = command.querySelectorAll('.tab');
            
            tabContents.forEach(content => content.classList.remove('active'));
            tabs.forEach(tab => tab.classList.remove('active'));
            
            // Show selected tab
            const selectedContent = command.querySelector('.' + tabName + '-content');
            const selectedTab = command.querySelector('[onclick*="' + tabName + '"]');
            
            if (selectedContent) selectedContent.classList.add('active');
            if (selectedTab) selectedTab.classList.add('active');
        }

        // Smooth scrolling for navigation links
        document.querySelectorAll('.nav-item').forEach(link => {
            link.addEventListener('click', function(e) {
                e.preventDefault();
                const targetId = this.getAttribute('href').substring(1);
                const targetElement = document.getElementById(targetId);
                
                if (targetElement) {
                    // Update active nav item
                    document.querySelectorAll('.nav-item').forEach(item => item.classList.remove('active'));
                    this.classList.add('active');
                    
                    // Smooth scroll to target
                    targetElement.scrollIntoView({
                        behavior: 'smooth',
                        block: 'start'
                    });
                    
                    // Close mobile sidebar
                    if (window.innerWidth <= 768) {
                        document.getElementById('sidebar').classList.remove('open');
                    }
                }
            });
        });

        // Highlight active section on scroll
        window.addEventListener('scroll', () => {
            const sections = document.querySelectorAll('.command-card');
            const navItems = document.querySelectorAll('.nav-item');
            
            let current = '';
            sections.forEach(section => {
                const sectionTop = section.offsetTop - 100;
                if (window.pageYOffset >= sectionTop) {
                    current = section.getAttribute('id');
                }
            });
            
            navItems.forEach(item => {
                item.classList.remove('active');
                if (item.getAttribute('href') === '#' + current) {
                    item.classList.add('active');
                }
            });
        });

        // Initialize first tab as active for commands with tabs
        document.addEventListener('DOMContentLoaded', () => {
            document.querySelectorAll('.command-card').forEach(command => {
                const firstTab = command.querySelector('.tab');
                const firstContent = command.querySelector('.tab-content');
                
                if (firstTab && firstContent) {
                    firstTab.classList.add('active');
                    firstContent.classList.add('active');
                }
            });
        });

        // Close mobile menu when clicking outside
        document.addEventListener('click', (e) => {
            const sidebar = document.getElementById('sidebar');
            const menuBtn = document.querySelector('.mobile-menu-btn');
            
            if (window.innerWidth <= 768 && 
                !sidebar.contains(e.target) && 
                !menuBtn.contains(e.target) && 
                sidebar.classList.contains('open')) {
                sidebar.classList.remove('open');
            }
        });
    </script>
</body>
</html>`)

	return html.String()
}

// isSystemCommand checks if a command is a system command
func (h *HelpHTMLCommand) isSystemCommand(name string) bool {
	systemCommands := []string{"whoami", "hostname", "ver", "clear", "echo"}
	for _, cmd := range systemCommands {
		if cmd == name {
			return true
		}
	}
	return false
}

// isFilesystemCommand checks if a command is a filesystem command
func (h *HelpHTMLCommand) isFilesystemCommand(name string) bool {
	fsCommands := []string{"pwd", "ls", "dir", "echo", "cd", "cat", "mkdir", "rm", "rmdir", "cp", "mv"}
	for _, cmd := range fsCommands {
		if cmd == name {
			return true
		}
	}
	return false
}

// isNetworkCommand checks if a command is a network command
func (h *HelpHTMLCommand) isNetworkCommand(name string) bool {
	networkCommands := []string{"ping", "tracert", "nslookup", "netstat", "portscan", "sniff", "wget", "arp", "route", "speedtest", "ipconfig", "netdiscover"}
	for _, cmd := range networkCommands {
		if cmd == name {
			return true
		}
	}
	return false
}

// isHelpCommand checks if a command is a help command
func (h *HelpHTMLCommand) isHelpCommand(name string) bool {
	helpCommands := []string{"help", "lookup", "history", "helphtml", "exit"}
	for _, cmd := range helpCommands {
		if cmd == name {
			return true
		}
	}
	return false
}

// isFastCPCommand checks if a command is a FastCP command
func (h *HelpHTMLCommand) isFastCPCommand(name string) bool {
	fastcpCommands := []string{"fastcp-send", "fastcp-recv", "fastcp-backup", "fastcp-restore", "fastcp-dedup"}
	for _, cmd := range fastcpCommands {
		if cmd == name {
			return true
		}
	}
	return false
}

// getEnhancedHTMLHelp returns comprehensive HTML help with tabs for each command
func (h *HelpHTMLCommand) getEnhancedHTMLHelp(commandName string) string {
	// Create tabbed interface for detailed help
	var help strings.Builder

	help.WriteString(fmt.Sprintf(`
                        <div class="tabs">
                            <button class="tab" onclick="showTab('options', '%s')">Options</button>
                            <button class="tab" onclick="showTab('examples', '%s')">Examples</button>
                            <button class="tab" onclick="showTab('usecases', '%s')">Use Cases</button>
                        </div>
                        
                        <div class="tab-content options-content">
                            <div class="options-grid">
                                %s
                            </div>
                        </div>
                        
                        <div class="tab-content examples-content">
                            <div class="examples-grid">
                                %s
                            </div>
                        </div>
                        
                        <div class="tab-content usecases-content">
                            <div class="use-cases-grid">
                                %s
                            </div>
                        </div>
`, commandName, commandName, commandName, h.getOptionsHTML(commandName), h.getExamplesHTML(commandName), h.getUseCasesHTML(commandName)))

	return help.String()
}

// getOptionsHTML returns HTML for command options
func (h *HelpHTMLCommand) getOptionsHTML(commandName string) string {
	switch commandName {
	case "firewall":
		return `
                                <div class="option-item">
                                    <div class="option-flag">status</div>
                                    <div class="option-description">Show current firewall status and configuration</div>
                                </div>
                                <div class="option-item">
                                    <div class="option-flag">enable</div>
                                    <div class="option-description">Enable the system firewall (requires admin privileges)</div>
                                </div>
                                <div class="option-item">
                                    <div class="option-flag">disable</div>
                                    <div class="option-description">Disable the system firewall (requires admin privileges)</div>
                                </div>
                                <div class="option-item">
                                    <div class="option-flag">rules list</div>
                                    <div class="option-description">List all configured firewall rules</div>
                                </div>`
	case "perf":
		return `
                                <div class="option-item">
                                    <div class="option-flag">analyze</div>
                                    <div class="option-description">Perform comprehensive system performance analysis</div>
                                </div>
                                <div class="option-item">
                                    <div class="option-flag">monitor</div>
                                    <div class="option-description">Start real-time performance monitoring</div>
                                </div>
                                <div class="option-item">
                                    <div class="option-flag">report</div>
                                    <div class="option-description">Generate detailed performance report</div>
                                </div>
                                <div class="option-item">
                                    <div class="option-flag">baseline create &lt;name&gt;</div>
                                    <div class="option-description">Create a new performance baseline</div>
                                </div>
                                <div class="option-item">
                                    <div class="option-flag">baseline list</div>
                                    <div class="option-description">List all saved baselines</div>
                                </div>`
	case "server":
		return `
                                <div class="option-item">
                                    <div class="option-flag">health</div>
                                    <div class="option-description">Check overall server health status</div>
                                </div>
                                <div class="option-item">
                                    <div class="option-flag">services list</div>
                                    <div class="option-description">List all system services</div>
                                </div>
                                <div class="option-item">
                                    <div class="option-flag">services start &lt;name&gt;</div>
                                    <div class="option-description">Start a specific service (requires admin)</div>
                                </div>
                                <div class="option-item">
                                    <div class="option-flag">services stop &lt;name&gt;</div>
                                    <div class="option-description">Stop a specific service (requires admin)</div>
                                </div>
                                <div class="option-item">
                                    <div class="option-flag">users</div>
                                    <div class="option-description">List all active users and sessions</div>
                                </div>`
	case "remote":
		return `
                                <div class="option-item">
                                    <div class="option-flag">list</div>
                                    <div class="option-description">List all configured remote servers</div>
                                </div>
                                <div class="option-item">
                                    <div class="option-flag">add &lt;name&gt; &lt;user@host&gt;</div>
                                    <div class="option-description">Add a new remote server configuration</div>
                                </div>
                                <div class="option-item">
                                    <div class="option-flag">remove &lt;name&gt;</div>
                                    <div class="option-description">Remove a remote server configuration</div>
                                </div>
                                <div class="option-item">
                                    <div class="option-flag">exec &lt;server&gt; "&lt;command&gt;"</div>
                                    <div class="option-description">Execute command on remote server</div>
                                </div>`
	default:
		return h.getDetailedHTMLHelp(commandName)
	}
}

// getExamplesHTML returns HTML for command examples
func (h *HelpHTMLCommand) getExamplesHTML(commandName string) string {
	switch commandName {
	case "firewall":
		return `
                                <div class="example-item">
                                    <div class="example-command">firewall status</div>
                                    <div class="example-description">Check if firewall is enabled and show basic info</div>
                                </div>
                                <div class="example-item">
                                    <div class="example-command">firewall enable</div>
                                    <div class="example-description">Enable Windows Defender Firewall</div>
                                </div>
                                <div class="example-item">
                                    <div class="example-command">firewall rules list</div>
                                    <div class="example-description">List all configured firewall rules</div>
                                </div>`
	case "perf":
		return `
                                <div class="example-item">
                                    <div class="example-command">perf analyze</div>
                                    <div class="example-description">Quick performance analysis</div>
                                </div>
                                <div class="example-item">
                                    <div class="example-command">perf baseline create prod-baseline</div>
                                    <div class="example-description">Create baseline named 'prod-baseline'</div>
                                </div>
                                <div class="example-item">
                                    <div class="example-command">perf report</div>
                                    <div class="example-description">Generate comprehensive performance report</div>
                                </div>`
	case "server":
		return `
                                <div class="example-item">
                                    <div class="example-command">server health</div>
                                    <div class="example-description">Check overall server health</div>
                                </div>
                                <div class="example-item">
                                    <div class="example-command">server services list</div>
                                    <div class="example-description">List all system services</div>
                                </div>
                                <div class="example-item">
                                    <div class="example-command">server services start "Print Spooler"</div>
                                    <div class="example-description">Start the Print Spooler service</div>
                                </div>`
	case "remote":
		return `
                                <div class="example-item">
                                    <div class="example-command">remote add web1 admin@192.168.1.10</div>
                                    <div class="example-description">Add server with default settings</div>
                                </div>
                                <div class="example-item">
                                    <div class="example-command">remote exec web1 "uptime"</div>
                                    <div class="example-description">Check uptime on web1 server</div>
                                </div>
                                <div class="example-item">
                                    <div class="example-command">remote list</div>
                                    <div class="example-description">List all configured servers</div>
                                </div>`
	default:
		return `<div class="example-item"><div class="example-command">` + commandName + `</div><div class="example-description">Basic usage example</div></div>`
	}
}

// getUseCasesHTML returns HTML for command use cases
func (h *HelpHTMLCommand) getUseCasesHTML(commandName string) string {
	switch commandName {
	case "firewall":
		return `
                                <div class="use-case-item">
                                    <div class="use-case-title">Security Management</div>
                                    <div class="use-case-description">Monitor and control system firewall settings</div>
                                </div>
                                <div class="use-case-item">
                                    <div class="use-case-title">Compliance Checking</div>
                                    <div class="use-case-description">Verify firewall status for security audits</div>
                                </div>
                                <div class="use-case-item">
                                    <div class="use-case-title">Troubleshooting</div>
                                    <div class="use-case-description">Diagnose network connectivity issues</div>
                                </div>`
	case "perf":
		return `
                                <div class="use-case-item">
                                    <div class="use-case-title">Performance Monitoring</div>
                                    <div class="use-case-description">Track system resource usage over time</div>
                                </div>
                                <div class="use-case-item">
                                    <div class="use-case-title">Bottleneck Detection</div>
                                    <div class="use-case-description">Identify CPU, memory, disk, or network bottlenecks</div>
                                </div>
                                <div class="use-case-item">
                                    <div class="use-case-title">Capacity Planning</div>
                                    <div class="use-case-description">Understand system limits and plan for scaling</div>
                                </div>`
	case "server":
		return `
                                <div class="use-case-item">
                                    <div class="use-case-title">System Administration</div>
                                    <div class="use-case-description">Monitor and manage server components</div>
                                </div>
                                <div class="use-case-item">
                                    <div class="use-case-title">Service Management</div>
                                    <div class="use-case-description">Control Windows/Linux services remotely</div>
                                </div>
                                <div class="use-case-item">
                                    <div class="use-case-title">Health Monitoring</div>
                                    <div class="use-case-description">Get real-time server health status</div>
                                </div>`
	case "remote":
		return `
                                <div class="use-case-item">
                                    <div class="use-case-title">Remote Administration</div>
                                    <div class="use-case-description">Manage multiple servers from one location</div>
                                </div>
                                <div class="use-case-item">
                                    <div class="use-case-title">Command Execution</div>
                                    <div class="use-case-description">Run commands across multiple servers</div>
                                </div>
                                <div class="use-case-item">
                                    <div class="use-case-title">Deployment Management</div>
                                    <div class="use-case-description">Execute deployment scripts remotely</div>
                                </div>`
	default:
		return `<div class="use-case-item"><div class="use-case-title">General Usage</div><div class="use-case-description">Common use cases for this command</div></div>`
	}
}

// getDetailedHTMLHelp returns comprehensive HTML help for each command (legacy function)
func (h *HelpHTMLCommand) getDetailedHTMLHelp(commandName string) string {
	switch commandName {
	case "sniff":
		return `
<div class="detailed-help">
    <div class="options-section">
        <div class="options-title">🔧 Command Options</div>
        <div class="option-item">
            <span class="option-flag">-i, --interface &lt;name&gt;</span>
            <span class="option-description">Network interface to monitor (default: eth0)</span>
        </div>
        <div class="option-item">
            <span class="option-flag">-c, --count &lt;number&gt;</span>
            <span class="option-description">Number of packets to capture (default: 10)</span>
        </div>
        <div class="option-item">
            <span class="option-flag">-p, --protocol &lt;proto&gt;</span>
            <span class="option-description">Filter by protocol (TCP, UDP, HTTP, HTTPS, DNS, SSH, FTP, etc.)</span>
        </div>
        <div class="option-item">
            <span class="option-flag">-s, --source &lt;ip&gt;</span>
            <span class="option-description">Filter by source IP address</span>
        </div>
        <div class="option-item">
            <span class="option-flag">-d, --dest &lt;ip&gt;</span>
            <span class="option-description">Filter by destination IP address</span>
        </div>
        <div class="option-item">
            <span class="option-flag">--port &lt;port&gt;</span>
            <span class="option-description">Filter by port number (matches source or destination)</span>
        </div>
        <div class="option-item">
            <span class="option-flag">-v, --verbose</span>
            <span class="option-description">Show detailed packet information</span>
        </div>
        <div class="option-item">
            <span class="option-flag">--hex</span>
            <span class="option-description">Display hexadecimal payload dump</span>
        </div>
        <div class="option-item">
            <span class="option-flag">--save &lt;file&gt;</span>
            <span class="option-description">Save capture to file</span>
        </div>
        <div class="option-item">
            <span class="option-flag">--continuous</span>
            <span class="option-description">Continuous capture mode</span>
        </div>
        <div class="option-item">
            <span class="option-flag">-t, --timeout &lt;seconds&gt;</span>
            <span class="option-description">Capture timeout for continuous mode</span>
        </div>
    </div>
    
    <div class="examples-section">
        <div class="examples-title">💡 Usage Examples</div>
        <div class="example-item">
            <div class="example-command">sniff -c 10</div>
            <div class="example-description">Capture 10 packets from the default interface</div>
        </div>
        <div class="example-item">
            <div class="example-command">sniff -p HTTP -v</div>
            <div class="example-description">Capture HTTP packets with detailed verbose output</div>
        </div>
        <div class="example-item">
            <div class="example-command">sniff -s 192.168.1.100 --hex</div>
            <div class="example-description">Capture packets from specific IP with hexadecimal payload dump</div>
        </div>
        <div class="example-item">
            <div class="example-command">sniff --port 80 -c 5</div>
            <div class="example-description">Capture 5 packets on port 80 (HTTP traffic)</div>
        </div>
        <div class="example-item">
            <div class="example-command">sniff -p TCP -d 8.8.8.8 --save capture.pcap</div>
            <div class="example-description">Capture TCP packets to Google DNS and save to file</div>
        </div>
    </div>
    
    <div class="use-cases-section">
        <div class="use-cases-title">🎯 Common Use Cases</div>
        <div class="use-case-item">
            <div class="use-case-title">Network Troubleshooting</div>
            <div class="use-case-description">Monitor network traffic to identify connectivity issues, packet loss, or unusual network behavior</div>
        </div>
        <div class="use-case-item">
            <div class="use-case-title">Security Analysis</div>
            <div class="use-case-description">Detect suspicious network activity, unauthorized connections, or potential security breaches</div>
        </div>
        <div class="use-case-item">
            <div class="use-case-title">Protocol Analysis</div>
            <div class="use-case-description">Analyze specific protocols (HTTP, DNS, SSH) to understand application behavior and performance</div>
        </div>
        <div class="use-case-item">
            <div class="use-case-title">Performance Monitoring</div>
            <div class="use-case-description">Monitor network performance, bandwidth usage, and identify bottlenecks in network communication</div>
        </div>
    </div>
</div>`

	case "wget":
		return `
<div class="detailed-help">
    <div class="options-section">
        <div class="options-title">🔧 Command Options</div>
        <div class="option-item">
            <span class="option-flag">-v, --verbose</span>
            <span class="option-description">Show detailed download progress and connection information</span>
        </div>
        <div class="option-item">
            <span class="option-flag">&lt;url&gt;</span>
            <span class="option-description">URL to download from (required)</span>
        </div>
        <div class="option-item">
            <span class="option-flag">[filename]</span>
            <span class="option-description">Optional filename (auto-detected from URL if not provided)</span>
        </div>
    </div>
    
    <div class="examples-section">
        <div class="examples-title">💡 Usage Examples</div>
        <div class="example-item">
            <div class="example-command">wget https://example.com/file.zip</div>
            <div class="example-description">Download a file with automatic filename detection</div>
        </div>
        <div class="example-item">
            <div class="example-command">wget -v https://api.github.com/users</div>
            <div class="example-description">Download with verbose output showing progress and speed</div>
        </div>
        <div class="example-item">
            <div class="example-command">wget https://example.com/data.json mydata.json</div>
            <div class="example-description">Download with custom filename</div>
        </div>
        <div class="example-item">
            <div class="example-command">wget https://releases.ubuntu.com/20.04/ubuntu-20.04.iso</div>
            <div class="example-description">Download large files like ISO images</div>
        </div>
    </div>
    
    <div class="use-cases-section">
        <div class="use-cases-title">🎯 Common Use Cases</div>
        <div class="use-case-item">
            <div class="use-case-title">Software Downloads</div>
            <div class="use-case-description">Download software packages, installers, and updates from the internet</div>
        </div>
        <div class="use-case-item">
            <div class="use-case-title">API Data Retrieval</div>
            <div class="use-case-description">Download JSON, XML, or other data from REST APIs for processing</div>
        </div>
        <div class="use-case-item">
            <div class="use-case-title">Backup Downloads</div>
            <div class="use-case-description">Download backup files, databases, or archives from remote servers</div>
        </div>
        <div class="use-case-item">
            <div class="use-case-title">Media Downloads</div>
            <div class="use-case-description">Download images, videos, documents, and other media files</div>
        </div>
    </div>
</div>`

	case "arp":
		return `
<div class="detailed-help">
    <div class="options-section">
        <div class="options-title">🔧 Command Options</div>
        <div class="option-item">
            <span class="option-flag">-a, --all</span>
            <span class="option-description">Show all ARP entries in the table</span>
        </div>
        <div class="option-item">
            <span class="option-flag">-d, --delete &lt;ip&gt;</span>
            <span class="option-description">Delete ARP entry for specified IP address</span>
        </div>
        <div class="option-item">
            <span class="option-flag">[ip_address]</span>
            <span class="option-description">Show ARP entry for specific IP address</span>
        </div>
    </div>
    
    <div class="examples-section">
        <div class="examples-title">💡 Usage Examples</div>
        <div class="example-item">
            <div class="example-command">arp -a</div>
            <div class="example-description">Display all ARP entries with MAC addresses and types</div>
        </div>
        <div class="example-item">
            <div class="example-command">arp 192.168.1.1</div>
            <div class="example-description">Show ARP entry for the gateway router</div>
        </div>
        <div class="example-item">
            <div class="example-command">arp -d 192.168.1.100</div>
            <div class="example-description">Delete ARP entry for a specific host</div>
        </div>
    </div>
    
    <div class="use-cases-section">
        <div class="use-cases-title">🎯 Common Use Cases</div>
        <div class="use-case-item">
            <div class="use-case-title">Network Troubleshooting</div>
            <div class="use-case-description">Resolve IP-to-MAC address mapping issues and connectivity problems</div>
        </div>
        <div class="use-case-item">
            <div class="use-case-title">Security Monitoring</div>
            <div class="use-case-description">Detect ARP spoofing attacks and unauthorized devices on the network</div>
        </div>
        <div class="use-case-item">
            <div class="use-case-title">Network Discovery</div>
            <div class="use-case-description">Identify active devices and their MAC addresses on the local network</div>
        </div>
    </div>
</div>`

	case "route":
		return `
<div class="detailed-help">
    <div class="options-section">
        <div class="options-title">🔧 Command Options</div>
        <div class="option-item">
            <span class="option-flag">print, show</span>
            <span class="option-description">Display routing table (default action)</span>
        </div>
        <div class="option-item">
            <span class="option-flag">add &lt;dest&gt; &lt;gateway&gt;</span>
            <span class="option-description">Add a new route to the routing table</span>
        </div>
        <div class="option-item">
            <span class="option-flag">delete &lt;dest&gt; [gateway]</span>
            <span class="option-description">Delete a route from the routing table</span>
        </div>
        <div class="option-item">
            <span class="option-flag">-4, --ipv4</span>
            <span class="option-description">Show only IPv4 routes</span>
        </div>
        <div class="option-item">
            <span class="option-flag">-6, --ipv6</span>
            <span class="option-description">Show only IPv6 routes</span>
        </div>
    </div>
    
    <div class="examples-section">
        <div class="examples-title">💡 Usage Examples</div>
        <div class="example-item">
            <div class="example-command">route</div>
            <div class="example-description">Display the complete routing table</div>
        </div>
        <div class="example-item">
            <div class="example-command">route -4</div>
            <div class="example-description">Show only IPv4 routing entries</div>
        </div>
        <div class="example-item">
            <div class="example-command">route add 10.0.0.0/8 192.168.1.1</div>
            <div class="example-description">Add route for 10.x.x.x network via gateway</div>
        </div>
        <div class="example-item">
            <div class="example-command">route delete 10.0.0.0/8</div>
            <div class="example-description">Remove route for 10.x.x.x network</div>
        </div>
    </div>
    
    <div class="use-cases-section">
        <div class="use-cases-title">🎯 Common Use Cases</div>
        <div class="use-case-item">
            <div class="use-case-title">Network Configuration</div>
            <div class="use-case-description">Configure routing for multi-homed systems and complex network topologies</div>
        </div>
        <div class="use-case-item">
            <div class="use-case-title">VPN Setup</div>
            <div class="use-case-description">Add routes for VPN connections and remote network access</div>
        </div>
        <div class="use-case-item">
            <div class="use-case-title">Network Troubleshooting</div>
            <div class="use-case-description">Diagnose routing issues and verify network path configurations</div>
        </div>
    </div>
</div>`

	case "speedtest":
		return `
<div class="detailed-help">
    <div class="options-section">
        <div class="options-title">🔧 Command Options</div>
        <div class="option-item">
            <span class="option-flag">-s, --simple</span>
            <span class="option-description">Simple output format with just the results</span>
        </div>
        <div class="option-item">
            <span class="option-flag">-q, --quiet</span>
            <span class="option-description">Minimal output during testing</span>
        </div>
        <div class="option-item">
            <span class="option-flag">--download-only</span>
            <span class="option-description">Test download speed only</span>
        </div>
        <div class="option-item">
            <span class="option-flag">--upload-only</span>
            <span class="option-description">Test upload speed only</span>
        </div>
    </div>
    
    <div class="examples-section">
        <div class="examples-title">💡 Usage Examples</div>
        <div class="example-item">
            <div class="example-command">speedtest</div>
            <div class="example-description">Run complete speed test with detailed output</div>
        </div>
        <div class="example-item">
            <div class="example-command">speedtest -s</div>
            <div class="example-description">Quick speed test with simple results</div>
        </div>
        <div class="example-item">
            <div class="example-command">speedtest --download-only</div>
            <div class="example-description">Test only download speed</div>
        </div>
        <div class="example-item">
            <div class="example-command">speedtest --upload-only -q</div>
            <div class="example-description">Test only upload speed quietly</div>
        </div>
    </div>
    
    <div class="use-cases-section">
        <div class="use-cases-title">🎯 Common Use Cases</div>
        <div class="use-case-item">
            <div class="use-case-title">Internet Performance Monitoring</div>
            <div class="use-case-description">Regular testing of internet connection speed and quality</div>
        </div>
        <div class="use-case-item">
            <div class="use-case-title">ISP Verification</div>
            <div class="use-case-description">Verify that ISP is delivering promised bandwidth speeds</div>
        </div>
        <div class="use-case-item">
            <div class="use-case-title">Network Troubleshooting</div>
            <div class="use-case-description">Diagnose slow internet connections and network performance issues</div>
        </div>
    </div>
</div>`

	case "portscan":
		return `
<div class="detailed-help">
    <div class="options-section">
        <div class="options-title">🔧 Command Options</div>
        <div class="option-item">
            <span class="option-flag">&lt;host&gt;</span>
            <span class="option-description">Target host to scan (IP address or hostname)</span>
        </div>
        <div class="option-item">
            <span class="option-flag">-p, --ports &lt;range&gt;</span>
            <span class="option-description">Specific ports (e.g., 80,443,22-25)</span>
        </div>
        <div class="option-item">
            <span class="option-flag">-t, --timeout &lt;ms&gt;</span>
            <span class="option-description">Connection timeout in milliseconds</span>
        </div>
        <div class="option-item">
            <span class="option-flag">-c, --concurrency &lt;num&gt;</span>
            <span class="option-description">Number of concurrent connections</span>
        </div>
        <div class="option-item">
            <span class="option-flag">--top-ports &lt;n&gt;</span>
            <span class="option-description">Scan top N most common ports</span>
        </div>
    </div>
    
    <div class="examples-section">
        <div class="examples-title">💡 Usage Examples</div>
        <div class="example-item">
            <div class="example-command">portscan google.com</div>
            <div class="example-description">Scan common ports on Google's servers</div>
        </div>
        <div class="example-item">
            <div class="example-command">portscan 192.168.1.1 -p 80,443,22</div>
            <div class="example-description">Scan specific ports on local router</div>
        </div>
        <div class="example-item">
            <div class="example-command">portscan example.com --top-ports 100</div>
            <div class="example-description">Scan top 100 most common ports</div>
        </div>
        <div class="example-item">
            <div class="example-command">portscan 10.0.0.1 -p 1-1000 -c 50</div>
            <div class="example-description">Fast scan of ports 1-1000 with high concurrency</div>
        </div>
    </div>
    
    <div class="use-cases-section">
        <div class="use-cases-title">🎯 Common Use Cases</div>
        <div class="use-case-item">
            <div class="use-case-title">Security Assessment</div>
            <div class="use-case-description">Identify open ports and potential security vulnerabilities</div>
        </div>
        <div class="use-case-item">
            <div class="use-case-title">Network Discovery</div>
            <div class="use-case-description">Discover services running on network hosts</div>
        </div>
        <div class="use-case-item">
            <div class="use-case-title">Service Verification</div>
            <div class="use-case-description">Verify that services are running on expected ports</div>
        </div>
    </div>
</div>`

	case "ping":
		return `
<div class="detailed-help">
    <div class="options-section">
        <div class="options-title">🔧 Command Options</div>
        <div class="option-item">
            <span class="option-flag">&lt;host&gt;</span>
            <span class="option-description">Target host to ping (IP address or hostname)</span>
        </div>
        <div class="option-item">
            <span class="option-flag">-c, --count &lt;number&gt;</span>
            <span class="option-description">Number of ping packets to send</span>
        </div>
        <div class="option-item">
            <span class="option-flag">-t, --timeout &lt;ms&gt;</span>
            <span class="option-description">Timeout for each ping in milliseconds</span>
        </div>
        <div class="option-item">
            <span class="option-flag">-i, --interval &lt;ms&gt;</span>
            <span class="option-description">Interval between pings in milliseconds</span>
        </div>
    </div>
    
    <div class="examples-section">
        <div class="examples-title">💡 Usage Examples</div>
        <div class="example-item">
            <div class="example-command">ping google.com</div>
            <div class="example-description">Basic connectivity test to Google</div>
        </div>
        <div class="example-item">
            <div class="example-command">ping -c 5 8.8.8.8</div>
            <div class="example-description">Send 5 ping packets to Google DNS</div>
        </div>
        <div class="example-item">
            <div class="example-command">ping -t 2000 example.com</div>
            <div class="example-description">Ping with 2 second timeout</div>
        </div>
    </div>
    
    <div class="use-cases-section">
        <div class="use-cases-title">🎯 Common Use Cases</div>
        <div class="use-case-item">
            <div class="use-case-title">Connectivity Testing</div>
            <div class="use-case-description">Test basic network connectivity to hosts and services</div>
        </div>
        <div class="use-case-item">
            <div class="use-case-title">Latency Measurement</div>
            <div class="use-case-description">Measure network latency and response times</div>
        </div>
        <div class="use-case-item">
            <div class="use-case-title">Network Troubleshooting</div>
            <div class="use-case-description">Diagnose network connectivity issues and packet loss</div>
        </div>
    </div>
</div>`

	case "sysinfo":
		return `
<div class="detailed-help">
    <div class="options-section">
        <div class="options-title">🔧 Command Options</div>
        <div class="option-item">
            <span class="option-flag">-v, --verbose</span>
            <span class="option-description">Show detailed system information</span>
        </div>
        <div class="option-item">
            <span class="option-flag">--cpu</span>
            <span class="option-description">Show only CPU information</span>
        </div>
        <div class="option-item">
            <span class="option-flag">--memory</span>
            <span class="option-description">Show only memory information</span>
        </div>
        <div class="option-item">
            <span class="option-flag">--disk</span>
            <span class="option-description">Show only disk information</span>
        </div>
        <div class="option-item">
            <span class="option-flag">--network</span>
            <span class="option-description">Show only network information</span>
        </div>
    </div>
    
    <div class="examples-section">
        <div class="examples-title">💡 Usage Examples</div>
        <div class="example-item">
            <div class="example-command">sysinfo</div>
            <div class="example-description">Display comprehensive system information</div>
        </div>
        <div class="example-item">
            <div class="example-command">sysinfo -v</div>
            <div class="example-description">Show detailed verbose system information</div>
        </div>
        <div class="example-item">
            <div class="example-command">sysinfo --cpu</div>
            <div class="example-description">Show only CPU details and specifications</div>
        </div>
        <div class="example-item">
            <div class="example-command">sysinfo --memory</div>
            <div class="example-description">Show only memory usage and specifications</div>
        </div>
    </div>
    
    <div class="use-cases-section">
        <div class="use-cases-title">🎯 Common Use Cases</div>
        <div class="use-case-item">
            <div class="use-case-title">System Monitoring</div>
            <div class="use-case-description">Monitor system resources, performance, and health status</div>
        </div>
        <div class="use-case-item">
            <div class="use-case-title">Hardware Inventory</div>
            <div class="use-case-description">Get detailed hardware specifications and system configuration</div>
        </div>
        <div class="use-case-item">
            <div class="use-case-title">Troubleshooting</div>
            <div class="use-case-description">Diagnose system issues and resource bottlenecks</div>
        </div>
    </div>
</div>`

	case "ls":
		return `
<div class="detailed-help">
    <div class="options-section">
        <div class="options-title">🔧 Command Options</div>
        <div class="option-item">
            <span class="option-flag">-l, --long</span>
            <span class="option-description">Use long listing format with detailed information</span>
        </div>
        <div class="option-item">
            <span class="option-flag">-a, --all</span>
            <span class="option-description">Show hidden files and directories</span>
        </div>
        <div class="option-item">
            <span class="option-flag">-h, --human-readable</span>
            <span class="option-description">Show file sizes in human readable format</span>
        </div>
        <div class="option-item">
            <span class="option-flag">-R, --recursive</span>
            <span class="option-description">List directories recursively</span>
        </div>
        <div class="option-item">
            <span class="option-flag">-t, --time</span>
            <span class="option-description">Sort by modification time</span>
        </div>
        <div class="option-item">
            <span class="option-flag">-S, --size</span>
            <span class="option-description">Sort by file size</span>
        </div>
    </div>
    
    <div class="examples-section">
        <div class="examples-title">💡 Usage Examples</div>
        <div class="example-item">
            <div class="example-command">ls</div>
            <div class="example-description">List files and directories in current directory</div>
        </div>
        <div class="example-item">
            <div class="example-command">ls -la</div>
            <div class="example-description">Long format listing including hidden files</div>
        </div>
        <div class="example-item">
            <div class="example-command">ls -lh /home</div>
            <div class="example-description">List /home directory with human-readable sizes</div>
        </div>
        <div class="example-item">
            <div class="example-command">ls -lt</div>
            <div class="example-description">List files sorted by modification time</div>
        </div>
    </div>
    
    <div class="use-cases-section">
        <div class="use-cases-title">🎯 Common Use Cases</div>
        <div class="use-case-item">
            <div class="use-case-title">File Management</div>
            <div class="use-case-description">Browse and explore directory contents and file information</div>
        </div>
        <div class="use-case-item">
            <div class="use-case-title">System Administration</div>
            <div class="use-case-description">Check file permissions, ownership, and system directory contents</div>
        </div>
        <div class="use-case-item">
            <div class="use-case-title">Development</div>
            <div class="use-case-description">Navigate project directories and examine file structures</div>
        </div>
    </div>
</div>`

	case "cp":
		return `
<div class="detailed-help">
    <div class="options-section">
        <div class="options-title">🔧 Command Options</div>
        <div class="option-item">
            <span class="option-flag">-r, --recursive</span>
            <span class="option-description">Copy directories recursively</span>
        </div>
        <div class="option-item">
            <span class="option-flag">-v, --verbose</span>
            <span class="option-description">Show detailed copy operations</span>
        </div>
        <div class="option-item">
            <span class="option-flag">-f, --force</span>
            <span class="option-description">Force overwrite existing files</span>
        </div>
        <div class="option-item">
            <span class="option-flag">-p, --preserve</span>
            <span class="option-description">Preserve file attributes and timestamps</span>
        </div>
    </div>
    
    <div class="examples-section">
        <div class="examples-title">💡 Usage Examples</div>
        <div class="example-item">
            <div class="example-command">cp file.txt backup.txt</div>
            <div class="example-description">Copy a single file</div>
        </div>
        <div class="example-item">
            <div class="example-command">cp -r folder/ backup_folder/</div>
            <div class="example-description">Copy directory recursively</div>
        </div>
        <div class="example-item">
            <div class="example-command">cp -v *.txt /backup/</div>
            <div class="example-description">Copy all text files with verbose output</div>
        </div>
    </div>
    
    <div class="use-cases-section">
        <div class="use-cases-title">🎯 Common Use Cases</div>
        <div class="use-case-item">
            <div class="use-case-title">File Backup</div>
            <div class="use-case-description">Create backups of important files and directories</div>
        </div>
        <div class="use-case-item">
            <div class="use-case-title">File Distribution</div>
            <div class="use-case-description">Copy files to multiple locations or systems</div>
        </div>
        <div class="use-case-item">
            <div class="use-case-title">Development</div>
            <div class="use-case-description">Copy project files, templates, and configurations</div>
        </div>
    </div>
</div>`

	case "lookup":
		return `
<div class="detailed-help">
    <div class="options-section">
        <div class="options-title">🔧 Command Options</div>
        <div class="option-item">
            <span class="option-flag">-s, --similar</span>
            <span class="option-description">Show similar commands using fuzzy matching</span>
        </div>
        <div class="option-item">
            <span class="option-flag">-c, --categories</span>
            <span class="option-description">Show all command categories</span>
        </div>
        <div class="option-item">
            <span class="option-flag">-t, --task &lt;task&gt;</span>
            <span class="option-description">Get task-based command suggestions</span>
        </div>
        <div class="option-item">
            <span class="option-flag">-m, --menu</span>
            <span class="option-description">Show interactive menu for command exploration</span>
        </div>
        <div class="option-item">
            <span class="option-flag">[query]</span>
            <span class="option-description">Search term for command lookup</span>
        </div>
    </div>
    
    <div class="examples-section">
        <div class="examples-title">💡 Usage Examples</div>
        <div class="example-item">
            <div class="example-command">lookup ping</div>
            <div class="example-description">Find commands related to 'ping' - shows exact, partial, and similar matches</div>
        </div>
        <div class="example-item">
            <div class="example-command">lookup pin</div>
            <div class="example-description">Find all commands containing 'pin' - will find 'ping' and similar commands</div>
        </div>
        <div class="example-item">
            <div class="example-command">lookup network</div>
            <div class="example-description">Find all network-related commands by description matching</div>
        </div>
        <div class="example-item">
            <div class="example-command">lookup -c</div>
            <div class="example-description">Show all command categories with command counts</div>
        </div>
        <div class="example-item">
            <div class="example-command">lookup -t network</div>
            <div class="example-description">Get task-based suggestions for network operations</div>
        </div>
        <div class="example-item">
            <div class="example-command">lookup -m</div>
            <div class="example-description">Open interactive menu for command exploration</div>
        </div>
    </div>
    
    <div class="use-cases-section">
        <div class="use-cases-title">🎯 Common Use Cases</div>
        <div class="use-case-item">
            <div class="use-case-title">Command Discovery</div>
            <div class="use-case-description">Find commands when you only remember part of the name or functionality</div>
        </div>
        <div class="use-case-item">
            <div class="use-case-title">Task-Based Help</div>
            <div class="use-case-description">Get command recommendations based on what you want to accomplish</div>
        </div>
        <div class="use-case-item">
            <div class="use-case-title">Learning Tool</div>
            <div class="use-case-description">Explore available commands and learn about SuperShell capabilities</div>
        </div>
        <div class="use-case-item">
            <div class="use-case-title">Quick Reference</div>
            <div class="use-case-description">Browse commands by category or search for specific functionality</div>
        </div>
    </div>
</div>`

	case "killtask":
		return `
<div class="detailed-help">
    <div class="options-section">
        <div class="options-title">🔧 Command Options</div>
        <div class="option-item">
            <span class="option-flag">-f, --force</span>
            <span class="option-description">Force terminate processes immediately (SIGKILL on Unix)</span>
        </div>
        <div class="option-item">
            <span class="option-flag">-t, --tree</span>
            <span class="option-description">Terminate process tree including child processes</span>
        </div>
        <div class="option-item">
            <span class="option-flag">&lt;pid&gt;</span>
            <span class="option-description">Process ID to terminate</span>
        </div>
        <div class="option-item">
            <span class="option-flag">&lt;process_name&gt;</span>
            <span class="option-description">Process name to terminate (e.g., notepad.exe)</span>
        </div>
    </div>
    
    <div class="examples-section">
        <div class="examples-title">💡 Usage Examples</div>
        <div class="example-item">
            <div class="example-command">killtask 1234</div>
            <div class="example-description">Terminate process with PID 1234</div>
        </div>
        <div class="example-item">
            <div class="example-command">killtask notepad</div>
            <div class="example-description">Terminate all notepad processes</div>
        </div>
        <div class="example-item">
            <div class="example-command">killtask -f chrome</div>
            <div class="example-description">Force terminate all Chrome processes immediately</div>
        </div>
        <div class="example-item">
            <div class="example-command">killtask -t explorer</div>
            <div class="example-description">Terminate Explorer and all child processes</div>
        </div>
        <div class="example-item">
            <div class="example-command">killtask 1234 5678 notepad</div>
            <div class="example-description">Terminate multiple processes in one command</div>
        </div>
    </div>
    
    <div class="use-cases-section">
        <div class="use-cases-title">🎯 Common Use Cases</div>
        <div class="use-case-item">
            <div class="use-case-title">Process Management</div>
            <div class="use-case-description">Terminate unresponsive or unwanted processes</div>
        </div>
        <div class="use-case-item">
            <div class="use-case-title">System Cleanup</div>
            <div class="use-case-description">Clean up multiple processes or process trees</div>
        </div>
        <div class="use-case-item">
            <div class="use-case-title">Security Response</div>
            <div class="use-case-description">Quickly terminate suspicious or malicious processes</div>
        </div>
        <div class="use-case-item">
            <div class="use-case-title">Resource Management</div>
            <div class="use-case-description">Free up system resources by terminating resource-heavy processes</div>
        </div>
    </div>
</div>`

	case "help":
		return `
<div class="detailed-help">
    <div class="options-section">
        <div class="options-title">🔧 Command Options</div>
        <div class="option-item">
            <span class="option-flag">[command]</span>
            <span class="option-description">Get detailed help for a specific command</span>
        </div>
    </div>
    
    <div class="examples-section">
        <div class="examples-title">💡 Usage Examples</div>
        <div class="example-item">
            <div class="example-command">help</div>
            <div class="example-description">Show all available commands with descriptions</div>
        </div>
        <div class="example-item">
            <div class="example-command">help ping</div>
            <div class="example-description">Get detailed help for the ping command</div>
        </div>
        <div class="example-item">
            <div class="example-command">help sniff</div>
            <div class="example-description">Get comprehensive help for the sniff command with all options</div>
        </div>
    </div>
    
    <div class="use-cases-section">
        <div class="use-cases-title">🎯 Common Use Cases</div>
        <div class="use-case-item">
            <div class="use-case-title">Command Reference</div>
            <div class="use-case-description">Quick reference for command syntax and options</div>
        </div>
        <div class="use-case-item">
            <div class="use-case-title">Learning Tool</div>
            <div class="use-case-description">Learn about available commands and their capabilities</div>
        </div>
        <div class="use-case-item">
            <div class="use-case-title">Troubleshooting</div>
            <div class="use-case-description">Get help when commands aren't working as expected</div>
        </div>
    </div>
</div>`

	case "ver":
		return `
<div class="detailed-help">
    <div class="options-section">
        <div class="options-title">🔧 Command Options</div>
        <div class="option-item">
            <span class="option-flag">-v, --verbose</span>
            <span class="option-description">Show detailed version information with features and runtime details</span>
        </div>
    </div>
    
    <div class="examples-section">
        <div class="examples-title">💡 Usage Examples</div>
        <div class="example-item">
            <div class="example-command">ver</div>
            <div class="example-description">Show basic version information</div>
        </div>
        <div class="example-item">
            <div class="example-command">ver -v</div>
            <div class="example-description">Show comprehensive version details with features and system info</div>
        </div>
    </div>
    
    <div class="use-cases-section">
        <div class="use-cases-title">🎯 Common Use Cases</div>
        <div class="use-case-item">
            <div class="use-case-title">Version Checking</div>
            <div class="use-case-description">Verify SuperShell version for compatibility or support</div>
        </div>
        <div class="use-case-item">
            <div class="use-case-title">System Information</div>
            <div class="use-case-description">Get runtime and build information for troubleshooting</div>
        </div>
        <div class="use-case-item">
            <div class="use-case-title">Feature Discovery</div>
            <div class="use-case-description">See what features are available in your version</div>
        </div>
    </div>
</div>`

	default:
		// For commands without detailed help, provide basic structure
		return `
<div class="detailed-help">
    <div class="examples-section">
        <div class="examples-title">💡 Basic Usage</div>
        <div class="example-item">
            <div class="example-description">This command provides essential functionality for SuperShell operations. Use the command with its available options as shown in the usage syntax above.</div>
        </div>
    </div>
</div>`
	}
}
