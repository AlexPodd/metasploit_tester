package reportGenerate

import (
	"fmt"
	"os"
	"strings"

	"github.com/AlexPodd/metasploit_tester_console/internal/domain"
)

func GenerateReport(rep *domain.Report) (string, error) {
	var sb strings.Builder

	total := len(rep.Results)
	successCount := 0
	for _, res := range rep.Results {
		if res.Success {
			successCount++
		}
	}

	// Заголовок
	sb.WriteString(fmt.Sprintf("Испытание прошло успешно: %d/%d эксплойтов могли сработать.\n\n", successCount, total))

	// По каждому эксплойту
	for i, res := range rep.Results {
		status := "провал"
		if res.Success {
			status = "успех"
		}
		sb.WriteString(fmt.Sprintf("%d. %s (%s)\n", i+1, res.ExploitName, status))
		sb.WriteString(res.Output + "\n\n")
	}

	timestamp := rep.Timestamp.Format("2006-01-02T15-04-05")
	filename := fmt.Sprintf("report_%s.txt", timestamp)

	err := os.WriteFile(filename, []byte(sb.String()), 0644)
	if err != nil {
		return "", fmt.Errorf("ошибка при сохранении отчета: %w", err)
	}

	return filename, nil
}
