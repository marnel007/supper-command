package ui

import (
	"bufio"
	"fmt"
	"os"
)

type Prompt struct {
	// Add fields for customization later
}

func NewPrompt() *Prompt {
	return &Prompt{}
}

func (p *Prompt) Read() string {
	fmt.Print("supershell> ") // TODO: Use color
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return input[:len(input)-1]
}
