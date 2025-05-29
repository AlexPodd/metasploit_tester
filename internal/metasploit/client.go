package metasploit

import (
	"log"
	"time"

	"github.com/AlexPodd/metasploit_tester/internal/domain"
	"github.com/fpr1m3/go-msf-rpc/rpc"
)

type Client struct {
	InstanceMSF *rpc.Metasploit
	Report      *domain.Report
}

func (client *Client) Login(host, login, password string) (err error) {
	client.InstanceMSF, err = rpc.New(host, login, password)
	if err != nil {
		log.Print(err)
	}
	return err
}

type ExploitResult struct {
	ExploitName string
	Success     bool
	Output      string
}

func (client *Client) Execute(exploits []domain.Exploit, progressChan chan<- float32) error {
	console, err := client.InstanceMSF.ConsoleCreate()
	if err != nil {
		return err
	}
	consoleId := console.Id
	defer client.InstanceMSF.ConsoleDestroy(consoleId)

	for i, exploit := range exploits {
		command := "use " + exploit.Path
		_, err = client.InstanceMSF.ConsoleWrite(consoleId, command+"\n")
		if err != nil {
			return err
		}

		time.Sleep(500 * time.Millisecond)
		res, err := client.InstanceMSF.ConsoleRead(consoleId)
		if err != nil {
			return err
		}
		log.Print("Output after 'use': ", res.Data)

		_, err = client.InstanceMSF.ConsoleWrite(consoleId, "run\n")
		if err != nil {
			return err
		}

		for {
			time.Sleep(300 * time.Millisecond)
			res, err = client.InstanceMSF.ConsoleRead(consoleId)
			if err != nil {
				return err
			}
			if res.Data != "" {
				log.Print("Output during 'run': ", res.Data)
			}
			if !res.Busy {
				break
			}
		}
		progressChan <- float32(i + 1)
	}
	close(progressChan)
	return nil
}
