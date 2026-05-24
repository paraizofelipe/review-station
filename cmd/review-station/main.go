package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type emptyModel struct{}

func (m emptyModel) Init() tea.Cmd                           { return tea.Quit }
func (m emptyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m emptyModel) View() string                            { return "" }

func main() {
	p := tea.NewProgram(emptyModel{}, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "erro: %v\n", err)
		os.Exit(1)
	}
}
