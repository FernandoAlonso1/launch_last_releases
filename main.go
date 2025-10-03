package main

import (
	"archive/zip"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type FileRecord struct {
	Name        string
	ModTime     time.Time
	ArchivePath string
	ArchiveName string
	Size        int64
}

// findArchiveFiles находит все файлы архивов в указанной директории
func findArchiveFiles(dir string) ([]string, error) {
	var archives []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && isArchiveFile(path) {
			archives = append(archives, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return archives, nil
}

// isArchiveFile проверяет, является ли файл архивом (по расширению)
func isArchiveFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".zip" // Можно добавить другие форматы: .rar, .7z и т.д.
}

// extractFileInfoFromArchive извлекает информацию о файлах из архива
func extractFileInfoFromArchive(archivePath string) ([]FileRecord, error) {
	var files []FileRecord

	// Получаем дату модификации самого архива
	archiveInfo, err := os.Stat(archivePath)
	if err != nil {
		return nil, err
	}

	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	for _, f := range r.File {
		// Пропускаем директории
		if f.FileInfo().IsDir() {
			continue
		}

		files = append(files, FileRecord{
			Name:        f.Name,
			ModTime:     archiveInfo.ModTime(), // Используем дату архива
			ArchivePath: archivePath,
			ArchiveName: filepath.Base(archivePath),
			Size:        f.FileInfo().Size(),
		})
	}

	return files, nil
}

// determineLatestReleases определяет последние релизы файлов
func determineLatestReleases(allFiles map[string][]FileRecord) map[string]FileRecord {
	latest := make(map[string]FileRecord)

	for filename, versions := range allFiles {
		if len(versions) == 0 {
			continue
		}

		// Сортируем версии по дате (от старых к новым)
		sort.Slice(versions, func(i, j int) bool {
			return versions[i].ModTime.Before(versions[j].ModTime)
		})

		// Берем последнюю версию (самую новую)
		latest[filename] = versions[len(versions)-1]
	}

	return latest
}

// writeResultsToFile записывает результаты в текстовый файл
func writeResultsToFile(filename string, results map[string]FileRecord) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Записываем заголовок
	header := fmt.Sprintf("%-50s %-20s %-10s %s\n", "Файл", "Дата релиза", "Размер", "Архив")
	divider := strings.Repeat("-", len(header)) + "\n"

	if _, err := file.WriteString(header); err != nil {
		return err
	}
	if _, err := file.WriteString(divider); err != nil {
		return err
	}

	// Записываем данные
	for _, record := range results {

		line := fmt.Sprintf("%-50s %-20s %-10d %s\n",
			truncateString(record.Name, 50),
			record.ModTime.Format("2006-01-02 15:04:05"),
			record.Size,
			record.ArchiveName)

		if _, err := file.WriteString(line); err != nil {
			return err
		}
	}

	return nil
}

// truncateString обрезает строку до указанной длины
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func main() {
	// Конфигурационные параметры
	archiveDir := "X:/NETDBS"             // Директория с архивами
	outputFile := "./latest_releases.txt" // Файл для записи результатов

	fmt.Printf("Поиск архивов в директории: %s\n", archiveDir)

	// Получаем список всех архивов
	archives, err := findArchiveFiles(archiveDir)
	if err != nil {
		log.Fatalf("Ошибка поиска архивов: %v", err)
	}

	if len(archives) == 0 {
		log.Fatal("Архивы не найдены")
	}

	fmt.Printf("Найдено архивов: %d\n", len(archives))

	// Собираем информацию о всех файлах во всех архивах
	allFiles := make(map[string][]FileRecord)

	for _, archivePath := range archives {
		files, err := extractFileInfoFromArchive(archivePath)
		if err != nil {
			log.Printf("Ошибка обработки архива %s: %v", archivePath, err)
			continue
		}

		for _, file := range files {
			allFiles[file.Name] = append(allFiles[file.Name], file)
		}
	}

	// Определяем последние релизы файлов
	latestReleases := determineLatestReleases(allFiles)

	// Сохраняем результаты в файл
	err = writeResultsToFile(outputFile, latestReleases)
	if err != nil {
		log.Fatalf("Ошибка записи результатов: %v", err)
	}

	fmt.Printf("Результаты сохранены в файл: %s\n", outputFile)
	fmt.Printf("Обработано уникальных файлов: %d\n", len(latestReleases))
}
