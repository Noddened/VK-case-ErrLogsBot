package usecases

import (
	"bufio"
	"context"
	"io"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"log/slog"

	"github.com/Noddened/ErrLogsBot/internal/adapters"
)

var (
	mtx sync.Mutex
)

// мониторинг logs.log в горутине
func StartLogMonitoring(ctx context.Context, adapter *adapters.TelegramAdapter, filters []string, logger *slog.Logger) {
	go func() {
		logFilePath := "payload/logs.log"

		// Создаём файл, если его нет
		if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
			file, err := os.Create(logFilePath)
			if err != nil {
				logger.Error("Не удалось создать logs.log", "error", err)
				return
			}
			file.Close()
			logger.Info("Создан пустой logs.log для мониторинга")
		}

		var pos int64 = 0

		for {
			select {
			case <-ctx.Done():
				logger.Info("Мониторинг логов остановлен")
				return
			default:

				// Поллинг: проверяем размер файла
				info, err := os.Stat(logFilePath)
				if err != nil {
					logger.Error("Ошибка stat logs.log", "error", err)
					time.Sleep(100 * time.Millisecond)
					continue
				}

				// Если размер файла меньше позиции, значит файл был переписан/очищен
				// Сбрасываем позицию на начало
				if info.Size() < pos {
					logger.Info("Файл logs.log был переписан, сбрасываем позицию", "oldPos", pos, "newSize", info.Size())
					pos = 0
				}

				if info.Size() > pos {
					// Файл вырос — открываем и читаем новые строки
					file, err := os.Open(logFilePath)
					if err != nil {
						logger.Error("Не удалось открыть logs.log", "error", err)
						time.Sleep(100 * time.Millisecond)
						continue
					}

					// Ищем новый контент начиная с предыдущей позиции
					_, err = file.Seek(pos, io.SeekStart)
					if err != nil {
						logger.Error("Ошибка seek в logs.log", "error", err)
						file.Close()
						time.Sleep(500 * time.Millisecond)
						continue
					}

					newBytes := make([]byte, info.Size()-pos)
					n, err := file.Read(newBytes)
					file.Close()

					if err != nil && err != io.EOF {
						logger.Error("Ошибка чтения новых байт", "error", err)
						time.Sleep(500 * time.Millisecond)
						continue
					}

					// Обновляем позицию
					pos = info.Size()

					// Сканируем новые строки

					newReader := bufio.NewScanner(strings.NewReader(string(newBytes[:n])))
					mtx.Lock()
					for newReader.Scan() {

						line := strings.TrimSpace(newReader.Text())
						if line == "" {
							continue
						}
						if matchFilters(line, filters) {
							adapter.Broadcast(line)
							logger.Debug("Отправлена строка в Broadcast", "line", line)
						}
					}
					mtx.Unlock()
					if err := newReader.Err(); err != nil {
						logger.Error("Ошибка сканнера новых байт", "error", err)
					}
				}

				time.Sleep(50 * time.Millisecond) // Поллинг интервал
			}
		}
	}()
}

// проверяет строку на совпадение с фильтрами
func matchFilters(line string, filters []string) bool {
	for _, f := range filters {
		if matched, _ := regexp.MatchString(f, line); matched {
			return true
		}
	}
	return false
}
