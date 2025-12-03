package logger

import (
	"log/slog"
	"os"
)

func SetupLogger() *slog.Logger {
	// В логгере также раздел на прод - json-логи для парсинга и локалку - вывод в консоль
	if env := os.Getenv("ENV"); env == "prod" {
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	}
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))
}
