package ciinterface

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type RunProgressModel struct {
	total      int
	completed  int
	progressCh chan int
	message    string
	app        *App
	done       bool
}

func NewRunProgressModel(total int, app *App, ch chan int) *RunProgressModel {
	return &RunProgressModel{
		total:      total,
		progressCh: ch,
		message:    "Запуск эксплоитов...",
		app:        app,
	}
}

type progressMsg int

// Init запускает горутину, которая читает из канала и шлёт сообщения в Bubble Tea
func (m *RunProgressModel) Init() tea.Cmd {
	go func() {
		for val := range m.progressCh {
			// здесь можно просто обновлять состояние через message, если нужно
			m.completed = val
		}
		m.done = true
	}()
	return tick()
}

func tick() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

type tickMsg struct{}

func (m *RunProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tickMsg:
		// Считываем прогресс из канала, если есть
		select {
		case val, ok := <-m.progressCh:
			if ok {
				m.completed = val
			} else {
				// Канал закрыт — все эксплоиты завершены
				m.done = true
				return m.app.finalPage, nil // переходим на страницу выбора
			}
		default:
		}
		// Продолжаем тикать
		if !m.done {
			return m, tick()
		}

	case progressMsg:
		// На случай, если будешь отправлять прогресс через progressMsg
		m.completed = int(msg)

	}

	return m, nil
}

func (m *RunProgressModel) View() string {
	barWidth := 30
	filled := int(float64(m.completed) / float64(m.total) * float64(barWidth))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
	return fmt.Sprintf("%s\n[%s] %d/%d", m.message, bar, m.completed, m.total)
}
