package ciinterface

import (
	"encoding/json"
	"os"

	"github.com/AlexPodd/metasploit_tester_console/internal/domain"
	filesystem "github.com/AlexPodd/metasploit_tester_console/internal/fileSystem"
	"github.com/AlexPodd/metasploit_tester_console/internal/reportGenerate"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type App struct {
	currentPage tea.Model
	history     []tea.Model

	writer     *filesystem.ExploitWriter
	chosePage  *ChoseExploit
	homePage   *MainMenu
	configMenu *ConfigModel
	addPage    *AddExploitPage
	metasploit domain.MetaSploitInterface
	finalPage  *SuccessPage
}

func NewApp(metasploit domain.MetaSploitInterface) *App {
	return &App{
		history:    []tea.Model{},
		metasploit: metasploit,
	}
}

func (a *App) Run(ch chan int) {
	exploits := a.chosePage.ChosenExploits()

	data, err := os.ReadFile(a.configMenu.selectedFile)
	if err != nil {
		panic(err)
	}
	var exploitsConfig []domain.ConfigExploit
	if err := json.Unmarshal(data, &exploitsConfig); err != nil {
		panic(err)
	}

	report, err := a.metasploit.Run(exploits, exploitsConfig, ch)
	if err != nil {
		panic(err)
	}

	pathRep, err := reportGenerate.GenerateReport(report)
	if err != nil {
		panic(err)
	}

	a.finalPage = NewSuccessPage(pathRep)
	close(ch)
}

func (a *App) Init() tea.Cmd {
	a.history = []tea.Model{}
	a.writer = &filesystem.ExploitWriter{}
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

func (a *App) AddNewExploit(srcPath, dstDir string) error {
	err := a.writer.AddExploit(srcPath, dstDir)
	if err != nil {
		return err
	}
	exploits, err := a.homePage.scanner.WalkDir()
	if err != nil {
		a.currentPage = NewErrorModel(err.Error(), a)
	}

	items := make([]list.Item, len(exploits))
	for i := range exploits {
		items[i] = exploits[i]
	}

	a.chosePage = NewChoseExploit(exploits, items, a)
	return err
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
