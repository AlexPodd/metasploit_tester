package ciinterface

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	configStyle       = lipgloss.NewStyle().Padding(1, 2)
	configTitleStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFDF5")).Background(lipgloss.Color("#25A065")).Padding(0, 1)
	selectedFileStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575")).Bold(true)
	errorStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Bold(true)
	helpStyleConfig   = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Italic(true)
)

type ConfigModel struct {
	filepicker   filepicker.Model
	selectedFile string
	quitting     bool
	err          error
	app          *App
}

type clearErrorMsg struct{}

func clearErrorAfter(t time.Duration) tea.Cmd {
	return tea.Tick(t, func(_ time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}

func newConfigModel(app *App) *ConfigModel {
	fp := filepicker.New()
	fp.AllowedTypes = []string{".json"}
	fp.AutoHeight = false
	fp.SetHeight(15)
	home, _ := os.UserHomeDir()
	fp.CurrentDirectory = home

	return &ConfigModel{
		filepicker: fp,
		app:        app,
	}
}

func (m *ConfigModel) Init() tea.Cmd {
	return m.filepicker.Init()
}

func (m *ConfigModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "esc":
			return m.app.homePage, nil
		}
	case clearErrorMsg:
		m.err = nil
	}

	var cmd tea.Cmd
	m.filepicker, cmd = m.filepicker.Update(msg)

	if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
		m.selectedFile = path
	}

	if didSelect, path := m.filepicker.DidSelectDisabledFile(msg); didSelect {
		m.err = errors.New(path + " is not valid.")
		m.selectedFile = ""
		return m, tea.Batch(cmd, clearErrorAfter(2*time.Second))
	}

	return m, cmd
}

func (m *ConfigModel) View() string {
	if m.quitting {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(configTitleStyle.Render("Конфигурация для эксплоитов") + "\n\n")

	if m.err != nil {
		sb.WriteString(errorStyle.Render(m.err.Error()) + "\n\n")
	} else if m.selectedFile != "" {
		sb.WriteString("Выбранный файл: " + selectedFileStyle.Render(m.selectedFile) + "\n\n")
	} else {
		sb.WriteString("Выберите файл конфигурации:\n\n")
	}

	sb.WriteString(m.filepicker.View() + "\n\n")
	sb.WriteString(helpStyleConfig.Render("↑/↓ - навигация • Enter - выбор • esc - назад • q - закрыть приложение"))

	return configStyle.Render(sb.String())
}
