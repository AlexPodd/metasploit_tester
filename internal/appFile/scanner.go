package appFile

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/AlexPodd/metasploit_tester/internal/domain"
)

type Scanner struct {
	catalogs []string
	exploits []domain.Exploit
}

type ExploitWriter struct {
	BasePath string
}

func (scanner *Scanner) WalkDir() ([]domain.Exploit, *map[string]struct{}, error) {
	var allExploits []domain.Exploit

	homeDir, _ := os.UserHomeDir()
	modulesBase := filepath.Join(homeDir, ".msf4", "modules")
	exploitsBase := filepath.Join(modulesBase, "exploits")
	set := make(map[string]struct{})

	err := filepath.Walk(exploitsBase, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if _, exists := set[info.Name()]; !exists {
				set[info.Name()] = struct{}{}
			}
			return nil
		}

		// Получаем путь относительно каталога modules
		relPath, err := filepath.Rel(modulesBase, path)
		if err != nil {
			return err
		}

		// Теги — относительный путь от exploitsBase, можно оставить как есть
		rel, err := filepath.Rel(exploitsBase, path)
		if err != nil {
			return err
		}
		parts := strings.Split(rel, string(os.PathSeparator))
		tags := parts[:len(parts)-1]

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		exploit := parseExploit(relPath, tags, string(content))
		allExploits = append(allExploits, exploit)

		return nil
	})

	if err != nil {
		return nil, &set, err
	}

	return allExploits, &set, nil
}

func parseExploit(path string, tags []string, content string) domain.Exploit {
	getField := func(field string) string {
		re := regexp.MustCompile(field + `\s*=>\s*['"](.+?)['"]`)
		match := re.FindStringSubmatch(content)
		if len(match) > 1 {
			return match[1]
		}
		return ""
	}

	getArrayField := func(field string) []string {
		re := regexp.MustCompile(field + `\s*=>\s*\[\s*(.+?)\s*\]`)
		match := re.FindStringSubmatch(content)
		if len(match) > 1 {
			items := regexp.MustCompile(`['"](.+?)['"]`).FindAllStringSubmatch(match[1], -1)
			var result []string
			for _, item := range items {
				result = append(result, item[1])
			}
			return result
		}
		return nil
	}

	return domain.Exploit{
		Path:        path,
		Tags:        tags,
		Name:        getField("['\"]Name['\"]"),
		Description: getField("['\"]Description['\"]"),
		Authors:     getArrayField("['\"]Author['\"]"),
		Platform:    getField("['\"]Platform['\"]"),
		Targets:     getArrayField("['\"]Targets['\"]"),
		Disclosure:  getField("['\"]DisclosureDate['\"]"),
	}
}

func NewExploitWriter() (*ExploitWriter, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	basePath := filepath.Join(homeDir, ".msf4", "modules", "exploits")
	return &ExploitWriter{BasePath: basePath}, nil
}

func (w *ExploitWriter) AddExploit(relPath, content string) error {
	fullPath := filepath.Join(w.BasePath, relPath)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(fullPath, []byte(content), 0644)
}
