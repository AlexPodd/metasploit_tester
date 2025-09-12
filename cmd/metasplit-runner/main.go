package main

import (
	"fmt"
	"log"
	"os"

	ciinterface "github.com/AlexPodd/metasploit_tester_console/internal/CLInterface"
	"github.com/AlexPodd/metasploit_tester_console/internal/metasploit"
	tea "github.com/charmbracelet/bubbletea"
)

type Config struct {
	Host     string `json"host"`
	User     string `json: "user"`
	Password string `json: "password"`
}

func main() {
	config, err := loadConfig("config.json")
	if err != nil {
		log.Fatal(err)
		return
	}

	metasploit := &metasploit.MetaSploitRPC{}
	err = metasploit.Login(config.Host, config.User, config.Password)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer metasploit.InstanceMSF.Logout()

	app := ciinterface.NewApp(metasploit)
	p := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Ошибка запуска приложения: %v\n", err)
		os.Exit(1)
	}
}
