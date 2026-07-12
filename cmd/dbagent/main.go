package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
	"quasi.db_analysis_agent/internal/tui"
)

func main() {

	// tree := llm.BuildTreeOfDocs()
	// fmt.Println(tree)
	godotenv.Load()

	baseURL := os.Getenv("LLM_PROVIDER_URL")

	model := tui.NewAppModel(baseURL)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running TUI: c%v\n", err)
		os.Exit(1)
	}
}
