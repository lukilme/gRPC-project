package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	tui "ifpb.com/client-tui/internal/tui"
)

func main() {
	p := tea.NewProgram(tui.InitialModel())
	if err := p.Start(); err != nil {
		fmt.Println("erro:", err)
		os.Exit(1)
	}
}
