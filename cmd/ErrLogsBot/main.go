package main

import (
	"log/slog"
	"os"

	"github.com/Noddened/ErrLogsBot/internal/adapters"
	"github.com/Noddened/ErrLogsBot/internal/config"
	"github.com/Noddened/ErrLogsBot/internal/logger"
	"github.com/Noddened/ErrLogsBot/payload"
)

func main() {
	// Настройка логгера
	log := logger.SetupLogger()
	slog.SetDefault(log)
	slog.Info("=== Запуск ErrLogsBot ===")
	// Загрузка конфига
	cfg, err := config.LoadConfig(log)
	if err != nil {
		slog.Error("Не удалось загрузить конфиг", err)
		os.Exit(1)
	}

	// Запуск генератора логов (по таску) + TODO: сделать очистку файла со временем
	//...
	payload.StartLogGenerator()

	// Запуск самого бота
	tgAdapter, err := adapters.NewTelegramAdapter(cfg)
	if err != nil {
		slog.Error("Не удалось создать TG адаптер", "error", err)
		os.Exit(1)
	}
	// фильтр по файлу конфигурации
	filters, err := config.LoadFilters(log)
	// Завершение работы по сигналу

	/*
		Дополнительно:
		1) сделать тихий режим отправки сообщений в тг
		2) сделать так, чтобы логи продолжали генерироваться, но бот засыпал и просыпался по сигналу
	*/

}
