package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Noddened/ErrLogsBot/internal/adapters"
	"github.com/Noddened/ErrLogsBot/internal/config"
	"github.com/Noddened/ErrLogsBot/internal/logger"
	"github.com/Noddened/ErrLogsBot/internal/usecases"
	"github.com/Noddened/ErrLogsBot/payload"
)

func main() {
	os.Remove("payload/logs.log")
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

	// Получаем адаптера тг бота
	tgAdapter, err := adapters.NewTelegramAdapter(cfg)
	if err != nil {
		slog.Error("Не удалось создать TG адаптер", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запуск генератора логов (по таску) + очистка файла со временем
	payload.StartLogGenerator(ctx, log)
	// Монитор логов
	usecases.StartLogMonitoring(ctx, tgAdapter, cfg.Filters, log)

	// Запуск самого бота в горутине (чтобы не блокировать main)
	go tgAdapter.Start()

	// Завершение работы по сигналу ОС
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals

	slog.Info("Получен сигнал завершения...")
	cancel()

	// Даём горутинам время на завершение
	time.Sleep(2 * time.Second)
	slog.Info("ErrLogsBot завершён")
}
