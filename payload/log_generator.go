package payload

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"sync"
	"time"
)

const (
	logFileName = "payload/logs.log"
	MaxLines    = 200             // Увеличено, чтобы реже происходила очистка файла
	maxSize     = 5 * 1024 * 1024 // 5мб
)

var (
	messages = []string{
		"User logged in",
		"User logged out",
		"File uploaded",
		"Error processing request",
		"Database connection established",
		"Invalid input received",
	}
	genMu sync.Mutex
)

// StartLogGenerator – запускает генерацию логов в горутине
func StartLogGenerator(ctx context.Context, logger *slog.Logger) {
	rand.Seed(time.Now().UnixNano())

	go func() {
		ticker := time.NewTicker(1 * time.Second) // Генерируем логи реже чтобы не spam Telegram
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				logger.Info("Генератор логов остановлен")
				return
			case <-ticker.C:
				genMu.Lock()
				generateLog(logger)
				genMu.Unlock()
			}
		}
	}()
}

// генерирует одну строку + проверка размера
func generateLog(logger *slog.Logger) {
	lines, err := readLines(logFileName)
	if err != nil {
		logger.Error("Ошибка чтения logs.log", "error", err)
		return
	}
	// Очистка по линиям
	if len(lines) >= MaxLines {
		lines = lines[len(lines)-MaxLines+1:]
	}

	// Очистка по размеру
	info, _ := os.Stat(logFileName)

	if info != nil && info.Size() > maxSize {
		lines = lines[len(lines)/2:] // Оставляем половину
		logger.Info("Очистка logs.log по размеру")
	}
	logLevels := []string{"DEBUG", "INFO", "ERROR"}
	logLine := fmt.Sprintf("%s [%s] %s", time.Now().Format(time.RFC3339), logLevels[rand.Intn(len(logLevels))], messages[rand.Intn(len(messages))])
	lines = append(lines, logLine)
	lines = append(lines, "") // Добавляем пустую строку после каждого лога

	if err := writeLines(lines, logFileName); err != nil {
		logger.Error("Ошибка записи в logs.log", "error", err)
	}
}

func readLines(fileName string) ([]string, error) {
	file, err := os.Open(fileName)
	if os.IsNotExist(err) {
		return []string{}, nil
	} else if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func writeLines(lines []string, fileName string) error {
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(writer, line)
	}
	return writer.Flush()
}
