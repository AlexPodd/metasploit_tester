package ciinterface

import (
	"fmt"

	filesystem "github.com/AlexPodd/metasploit_tester_console/internal/fileSystem"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	menuStyle      = lipgloss.NewStyle().Padding(1, 2)
	titleStyleMenu = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)
	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF7F50")).
			Bold(true)
	optionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Italic(true)
)

type MainMenu struct {
	scanner filesystem.Scanner
	choice  int
	app     *App
	options []string
}

func NewMainMenu(app *App) *MainMenu {
	return &MainMenu{
		choice: 0,
		app:    app,
		options: []string{
			"Запустить эксплоиты",
			"Выбрать эксплоиты",
			"Выбрать конфигурацию",
			"Добавить новый эксплоит",
		},
	}
}

func (m *MainMenu) Init() tea.Cmd {
	m.scanner = filesystem.Scanner{}
	return nil
}

func (m *MainMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.choice > 0 {
				m.choice--
			}
		case "down", "j":
			if m.choice < len(m.options)-1 {
				m.choice++
			}
		case "enter", " ":
			switch m.choice {
			case 0:
				if m.app.configMenu.selectedFile == "" {
					return NewErrorModel("Выберите конфигурацию для эксплоитов", m.app), nil
				}
				if len(m.app.chosePage.ChosenExploits()) == 0 {
					return NewErrorModel("Выберите эксплоиты для запуска", m.app), nil
				}
				progressCh := make(chan int, len(m.app.chosePage.ChosenExploits()))

				go func() {
					m.app.Run(progressCh)
				}()

				return NewRunProgressModel(len(m.app.chosePage.ChosenExploits()), m.app, progressCh), nil
			case 1:
				return m.app.chosePage, nil
			case 2:
				return m.app.configMenu, nil
			case 3:
				return m.app.addPage, nil
			}
		case "q", "esc":
			return nil, tea.Quit
		}
	}
	return m, nil
}

func (m *MainMenu) View() string {
	s := titleStyleMenu.Render("Metasploit Tester") + "\n\n"
	s += "Выберите раздел:\n\n"

	for i, option := range m.options {
		cursor := " "
		if m.choice == i {
			cursor = ">"
			s += fmt.Sprintf("%s %s\n", cursorStyle.Render(cursor), optionStyle.Render(option))
		} else {
			s += fmt.Sprintf("%s %s\n", cursor, optionStyle.Render(option))
		}
	}

	s += "\n" + helpStyle.Render("↑/↓ - навигация • Enter - выбор • q - выход") + "\n"
	return menuStyle.Render(s)
}
