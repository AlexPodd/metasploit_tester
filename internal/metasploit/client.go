package metasploit

import (
	"fmt"
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
	client.Report = &domain.Report{}
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

func (client *Client) Execute(exploits []domain.Exploit, progressChan chan<- float32) (*domain.Report, error) {
	console, err := client.InstanceMSF.ConsoleCreate()
	if err != nil {
		return nil, err
	}
	consoleId := console.Id
	defer client.InstanceMSF.ConsoleDestroy(consoleId)

	for i, exploit := range exploits {
		var output string
		command := "use " + exploit.Path
		_, err = client.InstanceMSF.ConsoleWrite(consoleId, command+"\n")
		if err != nil {
			return nil, err
		}

		time.Sleep(500 * time.Millisecond)
		res, err := client.InstanceMSF.ConsoleRead(consoleId)
		if err != nil {
			return nil, err
		}
		log.Print("Output after 'use': ", res.Data)
		output += res.Data

		for key, value := range exploit.Params {
			setCmd := fmt.Sprintf("set %s %s\n", key, value)
			_, err = client.InstanceMSF.ConsoleWrite(consoleId, setCmd)
			if err != nil {
				return nil, err
			}
			time.Sleep(300 * time.Millisecond)

			res, err = client.InstanceMSF.ConsoleRead(consoleId)
			if err != nil {
				return nil, err
			}

			if res.Data != "" {
				log.Print("Output during 'set': ", res.Data)
				output += res.Data
			}
			time.Sleep(500 * time.Millisecond)
		}

		_, err = client.InstanceMSF.ConsoleWrite(consoleId, "run\n")
		if err != nil {
			return nil, err
		}

		for {
			time.Sleep(300 * time.Millisecond)
			res, err = client.InstanceMSF.ConsoleRead(consoleId)
			if err != nil {
				return nil, err
			}
			if res.Data != "" {
				log.Print("Output during 'run': ", res.Data)
				output += res.Data
			}
			if !res.Busy {
				break
			}
		}

		exploitRes := domain.ExploitResult{
			ExploitName: exploit.Name,
			Output:      output,
		}
		client.Report.Results = append(client.Report.Results, exploitRes)

		progressChan <- float32(i + 1)
	}
	client.Report.Timestamp = time.Now()

	close(progressChan)
	return client.Report, nil
}
