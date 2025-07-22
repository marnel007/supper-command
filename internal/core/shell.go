package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	prompt "github.com/c-bata/go-prompt"
)

type Shell struct{}

func NewShell() *Shell {
	// Register built-in commands
	Register(&HelpCommand{})
	Register(&ClearCommand{})
	Register(&EchoCommand{})
	Register(&PwdCommand{})
	Register(&LsCommand{})
	Register(&CdCommand{})
	Register(&ExitCommand{})
	Register(&CatCommand{})
	Register(&MkdirCommand{})
	Register(&RmCommand{})
	Register(&RmdirCommand{})
	Register(&CpCommand{})
	Register(&MvCommand{})
	Register(&WhoamiCommand{})
	Register(&HostnameCommand{})
	Register(&VerCommand{})
	Register(&DirCommand{})
	Register(&PingCommand{})
	Register(&NslookupCommand{})
	Register(&TracertCommand{})
	Register(&WgetCommand{})
	Register(&IpconfigCommand{})
	return &Shell{}
}

func (s *Shell) Run() {
	p := prompt.New(
		executor,
		completer,
		prompt.OptionPrefix(getPrompt()),
		prompt.OptionTitle("SuperShell"),
		prompt.OptionInputTextColor(prompt.White),              // Prompt and input in bright white
		prompt.OptionSuggestionBGColor(prompt.Black),           // Suggestions background: black
		prompt.OptionSuggestionTextColor(prompt.White),         // Suggestions text: white
		prompt.OptionSelectedSuggestionBGColor(prompt.Blue),    // Selected suggestion: blue
		prompt.OptionSelectedSuggestionTextColor(prompt.White), // Selected suggestion text: white
		prompt.OptionPreviewSuggestionTextColor(prompt.Yellow), // Preview suggestion: yellow
	)
	p.Run()
}

func getPrompt() string {
	cwd, _ := os.Getwd()
	return strings.ReplaceAll(cwd, "/", "\\") + "> "
}

// This function is called when the user presses Enter
func executor(in string) {
	in = strings.TrimSpace(in)
	if in == "" {
		return
	}
	if in == "exit" || in == "quit" {
		fmt.Println("Goodbye!")
		os.Exit(0)
	}
	output := Dispatch(in)
	if output != "" {
		fmt.Println(output)
	}
}

// This function provides tab completion
func completer(d prompt.Document) []prompt.Suggest {
	// Gather all commands and aliases
	var suggestions []prompt.Suggest
	for name, cmd := range commandRegistry {
		suggestions = append(suggestions, prompt.Suggest{Text: name, Description: cmd.Description()})
	}

	args := strings.Fields(d.TextBeforeCursor())
	if len(args) == 0 {
		return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)
	}
	if len(args) == 1 {
		// Complete command names
		return prompt.FilterHasPrefix(suggestions, args[0], true)
	}
	// Complete file/dir names for arguments
	toComplete := args[len(args)-1]
	dir, filePrefix := filepath.Split(toComplete)
	if dir == "" {
		dir = "."
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var fileSugg []prompt.Suggest
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, filePrefix) {
			if entry.IsDir() {
				fileSugg = append(fileSugg, prompt.Suggest{Text: filepath.Join(dir, name) + string(os.PathSeparator), Description: "Directory"})
			} else {
				fileSugg = append(fileSugg, prompt.Suggest{Text: filepath.Join(dir, name), Description: "File"})
			}
		}
	}
	return prompt.FilterHasPrefix(fileSugg, filePrefix, true)
}
