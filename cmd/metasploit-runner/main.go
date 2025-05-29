package main

import (
	"log"

	"github.com/AlexPodd/metasploit_tester/internal/app"
	"github.com/AlexPodd/metasploit_tester/internal/metasploit"
	"github.com/AlexPodd/metasploit_tester/internal/ui"
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

	client := metasploit.Client{}
	err = client.Login(config.Host, config.User, config.Password)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer client.InstanceMSF.Logout()

	scanner := app.Scanner{}
	exploits, setOfTag, err := scanner.WalkDir()
	if err != nil {
		log.Print(err)
	}
	win := ui.NewMainWindow(exploits, setOfTag, &client)
	err = win.Start()
	if err != nil {
		log.Fatal(err)
	}

}
