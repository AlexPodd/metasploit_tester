package ciinterface

import (
	"github.com/AlexPodd/metasploit_tester_console/internal/domain"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type App struct {
	currentPage tea.Model
	history     []tea.Model

	chosePage  *ChoseExploit
	homePage   *MainMenu
	configMenu *ConfigModel
	addPage    *AddExploitPage
	metasploit domain.MetaSploitInterface
}

func NewApp(metasploit domain.MetaSploitInterface) *App {
	return &App{
		history:    []tea.Model{},
		metasploit: metasploit,
	}
}

func (a *App) Run() {

	print(a.chosePage.ChosenExploits())
	//a.metasploit.Run()
}

func (a *App) Init() tea.Cmd {
	a.history = []tea.Model{}

	a.homePage = NewMainMenu(a)
	a.currentPage = a.homePage

	exploits, err := a.homePage.scanner.WalkDir()
	if err != nil {
		a.currentPage = NewErrorModel(err.Error(), a)
	}

	items := make([]list.Item, len(exploits))
	for i := range exploits {
		items[i] = exploits[i]
	}
	a.chosePage = NewChoseExploit(exploits, items, a)

	a.configMenu = newConfigModel(a)
	a.addPage = newAddExploit(a)
	return tea.Batch(
		a.currentPage.Init(),
		a.configMenu.Init(),
		a.addPage.Init(),
		a.chosePage.Init(),
	)
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return a, tea.Quit
		}
	}

	newPage, cmd := a.currentPage.Update(msg)

	if newPage != a.currentPage {
		a.history = append(a.history, a.currentPage)
		a.currentPage = newPage
		if initCmd := newPage.Init(); initCmd != nil {
			return a, tea.Batch(cmd, initCmd)
		}
	}

	return a, cmd
}

func (a *App) View() string {
	return a.currentPage.View()
}
