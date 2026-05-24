package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"paraizofelipe/review-station/internal/config"
	"paraizofelipe/review-station/internal/gitlab"
	"paraizofelipe/review-station/internal/model"
)

func main() {
	cfgPath := config.DefaultPath()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao carregar config em %s: %v\n\n", cfgPath, err)
		fmt.Fprintf(os.Stderr, "Crie o arquivo:\n\n")
		fmt.Fprintf(os.Stderr, "[gitlab]\nbase_url = \"https://gitlab.com\"\ntoken = \"glpat-...\"\n\n[[repo]]\nname = \"meu-projeto\"\npath = \"org/meu-projeto\"\nlocal = \"~/projects/meu-projeto\"\n")
		os.Exit(1)
	}

	client := gitlab.NewClient(cfg.GitLab.BaseURL, cfg.GitLab.Token)
	m := model.New(cfg, client)

	p := tea.NewProgram(m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "erro: %v\n", err)
		os.Exit(1)
	}
}
