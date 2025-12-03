package config

import (
	"log/slog"
	"os"

	"github.com/Noddened/ErrLogsBot/internal/validators"
	"github.com/joho/godotenv"
)

// конфигурация приложения
// TODO: подумать над реализацией получения данных: через канал или личные сообщения
// *Через сообщения личные
type Config struct {
	BotToken string
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

	return &Config{
		BotToken: botToken,
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
