package internal

import (
	"fun/internal/menu"

	tea "github.com/charmbracelet/bubbletea"
)

func Run() error {
	mm, err := menu.NewMenuModel()
	if err != nil {
		return err
	}

	_, err = tea.NewProgram(mm, tea.WithAltScreen()).Run()
	return err
}
