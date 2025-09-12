package ciinterface

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// SuccessPage отображает сообщение об успешном завершении и путь к отчету
type SuccessPage struct {
	reportPath string
	done       bool
}

// Создание новой страницы успеха с указанием пути
func NewSuccessPage(path string) *SuccessPage {
	return &SuccessPage{
		reportPath: path,
	}
}

// Init запускает таймер на 5 секунд, после которого программа завершится
func (s *SuccessPage) Init() tea.Cmd {
	return tea.Tick(5*time.Second, func(time.Time) tea.Msg {
		return exitMsg{}
	})
}

// Update обрабатывает сообщения
func (s *SuccessPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case exitMsg:
		s.done = true
		return s, tea.Quit
	}
	return s, nil
}

// View — выводим сообщение с прогрессом
func (s *SuccessPage) View() string {
	lines := []string{
		"✅ Программа сработала успешно!",
		fmt.Sprintf("Отчет сохранен: %s", s.reportPath),
		"Программа закроется через 5 секунд...",
	}
	return strings.Join(lines, "\n")
}

// exitMsg — сообщение для завершения программы
type exitMsg struct{}
