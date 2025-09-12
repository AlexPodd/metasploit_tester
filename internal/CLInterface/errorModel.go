package ciinterface

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type ErrorModel struct {
	Message string
	app     *App
}

func NewErrorModel(message string, app *App) *ErrorModel {
	return &ErrorModel{
		Message: message,
		app:     app,
	}
}

func (m *ErrorModel) Init() tea.Cmd {
	return nil
}

func (m *ErrorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m.app.homePage, nil
		}

	}
	return m, nil
}

func (m *ErrorModel) View() string {
	return fmt.Sprintf(m.Message+" \nTo quit sooner press ctrl-c, or press ctrl-z to suspend...\n", m)
}
