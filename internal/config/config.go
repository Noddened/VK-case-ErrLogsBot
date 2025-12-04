package config

import (
	"log/slog"
	"os"
	"strings"

	"github.com/Noddened/ErrLogsBot/internal/validators"
	"github.com/joho/godotenv"
)

// конфигурация приложения
// TODO: подумать над реализацией получения данных: через канал или личные сообщения
// *Через сообщения личные
type Config struct {
	BotToken string
	Filters  []string
}

func LoadConfig(logger *slog.Logger) (*Config, error) {
	// Для локального запуска
	if err := godotenv.Load(); err != nil {
		logger.Info("Файл .env не найден, будет использоваться переменная окружения")
	}

	botToken := os.Getenv("BOT_TOKEN")

	if botToken == "" {
		logger.Error("Необходимые переменные окружения отсутствуют",
			"BOT_TOKEN", maskToken(botToken))
		os.Exit(1)
	}

	// Валидируем токен бота
	if err := validators.ValidateBotToken(botToken); err != nil {
		logger.Error("Невалидный токен бота", "error", err)
		return nil, err
	}

	filters, err := LoadFilters(logger)
	if err != nil {
		logger.Warn("Не удалось загрузить фильтры – используем дефолтные")
		filters = []string{"ERROR", "Invalid input"} // Дефолт
	}

	return &Config{
		BotToken: botToken,
		Filters:  filters,
	}, nil
}

func maskToken(token string) string {
	if token == "" {
		return "***"
	}
	if len(token) <= 8 {
		return "***"
	}
	return token[:4] + "***" + token[len(token)-4:]
}

func LoadFilters(logger *slog.Logger) ([]string, error) {
	data, err := os.ReadFile("configs/filters.txt")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(data), "\n")
	var filters []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			filters = append(filters, line)
		}
	}
	logger.Info("Фильтры загружены", "count", len(filters))
	return filters, nil
}
